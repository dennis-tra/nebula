// Code generated by SQLBoiler 4.14.1 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"strconv"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/strmangle"
)

// M type is for providing columns and column values to UpdateAll.
type M map[string]interface{}

// ErrSyncFail occurs during insert when the record could not be retrieved in
// order to populate default value information. This usually happens when LastInsertId
// fails or there was a primary key configuration that was not resolvable.
var ErrSyncFail = errors.New("models: failed to synchronize data after insert")

type insertCache struct {
	query        string
	retQuery     string
	valueMapping []uint64
	retMapping   []uint64
}

type updateCache struct {
	query        string
	valueMapping []uint64
}

func makeCacheKey(cols boil.Columns, nzDefaults []string) string {
	buf := strmangle.GetBuffer()

	buf.WriteString(strconv.Itoa(cols.Kind))
	for _, w := range cols.Cols {
		buf.WriteString(w)
	}

	if len(nzDefaults) != 0 {
		buf.WriteByte('.')
	}
	for _, nz := range nzDefaults {
		buf.WriteString(nz)
	}

	str := buf.String()
	strmangle.PutBuffer(buf)
	return str
}

// Enum values for NetError
const (
	NetErrorUnknown                    string = "unknown"
	NetErrorIoTimeout                  string = "io_timeout"
	NetErrorNoRecentNetworkActivity    string = "no_recent_network_activity"
	NetErrorConnectionRefused          string = "connection_refused"
	NetErrorProtocolNotSupported       string = "protocol_not_supported"
	NetErrorPeerIDMismatch             string = "peer_id_mismatch"
	NetErrorNoRouteToHost              string = "no_route_to_host"
	NetErrorNetworkUnreachable         string = "network_unreachable"
	NetErrorNoGoodAddresses            string = "no_good_addresses"
	NetErrorContextDeadlineExceeded    string = "context_deadline_exceeded"
	NetErrorNoPublicIP                 string = "no_public_ip"
	NetErrorMaxDialAttemptsExceeded    string = "max_dial_attempts_exceeded"
	NetErrorMaddrReset                 string = "maddr_reset"
	NetErrorStreamReset                string = "stream_reset"
	NetErrorHostIsDown                 string = "host_is_down"
	NetErrorNegotiateSecurityProtocol  string = "negotiate_security_protocol"
	NetErrorNegotiateStreamMultiplexer string = "negotiate_stream_multiplexer"
	NetErrorResourceLimitExceeded      string = "resource_limit_exceeded"
	NetErrorWriteOnStream              string = "write_on_stream"
)

func AllNetError() []string {
	return []string{
		NetErrorUnknown,
		NetErrorIoTimeout,
		NetErrorNoRecentNetworkActivity,
		NetErrorConnectionRefused,
		NetErrorProtocolNotSupported,
		NetErrorPeerIDMismatch,
		NetErrorNoRouteToHost,
		NetErrorNetworkUnreachable,
		NetErrorNoGoodAddresses,
		NetErrorContextDeadlineExceeded,
		NetErrorNoPublicIP,
		NetErrorMaxDialAttemptsExceeded,
		NetErrorMaddrReset,
		NetErrorStreamReset,
		NetErrorHostIsDown,
		NetErrorNegotiateSecurityProtocol,
		NetErrorNegotiateStreamMultiplexer,
		NetErrorResourceLimitExceeded,
		NetErrorWriteOnStream,
	}
}

// Enum values for CrawlState
const (
	CrawlStateStarted   string = "started"
	CrawlStateCancelled string = "cancelled"
	CrawlStateFailed    string = "failed"
	CrawlStateSucceeded string = "succeeded"
)

func AllCrawlState() []string {
	return []string{
		CrawlStateStarted,
		CrawlStateCancelled,
		CrawlStateFailed,
		CrawlStateSucceeded,
	}
}

// Enum values for SessionState
const (
	SessionStateOpen    string = "open"
	SessionStatePending string = "pending"
	SessionStateClosed  string = "closed"
)

func AllSessionState() []string {
	return []string{
		SessionStateOpen,
		SessionStatePending,
		SessionStateClosed,
	}
}

// Enum values for VisitType
const (
	VisitTypeCrawl string = "crawl"
	VisitTypeDial  string = "dial"
)

func AllVisitType() []string {
	return []string{
		VisitTypeCrawl,
		VisitTypeDial,
	}
}
