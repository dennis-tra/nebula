package main

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/discover/v4wire"

	mapset "github.com/deckarep/golang-set/v2"

	"golang.org/x/net/context"

	"github.com/dennis-tra/nebula-crawler/discv4"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

var bootnodes = []string{
	"enode://cae041158dabbe836c20342f85e215062e44b2c24d8d6b4e665b7132e2b924d735ffbc1d0fe3ee3e8e484bd9745a43f9eb6daf333e5920b79b94028058d1eaa6@38.242.152.98:30303",
	"enode://96799ccd3127a6f492393c96d3ac6a1fbf5e1df1ea33fe21c955e70a5085989429c3a09b4e32bce1f58477b6a73850e97e29477707df9f43e439280093fa3437@185.202.238.34:30303?discport=30306",
	"enode://834e2cf212cf9d63dded7660d898223bc28893750a93e20c319e763c11298aa5c626572a051d3a9a79da9ba92a88863237723505e816230840a035934bf54539@131.153.232.204:30370",
	"enode://47d7374d0eaacbdc9b24daf28360a1e0fae9057215cd609b648aff7ce6aef154b47a1f8980b1cb78da6784f1e17d7642f79232c3fd34bf631ee145e90c165f46@15.204.152.148:30303",
	"enode://e27b71e7f8c4a4c0e525b64fd8ea160d106df8f06bd655ae735696ac82f4865ea6d137a2b18f1264395a2047bdd52c036b25d28de5a3f4cacf2984f1a3172d77@3.219.67.1:30313",
	"enode://e386566f35e376d6d9ae91678ac7c788fa0e558b449f66213207365e4c9b6173dd65b948a1f1e5b1178a7b04fe4d2f43b88054d9763405f994187cabdf39e6fd@44.213.54.129:30308",
	"enode://d860a01f9722d78051619d1e2351aba3f43f943f6f00718d1b9baa4101932a1f5011f16bb2b1bb35db20d6fe28fa0bf09636d26a87d31de9ec6203eeedb1f666@18.138.108.67:30303", // bootnode-aws-ap-southeast-1-001
	"enode://22a8232c3abc76a16ae9d6c3b164f98775fe226f0917b0ca871128a74a8e9630b458460865bab457221f1d448dd9791d24c4e5d88786180ac185df813a68d4de@3.209.45.79:30303",   // bootnode-aws-us-east-1-001
	"enode://2b252ab6a1d0f971d9722cb839a42cb81db019ba44c08754628ab4a823487071b5695317c8ccd085219c3a03af063495b2f1da8d18218da2d6a82981b45e6ffc@65.108.70.101:30303", // bootnode-hetzner-hel
	"enode://4aeb4ab6c14b23e2c4cfdce879c04b0748a20d8e9b59e25ded2a08143e265c6c25936e74cbc8e641e3312ca288673d91f2f93f8e277de3cfa444ecdaaf982052@157.90.35.166:30303", // bootnode-hetzner-fsn
}

func main() {
	enr := bootnodes[0]
	node, _ := enode.Parse(enode.ValidSchemes, enr)
	ipAddr, _ := netip.AddrFromSlice(node.IP())
	udpAddr := netip.AddrPortFrom(ipAddr, uint16(node.UDP()))

	priv, _ := ecdsa.GenerateKey(ethcrypto.S256(), crand.Reader)
	ps, _ := enode.OpenDB("")
	ethNode := enode.NewLocalNode(ps, priv)

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		panic(err)
	}
	config := discv4.DefaultClientConfig()
	config.DialTimeout = 5 * time.Second
	client := discv4.NewClient(priv, config)

	discv4Cfg := discover.Config{
		PrivateKey:   priv,
		ValidSchemes: enode.ValidSchemes,
	}

	d, _ := discover.ListenV4(conn, ethNode, discv4Cfg)
	defer d.Close()

	pi, _ := discv4.NewPeerInfo(node)

	c, err := client.Connect(context.TODO(), pi)
	if err != nil {
		fmt.Println(err)
	} else {
		hello, status, err := c.Identify()
		if err != nil {
			fmt.Println(err)
		}
		if hello != nil {
			fmt.Println("Name:", hello.Name)
			fmt.Println("Version:", hello.Version)
			fmt.Println("Caps:", hello.Caps)
		}
		if status != nil {
			fmt.Println("ForkID:", hex.EncodeToString(status.ForkID.Hash[:]))
			fmt.Println("NetworkID:", status.NetworkID)
		}
	}

	crawlBucketsSequentially(node, udpAddr, d)
	// crawlBucketsConcurrently(node, udpAddr, d)
	// crawlBucketsSmartly(node, udpAddr, d)
}

func crawlBucketsSmartly(node *enode.Node, udpAddr netip.AddrPort, d *discover.UDPv4) error {
	fmt.Println("Crawling", node.ID().String()[:16])

	type findNodeResult struct {
		nodes []*enode.Node
		set   mapset.Set[string]
		err   error
	}

	var wg sync.WaitGroup
	probes := 3

	var mu sync.RWMutex
	results := make(chan findNodeResult)
	for i := 0; i < probes; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			mu.Lock()
			defer mu.Unlock()

			targetKey, err := discv4.GenRandomPublicKey(node.ID(), 0)
			if err != nil {
				results <- findNodeResult{err: err}
				return
			}

			fmt.Printf("  Finding closest node in bucket %d to ID: %s\n", 0, targetKey.ID().String()[:16])
			nodes, err := d.FindNode(node.ID(), udpAddr, targetKey)

			ids := make([]string, 0, len(nodes))
			for _, c := range nodes {
				ids = append(ids, c.ID().String())
			}
			results <- findNodeResult{
				nodes: nodes,
				set:   mapset.NewThreadUnsafeSet(ids...),
				err:   err,
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var (
		sets []mapset.Set[string]
		errs []error
	)

	for result := range results {
		if result.err != nil {
			errs = append(errs, result.err)
			continue
		}

		// if a remote peer has BucketSize (16) peers in a bucket, it will send
		// them in two batches. First, it only sends [v4wire.MaxNeighbors] (12)
		// peers and then the remaining. The reason is that only 12 peers will
		// fit into a single UDP packet. Receiving only 12 peers from a remote
		// peer is a strong indication that the second packet arrived too late
		// as the response timeout is quite aggressive (500ms). This means
		// if one FindNode response only contains 12 peers we likely ran into a
		// timeout for the second packet which was then discarded. One solution
		// here is to probe the same bucket multiple times with different keys.
		//
		// An alternative explanation is that the peer really only has 12 peers
		// in its bucket. We can identify this case with our multiple probes.
		//if result.set.Cardinality() == v4wire.MaxNeighbors {
		//	hasMaxNeighborsResponse = true
		//}

		sets = append(sets, result.set)
	}

	if len(errs) == probes {
		return fmt.Errorf("probing bucket 0: %w", errors.Join(errs...))
	}

	switch determineStrategy(sets, errs) {
	case crawlStrategySingleProbe:
		fmt.Println("  Using single-probe strategy")
	case crawlStrategyMultiProbe:
		fmt.Println("  Using multi-probe strategy")
	case crawlStrategyRandomProbe:
		fmt.Println("  Using random-probe strategy")
	}

	return nil
}

type crawlStrategy string

const (
	crawlStrategySingleProbe crawlStrategy = "single-probe"
	crawlStrategyMultiProbe  crawlStrategy = "multi-probe"
	crawlStrategyRandomProbe crawlStrategy = "random-probe"
)

func determineStrategy(sets []mapset.Set[string], errs []error) crawlStrategy {
	// Calculate the average difference between two responses. If the response
	// sizes are always 16, one new peer will result in a symmetric difference
	// of cardinality 2. One peer in the first set that's not in the second and one
	// peer in the second that's not in the first set. We consider that it's the
	// happy path if the average symmetric difference is less than 2.
	avgSymDiff := float32(0)
	diffCount := float32(0)
	allNodes := mapset.NewThreadUnsafeSet[string]()
	for i := 0; i < len(sets); i++ {
		allNodes = allNodes.Union(sets[i])
		for j := i + 1; j < len(sets); j++ {
			diffCount += 1
			avgSymDiff += float32(sets[i].SymmetricDifference(sets[j]).Cardinality())
		}
	}
	if diffCount > 0 {
		avgSymDiff /= diffCount
	}

	switch {
	case avgSymDiff < 2:
		return crawlStrategySingleProbe
	case allNodes.Cardinality() > v4wire.MaxNeighbors:
		return crawlStrategyMultiProbe
	default:
		return crawlStrategyRandomProbe
	}
}

func crawlBucketsConcurrently(node *enode.Node, udpAddr netip.AddrPort, d *discover.UDPv4, probesPerBucket int) {
	fmt.Println("Crawling", node.ID().String()[:16])

	var wg sync.WaitGroup

	allNeighborsMu := sync.Mutex{}
	allNeighbors := map[string]*enode.Node{}
	for i := 1; i < 15; i++ {
		for j := 0; j < probesPerBucket; j++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				targetKey, _ := discv4.GenRandomPublicKey(node.ID(), i)
				fmt.Println("  Finding closest node to ID: ", targetKey.ID().String()[:16])
				closest, err := d.FindNode(node.ID(), udpAddr, targetKey)
				if err != nil {
					fmt.Println("")
					return
				}

				allNeighborsMu.Lock()
				defer allNeighborsMu.Unlock()

				neighbors := make([]string, 0, len(closest))
				for _, c := range closest {
					neighbors = append(neighbors, c.ID().String())
					allNeighbors[c.ID().String()] = c
				}
				sort.Strings(neighbors)
				for _, n := range neighbors {
					fmt.Println("    ", n[:16])
				}
			}()
		}
	}
	wg.Wait()
	fmt.Println("Unique neighbors found:", len(allNeighbors))
}

func crawlBucketsSequentially(node *enode.Node, udpAddr netip.AddrPort, d *discover.UDPv4) {
	fmt.Println("Crawling", node.ID().String()[:16])

	allNeighbors := map[string]struct{}{}
	for i := 0; i < 15; i++ {
		targetKey, _ := discv4.GenRandomPublicKey(node.ID(), i)
		fmt.Println("  Finding closest node to ID: ", targetKey.ID().String()[:16])
		closest, err := d.FindNode(node.ID(), udpAddr, targetKey)
		if err != nil {
			fmt.Println("")
			continue
		}

		neighbors := make([]string, 0, len(closest))
		for _, c := range closest {
			neighbors = append(neighbors, c.ID().String())
			allNeighbors[c.ID().String()] = struct{}{}
		}
		sort.Strings(neighbors)
		for _, n := range neighbors {
			fmt.Println("    ", n[:16])
		}
	}
	fmt.Println("Unique neighbors found:", len(allNeighbors))
}
