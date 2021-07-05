package monitor

import (
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/morphchain"
)

func TestGetDiff(t *testing.T) {
	tests := []struct {
		nm                   morphchain.NetmapInfo
		cand                 morphchain.NetmapCandidatesInfo
		wantNew, wantDropped int
	}{
		// empty Network Map
		{
			nm: morphchain.NetmapInfo{
				Nodes: generateNodes(0, 0),
			},
			cand: morphchain.NetmapCandidatesInfo{
				Nodes: generateNodes(0, 0),
			},
			wantNew:     0,
			wantDropped: 0,
		},
		// disjoint Network Maps
		{
			nm: morphchain.NetmapInfo{
				Nodes: generateNodes(0, 5),
			},
			cand: morphchain.NetmapCandidatesInfo{
				Nodes: generateNodes(5, 10),
			},
			wantNew:     5,
			wantDropped: 5,
		},
		// intersecting Network Maps
		{
			nm: morphchain.NetmapInfo{
				Nodes: generateNodes(0, 5),
			},
			cand: morphchain.NetmapCandidatesInfo{
				Nodes: generateNodes(3, 6),
			},
			wantNew:     1,
			wantDropped: 3,
		},
	}

	var (
		gotNew, gotDropped []*morphchain.Node
	)

	for _, test := range tests {
		gotNew, gotDropped = getDiff(test.nm, test.cand)

		require.Equal(t, test.wantNew, len(gotNew))
		require.Equal(t, test.wantDropped, len(gotDropped))
	}
}

func generateNodes(start, finish int) []*morphchain.Node {
	nodes := make([]*morphchain.Node, 0, finish-start)

	for i := start; i < finish; i++ {
		privKey, _ := keys.NewPrivateKey()

		nodes = append(
			nodes,
			&morphchain.Node{
				ID:        uint64(i),
				PublicKey: privKey.PublicKey(),
				Address:   strconv.Itoa(i),
			},
		)
	}

	return nodes
}
