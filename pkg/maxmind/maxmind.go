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

type Client struct {
	reader *geoip2.Reader
}

// NewClient initializes a new maxmind database client from the embedded database
func NewClient() (*Client, error) {
	reader, err := geoip2.FromBytes(geoLite2Country)
	if err != nil {
		return nil, errors.Wrap(err, "geoip from bytes")
	}

	return &Client{
		reader: reader,
	}, nil
}

func (c *Client) AddrCountry(addr string) (string, error) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return "", fmt.Errorf("invalid address %s", addr)
	}
	record, err := c.reader.City(ip)
	if err != nil {
		return "", err
	}
	return record.Country.IsoCode, nil
}

func (c *Client) MaddrCountry(ctx context.Context, maddr ma.Multiaddr) (map[string]string, error) {
	resolved := resolveAddrs(ctx, maddr)
	if len(resolved) == 0 {
		return nil, fmt.Errorf("could not resolve multi address %s", maddr)
	}

	countries := map[string]string{}
	for _, addr := range resolved {
		country, err := c.AddrCountry(addr)
		if err != nil {
			log.Debugln("could not derive country for address", addr)
		}
		countries[addr] = country
	}
	return countries, nil
}

func (c *Client) Close() error {
	return c.reader.Close()
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
