package morphchain

import (
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
)

// Committeer must provide side chain committee
// public keys.
type Committeer interface {
	GetCommittee() (keys.PublicKeys, error)
}

type (
	AlphabetFetcher struct {
		c Committeer
	}

	AlphabetFetcherArgs struct {
		Committeer Committeer
	}
)

func NewAlphabetFetcher(p AlphabetFetcherArgs) (*AlphabetFetcher, error) {
	return &AlphabetFetcher{c: p.Committeer}, nil
}

func (a AlphabetFetcher) FetchAlphabet() (keys.PublicKeys, error) {
	return a.c.GetCommittee()
}
