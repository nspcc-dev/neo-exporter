package monitor

import (
	"fmt"
	"math"
	"math/big"

	"github.com/nspcc-dev/neo-go/pkg/rpcclient/nep17"
	"github.com/nspcc-dev/neo-go/pkg/util"
)

type (
	// Nep17Fetcher allows to fetch balances from passed contract and account.
	Nep17Fetcher struct {
		cli   nep17.Invoker
		cache map[util.Uint160]*big.Float
	}
)

// NewNep17BalanceFetcher is a constructor for Nep17Fetcher.
func NewNep17BalanceFetcher(cli nep17.Invoker) (*Nep17Fetcher, error) {
	return &Nep17Fetcher{
		cli:   cli,
		cache: make(map[util.Uint160]*big.Float),
	}, nil
}

func (b *Nep17Fetcher) decimals(tokenHash util.Uint160) (*big.Float, error) {
	res, ok := b.cache[tokenHash]
	if ok {
		return res, nil
	}

	dec, err := nep17.NewReader(b.cli, tokenHash).Decimals()
	if err != nil {
		return nil, err
	}

	res = big.NewFloat(math.Pow10(dec))
	b.cache[tokenHash] = res

	return res, nil
}

func (b *Nep17Fetcher) format(tokenHash util.Uint160, balance *big.Int) (float64, error) {
	multiplier, err := b.decimals(tokenHash)
	if err != nil {
		return 0, err
	}

	var bigFloat big.Float
	bigFloat.SetInt(balance)
	bigFloat.Quo(&bigFloat, multiplier)

	res, _ := bigFloat.Float64()

	return res, nil
}

// Fetch returns the token balance of the given account.
func (b *Nep17Fetcher) Fetch(tokenHash util.Uint160, account util.Uint160) (float64, error) {
	balance, err := nep17.NewReader(b.cli, tokenHash).BalanceOf(account)
	if err != nil {
		return 0, fmt.Errorf("balanceOf: %w", err)
	}

	res, err := b.format(tokenHash, balance)
	if err != nil {
		return 0, fmt.Errorf("format: %w", err)
	}

	return res, nil
}

// FetchTotalSupply returns total token supply currently available.
func (b *Nep17Fetcher) FetchTotalSupply(tokenHash util.Uint160) (float64, error) {
	balance, err := nep17.NewReader(b.cli, tokenHash).TotalSupply()
	if err != nil {
		return 0, err
	}

	res, err := b.format(tokenHash, balance)
	if err != nil {
		return 0, fmt.Errorf("format: %w", err)
	}

	return res, nil
}

// Symbol returns a short token identifier.
func (b *Nep17Fetcher) Symbol(tokenHash util.Uint160) (string, error) {
	symbol, err := nep17.NewReader(b.cli, tokenHash).Symbol()
	if err != nil {
		return "", err
	}

	return symbol, nil
}
