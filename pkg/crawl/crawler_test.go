package crawl

import (
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
	assert.Equal(t, "crawler-01", crawler.Service.Identifier())

	// increments crawler service number
	crawler, err = NewCrawler(nil, conf)
	require.NoError(t, err)
	assert.Equal(t, "crawler-02", crawler.Service.Identifier())
}

func TestCrawler_StartCrawling_stopsOnShutdown(t *testing.T) {
	conf := &config.Config{
		DialTimeout: 0,
		Protocols:   nil,
	}
	crawler, err := NewCrawler(nil, conf)
	require.NoError(t, err)

	go crawler.StartCrawling(queue.NewFIFO(), queue.NewFIFO())

	time.Sleep(time.Millisecond * 100)

	crawler.Shutdown()
}
