package morphchain

import (
	"github.com/nspcc-dev/neo-go/pkg/crypto/hash"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/gas"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/nep17"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/notary"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/pool"
)

type (
	BalanceFetcher struct {
		cli *pool.Pool
	}

	BalanceFetcherArgs struct {
		Cli *pool.Pool
	}
)

func NewBalanceFetcher(p BalanceFetcherArgs) (*BalanceFetcher, error) {
	return &BalanceFetcher{
		cli: p.Cli,
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
	res, err := gas.NewReader(b.cli).BalanceOf(sh)
	if err != nil {
		return 0, err
	}

	return res.Int64(), nil
}

func (b BalanceFetcher) FetchNotaryByScriptHash(sh util.Uint160) (int64, error) {
	res, err := notary.NewReader(b.cli).BalanceOf(sh)
	if err != nil {
		return 0, err
	}

	return res.Int64(), nil
}

func (b BalanceFetcher) FetchNEP17TotalSupply(tokenHash util.Uint160) (int64, error) {
	res, err := nep17.NewReader(b.cli, tokenHash).TotalSupply()
	if err != nil {
		return 0, err
	}

	return res.Int64(), nil
}
