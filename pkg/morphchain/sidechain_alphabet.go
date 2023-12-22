package morphchain

import (
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
)

// Committeer must provide side chain committee public keys.
type Committeer interface {
	GetCommittee() (keys.PublicKeys, error)
}

type (
	SideChainAlphabetFetcher struct {
		committeer Committeer
	}
)

func NewSideChainAlphabetFetcher(committeer Committeer) *SideChainAlphabetFetcher {
	return &SideChainAlphabetFetcher{
		committeer: committeer,
	}
}

func (a SideChainAlphabetFetcher) FetchAlphabet() (keys.PublicKeys, error) {
	return a.committeer.GetCommittee()
}
