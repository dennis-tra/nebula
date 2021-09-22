package crawl

import (
	"strings"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
	ma "github.com/multiformats/go-multiaddr"
)

// TotalErrors counts the total amount of errors - equivalent to undialable peers during this crawl.
func (s *Scheduler) TotalErrors() int {
	sum := 0
	for _, count := range s.errors {
		sum += count
	}
	return sum
}

func determineDialError(err error) string {
	for key, errStr := range knownErrors {
		if strings.Contains(err.Error(), errStr) {
			return key
		}
	}
	return models.DialErrorUnknown
}

func addrsToMaddrs(addrs []string) ([]ma.Multiaddr, error) {
	maddrs := make([]ma.Multiaddr, len(addrs))
	for i, addr := range addrs {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		maddrs[i] = maddr
	}
	return maddrs, nil
}

func maddrsToAddrs(maddrs []ma.Multiaddr) []string {
	addrs := make([]string, len(maddrs))
	for i, maddr := range maddrs {
		addrs[i] = maddr.String()
	}
	return addrs
}
