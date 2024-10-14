package discv4

import (
	"fmt"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
)

func Test_determineStrategy(t *testing.T) {
	tests := []struct {
		name string
		sets [][]string
		errs []error
		want CrawlStrategy
	}{
		{
			// simulates, we received the same response three times (success case)
			name: "all same (3)",
			sets: [][]string{
				{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
				{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
				{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
			},
			want: crawlStrategySingleProbe,
		},
		{
			// simulates, we received the same response two times but also one
			// error. This indicates a flaky connection. Just issue one probe
			// for each bucket but also retry if failed.
			name: "all same with error (2)",
			sets: [][]string{
				{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
				{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
			},
			errs: []error{
				fmt.Errorf("some error"),
			},
			want: crawlStrategySingleProbe,
		},
		{
			// simulates: remote peer replaced a node in its RT during probing
			name: "single diff full responses (3)",
			sets: [][]string{
				{"A", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
				{"B", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
				{"B", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
			},
			want: crawlStrategySingleProbe,
		},
		{
			name: "single diff full responses (2)",
			sets: [][]string{
				{"A", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
				{"B", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
			},
			want: crawlStrategyMultiProbe,
		},
		{
			name: "partial response, full bucket",
			sets: [][]string{
				{ /* missing */ "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
				{"0", "1", "2" /* missing */, "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
				{"0", "1", "2", "3", "4", "5" /* missing */, "9", "10", "11", "12", "13", "14", "15"},
			},
			want: crawlStrategyMultiProbe,
		},
		{
			// simulates: the weird node that only returns three peers for each
			// request and all of them are different
			name: "weird node (all different responses)",
			sets: [][]string{
				{"0", "1", "2"},
				{"3", "4", "5"},
				{"6", "7", "8"},
			},
			want: crawlStrategyRandomProbe,
		},
		{
			// simulates: the weird node that only returns three peers for each
			// request and all of them are different
			name: "weird node (single overlap responses)",
			sets: [][]string{
				{"0", "1", "2"},
				{"3", "4", "0"},
				{"6", "4", "8"},
			},
			want: crawlStrategyRandomProbe,
		},
		{
			name: "more than 16 peers in each bucket",
			sets: [][]string{
				{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
				{"16", "17", "18", "19", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
				{"16", "17", "20", "21", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
			},
			want: crawlStrategyMultiProbe,
		},
		{
			name: "partially filled bucket",
			sets: [][]string{
				{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13"},
				{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13"},
				{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13"},
			},
			want: crawlStrategySingleProbe,
		},
		{
			name: "received v4wire.MaxNeighbors responses, full bucket",
			sets: [][]string{
				{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
				{"2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13"},
				{"5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
			},
			want: crawlStrategyMultiProbe,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sets []mapset.Set[peer.ID]
			for i, s := range tt.sets {
				sets = append(sets, mapset.NewThreadUnsafeSet[peer.ID]())
				for _, item := range s {
					sets[i].Add(peer.ID(item))
				}
			}
			got := determineStrategy(sets)
			assert.Equal(t, tt.want, got)
		})
	}
}
