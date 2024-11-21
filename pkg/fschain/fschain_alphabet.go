package fschain

import (
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
)

// Committeer provides FS chain committee public keys.
type Committeer interface {
	GetCommittee() (keys.PublicKeys, error)
}

type (
	FSChainAlphabetFetcher struct {
		committeer Committeer
	}
)

func NewFSChainAlphabetFetcher(committeer Committeer) *FSChainAlphabetFetcher {
	return &FSChainAlphabetFetcher{
		committeer: committeer,
	}
}

func (a FSChainAlphabetFetcher) FetchAlphabet() (keys.PublicKeys, error) {
	return a.committeer.GetCommittee()
}
