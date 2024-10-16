package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	mapset "github.com/deckarep/golang-set/v2"
)

func Test_determineStrategy(t *testing.T) {
	tests := []struct {
		name string
		sets []mapset.Set[string]
		errs []error
		want crawlStrategy
	}{
		{
			// simulates, we received the same response three times (success case)
			name: "all same (3)",
			sets: []mapset.Set[string]{
				mapset.NewThreadUnsafeSet("0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
				mapset.NewThreadUnsafeSet("0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
				mapset.NewThreadUnsafeSet("0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
			},
			want: crawlStrategySingleProbe,
		},
		{
			// simulates, we received the same response two times but also one
			// error. This indicates a flaky connection. Just issue one probe
			// for each bucket but also retry if failed.
			name: "all same with error (2)",
			sets: []mapset.Set[string]{
				mapset.NewThreadUnsafeSet("0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
				mapset.NewThreadUnsafeSet("0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
			},
			errs: []error{
				fmt.Errorf("some error"),
			},
			want: crawlStrategySingleProbe,
		},
		{
			// simulates: remote peer replaced a node in its RT during probing
			name: "single diff full responses (3)",
			sets: []mapset.Set[string]{
				mapset.NewThreadUnsafeSet("A", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
				mapset.NewThreadUnsafeSet("B", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
				mapset.NewThreadUnsafeSet("B", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
			},
			want: crawlStrategySingleProbe,
		},
		{
			name: "single diff full responses (2)",
			sets: []mapset.Set[string]{
				mapset.NewThreadUnsafeSet("A", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
				mapset.NewThreadUnsafeSet("B", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
			},
			want: crawlStrategyMultiProbe,
		},
		{
			name: "partial response, full bucket",
			sets: []mapset.Set[string]{
				mapset.NewThreadUnsafeSet( /* missing */ "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
				mapset.NewThreadUnsafeSet("0", "1", "2" /* missing */, "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
				mapset.NewThreadUnsafeSet("0", "1", "2", "3", "4", "5" /* missing */, "9", "10", "11", "12", "13", "14", "15"),
			},
			want: crawlStrategyMultiProbe,
		},
		{
			// simulates: the weird node that only returns three peers for each
			// request and all of them are different
			name: "weird node (all different responses)",
			sets: []mapset.Set[string]{
				mapset.NewThreadUnsafeSet("0", "1", "2"),
				mapset.NewThreadUnsafeSet("3", "4", "5"),
				mapset.NewThreadUnsafeSet("6", "7", "8"),
			},
			want: crawlStrategyRandomProbe,
		},
		{
			// simulates: the weird node that only returns three peers for each
			// request and all of them are different
			name: "weird node (single overlap responses)",
			sets: []mapset.Set[string]{
				mapset.NewThreadUnsafeSet("0", "1", "2"),
				mapset.NewThreadUnsafeSet("3", "4", "0"),
				mapset.NewThreadUnsafeSet("6", "4", "8"),
			},
			want: crawlStrategyRandomProbe,
		},
		{
			name: "more than 16 peers in each bucket",
			sets: []mapset.Set[string]{
				mapset.NewThreadUnsafeSet("0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
				mapset.NewThreadUnsafeSet("16", "17", "18", "19", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
				mapset.NewThreadUnsafeSet("16", "17", "20", "21", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
			},
			want: crawlStrategyMultiProbe,
		},
		{
			name: "partially filled bucket",
			sets: []mapset.Set[string]{
				mapset.NewThreadUnsafeSet("0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13"),
				mapset.NewThreadUnsafeSet("0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13"),
				mapset.NewThreadUnsafeSet("0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13"),
			},
			want: crawlStrategySingleProbe,
		},
		{
			name: "received v4wire.MaxNeighbors responses, full bucket",
			sets: []mapset.Set[string]{
				mapset.NewThreadUnsafeSet("0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"),
				mapset.NewThreadUnsafeSet("2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13"),
				mapset.NewThreadUnsafeSet("5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"),
			},
			want: crawlStrategyMultiProbe,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineStrategy(tt.sets, tt.errs)
			assert.Equal(t, tt.want, got)
		})
	}
}
