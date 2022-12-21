package db

import (
	"strings"

	"github.com/libp2p/go-libp2p/p2p/net/swarm"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

// KnownErrors contains a list of known errors. Property key + string to match for
var KnownErrors = map[string]string{
	models.NetErrorIoTimeout:                  "i/o timeout",
	models.NetErrorNoRecentNetworkActivity:    "no recent network activity",
	models.NetErrorConnectionRefused:          "connection refused",
	models.NetErrorProtocolNotSupported:       "protocol not supported",
	models.NetErrorPeerIDMismatch:             "peer id mismatch",
	models.NetErrorNoRouteToHost:              "no route to host",
	models.NetErrorNetworkUnreachable:         "network is unreachable",
	models.NetErrorNoGoodAddresses:            "no good addresses",
	models.NetErrorContextDeadlineExceeded:    "context deadline exceeded",
	models.NetErrorNoPublicIP:                 "no public IP address",
	models.NetErrorMaxDialAttemptsExceeded:    "max dial attempts exceeded",
	models.NetErrorHostIsDown:                 "host is down",
	models.NetErrorStreamReset:                "stream reset",
	models.NetErrorNegotiateSecurityProtocol:  "failed to negotiate security protocol",
	models.NetErrorNegotiateStreamMultiplexer: "failed to negotiate stream multiplexer",
	models.NetErrorResourceLimitExceeded:      "resource limit exceeded",
	models.NetErrorWriteOnStream:              "Write on stream",
}

// NetError extracts the appropriate error type from the given error.
func NetError(err error) string {
	if netErr, ok := err.(*swarm.DialError); ok && netErr.Cause != nil {
		return NetError(netErr.Cause)
	}

	for key, errStr := range KnownErrors {
		if strings.Contains(err.Error(), errStr) {
			return key
		}
	}

	return models.NetErrorUnknown
}
