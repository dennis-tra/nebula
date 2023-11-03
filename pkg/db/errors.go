package db

import (
	"errors"
	"strings"

	"github.com/libp2p/go-libp2p/p2p/net/swarm"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

// KnownErrors contains a list of known errors. Property key + string to match for
var KnownErrors = map[string]string{
	"i/o timeout":                                models.NetErrorIoTimeout,
	"RPC timeout":                                models.NetErrorIoTimeout,
	"no recent network activity":                 models.NetErrorNoRecentNetworkActivity,
	"connection refused":                         models.NetErrorConnectionRefused,
	"connection reset by peer":                   models.NetErrorConnectionResetByPeer,
	"protocol not supported":                     models.NetErrorProtocolNotSupported,
	"peer id mismatch":                           models.NetErrorPeerIDMismatch,
	"no route to host":                           models.NetErrorNoRouteToHost,
	"network is unreachable":                     models.NetErrorNetworkUnreachable,
	"no good addresses":                          models.NetErrorNoGoodAddresses,
	"context deadline exceeded":                  models.NetErrorIoTimeout, // formerly NetErrorContextDeadlineExceeded
	"no public IP address":                       models.NetErrorNoPublicIP,
	"max dial attempts exceeded":                 models.NetErrorMaxDialAttemptsExceeded,
	"host is down":                               models.NetErrorHostIsDown,
	"stream reset":                               models.NetErrorStreamReset,
	"failed to negotiate security protocol: EOF": models.NetErrorNegotiateSecurityProtocol, // connect retry logic in discv5 relies on the ": EOF" suffix.
	"failed to negotiate stream multiplexer":     models.NetErrorNegotiateStreamMultiplexer,
	"resource limit exceeded":                    models.NetErrorResourceLimitExceeded,
	"Write on stream":                            models.NetErrorWriteOnStream,
}

var ErrorStr = map[string]string{}

func init() {
	for errStr, dbErr := range KnownErrors {
		ErrorStr[dbErr] = errStr
	}
}

// Because looping through a map doesn't have a deterministic order, we define
// the order here. E.g. "peer id mismatch" should be checked before
// "failed to negotiate security protocol" because the former is always part of
// the latter.
var knownErrorsPrecedence = []string{
	"i/o timeout",
	"RPC timeout",
	"no recent network activity",
	"connection refused",
	"connection reset by peer",
	"protocol not supported",
	"peer id mismatch",
	"no route to host",
	"network is unreachable",
	"no good addresses",
	"context deadline exceeded",
	"no public IP address",
	"max dial attempts exceeded",
	"host is down",
	"stream reset",
	"failed to negotiate security protocol: EOF",
	"failed to negotiate stream multiplexer",
	"resource limit exceeded",
	"Write on stream",
}

// NetError extracts the appropriate error type from the given error.
func NetError(err error) string {
	unwrapped := errors.Unwrap(err)
	if unwrapped != nil {
		errStr := NetError(unwrapped)
		if errStr != models.NetErrorUnknown {
			return errStr
		}
	} else if netErr, ok := err.(*swarm.DialError); ok && netErr.Cause != nil {
		errStr := NetError(netErr.Cause)
		if errStr != models.NetErrorUnknown {
			return errStr
		}
	}

	for _, errStr := range knownErrorsPrecedence {
		if strings.Contains(err.Error(), errStr) {
			return KnownErrors[errStr]
		}
	}

	return models.NetErrorUnknown
}
