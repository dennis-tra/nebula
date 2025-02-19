package db

import (
	"fmt"
	"testing"

	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"github.com/dennis-tra/nebula-crawler/utils"
)

func TestKnownErrorsMatchKnownErrorsPrecedence(t *testing.T) {
	assert.Equal(t, len(KnownErrors), len(knownErrorsPrecedence))

	for _, errStr := range knownErrorsPrecedence {
		_, found := KnownErrors[errStr]
		assert.True(t, found, "%s not in KnownErrors", errStr)
	}
}

func TestDialErrors(t *testing.T) {
	type args struct {
		maddrs []ma.Multiaddr
		err    error
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty",
			args: args{},
			want: []string{},
		},
		{
			name: "single timeout",
			args: args{
				maddrs: []ma.Multiaddr{
					utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1234"),
				},
				err: &swarm.DialError{
					DialErrors: []swarm.TransportError{
						{
							Address: utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1234"),
							Cause:   fmt.Errorf("i/o timeout"),
						},
					},
				},
			},
			want: []string{"io_timeout"},
		},
		{
			name: "single timeout",
			args: args{
				maddrs: []ma.Multiaddr{
					utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1234"),
				},
				err: &swarm.DialError{
					DialErrors: []swarm.TransportError{
						{
							Address: utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1234"),
							Cause:   fmt.Errorf("i/o timeout"),
						},
					},
				},
			},
			want: []string{"io_timeout"},
		},
		{
			name: "address not dialed",
			args: args{
				maddrs: []ma.Multiaddr{
					utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1234"),
					utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1235"),
				},
				err: &swarm.DialError{
					DialErrors: []swarm.TransportError{
						{
							Address: utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1234"),
							Cause:   fmt.Errorf("i/o timeout"),
						},
					},
				},
			},
			want: []string{"io_timeout", "not_dialed"},
		},
		{
			name: "error for non existing address",
			args: args{
				maddrs: []ma.Multiaddr{
					utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1234"),
					utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1235"),
				},
				err: &swarm.DialError{
					DialErrors: []swarm.TransportError{
						{
							Address: utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/4564"),
							Cause:   fmt.Errorf("i/o timeout"),
						},
					},
				},
			},
			want: []string{"not_dialed", "not_dialed"},
		},
		{
			name: "user canceled",
			args: args{
				maddrs: []ma.Multiaddr{
					utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1234"),
					utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1235"),
				},
				err: context.Canceled,
			},
			want: []string{"canceled", "canceled"},
		},
		{
			name: "no public ip",
			args: args{
				maddrs: []ma.Multiaddr{
					utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1234"),
					utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1235"),
				},
				err: ErrNoPublicIP,
			},
			want: []string{"not_dialed", "not_dialed"},
		},
		{
			name: "dial canceled",
			args: args{
				maddrs: []ma.Multiaddr{
					utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1234"),
				},
				err: &swarm.DialError{
					DialErrors: []swarm.TransportError{
						{
							Address: utils.MustMultiaddr(t, "/ip4/127.0.0.1/tcp/1234"),
							Cause:   context.Canceled,
						},
					},
				},
			},
			want: []string{"canceled"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, MaddrErrors(tt.args.maddrs, tt.args.err), "DialErrors(%v, %v)", tt.args.maddrs, tt.args.err)
		})
	}
}
