package crawl

import (
	"strings"
	"time"

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

// knownErrors contains a list of known errors. Property key + string to match for
var knownErrors = map[string]string{
	models.DialErrorIoTimeout:               "i/o timeout",
	models.DialErrorConnectionRefused:       "connection refused",
	models.DialErrorProtocolNotSupported:    "protocol not supported",
	models.DialErrorPeerIDMismatch:          "peer id mismatch",
	models.DialErrorNoRouteToHost:           "no route to host",
	models.DialErrorNetworkUnreachable:      "network is unreachable",
	models.DialErrorNoGoodAddresses:         "no good addresses",
	models.DialErrorContextDeadlineExceeded: "context deadline exceeded",
	models.DialErrorNoPublicIP:              "no public IP address",
	models.DialErrorMaxDialAttemptsExceeded: "max dial attempts exceeded",
}

// millisSince returns the number of milliseconds between now and the given time.
func millisSince(start time.Time) float64 {
	return float64(time.Since(start)) / float64(time.Millisecond)
}
