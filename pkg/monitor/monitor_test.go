package monitor

import (
	"strconv"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/stretchr/testify/require"
)

func TestGetDiff(t *testing.T) {
	tests := []struct {
		nm                   NetmapInfo
		cand                 NetmapCandidatesInfo
		wantNew, wantDropped int
	}{
		// empty Network Map
		{
			nm: NetmapInfo{
				Nodes: generateNodes(0, 0),
			},
			cand: NetmapCandidatesInfo{
				Nodes: generateNodes(0, 0),
			},
			wantNew:     0,
			wantDropped: 0,
		},
		// disjoint Network Maps
		{
			nm: NetmapInfo{
				Nodes: generateNodes(0, 5),
			},
			cand: NetmapCandidatesInfo{
				Nodes: generateNodes(5, 10),
			},
			wantNew:     5,
			wantDropped: 5,
		},
		// intersecting Network Maps
		{
			nm: NetmapInfo{
				Nodes: generateNodes(0, 5),
			},
			cand: NetmapCandidatesInfo{
				Nodes: generateNodes(3, 6),
			},
			wantNew:     1,
			wantDropped: 3,
		},
	}

	var (
		gotNew, gotDropped []*Node
	)

	for _, test := range tests {
		gotNew, gotDropped = getDiff(test.nm, test.cand)

		require.Equal(t, test.wantNew, len(gotNew))
		require.Equal(t, test.wantDropped, len(gotDropped))
	}
}

func generateNodes(start, finish int) []*Node {
	nodes := make([]*Node, 0, finish-start)

	for i := start; i < finish; i++ {
		privKey, _ := keys.NewPrivateKey()

		nodes = append(
			nodes,
			&Node{
				ID:        uint64(i),
				PublicKey: privKey.PublicKey(),
				Address:   strconv.Itoa(i),
			},
		)
	}

	return nodes
}
