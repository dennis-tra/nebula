package crawl

import (
	"time"

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

// millisSince returns the number of milliseconds between now and the given time.
func millisSince(start time.Time) float64 {
	return float64(time.Since(start)) / float64(time.Millisecond)
}
