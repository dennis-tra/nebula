package core

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

// DialResult captures data that is gathered from pinging a single peer.
type DialResult[I PeerInfo[I]] struct {
	// The dialer that generated this result
	DialerID string

	// The dialed peer
	Info I

	// If error is set, the peer was not dialable
	Error error

	// The above error transferred to a known error
	DialError string

	// When was the dial started
	DialStartTime time.Time

	// When did this crawl end
	DialEndTime time.Time
}

func (r DialResult[I]) PeerInfo() I {
	return r.Info
}

func (r DialResult[I]) LogEntry() *log.Entry {
	logEntry := log.WithFields(log.Fields{
		"dialerID": r.DialerID,
		"remoteID": r.Info.ID().ShortString(),
		"alive":    r.Error == nil,
		"dialDur":  r.DialDuration(),
	})

	if r.Error != nil {
		if r.DialError == models.NetErrorUnknown {
			logEntry = logEntry.WithError(r.Error)
		} else {
			logEntry = logEntry.WithField("dialErr", r.DialError)
		}
	}

	return logEntry
}

// DialDuration returns the time it took to dial the peer
func (r DialResult[I]) DialDuration() time.Duration {
	return r.DialEndTime.Sub(r.DialStartTime)
}

// DialWriter handles the insert/upsert/update operations for a particular crawl result.
type DialWriter[I PeerInfo[I]] struct {
	id  string
	dbc *db.DBClient
}

func NewDialWriter[I PeerInfo[I]](id string, dbc *db.DBClient) *DialWriter[I] {
	return &DialWriter[I]{
		id:  id,
		dbc: dbc,
	}
}

// Work takes a crawl result (persist job) and inserts a denormalized database entry of the results.
func (w *DialWriter[I]) Work(ctx context.Context, task DialResult[I]) (WriteResult, error) {
	logEntry := task.LogEntry()
	if task.Error != nil {
		if task.DialError == models.NetErrorUnknown {
			logEntry = logEntry.WithError(task.Error)
		} else {
			logEntry = logEntry.WithField("error", task.DialError)
		}
	}

	start := time.Now()
	ivr, err := w.dbc.PersistDialVisit(
		task.Info.ID(),
		task.Info.Addrs(),
		task.DialDuration(),
		task.DialStartTime,
		task.DialEndTime,
		task.DialError,
	)
	if err != nil {
		logEntry.WithError(err).Warnln("Could not write dial result")
	}

	return WriteResult{
		InsertVisitResult: ivr,
		WriterID:          w.id,
		PeerID:            task.Info.ID(),
		Duration:          time.Since(start),
		Error:             err,
	}, nil
}
