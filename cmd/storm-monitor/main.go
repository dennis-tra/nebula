package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/net"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	protocol "github.com/libp2p/go-libp2p-protocol"
	routing "github.com/libp2p/go-libp2p-routing"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
)

func main() {
	ctx := context.Background()
	var dualDHT *dual.DHT
	h, err := libp2p.New(
		ctx,
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			var err error
			dualDHT, err = dual.New(ctx, h)
			return dualDHT, err
		}),
	)
	if err != nil {
		fmt.Println(strings.ReplaceAll(err.Error(), "\n", " "))
		return
	}

	fmt.Println("My ID is:", h.ID())
	sender := net.NewMessageSenderImpl(h, []protocol.ID{"/ipfs/kad/1.0.0"})
	msger, err := pb.NewProtocolMessenger(sender)
	if err != nil {
		fmt.Println(strings.ReplaceAll(err.Error(), "\n", " "))
		return
	}
	file, err := os.Open("../../analysis/mixed/storm.list")
	if err != nil {
		fmt.Println(strings.ReplaceAll(err.Error(), "\n", " "))
		return
	}
	defer file.Close()

	res := make(map[int]int)
	resLock := &sync.RWMutex{}

	resP := make([]time.Duration, 0)
	resR := make([]time.Duration, 0)
	timeLock := &sync.RWMutex{}

	scanner := bufio.NewScanner(file)
	i := 0
	var wg sync.WaitGroup
	for scanner.Scan() {
		line := scanner.Text()
		pis := strings.Split(line, " ")
		id, err := peer.IDB58Decode(pis[0])
		if err != nil {
			fmt.Println(strings.ReplaceAll(err.Error(), "\n", " "))
			return
		}
		addr, err := multiaddr.NewMultiaddr(pis[1])
		if err != nil {
			fmt.Println(strings.ReplaceAll(err.Error(), "\n", " "))
			return
		}
		pi := peer.AddrInfo{Addrs: []multiaddr.Multiaddr{addr}, ID: id}
		wg.Add(1)
		go func(pi peer.AddrInfo, i int) {
			defer wg.Done()
			err = h.Connect(ctx, pi)
			if err != nil {
				resLock.Lock()
				defer resLock.Unlock()
				res[i] = 0
				fmt.Printf("Routine %v \t| fail to connect: %v\n", i, strings.ReplaceAll(err.Error(), "\n", " "))
				return
			}
			protos, _ := h.Peerstore().GetProtocols(id)
			exists := false
			for _, proto := range protos {
				if proto == "/ipfs/kad/1.0.0" {
					exists = true
					break
				}
			}
			fmt.Printf("Routine %v \t| start monitoring %v | support kad %v | protos supported %v\n", i, pi.ID, exists, protos)
			j := 0
			for {
				// Random cid
				pref := cid.Prefix{
					Version:  0,
					Codec:    cid.DagProtobuf,
					MhType:   multihash.SHA2_256,
					MhLength: -1, // Default length
				}
				seed := make([]byte, 32)
				rand.Read(seed)
				randomRoot, _ := pref.Sum(seed)
				t0 := time.Now()
				err = msger.PutProvider(ctx, id, randomRoot.Hash(), h)
				if err != nil {
					resLock.Lock()
					defer resLock.Unlock()
					res[i] = j
					fmt.Printf("Routine %v \t| fail to put provider record after %v successful attempts: %v\n", i, j, strings.ReplaceAll(err.Error(), "\n", " "))
					return
				}
				t1 := time.Now()
				time.Sleep(1 * time.Second)
				t2 := time.Now()
				pvds1, _, err := msger.GetProviders(ctx, id, randomRoot.Hash())
				if err != nil {
					resLock.Lock()
					defer resLock.Unlock()
					res[i] = j
					fmt.Printf("Routine %v \t| fail to fetch provider record after %v successful attempts: %v\n", i, j, strings.ReplaceAll(err.Error(), "\n", " "))
					return
				}
				t3 := time.Now()
				if len(pvds1) == 0 {
					resLock.Lock()
					defer resLock.Unlock()
					res[i] = j
					fmt.Printf("Routine %v \t| fetched empty provider record after %v successful attempts\n", i, j)
					return
				} else {
					pvd := pvds1[0]
					if pvd.ID != h.ID() {
						resLock.Lock()
						defer resLock.Unlock()
						res[i] = j
						fmt.Printf("Routine %v \t| fetched wrong provider record after %v successful attempts\n", i, j)
						return
					}
				}
				j++
				fmt.Printf("Routine %v \t| successful put/fetch provider record total %v successful attempts, time taken in publish %v in fetch %v\n", i, j, t1.Sub(t0), t3.Sub(t2))
				resLock.Lock()
				res[i] = j
				resLock.Unlock()
				timeLock.Lock()
				resP = append(resP, t1.Sub(t0))
				resR = append(resR, t3.Sub(t2))
				timeLock.Unlock()
				if j >= 500 {
					return
				}
				time.Sleep(1 * time.Second)
			}
		}(pi, i)
		i += 1
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("Monitoring cancelled, print result.")
		sum := 0
		resLock.RLock()
		defer resLock.RUnlock()
		for routine, success := range res {
			fmt.Printf("Routine %v success count %v\n", routine, success)
			sum += success
		}
		fmt.Printf("Average success count: %v\n", float64(sum)/float64(len(res)))
		out, err := os.Create("../../analysis/mixed/storm.time")
		if err != nil {
			fmt.Println(strings.ReplaceAll(err.Error(), "\n", " "))
			return
		}
		defer out.Close()
		timeLock.RLock()
		defer timeLock.RUnlock()
		w := bufio.NewWriter(out)
		w.Write([]byte("publish\n"))
		for _, t := range resP {
			w.Write([]byte(t.String()))
			w.Write([]byte("\n"))
		}
		w.Write([]byte("fetch\n"))
		for _, t := range resR {
			w.Write([]byte(t.String()))
			w.Write([]byte("\n"))
		}
		panic("monitoring cancelled")
	}()
	wg.Wait()
	fmt.Println("Monitoring done, print result.")
	sum := 0
	for routine, success := range res {
		fmt.Printf("Routine %v success count %v\n", routine, success)
		sum += success
	}
	fmt.Printf("Average success count: %v\n", float64(sum)/float64(len(res)))
	out, err := os.Create("../../analysis/mixed/storm.time")
	if err != nil {
		fmt.Println(strings.ReplaceAll(err.Error(), "\n", " "))
		return
	}
	defer out.Close()
	timeLock.RLock()
	defer timeLock.RUnlock()
	w := bufio.NewWriter(out)
	w.Write([]byte("publish\n"))
	for _, t := range resP {
		w.Write([]byte(t.String()))
		w.Write([]byte("\n"))
	}
	w.Write([]byte("fetch\n"))
	for _, t := range resR {
		w.Write([]byte(t.String()))
		w.Write([]byte("\n"))
	}
	return
}
