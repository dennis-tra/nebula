package maxmind

import (
	"context"
	_ "embed"
	"fmt"
	"net"

	"github.com/friendsofgo/errors"
	ma "github.com/multiformats/go-multiaddr"
	madns "github.com/multiformats/go-multiaddr-dns"
	"github.com/oschwald/geoip2-golang"
	log "github.com/sirupsen/logrus"
)

//go:embed GeoLite2-Country.mmdb
var geoLite2Country []byte

//go:embed GeoLite2-ASN.mmdb
var geoLite2ASN []byte

type Client struct {
	countryReader *geoip2.Reader
	asnReader     *geoip2.Reader
}

// NewClient initializes a new maxmind database client from the embedded database
func NewClient() (*Client, error) {
	countryReader, err := geoip2.FromBytes(geoLite2Country)
	if err != nil {
		return nil, errors.Wrap(err, "geoip from bytes")
	}

	asnReader, err := geoip2.FromBytes(geoLite2ASN)
	if err != nil {
		return nil, errors.Wrap(err, "geoip from bytes")
	}

	return &Client{
		countryReader: countryReader,
		asnReader:     asnReader,
	}, nil
}

type AddrInfo struct {
	Country   string
	Continent string
	ASN       uint
}

// MaddrInfo resolve the give multi address to its corresponding
// IP addresses (it could be multiple due to protocols like dnsaddr)
// and returns a map of the form IP-address -> Country ISO code.
func (c *Client) MaddrInfo(ctx context.Context, maddr ma.Multiaddr) (map[string]*AddrInfo, error) {
	resolved := resolveAddrs(ctx, maddr)
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
func resolveAddrs(ctx context.Context, maddr ma.Multiaddr) []string {
	resolveSteps := 0

	// Recursively resolve all addrs.
	//
	// While the toResolve list is non-empty:
	// * Pop an address off.
	// * If the address is fully resolved, add it to the resolved list.
	// * Otherwise, resolve it and add the results to the "to resolve" list.
	toResolve := []ma.Multiaddr{maddr}
	resolved := []ma.Multiaddr{}
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

		// otherwise, resolve it
		resaddrs, err := madns.DefaultResolver.Resolve(ctx, addr)
		if err != nil {
			log.Debugf("error resolving %s: %s", addr, err)
			continue
		}

		// add the results to the toResolve list.
		for _, res := range resaddrs {
			toResolve = append(toResolve, res)
		}
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
