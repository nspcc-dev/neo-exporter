package morphchain

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/core/native/nativenames"
	"github.com/nspcc-dev/neo-go/pkg/crypto/hash"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
)

type (
	BalanceFetcher struct {
		cli *client.Client
		gas util.Uint160
	}

	BalanceFetcherArgs struct {
		Endpoint    string
		DialTimeout time.Duration
	}
)

const (
	successState  = "HALT"
	balanceMethod = "balanceOf"
)

func NewBalanceFetcher(ctx context.Context, p BalanceFetcherArgs) (*BalanceFetcher, error) {
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

	gas, err := cli.GetNativeContractHash(nativenames.Gas)
	if err != nil {
		return nil, fmt.Errorf("can't get native GAS contract address: %w", err)
	}

	return &BalanceFetcher{
		cli: cli,
		gas: gas,
	}, nil
}

func (b BalanceFetcher) FetchGAS(key keys.PublicKey) (int64, error) {
	scriptHash := hash.Hash160(key.GetVerificationScript())

	return b.FetchGASByScriptHash(scriptHash)
}

func (b BalanceFetcher) FetchGASByScriptHash(sh util.Uint160) (int64, error) {
	output, err := b.cli.InvokeFunction(b.gas, balanceMethod,
		[]smartcontract.Parameter{
			{
				Type:  smartcontract.ByteArrayType,
				Value: sh.BytesBE(),
			},
		}, nil)
	if err != nil {
		return 0, fmt.Errorf("can't get balance of %s: %w", hex.EncodeToString(sh.BytesLE()), err)
	}

	if output.State != successState {
		return 0, fmt.Errorf("can't get balance of %s, state <%s>, error <%s>",
			hex.EncodeToString(sh.BytesLE()),
			output.State,
			output.FaultException)
	}

	stack := output.Stack
	ln := len(stack)

	if ln != 1 {
		return 0, fmt.Errorf("can't parse gas balance, got %d items on stack", ln)
	}

	bigInt, err := stack[0].TryInteger()
	if err != nil {
		return 0, fmt.Errorf("can't parse gas balance: %w", err)
	}

	return bigInt.Int64(), nil
}
