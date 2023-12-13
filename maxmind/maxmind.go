package maxmind

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/transport"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	ma "github.com/multiformats/go-multiaddr"
	madns "github.com/multiformats/go-multiaddr-dns"
	"github.com/oschwald/geoip2-golang"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	host          host.Host
	countryReader *geoip2.Reader
	asnReader     *geoip2.Reader
}

// NewClient initializes a new maxmind database client from the embedded database
func NewClient(asnDB string, countryDB string) (*Client, error) {
	asnData, err := os.ReadFile(asnDB)
	if err != nil {
		return nil, fmt.Errorf("read asn file %s: %w", asnDB, err)
	}

	asnReader, err := geoip2.FromBytes(asnData)
	if err != nil {
		return nil, fmt.Errorf("asn geoip from bytes: %w", err)
	}

	countryData, err := os.ReadFile(countryDB)
	if err != nil {
		return nil, fmt.Errorf("read country file %s: %w", asnDB, err)
	}

	countryReader, err := geoip2.FromBytes(countryData)
	if err != nil {
		return nil, fmt.Errorf("country geoip from bytes: %w", err)
	}

	h, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		return nil, fmt.Errorf("new libp2p host")
	}

	return &Client{
		host:          h,
		countryReader: countryReader,
		asnReader:     asnReader,
	}, nil
}

type AddrInfo struct {
	Country   string
	Continent string
	ASN       uint
}

// MaddrInfo resolves the give multi address to its corresponding
// IP addresses (it could be multiple due to protocols like dnsaddr)
// and returns a map of the form IP-address -> Country ISO code.
func (c *Client) MaddrInfo(ctx context.Context, maddr ma.Multiaddr) (map[string]*AddrInfo, error) {
	// give it a maximum of 10 seconds to resolve a single multi address
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resolved := c.resolveAddr(ctx, maddr)
	if len(resolved) == 0 {
		return nil, fmt.Errorf("could not resolve multi address %s", maddr)
	}

	infos := map[string]*AddrInfo{}
	for _, addr := range resolved {
		country, continent, err := c.AddrGeoInfo(addr)
		if err != nil {
			log.Debugln("could not derive Country for address", addr)
		}
		asn, _, err := c.AddrAS(addr)
		if err != nil {
			log.Debugln("could not derive Country for address", addr)
		}
		infos[addr] = &AddrInfo{Country: country, Continent: continent, ASN: asn}
	}
	return infos, nil
}

// AddrGeoInfo takes an IP address string and tries to derive the Country ISO code and continent code.
func (c *Client) AddrGeoInfo(addr string) (string, string, error) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return "", "", fmt.Errorf("invalid address %s", addr)
	}
	record, err := c.countryReader.Country(ip)
	if err != nil {
		return "", "", err
	}
	return record.Country.IsoCode, record.Continent.Code, nil
}

// AddrAS takes an IP address string and tries to derive the Autonomous System Number
func (c *Client) AddrAS(addr string) (uint, string, error) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return 0, "", fmt.Errorf("invalid address %s", addr)
	}
	record, err := c.asnReader.ASN(ip)
	if err != nil {
		return 0, "", err
	}
	return record.AutonomousSystemNumber, record.AutonomousSystemOrganization, nil
}

func (c *Client) Close() error {
	return c.countryReader.Close()
}

// resolveAddrs loops through the multi addresses of the given peer and recursively resolves
// the various DNS protocols (especially the dnsaddr protocol). This implementation is
// taken from:
// https://github.com/libp2p/go-libp2p/blob/9d3fd8bc4675b9cebf3102bdf62e56204c67ce5b/p2p/host/basic/basic_host.go#L676
func (c *Client) resolveAddr(ctx context.Context, maddr ma.Multiaddr) []string {
	// Recursively resolve all addrs.
	//
	// While the toResolve list is non-empty:
	// * Pop an address off.
	// * If the address is fully resolved, add it to the resolved list.
	// * Otherwise, resolve it and add the results to the "to resolve" list.
	toResolve := []ma.Multiaddr{maddr}
	resolved := make([]ma.Multiaddr, 0)
	resolveSteps := 0 // keep track of resolution iterations
	for len(toResolve) > 0 {
		// pop the last addr off.
		addr := toResolve[len(toResolve)-1]
		toResolve = toResolve[:len(toResolve)-1]

		// if it's resolved, add it to the resolved list.
		if !madns.Matches(addr) {
			resolved = append(resolved, addr)
			continue
		}

		resolveSteps++

		// We've resolved too many addresses. We can keep all the fully
		// resolved addresses, but we'll need to skip the rest.
		if resolveSteps >= 32 {
			log.Warnf("resolving too many addresses: %d/%d", resolveSteps, 32)
			continue
		}

		s := c.host.Network().(*swarm.Swarm)
		tpt := s.TransportForDialing(addr)
		resolver, ok := tpt.(transport.Resolver)
		if ok {
			resolvedAddrs, err := resolver.Resolve(ctx, addr)
			if err != nil {
				log.Warnf("Failed to resolve multiaddr %s by transport %v: %v", addr, tpt, err)
				continue
			}
			var added bool
			for _, a := range resolvedAddrs {
				if !addr.Equal(a) {
					toResolve = append(toResolve, a)
					added = true
				}
			}
			if added {
				continue
			}
		}

		// otherwise, resolve it
		resaddrs, err := madns.DefaultResolver.Resolve(ctx, addr)
		if err != nil {
			log.Infof("error resolving %s: %s", addr, err)
			continue
		}

		// add the results to the toResolve list.
		toResolve = append(toResolve, resaddrs...)
	}

	addrsMap := map[string]string{}
	for _, maddr := range resolved {
		for _, pr := range []int{ma.P_IP4, ma.P_IP6} { // DNS protocols are stripped via resolveAddrs above
			if addr, err := maddr.ValueForProtocol(pr); err == nil {
				addrsMap[addr] = addr
				break
			}
		}
	}

	var addrs []string
	for addr := range addrsMap {
		addrs = append(addrs, addr)
	}

	return addrs
}
