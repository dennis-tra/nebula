package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestSomething(t *testing.T) {
	db, err := sql.Open("postgres", "dbname=nebula user=nebula password=password sslmode=disable")
	require.NoError(t, err)
	rawVisit := &models.RawVisit{
		CrawlID:         null.IntFrom(1),
		VisitStartedAt:  time.Now(),
		VisitEndedAt:    time.Now(),
		ConnectDuration: null.StringFrom(fmt.Sprintf("%f seconds", time.Duration(time.Millisecond*5).Seconds())),
		CrawlDuration:   null.StringFrom(fmt.Sprintf("%f seconds", time.Duration(time.Millisecond*3).Seconds())),
		PeerMultiHash:   "QmSKVUFAyCddg2wDUdZVCfvqG5YCwwJTWY1HRmorebXcKG",
		Protocols:       []string{"quicknode", "/sbst/1.0.0"},
		MultiAddresses: []string{
			"/ip4/159.69.43.228/tcp/dennis",
			"/ip4/159.69.43.228/tcp/4001",
			"/ip4/159.69.43.228/udp/4001/quic",
			"/ip6/2a01:4f8:1c1c:22d6::1/tcp/1024",
			"/ip6/2a01:4f8:1c1c:22d6::1/tcp/4001",
			"/ip6/2a01:4f8:1c1c:22d6::1/udp/1025/quic",
			"/ip6/2a01:4f8:1c1c:22d6::1/udp/4001/quic",
		},
		Type:         models.VisitTypeCrawl,
		AgentVersion: null.StringFrom("storm"),
	}
	ctx := context.Background()
	err = rawVisit.Insert(ctx, db, boil.Infer())
	//conn, err := db.Conn(ctx)
	//require.NoError(t, err)
	//res, err := conn.ExecContext(ctx,
	//	saveVisitTxn,
	//	rawVisit.PeerMultiHash, // $1
	//	// rawVisit.Error,           // $2
	//	// rawVisit.VisitStartedAt,  // $3
	//	// rawVisit.VisitEndedAt,    // $4
	//	// rawVisit.DialDuration,    // $5
	//	// rawVisit.ConnectDuration, // $6
	//	// rawVisit.CrawlID,         // $7
	//	// rawVisit.CrawlDuration,   // $8
	//	// rawVisit.Type,            // $9
	//)
	assert.NoError(t, err)
	// fmt.Println(res)
}
