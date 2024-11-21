package fschain

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/core/native/noderoles"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
)

// Designater must provide main chain alphabet public keys.
type Designater interface {
	GetBlockCount() (uint32, error)
	GetDesignatedByRole(noderoles.Role, uint32) (keys.PublicKeys, error)
}

type (
	MainChainAlphabetFetcher struct {
		designater Designater
	}
)

func NewMainChainAlphabetFetcher(designater Designater) *MainChainAlphabetFetcher {
	return &MainChainAlphabetFetcher{
		designater: designater,
	}
}

func (a MainChainAlphabetFetcher) FetchAlphabet() (keys.PublicKeys, error) {
	height, err := a.designater.GetBlockCount()
	if err != nil {
		return nil, fmt.Errorf("can't get chain height: %w", err)
	}

	return a.designater.GetDesignatedByRole(noderoles.NeoFSAlphabet, height)
}
