package pool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/core/native/nativenames"
	"github.com/nspcc-dev/neo-go/pkg/core/native/noderoles"
	"github.com/nspcc-dev/neo-go/pkg/core/state"
	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/rpc/response/result"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
)

// Pool represent virtual connection to the Neo network to communicate
// with multiple Neo servers.
type Pool struct {
	ctx  context.Context
	mu   sync.RWMutex
	rpc  *client.Client
	opts client.Options

	lastHealthyTimestamp int64
	recheckInterval      time.Duration

	next      int
	endpoints []string
}

// PrmPool groups parameter to create Pool.
type PrmPool struct {
	Endpoints       []string
	DialTimeout     time.Duration
	RecheckInterval time.Duration
}

// defaultRecheckInterval stores the interval after which a connection health check is performed.
const defaultRecheckInterval = 5 * time.Second

// NewPool creates connection pool using parameters.
func NewPool(ctx context.Context, prm PrmPool) (*Pool, error) {
	recheck := prm.RecheckInterval
	if recheck <= 0 {
		recheck = defaultRecheckInterval
	}

	pool := &Pool{
		ctx:             ctx,
		endpoints:       prm.Endpoints,
		recheckInterval: recheck,
		opts:            client.Options{DialTimeout: prm.DialTimeout},
	}

	return pool, pool.establishNewConnection()
}

func (p *Pool) isCurrentHealthy() bool {
	if (time.Now().UTC().UnixNano() - atomic.LoadInt64(&p.lastHealthyTimestamp)) < p.recheckInterval.Nanoseconds() {
		return true
	}

	conn := p.conn()
	if conn == nil {
		return false
	}

	if _, err := conn.GetBlockCount(); err == nil {
		atomic.StoreInt64(&p.lastHealthyTimestamp, time.Now().UTC().UnixNano())
		return true
	}

	return false
}

// nextConnection returns healthy connection,
// the second resp value is true if current connection was updated.
// Returns error if there are not healthy connections.
func (p *Pool) nextConnection() (*client.Client, bool, error) {
	if p.isCurrentHealthy() {
		return p.conn(), false, nil
	}

	if err := p.establishNewConnection(); err != nil {
		return nil, false, err
	}

	return p.conn(), true, nil
}

// GetContractStateByID queries contract information, according to the contract ID.
func (p *Pool) GetContractStateByID(id int32) (*state.Contract, error) {
	conn, _, err := p.nextConnection()
	if err != nil {
		return nil, err
	}

	return conn.GetContractStateByID(id)
}

// GetNativeContractHash returns native contract hash by its name.
func (p *Pool) GetNativeContractHash(name string) (util.Uint160, error) {
	conn, _, err := p.nextConnection()
	if err != nil {
		return util.Uint160{}, err
	}

	return conn.GetNativeContractHash(name)
}

// NEP17BalanceOf invokes `balanceOf` NEP17 method on a specified contract.
func (p *Pool) NEP17BalanceOf(tokenHash, acc util.Uint160) (int64, error) {
	conn, _, err := p.nextConnection()
	if err != nil {
		return 0, err
	}

	return conn.NEP17BalanceOf(tokenHash, acc)
}

// NEP17TotalSupply invokes `totalSupply` NEP17 method on a specified contract.
func (p *Pool) NEP17TotalSupply(tokenHash util.Uint160) (int64, error) {
	conn, _, err := p.nextConnection()
	if err != nil {
		return 0, err
	}

	return conn.NEP17TotalSupply(tokenHash)
}

// InvokeFunction returns the results after calling the smart contract scripthash
// with the given operation and parameters.
// NOTE: this is test invoke and will not affect the blockchain.
func (p *Pool) InvokeFunction(contract util.Uint160, operation string, params []smartcontract.Parameter, signers []transaction.Signer) (*result.Invoke, error) {
	conn, _, err := p.nextConnection()
	if err != nil {
		return nil, err
	}

	return conn.InvokeFunction(contract, operation, params, signers)
}

// GetBlockCount returns the number of blocks in the main chain.
func (p *Pool) GetBlockCount() (uint32, error) {
	conn, _, err := p.nextConnection()
	if err != nil {
		return 0, err
	}

	return conn.GetBlockCount()
}

// GetDesignatedByRole invokes `getDesignatedByRole` method on a native RoleManagement contract.
func (p *Pool) GetDesignatedByRole(role noderoles.Role, height uint32) (keys.PublicKeys, error) {
	conn, _, err := p.nextConnection()
	if err != nil {
		return nil, err
	}

	return conn.GetDesignatedByRole(role, height)
}

// GetCommittee returns the current public keys of NEO nodes in committee.
func (p *Pool) GetCommittee() (keys.PublicKeys, error) {
	conn, _, err := p.nextConnection()
	if err != nil {
		return nil, err
	}

	return conn.GetCommittee()
}

// ProbeNotary checks if native `Notary` contract is presented on chain.
func (p *Pool) ProbeNotary() bool {
	conn, _, err := p.nextConnection()
	if err != nil {
		return false
	}

	_, err = conn.GetNativeContractHash(nativenames.Notary)
	return err == nil
}

func (p *Pool) conn() *client.Client {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.rpc
}

func (p *Pool) establishNewConnection() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var err error

	for i := p.next; i < p.next+len(p.endpoints); i++ {
		index := i % len(p.endpoints)
		if p.rpc, err = neoGoClient(p.ctx, p.endpoints[index], p.opts); err == nil {
			p.next = (index + 1) % len(p.endpoints)
			return nil
		}
	}

	return fmt.Errorf("no healthy client")
}

func neoGoClient(ctx context.Context, endpoint string, opts client.Options) (*client.Client, error) {
	cli, err := client.New(ctx, endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("can't create neo-go client: %w", err)
	}

	err = cli.Init()
	if err != nil {
		return nil, fmt.Errorf("can't init neo-go client: %w", err)
	}

	return cli, nil
}
