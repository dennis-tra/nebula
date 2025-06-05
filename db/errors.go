package db

import (
	"errors"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	ma "github.com/multiformats/go-multiaddr"
	"golang.org/x/net/context"

	pgmodels "github.com/dennis-tra/nebula-crawler/db/models/pg"
)

var ErrNoPublicIP = fmt.Errorf("skipping node as it has no public IP address") // change knownErrs map if changing this msg

func MaddrErrors(maddrs []ma.Multiaddr, err error) []string {
	if err == nil {
		return []string{}
	}
	dialErrors := make([]string, len(maddrs))

	var serr *swarm.DialError
	if errors.As(err, &serr) {
		maddrErr := make(map[string]error, len(serr.DialErrors))
		for _, derr := range serr.DialErrors {
			maddrErr[string(derr.Address.Bytes())] = derr.Cause
		}

		for i, maddr := range maddrs {
			derr, ok := maddrErr[string(maddr.Bytes())]
			if ok {
				if errors.Is(derr, context.Canceled) {
					dialErrors[i] = "canceled"
				} else {
					dialErrors[i] = NetError(derr)
				}
			} else {
				dialErrors[i] = "not_dialed"
			}
		}
	} else {
		var commonErr string
		if errors.Is(err, ErrNoPublicIP) {
			commonErr = "not_dialed"
		} else if errors.Is(err, context.Canceled) {
			commonErr = "canceled"
		} else {
			commonErr = NetError(err)
		}
		for i := range maddrs {
			dialErrors[i] = commonErr
		}
	}

	return dialErrors
}

// KnownErrors contains a list of known errors. Property key + string to match for
var KnownErrors = map[string]string{
	"i/o timeout":                                         pgmodels.NetErrorIoTimeout,
	"RPC timeout":                                         pgmodels.NetErrorIoTimeout,
	"no recent network activity":                          pgmodels.NetErrorIoTimeout, // formerly NetErrorNoRecentNetworkActivity (equivalent to a timeout)
	"handshake did not complete in time":                  pgmodels.NetErrorIoTimeout, // quic error
	"connection refused":                                  pgmodels.NetErrorConnectionRefused,
	"CONNECTION_REFUSED":                                  pgmodels.NetErrorConnectionRefused,
	"connection reset by peer":                            pgmodels.NetErrorConnectionResetByPeer,
	"protocol not supported":                              pgmodels.NetErrorProtocolNotSupported,
	"protocols not supported":                             pgmodels.NetErrorProtocolNotSupported,
	"peer id mismatch":                                    pgmodels.NetErrorPeerIDMismatch,
	"peer IDs don't match":                                pgmodels.NetErrorPeerIDMismatch,
	"no route to host":                                    pgmodels.NetErrorNoRouteToHost,
	"network is unreachable":                              pgmodels.NetErrorNetworkUnreachable,
	"no good addresses":                                   pgmodels.NetErrorNoGoodAddresses,
	"context deadline exceeded":                           pgmodels.NetErrorIoTimeout, // formerly NetErrorContextDeadlineExceeded
	"no public IP address":                                pgmodels.NetErrorNoIPAddress,
	"max dial attempts exceeded":                          pgmodels.NetErrorMaxDialAttemptsExceeded,
	"host is down":                                        pgmodels.NetErrorHostIsDown,
	"stream reset":                                        pgmodels.NetErrorStreamReset,
	"stream closed":                                       pgmodels.NetErrorStreamReset,
	"failed to negotiate security protocol: EOF":          pgmodels.NetErrorNegotiateSecurityProtocol, // connect retry logic in discv5 relies on the ": EOF" suffix.
	"failed to negotiate security protocol":               pgmodels.NetErrorNegotiateSecurityProtocol,
	"failed to negotiate stream multiplexer":              pgmodels.NetErrorNegotiateStreamMultiplexer,
	"resource limit exceeded":                             pgmodels.NetErrorResourceLimitExceeded,
	"Write on stream":                                     pgmodels.NetErrorWriteOnStream,
	"can't assign requested address":                      pgmodels.NetErrorCantAssignRequestedAddress, // transient error
	"cannot assign requested address":                     pgmodels.NetErrorCantAssignRequestedAddress, // transient error
	"connection gated":                                    pgmodels.NetErrorConnectionGated,            // transient error
	"RESOURCE_LIMIT_EXCEEDED (201)":                       pgmodels.NetErrorCantConnectOverRelay,       // transient error
	"opening relay circuit: CONNECTION_FAILED (203)":      pgmodels.NetErrorCantConnectOverRelay,       // transient error
	"NO_RESERVATION (204)":                                pgmodels.NetErrorCantConnectOverRelay,       // permanent error
	"relay failed with a protocol error":                  pgmodels.NetErrorCantConnectOverRelay,
	"can't dial a p2p-circuit without specifying a relay": pgmodels.NetErrorNoIPAddress,
	"no transport for protocol":                           pgmodels.NetErrorNoTransportForProtocol,
	// devp2p errors
	"no good ip address":                 pgmodels.NetErrorNoIPAddress,
	"disconnect requested":               pgmodels.NetErrorDevp2pDisconnectRequested,
	"network error":                      pgmodels.NetErrorDevp2pNetworkError,
	"breach of protocol":                 pgmodels.NetErrorDevp2pBreachOfProtocol,
	"useless peer":                       pgmodels.NetErrorDevp2pUselessPeer,
	"too many peers":                     pgmodels.NetErrorDevp2pTooManyPeers,
	"already connected":                  pgmodels.NetErrorDevp2pAlreadyConnected,
	"incompatible p2p protocol version":  pgmodels.NetErrorDevp2pIncompatibleP2PProtocolVersion,
	"invalid node identity":              pgmodels.NetErrorDevp2pInvalidNodeIdentity,
	"client quitting":                    pgmodels.NetErrorDevp2pClientQuitting,
	"unexpected identity":                pgmodels.NetErrorDevp2pUnexpectedIdentity,
	"connected to self":                  pgmodels.NetErrorDevp2pConnectedToSelf,
	"read timeout":                       pgmodels.NetErrorDevp2pReadTimeout,
	"subprotocol error":                  pgmodels.NetErrorDevp2pSubprotocolError,
	"could not negotiate eth protocol":   pgmodels.NetErrorDevp2pEthprotocolError,
	"handshake failed: EOF":              pgmodels.NetErrorDevp2pHandshakeEOF,               // dependent on error string in discv4
	"malformed disconnect message":       pgmodels.NetErrorDevp2pMalformedDisconnectMessage, // dependent on error string in discv4
	"dial refused because of black hole": pgmodels.NetErrorBlackHoleRefused,
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
	"stream closed",
	"failed to negotiate security protocol: EOF",
	"failed to negotiate stream multiplexer",
	"resource limit exceeded",
	"Write on stream",
	"RESOURCE_LIMIT_EXCEEDED (201)",
	"NO_RESERVATION (204)",
	"no transport for protocol",
	"relay failed with a protocol error",
	"opening relay circuit: CONNECTION_FAILED (203)",
	"can't dial a p2p-circuit without specifying a relay",
	"too many peers",
	"no good ip address",
	"malformed disconnect message",
	"handshake did not complete in time",
	"disconnect requested",
	"network error",
	"breach of protocol",
	"useless peer",
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
	"dial refused because of black hole",
	"failed to negotiate security protocol",
	"no transport for protocol",
}

// NetError extracts the appropriate error type from the given error.
func NetError(err error) string {
	unwrapped := errors.Unwrap(err)
	if unwrapped != nil {
		errStr := NetError(unwrapped)
		if errStr != pgmodels.NetErrorUnknown {
			return errStr
		}
	} else if netErr, ok := err.(*swarm.DialError); ok && netErr.Cause != nil {
		errStr := NetError(netErr.Cause)
		if errStr != pgmodels.NetErrorUnknown {
			return errStr
		}
	}

	for _, errStr := range knownErrorsPrecedence {
		if strings.Contains(err.Error(), errStr) {
			return KnownErrors[errStr]
		}
	}

	return pgmodels.NetErrorUnknown
}
