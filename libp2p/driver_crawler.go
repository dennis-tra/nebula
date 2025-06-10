package libp2p

import (
	"encoding/binary"
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/record"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"github.com/libp2p/go-msgio"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/kubo"
	"github.com/dennis-tra/nebula-crawler/utils"
)

type PeerInfo struct {
	peer.AddrInfo
}

var _ core.PeerInfo[PeerInfo] = (*PeerInfo)(nil)

func (p PeerInfo) ID() peer.ID {
	return p.AddrInfo.ID
}

func (p PeerInfo) Addrs() []ma.Multiaddr {
	return p.AddrInfo.Addrs
}

func (p PeerInfo) Merge(other PeerInfo) PeerInfo {
	if p.AddrInfo.ID != other.AddrInfo.ID {
		panic("merge peer ID mismatch")
	}

	return PeerInfo{
		AddrInfo: peer.AddrInfo{
			ID:    p.AddrInfo.ID,
			Addrs: utils.MergeMaddrs(p.AddrInfo.Addrs, other.AddrInfo.Addrs),
		},
	}
}

func (p PeerInfo) DiscoveryPrefix() uint64 {
	kadID := kbucket.ConvertPeerID(p.AddrInfo.ID)
	return binary.BigEndian.Uint64(kadID[:8])
}

func (p PeerInfo) DeduplicationKey() string {
	return p.AddrInfo.ID.String()
}

type CrawlDriverConfig struct {
	Version        string
	WorkerCount    int
	Network        config.Network
	Protocols      []string
	DialTimeout    time.Duration
	CheckExposed   bool
	BootstrapPeers []peer.AddrInfo
	AddrDialType   config.AddrType
	MeterProvider  metric.MeterProvider
	TracerProvider trace.TracerProvider
	GossipSubPX    bool
	LogErrors      bool
}

func (cfg *CrawlDriverConfig) CrawlerConfig() *CrawlerConfig {
	crawlerCfg := DefaultCrawlerConfig()
	crawlerCfg.DialTimeout = cfg.DialTimeout
	crawlerCfg.CheckExposed = cfg.CheckExposed
	crawlerCfg.AddrDialType = cfg.AddrDialType
	crawlerCfg.GossipSubPX = cfg.GossipSubPX
	crawlerCfg.LogErrors = cfg.LogErrors
	return crawlerCfg
}

func (cfg *CrawlDriverConfig) WriterConfig() *core.CrawlWriterConfig {
	return &core.CrawlWriterConfig{}
}

type CrawlDriver struct {
	cfg   *CrawlDriverConfig
	hosts map[peer.ID]*Host
	dbc   db.Client

	// pxPeersChan receives peers that we get to know
	// via the GossipSub Peer Exchange PX mechanism
	pxPeersChan chan []PeerInfo

	// tasksChan will be read by the engine and allows
	// the driver to submit tasks (peers to crawl) to
	// the engine. In the case when GossipSubPX is disabled,
	// the channel will be closed immediately after the
	// bootstrap peers have been sent to it. If GossipSubPX
	// is enabled, the driver will sent new peers to it as
	// they are discovered via GossipSub.
	tasksChan chan PeerInfo

	// workerStateChan receives the strings "busy" or "idle"
	// from the workers. This is used by the GossipSub monitoring
	// go routine to determine if the crawl is still running.
	// If all workers are idle for more than 1s it'll stop listening
	// for GossipSub messages.
	workerStateChan chan string
	crawlerCount    int
	writerCount     int
}

var _ core.Driver[PeerInfo, core.CrawlResult[PeerInfo]] = (*CrawlDriver)(nil)

func NewCrawlDriver(dbc db.Client, cfg *CrawlDriverConfig) (*CrawlDriver, error) {
	// The Avail light clients verify the agent version:
	// https://github.com/availproject/avail-light/blob/0ddc5d50d6f3d7217c448d6d008846c6b8c4fec3/src/network/p2p/event_loop.rs#L296
	// Spoof it
	userAgent := "nebula/" + cfg.Version
	if cfg.Network == config.NetworkAvailTuringLC || cfg.Network == config.NetworkAvailMainnetLC {
		userAgent = "avail-light-client/light-client/1.12.13/rust-client"
	}

	hosts := make(map[peer.ID]*Host, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		h, err := newLibp2pHost(userAgent)
		if err != nil {
			return nil, fmt.Errorf("new libp2p host: %w", err)
		}
		hosts[h.ID()] = h
	}

	tasksChan := make(chan PeerInfo, len(cfg.BootstrapPeers))
	for _, addrInfo := range cfg.BootstrapPeers {
		addrInfo := addrInfo
		tasksChan <- PeerInfo{AddrInfo: addrInfo}
	}

	d := &CrawlDriver{
		cfg:             cfg,
		hosts:           hosts,
		dbc:             dbc,
		tasksChan:       tasksChan,
		pxPeersChan:     make(chan []PeerInfo),
		workerStateChan: make(chan string, cfg.WorkerCount),
		crawlerCount:    0,
		writerCount:     0,
	}

	if cfg.GossipSubPX {
		go d.monitorGossipSubPX()

		for _, h := range hosts {
			for _, protID := range pubsub.GossipSubDefaultProtocols {
				h.SetStreamHandler(protID, d.handleGossipSubStream)
			}
		}
	} else {
		close(tasksChan)
	}

	return d, nil
}

func (d *CrawlDriver) NewWorker() (core.Worker[PeerInfo, core.CrawlResult[PeerInfo]], error) {
	hostsList := make([]string, 0, len(d.hosts))
	for _, h := range d.hosts {
		hostsList = append(hostsList, string(h.ID()))
	}
	sort.Strings(hostsList)
	hostID := peer.ID(hostsList[d.crawlerCount%len(d.hosts)])

	ms := &msgSender{
		h:         d.hosts[hostID].Host,
		protocols: protocol.ConvertFromStrings(d.cfg.Protocols),
		timeout:   d.cfg.DialTimeout,
	}

	pm, err := pb.NewProtocolMessenger(ms)
	if err != nil {
		return nil, fmt.Errorf("new protocol messenger: %w", err)
	}

	c := &Crawler{
		id:        fmt.Sprintf("crawler-%02d", d.crawlerCount),
		host:      d.hosts[hostID],
		pm:        pm,
		psTopics:  make(map[string]struct{}),
		cfg:       d.cfg.CrawlerConfig(),
		client:    kubo.NewClient(),
		stateChan: d.workerStateChan,
	}

	d.crawlerCount += 1

	return c, nil
}

func (d *CrawlDriver) NewWriter() (core.Worker[core.CrawlResult[PeerInfo], core.WriteResult], error) {
	w := core.NewCrawlWriter[PeerInfo](fmt.Sprintf("writer-%02d", d.writerCount), d.dbc, d.cfg.WriterConfig())
	d.writerCount += 1
	return w, nil
}

func (d *CrawlDriver) Tasks() <-chan PeerInfo {
	return d.tasksChan
}

func (d *CrawlDriver) Close() {
	shutdown := make(chan struct{})
	go func() {
		d.shutdown()
		close(shutdown)
	}()

	select {
	case <-shutdown:
	case <-time.After(10 * time.Second):
		log.Warnln("shutdown timed out")
	}
}

func (d *CrawlDriver) shutdown() {
	var wgHostClose sync.WaitGroup
	for _, h := range d.hosts {
		wgHostClose.Add(1)
		go func(h *Host) {
			defer wgHostClose.Done()
			if err := h.Close(); err != nil {
				log.WithError(err).Warnln("failed to close host")
			}
		}(h)
	}

	var wgTasksClose sync.WaitGroup
	wgTasksClose.Add(1)
	go func() {
		for range d.tasksChan {
		}
		wgTasksClose.Done()
	}()

	wgHostClose.Wait()
	close(d.pxPeersChan)
	wgTasksClose.Wait()
}

func (d *CrawlDriver) monitorGossipSubPX() {
	defer close(d.tasksChan)

	timeout := time.Second
	timer := time.NewTimer(math.MaxInt64)

	busyWorkers := 0
LOOP:
	for {
		select {
		case <-timer.C:
			log.Infof("All workers idle for %s. Stop monitoring gossipsub streams.", timeout)
			break LOOP
		case state := <-d.workerStateChan:
			switch state {
			case "busy":
				busyWorkers += 1
			case "idle":
				busyWorkers -= 1
			}
			if busyWorkers == 0 {
				timer.Reset(timeout)
			} else {
				timer.Reset(math.MaxInt64)
			}
		case pxPeers, more := <-d.pxPeersChan:
			if !more {
				break LOOP
			}

			log.Infof("Discovered %d peers via gossipsub\n", len(pxPeers))
			for _, pxPeer := range pxPeers {
				d.tasksChan <- pxPeer
			}
		}
	}
}

func newLibp2pHost(userAgent string) (*Host, error) {
	// Configure the resource manager to not limit anything
	// Don't use a connection manager that could potentially
	// prune any connections. We clean up after ourselves.
	cm := connmgr.NullConnMgr{}
	rm := network.NullResourceManager{}

	// Initialize a single libp2p node that's shared between all crawlers.
	h, err := libp2p.New(
		libp2p.UserAgent(userAgent),
		libp2p.ResourceManager(&rm),
		libp2p.ConnectionManager(cm),
		libp2p.DisableMetrics(),
		libp2p.SwarmOpts(swarm.WithReadOnlyBlackHoleDetector()),
		libp2p.UDPBlackHoleSuccessCounter(nil),
		libp2p.IPv6BlackHoleSuccessCounter(nil),
	)
	if err != nil {
		return nil, fmt.Errorf("new libp2p host: %w", err)
	}

	return WrapHost(h)
}

// handleGossipSubStream manages a GossipSub stream between two peers. It first
// read the "hello" gossip sub message which contains all active subscriptions
// by the remote peer. Then we open and outgoing stream to the remote peer and
// just mirror back the subscriptions. Then we try to graft the remote peer on
// all topics in the hope that we are the one that exceeds Dhi for the mesh of
// the remote peer. In that case the peer will send a PRUNE message to us that
// contains signed peer records which we'll feed into the [core.Engine] to
// continue the crawl.
func (d *CrawlDriver) handleGossipSubStream(incomingStream network.Stream) {
	defer incomingStream.Reset()

	remoteID := incomingStream.Conn().RemotePeer()
	localID := incomingStream.Conn().LocalPeer()

	// read hello RPC from the remote to get to know all subscriptions
	helloRPC, err := readRPC(incomingStream)
	if err != nil {
		return
	}

	outgoingStream, err := d.hosts[localID].NewStream(context.Background(), remoteID, pubsub.GossipSubDefaultProtocols...)
	if err != nil {
		return
	}
	defer outgoingStream.Reset()

	// let the remote know that we're interested in the exact same topics
	if err = writeRPC(outgoingStream, helloRPC); err != nil {
		return
	}

	graftRPC := &pubsub_pb.RPC{
		Control: &pubsub_pb.ControlMessage{
			Graft: make([]*pubsub_pb.ControlGraft, len(helloRPC.GetSubscriptions())),
		},
	}
	for i, sub := range helloRPC.GetSubscriptions() {
		graftRPC.Control.Graft[i] = &pubsub_pb.ControlGraft{
			TopicID: sub.Topicid,
		}
	}

	// graft the remote peer on all topics we got to know in the hello RPC
	if err = writeRPC(outgoingStream, graftRPC); err != nil {
		return
	}

	// wait until we get a PRUNE
	for {
		rpc, err := readRPC(incomingStream)
		if err != nil {
			return
		}

		pxPeers := make([]PeerInfo, 0)
		for _, prune := range rpc.GetControl().GetPrune() {
			for _, p := range prune.GetPeers() {
				addrInfo, err := parseSignedPeerRecord(p.SignedPeerRecord)
				if err != nil {
					log.WithError(err).Debugln("failed to parse signed peer record")
					continue
				}
				pxPeers = append(pxPeers, PeerInfo{AddrInfo: *addrInfo})
			}
		}

		if len(pxPeers) > 0 {
			d.pxPeersChan <- pxPeers
			return
		}
	}
}

func readRPC(s network.Stream) (*pubsub_pb.RPC, error) {
	data, err := msgio.NewVarintReader(s).ReadMsg()
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	rpc := &pubsub_pb.RPC{}
	if err = rpc.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rpc message: %w", err)
	}

	return rpc, nil
}

func writeRPC(s network.Stream, rpc *pubsub_pb.RPC) error {
	data, err := rpc.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal rpc message: %w", err)
	}

	if err = msgio.NewVarintWriter(s).WriteMsg(data); err != nil {
		return fmt.Errorf("failed to read message: %w", err)
	}

	return nil
}

// parseSignedPeerRecord extracts peer information from a signed peer record.
// It validates and unmarshals the record to return a peer.AddrInfo instance.
// Returns an error if the record is invalid or unmarshalling fails.
func parseSignedPeerRecord(signedPeerRecord []byte) (*peer.AddrInfo, error) {
	envelope, err := record.UnmarshalEnvelope(signedPeerRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal signed peer record: %w", err)
	}

	pid, err := peer.IDFromPublicKey(envelope.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive peer ID: %s", err)
	}
	r, err := envelope.Record()
	if err != nil {
		return nil, fmt.Errorf("failed to obtain record: %w", err)
	}

	rec, ok := r.(*peer.PeerRecord)
	if !ok {
		return nil, fmt.Errorf("not a peer record")
	}

	addrInfo := peer.AddrInfo{
		ID:    pid,
		Addrs: rec.Addrs,
	}

	return &addrInfo, nil
}

// openInboundGossipSubStreams counts inbound GossipSub protocol streams.
// Returns the total number of open inbound GossipSub streams.
func openInboundGossipSubStreams(h host.Host, pid peer.ID) int {
	openStreams := 0
	for _, conn := range h.Network().ConnsToPeer(pid) {
		for _, stream := range conn.GetStreams() {
			if stream.Stat().Direction != network.DirInbound {
				continue
			}
			switch stream.Protocol() {
			case pubsub.GossipSubID_v10:
			case pubsub.GossipSubID_v11:
			case pubsub.GossipSubID_v12:
			case pubsub.FloodSubID:
			default:
				continue
			}
			openStreams += 1
		}
	}
	return openStreams
}
