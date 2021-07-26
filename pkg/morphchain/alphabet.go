package morphchain

import (
	"context"
	"fmt"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
)

type (
	AlphabetFetcher struct {
		cli *client.Client
	}

	AlphabetFetcherArgs struct {
		Endpoint    string
		DialTimeout time.Duration
	}
)

func NewAlphabetFetcher(ctx context.Context, p AlphabetFetcherArgs) (*AlphabetFetcher, error) {
	cli, err := client.New(ctx, p.Endpoint, client.Options{
		DialTimeout: p.DialTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("can't create neo-go client: %w", err)
	}

	err = cli.Init()
	if err != nil {
		return nil, fmt.Errorf("can't init neo-go client: %w", err)
	}

	return &AlphabetFetcher{cli: cli}, nil
}

func (a AlphabetFetcher) FetchAlphabet() (keys.PublicKeys, error) {
	return a.cli.GetCommittee()
}
