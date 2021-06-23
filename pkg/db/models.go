package db

import (
	"time"

	"gorm.io/gorm"
)

type Crawl struct {
	gorm.Model
	StartedAt       time.Time
	FinishedAt      time.Time
	CrawledPeers    uint
	DialablePeers   uint
	UndialablePeers uint
	// neighbor_count_distribution time.Time
	// connect_duration_distribution JSONB NOT NULL
}

type Peer struct {
	gorm.Model
	ID             string
	FirstDial      time.Time
	LastDial       time.Time
	NextDial       time.Time
	FailedDial     *time.Time
	Dials          int
	MultiAddresses []MultiAddress `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type MultiAddress struct {
	gorm.Model
	MultiAddress string
	PeerID       string
}

// PeerProperty captures the number of peers with the
// given property of a particular crawl
type PeerProperty struct {
	gorm.Model
	Property  string
	PeerCount int
	CrawlID   int
	Crawl     Crawl
}
