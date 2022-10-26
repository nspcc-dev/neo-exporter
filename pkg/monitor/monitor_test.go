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

func TestComputeUniqueAlphabetKeys(t *testing.T) {
	tests := []struct {
		name               string
		sortedMainAlphabet []string
		sortedSideAlphabet []string
		expectedMainUnique []string
		expectedSideUnique []string
	}{
		{
			name: "empty",
		},
		{
			name:               "no unique",
			sortedMainAlphabet: []string{"a", "b"},
			sortedSideAlphabet: []string{"a", "b"},
		},
		{
			name:               "tail unique",
			sortedMainAlphabet: []string{"a", "b"},
			sortedSideAlphabet: []string{"a", "b", "c"},
			expectedSideUnique: []string{"c"},
		},
		{
			name:               "middle both unique",
			sortedMainAlphabet: []string{"a", "b", "d"},
			sortedSideAlphabet: []string{"a", "c", "d"},
			expectedMainUnique: []string{"b"},
			expectedSideUnique: []string{"c"},
		},
		{
			name:               "middle side unique",
			sortedMainAlphabet: []string{"a", "b", "d"},
			sortedSideAlphabet: []string{"a", "b", "c", "d"},
			expectedSideUnique: []string{"c"},
		},
		{
			name:               "middle main unique",
			sortedMainAlphabet: []string{"a", "b", "c", "d"},
			sortedSideAlphabet: []string{"a", "c", "d"},
			expectedMainUnique: []string{"b"},
		},
		{
			name:               "same length tail unique",
			sortedMainAlphabet: []string{"a", "b", "c"},
			sortedSideAlphabet: []string{"a", "b", "d"},
			expectedMainUnique: []string{"c"},
			expectedSideUnique: []string{"d"},
		},
		{
			name:               "same length all unique",
			sortedMainAlphabet: []string{"a", "c"},
			sortedSideAlphabet: []string{"b", "d"},
			expectedMainUnique: []string{"a", "c"},
			expectedSideUnique: []string{"b", "d"},
		},
		{
			name:               "all unique",
			sortedMainAlphabet: []string{"e", "f"},
			sortedSideAlphabet: []string{"a", "b", "c", "d"},
			expectedMainUnique: []string{"e", "f"},
			expectedSideUnique: []string{"a", "b", "c", "d"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			uniqueMain, uniqueSide := computeUniqueAlphabets(test.sortedMainAlphabet, test.sortedSideAlphabet)
			require.Equal(t, test.expectedMainUnique, uniqueMain)
			require.Equal(t, test.expectedSideUnique, uniqueSide)
		})
	}
}
