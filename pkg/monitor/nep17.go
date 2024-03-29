package monitor

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neo-exporter/pkg/model"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/gas"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/neo"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"go.uber.org/zap"
)

type (
	// NNSResolver helps to resolve NNS contract name to address.
	NNSResolver interface {
		ResolveFSContract(name string) (util.Uint160, error)
	}

	// Item describes task for [Nep17tracker].
	Item struct {
		Symbol   string
		Hash     util.Uint160
		Accounts []util.Uint160
		Total    bool
	}
)

var (
	errInvalidAddress = errors.New("invalid address")
)

// ParseNep17Tasks prepares tasks for [Nep17tracker].
func ParseNep17Tasks(balanceFetcher Nep17BalanceFetcher, items []model.Nep17Balance, nns NNSResolver) ([]Item, error) {
	var (
		result []Item
	)

	for _, it := range items {
		var (
			contract *util.Uint160
			err      error
		)

		task := Item{
			Total:    it.TotalSupply,
			Accounts: make([]util.Uint160, 0, len(it.BalanceOf)),
		}

		contract = nativeNep17ContractHash(it.Contract, nns)
		if contract == nil {
			contract, err = parseUint160(it.Contract)
			if err != nil {
				zap.L().Error("parse nep17 contract", zap.Error(err), zap.String("contract", it.Contract))
				return nil, fmt.Errorf("nep17 contract hash %s in invalid: %w", it.Contract, err)
			}
		}

		symbol, err := balanceFetcher.Symbol(*contract)
		if err != nil {
			return nil, fmt.Errorf("nep17 contract %s symbol: %w", it.Contract, err)
		}

		task.Symbol = symbol
		task.Hash = *contract

		for _, balanceOf := range it.BalanceOf {
			acc, err := parseUint160(balanceOf)
			if err != nil {
				zap.L().Error(
					"parse nep17 account",
					zap.Error(err),
					zap.String("contract", contract.StringLE()),
					zap.String("balanceOf", balanceOf),
				)
				continue
			}

			task.Accounts = append(task.Accounts, *acc)
		}

		result = append(result, task)
	}

	return result, nil
}

func nativeNep17ContractHash(name string, nns NNSResolver) *util.Uint160 {
	switch name {
	case "NEO", "neo":
		return &neo.Hash
	case "GAS", "gas":
		return &gas.Hash
	default:
		addr, err := nns.ResolveFSContract(name)
		if err == nil {
			return &addr
		}
	}

	return nil
}

func parseUint160(value string) (*util.Uint160, error) {
	addr, err := util.Uint160DecodeStringLE(value)
	if err == nil {
		return &addr, nil
	}

	addr, err = address.StringToUint160(value)
	if err == nil {
		return &addr, nil
	}

	return nil, fmt.Errorf("%w: %s", errInvalidAddress, value)
}
