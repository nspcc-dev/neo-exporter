package morphchain

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/core/native/nativenames"
	"github.com/nspcc-dev/neo-go/pkg/crypto/hash"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/pool"
)

type (
	BalanceFetcher struct {
		cli    *pool.Pool
		gas    util.Uint160
		notary util.Uint160
	}

	BalanceFetcherArgs struct {
		Cli *pool.Pool
	}
)

func NewBalanceFetcher(p BalanceFetcherArgs) (*BalanceFetcher, error) {
	gas, err := p.Cli.GetNativeContractHash(nativenames.Gas)
	if err != nil {
		return nil, fmt.Errorf("can't get native GAS contract address: %w", err)
	}

	notary, _ := p.Cli.GetNativeContractHash(nativenames.Notary)

	return &BalanceFetcher{
		cli:    p.Cli,
		gas:    gas,
		notary: notary,
	}, nil
}

func (b BalanceFetcher) FetchGAS(key keys.PublicKey) (int64, error) {
	scriptHash := hash.Hash160(key.GetVerificationScript())

	return b.FetchGASByScriptHash(scriptHash)
}

func (b BalanceFetcher) FetchNotary(key keys.PublicKey) (int64, error) {
	scriptHash := hash.Hash160(key.GetVerificationScript())

	return b.FetchNotaryByScriptHash(scriptHash)
}

func (b BalanceFetcher) FetchGASByScriptHash(sh util.Uint160) (int64, error) {
	return b.cli.NEP17BalanceOf(b.gas, sh)
}

func (b BalanceFetcher) FetchNotaryByScriptHash(sh util.Uint160) (int64, error) {
	return b.cli.NEP17BalanceOf(b.notary, sh)
}

func (b BalanceFetcher) FetchNEP17TotalSupply(tokenHash util.Uint160) (int64, error) {
	return b.cli.NEP17TotalSupply(tokenHash)
}
