package monitor

import (
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/nep17"
	"github.com/nspcc-dev/neo-go/pkg/util"
)

type (
	// Nep17Fetcher allows to fetch balances from passed contract and account.
	Nep17Fetcher struct {
		cli nep17.Invoker
	}
)

// NewNep17BalanceFetcher is a constructor for Nep17Fetcher.
func NewNep17BalanceFetcher(cli nep17.Invoker) (*Nep17Fetcher, error) {
	return &Nep17Fetcher{
		cli: cli,
	}, nil
}

// Fetch returns the token balance of the given account.
func (b Nep17Fetcher) Fetch(tokenHash util.Uint160, account util.Uint160) (int64, error) {
	res, err := nep17.NewReader(b.cli, tokenHash).BalanceOf(account)
	if err != nil {
		return 0, err
	}

	return res.Int64(), nil
}

// FetchTotalSupply returns total token supply currently available.
func (b Nep17Fetcher) FetchTotalSupply(tokenHash util.Uint160) (int64, error) {
	res, err := nep17.NewReader(b.cli, tokenHash).TotalSupply()
	if err != nil {
		return 0, err
	}

	return res.Int64(), nil
}

// Symbol returns a short token identifier.
func (b Nep17Fetcher) Symbol(tokenHash util.Uint160) (string, error) {
	symbol, err := nep17.NewReader(b.cli, tokenHash).Symbol()
	if err != nil {
		return "", err
	}

	return symbol, nil
}
