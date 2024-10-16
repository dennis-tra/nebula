package db

import (
	"errors"
	"strings"

	"github.com/libp2p/go-libp2p/p2p/net/swarm"

	"github.com/dennis-tra/nebula-crawler/db/models"
)

// KnownErrors contains a list of known errors. Property key + string to match for
var KnownErrors = map[string]string{
	"i/o timeout":                                models.NetErrorIoTimeout,
	"RPC timeout":                                models.NetErrorIoTimeout,
	"no recent network activity":                 models.NetErrorIoTimeout, // formerly NetErrorNoRecentNetworkActivity (equivalent to a timeout)
	"handshake did not complete in time":         models.NetErrorIoTimeout, // quic error
	"connection refused":                         models.NetErrorConnectionRefused,
	"connection reset by peer":                   models.NetErrorConnectionResetByPeer,
	"protocol not supported":                     models.NetErrorProtocolNotSupported,
	"protocols not supported":                    models.NetErrorProtocolNotSupported,
	"peer id mismatch":                           models.NetErrorPeerIDMismatch,
	"peer IDs don't match":                       models.NetErrorPeerIDMismatch,
	"no route to host":                           models.NetErrorNoRouteToHost,
	"network is unreachable":                     models.NetErrorNetworkUnreachable,
	"no good addresses":                          models.NetErrorNoGoodAddresses,
	"context deadline exceeded":                  models.NetErrorIoTimeout, // formerly NetErrorContextDeadlineExceeded
	"no public IP address":                       models.NetErrorNoIPAddress,
	"max dial attempts exceeded":                 models.NetErrorMaxDialAttemptsExceeded,
	"host is down":                               models.NetErrorHostIsDown,
	"stream reset":                               models.NetErrorStreamReset,
	"failed to negotiate security protocol: EOF": models.NetErrorNegotiateSecurityProtocol, // connect retry logic in discv5 relies on the ": EOF" suffix.
	"failed to negotiate stream multiplexer":     models.NetErrorNegotiateStreamMultiplexer,
	"resource limit exceeded":                    models.NetErrorResourceLimitExceeded,
	"Write on stream":                            models.NetErrorWriteOnStream,
	"can't assign requested address":             models.NetErrorCantAssignRequestedAddress, // transient error
	"cannot assign requested address":            models.NetErrorCantAssignRequestedAddress, // transient error
	"connection gated":                           models.NetErrorConnectionGated,            // transient error
	"RESOURCE_LIMIT_EXCEEDED (201)":              models.NetErrorCantConnectOverRelay,       // transient error
	"NO_RESERVATION (204)":                       models.NetErrorCantConnectOverRelay,       // permanent error
	// devp2p errors
	"disconnect requested":              models.NetErrorDevp2pDisconnectRequested,
	"network error":                     models.NetErrorDevp2pNetworkError,
	"breach of protocol":                models.NetErrorDevp2pBreachOfProtocol,
	"useless peer":                      models.NetErrorDevp2pUselessPeer,
	"too many peers":                    models.NetErrorDevp2pTooManyPeers,
	"already connected":                 models.NetErrorDevp2pAlreadyConnected,
	"incompatible p2p protocol version": models.NetErrorDevp2pIncompatibleP2PProtocolVersion,
	"invalid node identity":             models.NetErrorDevp2pInvalidNodeIdentity,
	"client quitting":                   models.NetErrorDevp2pClientQuitting,
	"unexpected identity":               models.NetErrorDevp2pUnexpectedIdentity,
	"connected to self":                 models.NetErrorDevp2pConnectedToSelf,
	"read timeout":                      models.NetErrorDevp2pReadTimeout,
	"subprotocol error":                 models.NetErrorDevp2pSubprotocolError,
	"could not negotiate eth protocol":  models.NetErrorDevp2pEthprotocolError,
	"handshake failed: EOF":             models.NetErrorDevp2pHandshakeEOF, // dependent on error string in discv4
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
	"cannot assign requested address",
	"can't assign requested address",
	"connection gated",
	"connection refused",
	"connection reset by peer",
	"protocol not supported",
	"protocols not supported",
	"peer id mismatch",
	"peer IDs don't match",
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
	"RESOURCE_LIMIT_EXCEEDED (201)",
	"NO_RESERVATION (204)",
	"handshake did not complete in time",
	"disconnect requested",
	"network error",
	"breach of protocol",
	"useless peer",
	"too many peers",
	"already connected",
	"incompatible p2p protocol version",
	"invalid node identity",
	"client quitting",
	"unexpected identity",
	"connected to self",
	"read timeout",
	"subprotocol error",
	"could not negotiate eth protocol",
	"handshake failed: EOF",
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
