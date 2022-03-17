package morphchain

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/core/native/noderoles"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
)

// Committeer must provide side chain committee
// public keys.
type Committeer interface {
	GetCommittee() (keys.PublicKeys, error)
}

// Designater must provide main chain alphabet
// public keys.
type Designater interface {
	GetBlockCount() (uint32, error)
	GetDesignatedByRole(noderoles.Role, uint32) (keys.PublicKeys, error)
}

type (
	AlphabetFetcher struct {
		c Committeer
		d Designater
	}

	AlphabetFetcherArgs struct {
		Committeer Committeer
		Designater Designater
	}
)

func NewAlphabetFetcher(p AlphabetFetcherArgs) (*AlphabetFetcher, error) {
	return &AlphabetFetcher{
		c: p.Committeer,
		d: p.Designater,
	}, nil
}

func (a AlphabetFetcher) FetchSideAlphabet() (keys.PublicKeys, error) {
	return a.c.GetCommittee()
}

func (a AlphabetFetcher) FetchMainAlphabet() (keys.PublicKeys, error) {
	height, err := a.d.GetBlockCount()
	if err != nil {
		return nil, fmt.Errorf("can't get block height: %w", err)
	}
	return a.d.GetDesignatedByRole(noderoles.NeoFSAlphabet, height)
}
