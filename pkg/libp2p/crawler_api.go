package libp2p

import (
	"context"
	"errors"
	"fmt"

	manet "github.com/multiformats/go-multiaddr/net"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/dennis-tra/nebula-crawler/pkg/api"
)

type APIResult struct {
	// Indicates if we actually found IP addresses to probe
	Attempted bool

	// The ID response object from the Kubo API
	ID *api.IDResponse

	// The Kubo routing table. Doesn't contain multi addresses. Don't use this to continue crawling.
	RoutingTable *api.RoutingTableResponse
}

func (c *Crawler) crawlAPI(ctx context.Context, pi PeerInfo) <-chan APIResult {
	resultCh := make(chan APIResult)

	// if Nebula is configured to not check for an exposed API return early
	if !c.cfg.CheckExposed {
		close(resultCh)
		return resultCh
	}

	go func() {
		crawledIPs := map[string]struct{}{}
		for _, maddr := range pi.Addrs() {

			// extract IP address from multi address
			ip, err := manet.ToIP(maddr)
			if err != nil {
				log.WithField("maddr", maddr).WithError(err).Debugln("Could not parse IP from Multiaddr")
				continue
			}

			// check if we have already crawled this IP address. A peer usually advertises the same IP address
			// multiple times. E.g., for TCP/QUIC transports. We don't want to crawl an IP twice or more.
			if _, alreadyCrawled := crawledIPs[ip.String()]; alreadyCrawled {
				continue
			}
			crawledIPs[ip.String()] = struct{}{}

			// declare responses
			var (
				idResp *api.IDResponse
				rtResp *api.RoutingTableResponse
			)

			// init timeout context
			tCtx, cancel := context.WithTimeout(ctx, api.RequestTimeout)

			// start both requests in parallel and stop if either fails
			errg := errgroup.Group{}
			errg.Go(func() error {
				idResp, err = c.client.ID(tCtx, ip.String())
				if err != nil {
					return fmt.Errorf("could not crawl ID api: %w", err)
				}
				return nil
			})

			// Only crawl routing table if we actually want to persist neighbors. The result from this API
			// call cannot be used to continue our crawls because the response does not contain multiaddresses
			// of remote peers.
			if c.cfg.TrackNeighbors {
				errg.Go(func() error {
					rtResp, err = c.client.RoutingTable(tCtx, ip.String())
					if err != nil {
						return fmt.Errorf("could not crawl routing table api: %w", err)
					}
					return nil
				})
			}

			// wait for an error or two successes
			err = errg.Wait()
			if errors.Is(err, context.Canceled) {
				cancel()
				break // properly closes the channel
			} else if err != nil {
				log.WithField("maddr", maddr).WithError(err).Debugln("Could not crawl api")
				cancel()
				continue
			}
			cancel()

			// Report result back
			result := APIResult{
				Attempted:    true,
				ID:           idResp,
				RoutingTable: rtResp,
			}

			select {
			case resultCh <- result:
			case <-ctx.Done():
			}

			// since we have what we want, close the channel and return
			close(resultCh)

			return
		}

		select {
		case resultCh <- APIResult{Attempted: len(crawledIPs) > 0}:
		case <-ctx.Done():
		}

		// if crawling the API didn't succeed, just close the channel to indicate that we're done
		close(resultCh)
	}()

	return resultCh
}
