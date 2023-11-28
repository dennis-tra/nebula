package kubo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	RequestTimeout = 5 * time.Second
	DefaultPort    = "5001"
)

type Client struct {
	http.Client
}

func NewClient() *Client {
	return &Client{}
}

type IDResponse struct {
	ID              string
	PublicKey       string
	Addresses       []string
	AgentVersion    string
	ProtocolVersion string
	Protocols       []string
}

func (c *Client) ID(ctx context.Context, host string) (*IDResponse, error) {
	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", host, DefaultPort),
		Path:   "/api/v0/id",
	}

	// Build request
	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	// Fire request
	resp, err := c.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("could not query api: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read id response body: %w", err)
	}

	idResp := IDResponse{}
	err = json.Unmarshal(data, &idResp)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal id response: %w", err)
	}

	return &idResp, nil
}

type RoutingTableResponse struct {
	Name    string   `json:"Name"`
	Buckets []Bucket `json:"Buckets"`
}

type Bucket struct {
	LastRefresh string       `json:"LastRefresh"`
	Peers       []BucketPeer `json:"Peers"`
}

type BucketPeer struct {
	ID            string `json:"ID"`
	Connected     bool   `json:"Connected"`
	AgentVersion  string `json:"AgentVersion"`
	LastUsefulAt  string `json:"LastUsefulAt"`
	LastQueriedAt string `json:"LastQueriedAt"`
}

func (c *Client) RoutingTable(ctx context.Context, host string) (*RoutingTableResponse, error) {
	q := url.Values{}
	q.Add("arg", "wan")

	u := url.URL{
		Scheme:   "http",
		Host:     fmt.Sprintf("%s:%s", host, DefaultPort),
		Path:     "/api/v0/stats/dht",
		RawQuery: q.Encode(),
	}

	// Build request
	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	// Fire request
	resp, err := c.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("could not query api: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read routing table response body: %w", err)
	}

	rtResult := RoutingTableResponse{}
	err = json.Unmarshal(data, &rtResult)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal routing table response: %w", err)
	}

	return &rtResult, nil
}
