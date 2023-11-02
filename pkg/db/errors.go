package db

import (
	"errors"
	"strings"

	"github.com/libp2p/go-libp2p/p2p/net/swarm"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

// KnownErrors contains a list of known errors. Property key + string to match for
var KnownErrors = map[string]string{
	"i/o timeout":                            models.NetErrorIoTimeout,
	"RPC timeout":                            models.NetErrorIoTimeout,
	"no recent network activity":             models.NetErrorNoRecentNetworkActivity,
	"connection refused":                     models.NetErrorConnectionRefused,
	"connection reset by peer":               models.NetErrorConnectionResetByPeer,
	"protocol not supported":                 models.NetErrorProtocolNotSupported,
	"peer id mismatch":                       models.NetErrorPeerIDMismatch,
	"no route to host":                       models.NetErrorNoRouteToHost,
	"network is unreachable":                 models.NetErrorNetworkUnreachable,
	"no good addresses":                      models.NetErrorNoGoodAddresses,
	"context deadline exceeded":              models.NetErrorIoTimeout, // formerly NetErrorContextDeadlineExceeded
	"no public IP address":                   models.NetErrorNoPublicIP,
	"max dial attempts exceeded":             models.NetErrorMaxDialAttemptsExceeded,
	"host is down":                           models.NetErrorHostIsDown,
	"stream reset":                           models.NetErrorStreamReset,
	"failed to negotiate security protocol":  models.NetErrorNegotiateSecurityProtocol,
	"failed to negotiate stream multiplexer": models.NetErrorNegotiateStreamMultiplexer,
	"resource limit exceeded":                models.NetErrorResourceLimitExceeded,
	"Write on stream":                        models.NetErrorWriteOnStream,
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

	for errStr, key := range KnownErrors {
		if strings.Contains(err.Error(), errStr) {
			return key
		}
	}

	return models.NetErrorUnknown
}
