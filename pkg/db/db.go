package db

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Client struct {
	*gorm.DB
}

func NewClient() (*Client, error) {
	dsn := "host=localhost user=postgres password=mysecretpassword dbname=nebula port=5432 sslmode=disable TimeZone=Europe/Berlin"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	err = db.AutoMigrate(&Crawl{}, &Peer{}, &PeerProperty{})
	if err != nil {
		return nil, err
	}

	return &Client{DB: db}, nil
}

// ExpiredPeers fetches all peers from the database
func (c *Client) ExpiredPeers() ([]*Peer, error) {
	var peers []*Peer
	tx := c.Where("next_dial <= ?", time.Now()).Find(&peers)
	return peers, tx.Error
}
