package core

import (
	"context"
	"errors"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/models"

	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"

	"github.com/dennis-tra/nebula-crawler/pkg/db"
)

// CrawlResult captures data that is gathered from crawling a single peer.
type CrawlResult[I PeerInfo] struct {
	// The crawler that generated this result
	CrawlerID string

	// Information about crawled peer
	Info I

	// The neighbors of the crawled peer
	RoutingTable *RoutingTable[I]

	// The agent version of the crawled peer
	Agent string

	// The protocols the peer supports
	Protocols []string

	// Indicates whether the above routing table information was queried through the API.
	// The API routing table does not include MultiAddresses, so we won't use them for further crawls.
	RoutingTableFromAPI bool

	// Any error that has occurred when connecting to the peer
	ConnectError error

	// The above error transferred to a known error
	ConnectErrorStr string

	// Any error that has occurred during fetching neighbor information
	CrawlError error

	// The above error transferred to a known error
	CrawlErrorStr string

	// When was the crawl started
	CrawlStartTime time.Time

	// When did this crawl end
	CrawlEndTime time.Time

	// When was the connection attempt made
	ConnectStartTime time.Time

	// As it can take some time to handle the result we track the timestamp explicitly
	ConnectEndTime time.Time

	// Additional properties of that specific peer we have crawled
	// Properties json.RawMessage

	// Whether kubos RPC API is exposed
	IsExposed null.Bool
}

var _ WorkResult[PeerInfo] = CrawlResult[PeerInfo]{}

func (r CrawlResult[I]) PeerInfo() I {
	return r.Info
}

func (r CrawlResult[I]) LogEntry() *log.Entry {
	logEntry := log.WithFields(log.Fields{
		"crawlerID":  r.CrawlerID,
		"remoteID":   r.Info.ID().ShortString(),
		"isDialable": r.ConnectError == nil && r.CrawlError == nil,
	})

	if r.ConnectError != nil {
		if r.ConnectErrorStr == models.NetErrorUnknown {
			logEntry = logEntry.WithError(r.ConnectError)
		} else {
			logEntry = logEntry.WithField("dialErr", r.ConnectErrorStr)
		}
	}

	if r.CrawlError != nil {
		// Log and count crawl errors
		if r.CrawlErrorStr == models.NetErrorUnknown {
			logEntry = logEntry.WithError(r.CrawlError)
		} else {
			logEntry = logEntry.WithField("crawlErr", r.CrawlErrorStr)
		}
	}

	return logEntry
}

// CrawlDuration returns the time it took to crawl to the peer (connecting + fetching neighbors)
func (r CrawlResult[I]) CrawlDuration() time.Duration {
	return r.CrawlEndTime.Sub(r.CrawlStartTime)
}

// ConnectDuration returns the time it took to connect to the peer. This includes dialing and the identity protocol.
func (r CrawlResult[I]) ConnectDuration() time.Duration {
	return r.ConnectEndTime.Sub(r.ConnectStartTime)
}

// CrawlWriter handles the insert/upsert/update operations for a particular crawl result.
type CrawlWriter[I PeerInfo] struct {
	id           string
	dbc          db.Client
	dbCrawlID    int
	writtenPeers int
}

var _ Worker[CrawlResult[PeerInfo], WriteResult] = (*CrawlWriter[PeerInfo])(nil)

func NewCrawlWriter[I PeerInfo](id string, dbc db.Client, dbCrawlID int) *CrawlWriter[I] {
	return &CrawlWriter[I]{
		id:           id,
		dbc:          dbc,
		dbCrawlID:    dbCrawlID,
		writtenPeers: 0,
	}
}

// Work takes a crawl result (persist job) and inserts a denormalized database entry of the results.
func (w *CrawlWriter[I]) Work(ctx context.Context, task CrawlResult[I]) (WriteResult, error) {
	if _, ok := w.dbc.(*db.NoopClient); ok {
		return WriteResult{
			InsertVisitResult: &db.InsertVisitResult{},
		}, nil
	}

	logEntry := log.WithFields(log.Fields{
		"writerID": w.id,
		"remoteID": task.Info.ID().ShortString(),
	})
	logEntry.Debugln("Storing peer")
	defer logEntry.Debugln("Stored peer")

	start := time.Now()

	ivr, err := w.dbc.PersistCrawlVisit(
		ctx,
		w.dbCrawlID,
		task.Info.ID(),
		task.Info.Addrs(),
		task.Protocols,
		task.Agent,
		task.ConnectDuration(),
		task.CrawlDuration(),
		task.CrawlStartTime,
		task.CrawlEndTime,
		task.ConnectErrorStr,
		task.CrawlErrorStr,
		task.IsExposed,
	)
	if err != nil && !errors.Is(ctx.Err(), context.Canceled) {
		logEntry.WithError(err).Warnln("Error inserting raw visit")
	} else {
		w.writtenPeers++
	}

	return WriteResult{
		InsertVisitResult: ivr,
		WriterID:          w.id,
		PeerID:            task.Info.ID(),
		Duration:          time.Since(start),
		Error:             err,
	}, nil
}
