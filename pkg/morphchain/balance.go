package morphchain

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/core/native/nativenames"
	"github.com/nspcc-dev/neo-go/pkg/crypto/hash"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/util"
)

type (
	BalanceFetcher struct {
		cli *client.Client
		gas util.Uint160
	}

	BalanceFetcherArgs struct {
		Cli *client.Client
	}
)

func NewBalanceFetcher(p BalanceFetcherArgs) (*BalanceFetcher, error) {
	gas, err := p.Cli.GetNativeContractHash(nativenames.Gas)
	if err != nil {
		return nil, fmt.Errorf("can't get native GAS contract address: %w", err)
	}

	return &BalanceFetcher{
		cli: p.Cli,
		gas: gas,
	}, nil
}

func (b BalanceFetcher) FetchGAS(key keys.PublicKey) (int64, error) {
	scriptHash := hash.Hash160(key.GetVerificationScript())

	return b.FetchGASByScriptHash(scriptHash)
}

func (b BalanceFetcher) FetchGASByScriptHash(sh util.Uint160) (int64, error) {
	return b.cli.NEP17BalanceOf(b.gas, sh)
}
