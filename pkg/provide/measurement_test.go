package provide

import (
	"fmt"
	"testing"

	pt "github.com/libp2p/go-libp2p-core/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeasurement_detectSpans_singleDial(t *testing.T) {
	local, err := pt.RandPeerID()
	require.NoError(t, err)
	remote, err := pt.RandPeerID()
	require.NoError(t, err)

	m := &Measurement{
		events: []Event{
			&DialStart{
				BaseEvent: NewBaseEvent(local, remote),
			},
			&DialEnd{
				BaseEvent: NewBaseEvent(local, remote),
			},
		},
	}
	spans := m.detectSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, SpanTypeDial, spans[0].Type)
	assert.Equal(t, remote, spans[0].PeerID)
	assert.Nil(t, spans[0].Error)
}

func TestMeasurement_detectSpans_multiDial(t *testing.T) {
	local, err := pt.RandPeerID()
	require.NoError(t, err)
	remote, err := pt.RandPeerID()
	require.NoError(t, err)

	m := &Measurement{
		events: []Event{
			&DialStart{
				BaseEvent: NewBaseEvent(local, remote),
			},
			&DialStart{
				BaseEvent: NewBaseEvent(local, remote),
			},
			&DialEnd{
				BaseEvent: NewBaseEvent(local, remote),
				Err:       fmt.Errorf("some err"),
			},
			&DialEnd{
				BaseEvent: NewBaseEvent(local, remote),
			},
		},
	}
	spans := m.detectSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, SpanTypeDial, spans[0].Type)
	assert.Equal(t, remote, spans[0].PeerID)
	assert.Nil(t, spans[0].Error)
}

func TestMeasurement_detectSpans_multiDial_errorLast(t *testing.T) {
	local, err := pt.RandPeerID()
	require.NoError(t, err)
	remote, err := pt.RandPeerID()
	require.NoError(t, err)

	m := &Measurement{
		events: []Event{
			&DialStart{
				BaseEvent: NewBaseEvent(local, remote),
			},
			&DialStart{
				BaseEvent: NewBaseEvent(local, remote),
			},
			&DialEnd{
				BaseEvent: NewBaseEvent(local, remote),
			},
			&DialEnd{
				BaseEvent: NewBaseEvent(local, remote),
				Err:       fmt.Errorf("some err"),
			},
		},
	}
	spans := m.detectSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, SpanTypeDial, spans[0].Type)
	assert.Equal(t, remote, spans[0].PeerID)
	assert.Nil(t, spans[0].Error)
}
