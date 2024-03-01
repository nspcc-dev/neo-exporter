package monitor

import (
	"fmt"
	"math"
	"math/big"

	"github.com/nspcc-dev/neo-go/pkg/rpcclient/nep17"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/notary"
	"github.com/nspcc-dev/neo-go/pkg/util"
)

type (
	// NotaryFetcher allows to fetch notary balances from account.
	NotaryFetcher struct {
		cli      nep17.Invoker
		decimals *big.Float
	}
)

const (
	decimals = 8
)

// NewNotaryFetcher is a constructor for NotaryFetcher.
func NewNotaryFetcher(cli nep17.Invoker) (*NotaryFetcher, error) {
	return &NotaryFetcher{
		cli:      cli,
		decimals: big.NewFloat(math.Pow10(decimals)),
	}, nil
}

func (b *NotaryFetcher) format(balance *big.Int) (float64, error) {
	var bigFloat big.Float
	bigFloat.SetInt(balance)
	bigFloat.Quo(&bigFloat, b.decimals)

	res, _ := bigFloat.Float64()

	return res, nil
}

// FetchNotary returns the notary balance of the given account.
func (b *NotaryFetcher) FetchNotary(account util.Uint160) (float64, error) {
	balance, err := notary.NewReader(b.cli).BalanceOf(account)
	if err != nil {
		return 0, fmt.Errorf("balanceOf: %w", err)
	}

	res, err := b.format(balance)
	if err != nil {
		return 0, fmt.Errorf("format: %w", err)
	}

	return res, nil
}
