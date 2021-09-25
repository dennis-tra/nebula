package ping

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	madns "github.com/multiformats/go-multiaddr-dns"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

var pingerID = atomic.NewInt32(0)

// Result captures data that is gathered from crawling a single peer.
type Result struct {
	// The pinger that generated this result
	PingerID string

	// The pinged peer
	Peer peer.AddrInfo

	// The latency to the particular peer as measured via ICM ping packets
	PingLatencies []*models.Latency
}

type Job struct {
	pi     peer.AddrInfo
	dbpeer *models.Peer
}

// Pinger encapsulates a libp2p host that crawls the network.
type Pinger struct {
	*service.Service
	config      *config.Config
	pingedPeers int
}

// NewPinger initializes a new pinger based on the given configuration.
func NewPinger(conf *config.Config) (*Pinger, error) {
	p := &Pinger{
		Service: service.New(fmt.Sprintf("pinger-%02d", pingerID.Load())),
		config:  conf,
	}
	pingerID.Inc()

	return p, nil
}

// StartPinging enters an endless loop and consumes measure jobs from the measure queue
// and publishes its result on the results queue until it is told to stop or the
// measure queue was closed.
func (c *Pinger) StartPinging(measureQueue *queue.FIFO, resultsQueue *queue.FIFO) {
	c.ServiceStarted()
	defer c.ServiceStopped()
	ctx := c.ServiceContext()

	for {
		// Give the shutdown signal precedence
		select {
		case <-c.SigShutdown():
			return
		default:
		}

		select {
		case elem, ok := <-measureQueue.Consume():
			if !ok {
				// The crawl queue was closed
				return
			}
			result := c.handleMeasureJob(ctx, elem.(Job))
			resultsQueue.Push(result)
		case <-c.SigShutdown():
			return
		}
	}
}

// handleMeasureJob takes a measure job and pings the given peer.
func (c *Pinger) handleMeasureJob(ctx context.Context, job Job) Result {
	logEntry := log.WithFields(log.Fields{
		"pingerID":  c.Identifier(),
		"targetID":  job.pi.ID.Pretty()[:16],
		"pingCount": c.pingedPeers,
	})
	logEntry.Debugln("Pinging peer")
	defer logEntry.Debugln("Pinged peer")

	latencies := c.measureLatencies(ctx, job)

	return Result{
		PingerID:      c.Identifier(),
		Peer:          job.pi,
		PingLatencies: latencies,
	}
}

// measureLatencies measures the ICM ping latency to all addresses of the given peer.
func (c *Pinger) measureLatencies(ctx context.Context, job Job) []*models.Latency {
	// TODO: The following three steps can probably be consolidated. In the current state it's quite messy.

	pi := job.pi

	// Only consider publicly reachable multi-addresses
	pi = filterPrivateMaddrs(pi)

	// Resolve DNS multi addresses to IP addresses (especially maddrs with the dnsaddr protocol)
	pi.Addrs = resolveAddrs(ctx, pi)

	// The following loops extract addresses from the AddrInfo multi-addresses.
	// The set of multi addresses could contain multiple maddrs with the
	// same IPv4/IPv6 addresses. This loop de-duplicates that.
	// TODO: To which address should the ping messages be sent? Currently it's to all found addresses.
	// TODO: The deduplication can probably be implemented a little bit prettier
	addrsMap := map[string]string{}
	for _, maddr := range pi.Addrs {
		for _, pr := range []int{ma.P_IP4, ma.P_IP6} { // DNS protocols are stripped via resolveAddrs above
			if addr, err := maddr.ValueForProtocol(pr); err == nil {
				addrsMap[addr] = addr
				break
			}
		}
	}

	// Exit early if there is no address
	if len(addrsMap) == 0 {
		return nil
	}

	// Start sending ping messages to all IP addresses in parallel.
	var wg sync.WaitGroup
	results := make(chan interface{})
	for addr := range addrsMap {

		// Configure the new pinger instance
		pinger, err := ping.NewPinger(addr)
		if err != nil {
			log.WithError(errors.Wrap(err, "new pinger")).Warnln("Error instantiating new pinger")
			continue
		}

		pinger.Timeout = time.Minute
		pinger.Count = 10

		// This Go routine allows reacting to context cancellations (e.g., user presses ^C)
		// The done channel is necessary to not leak this go routine after the pinger has finished.
		done := make(chan struct{})
		go func() {
			select {
			case <-done:
			case <-ctx.Done():
				pinger.Stop()
			}
		}()

		// This Go routine starts sending ICM pings to the address configured above.
		// After it has terminated (successfully or erroneously) it sends the result
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(done)
			// Blocks until finished.
			if err := pinger.Run(); err != nil {
				results <- err
			} else {
				results <- pinger.Statistics()
			}
		}()
	}

	// Since we're ranging over the results channel below we need to
	// know when the pinger Go routines are done. This the case
	// after the wg.Wait() call returns. We close the channel and
	// break out of the for loop below.
	go func() {
		wg.Wait()
		close(results)
	}()

	var latencies []*models.Latency
	for result := range results {
		switch res := result.(type) {
		case error:
			log.WithError(errors.Wrap(res, "pinger run")).Warnln("Error pinging peer")
		case *ping.Statistics:
			latencies = append(latencies, &models.Latency{
				PeerID:          job.dbpeer.ID,
				Address:         res.Addr,
				PingLatencySAvg: res.AvgRtt.Seconds(),
				PingLatencySSTD: res.StdDevRtt.Seconds(),
				PingLatencySMin: res.MinRtt.Seconds(),
				PingLatencySMax: res.MaxRtt.Seconds(),
				PingPacketsSent: res.PacketsSent,
				PingPacketsRecv: res.PacketsRecv,
				PingPacketsDupl: res.PacketsRecvDuplicates,
				PingPacketLoss:  res.PacketLoss,
			})
		}
	}

	return latencies
}

// filterPrivateMaddrs strips private multiaddrs from the given peer address information.
func filterPrivateMaddrs(pi peer.AddrInfo) peer.AddrInfo {
	filtered := peer.AddrInfo{
		ID:    pi.ID,
		Addrs: []ma.Multiaddr{},
	}

	// Just keep public multi addresses
	for _, maddr := range pi.Addrs {
		if manet.IsPrivateAddr(maddr) {
			continue
		}
		filtered.Addrs = append(filtered.Addrs, maddr) // TODO: Strip relays?
	}

	return filtered
}

// resolveAddrs loops through the multi addresses of the given peer and recursively resolves
// the various DNS protocols (especially the dnsaddr protocol). This implementation is
// taken from:
// https://github.com/libp2p/go-libp2p/blob/9d3fd8bc4675b9cebf3102bdf62e56204c67ce5b/p2p/host/basic/basic_host.go#L676
func resolveAddrs(ctx context.Context, pi peer.AddrInfo) []ma.Multiaddr {
	proto := ma.ProtocolWithCode(ma.P_P2P).Name
	p2paddr, err := ma.NewMultiaddr("/" + proto + "/" + pi.ID.Pretty())
	if err != nil {
		return []ma.Multiaddr{}
	}

	resolveSteps := 0

	// Recursively resolve all addrs.
	//
	// While the toResolve list is non-empty:
	// * Pop an address off.
	// * If the address is fully resolved, add it to the resolved list.
	// * Otherwise, resolve it and add the results to the "to resolve" list.
	toResolve := append(([]ma.Multiaddr)(nil), pi.Addrs...)
	resolved := make([]ma.Multiaddr, 0, len(pi.Addrs))
	for len(toResolve) > 0 {
		// pop the last addr off.
		addr := toResolve[len(toResolve)-1]
		toResolve = toResolve[:len(toResolve)-1]

		// if it's resolved, add it to the resolved list.
		if !madns.Matches(addr) {
			resolved = append(resolved, addr)
			continue
		}

		resolveSteps++

		// otherwise, resolve it
		reqaddr := addr.Encapsulate(p2paddr)
		resaddrs, err := madns.DefaultResolver.Resolve(ctx, reqaddr)
		if err != nil {
			log.Infof("error resolving %s: %s", reqaddr, err)
		}

		// add the results to the toResolve list.
		for _, res := range resaddrs {
			pi, err := peer.AddrInfoFromP2pAddr(res)
			if err != nil {
				log.Infof("error parsing %s: %s", res, err)
			}
			toResolve = append(toResolve, pi.Addrs...)
		}
	}

	return resolved
}
