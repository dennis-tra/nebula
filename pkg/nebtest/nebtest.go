package nebtest

import (
	"context"
	"strconv"
)

type PeerInfo struct {
	Id string
}

func (p *PeerInfo) ID() string {
	return p.Id
}

type CrawlResult struct {
	peerID string
}

func (cr *CrawlResult) PeerID() string {
	return cr.peerID
}

type Crawler struct{}

func (t *Crawler) Work(ctx context.Context, task *PeerInfo) (*CrawlResult, error) {
	return &CrawlResult{
		peerID: task.Id,
	}, nil
}

type Persister struct{}

func (p *Persister) Work(ctx context.Context, task *CrawlResult) (bool, error) {
	return true, nil
}

type Worker struct {
	WorkHook func(ctx context.Context, task string) (int, error)
}

func NewWorker() *Worker {
	return &Worker{}
}

func (w *Worker) Work(ctx context.Context, task string) (int, error) {
	if w.WorkHook != nil {
		return w.WorkHook(ctx, task)
	}
	return strconv.Atoi(task)
}
