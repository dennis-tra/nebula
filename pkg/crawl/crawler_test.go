package crawl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
)

func TestNewCrawler_correctInit(t *testing.T) {
	conf := &config.Config{
		DialTimeout: 0,
		Protocols:   nil,
	}
	crawler, err := NewCrawler(nil, conf)
	require.NoError(t, err)

	assert.NotNil(t, crawler.pm)
	assert.Equal(t, conf, crawler.config)
	assert.Equal(t, "crawler-01", crawler.id)

	// increments crawler service number
	crawler, err = NewCrawler(nil, conf)
	require.NoError(t, err)
	assert.Equal(t, "crawler-02", crawler.id)
}

func TestCrawler_StartCrawling_stopsOnShutdown(t *testing.T) {
	conf := &config.Config{
		DialTimeout: 0,
		Protocols:   nil,
	}
	crawler, err := NewCrawler(nil, conf)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go crawler.StartCrawling(ctx, queue.NewFIFO(), queue.NewFIFO())

	time.Sleep(time.Millisecond * 100)

	cancel()
	<-crawler.done
}
