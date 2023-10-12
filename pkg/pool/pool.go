package pool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/core/native/noderoles"
	"github.com/nspcc-dev/neo-go/pkg/core/state"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/neorpc/result"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/invoker"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/nep17"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/rolemgmt"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	rpcnns "github.com/nspcc-dev/neofs-contract/rpc/nns"
)

// Pool represent virtual connection to the Neo network to communicate
// with multiple Neo servers.
type Pool struct {
	ctx     context.Context
	mu      sync.RWMutex
	rpc     *rpcclient.Client
	invoker *invoker.Invoker
	opts    rpcclient.Options

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
		opts:            rpcclient.Options{DialTimeout: prm.DialTimeout},
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
// Returns error if there are no healthy connections.
func (p *Pool) nextConnection() (*rpcclient.Client, bool, error) {
	if p.isCurrentHealthy() {
		return p.conn(), false, nil
	}

	if err := p.establishNewConnection(); err != nil {
		return nil, false, err
	}

	return p.conn(), true, nil
}

// nextInvoker returns invoker wrapper on healthy connection.
// Returns error if there are no healthy connections.
func (p *Pool) nextInvoker() (*invoker.Invoker, error) {
	if p.isCurrentHealthy() {
		return p.invokerConn(), nil
	}

	if err := p.establishNewConnection(); err != nil {
		return nil, err
	}

	return p.invokerConn(), nil
}

// GetContractStateByID queries contract information, according to the contract ID.
func (p *Pool) GetContractStateByID(id int32) (*state.Contract, error) {
	conn, _, err := p.nextConnection()
	if err != nil {
		return nil, err
	}

	return conn.GetContractStateByID(id)
}

// NEP17BalanceOf invokes `balanceOf` NEP17 method on a specified contract.
func (p *Pool) NEP17BalanceOf(tokenHash, acc util.Uint160) (int64, error) {
	invokerConn, err := p.NextInvoker()
	if err != nil {
		return 0, err
	}

	res, err := nep17.NewReader(invokerConn, tokenHash).BalanceOf(acc)
	if err != nil {
		return 0, err
	}

	return res.Int64(), nil
}

// NEP17TotalSupply invokes `totalSupply` NEP17 method on a specified contract.
func (p *Pool) NEP17TotalSupply(tokenHash util.Uint160) (int64, error) {
	invokerConn, err := p.NextInvoker()
	if err != nil {
		return 0, err
	}

	res, err := nep17.NewReader(invokerConn, tokenHash).TotalSupply()
	if err != nil {
		return 0, err
	}

	return res.Int64(), nil
}

// Call returns the results after calling the smart contract scripthash
// with the given operation and parameters.
// NOTE: this is test invoke and will not affect the blockchain.
func (p *Pool) Call(contract util.Uint160, operation string, params ...any) (*result.Invoke, error) {
	conn, err := p.nextInvoker()
	if err != nil {
		return nil, err
	}

	return conn.Call(contract, operation, params...)
}

// CallAndExpandIterator creates a script containing a call of the specified method
// of a contract with given parameters (similar to how Call operates). But then this
// script contains additional code that expects that the result of the first call is
// an iterator.
func (p *Pool) CallAndExpandIterator(contract util.Uint160, method string, maxItems int, params ...any) (*result.Invoke, error) {
	conn, err := p.nextInvoker()
	if err != nil {
		return nil, err
	}

	return conn.CallAndExpandIterator(contract, method, maxItems, params...)
}

// TerminateSession closes the given session, returning an error if anything
// goes wrong. It's not strictly required to close the session (it'll expire on
// the server anyway), but it helps to release server resources earlier.
func (p *Pool) TerminateSession(_ uuid.UUID) error {
	return errors.New("unsupported")
}

// TraverseIterator allows to retrieve the next batch of items from the given
// iterator in the given session (previously returned from Call).
func (p *Pool) TraverseIterator(_ uuid.UUID, _ *result.Iterator, _ int) ([]stackitem.Item, error) {
	return nil, errors.New("unsupported")
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
	invokerConn, err := p.nextInvoker()
	if err != nil {
		return nil, err
	}

	return rolemgmt.NewReader(invokerConn).GetDesignatedByRole(role, height)
}

// GetCommittee returns the current public keys of NEO nodes in committee.
func (p *Pool) GetCommittee() (keys.PublicKeys, error) {
	conn, _, err := p.nextConnection()
	if err != nil {
		return nil, err
	}

	return conn.GetCommittee()
}

func (p *Pool) conn() *rpcclient.Client {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.rpc
}

func (p *Pool) invokerConn() *invoker.Invoker {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.invoker
}

func (p *Pool) establishNewConnection() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var err error

	for i := p.next; i < p.next+len(p.endpoints); i++ {
		index := i % len(p.endpoints)
		if p.rpc, err = neoGoClient(p.ctx, p.endpoints[index], p.opts); err == nil {
			p.invoker = invoker.New(p.rpc, nil)
			p.next = (index + 1) % len(p.endpoints)
			return nil
		}
	}

	return fmt.Errorf("no healthy client")
}

func neoGoClient(ctx context.Context, endpoint string, opts rpcclient.Options) (*rpcclient.Client, error) {
	cli, err := rpcclient.New(ctx, endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("can't create neo-go client: %w", err)
	}

	err = cli.Init()
	if err != nil {
		return nil, fmt.Errorf("can't init neo-go client: %w", err)
	}

	return cli, nil
}

// ResolveContract helps to take contract address by contract name. Name list can be taken from contract wrappers,
// for instance [rpcnns.NameNetmap].
func (p *Pool) ResolveContract(contractName string) (util.Uint160, error) {
	nnsHash, err := rpcnns.InferHash(p)
	if err != nil {
		return util.Uint160{}, fmt.Errorf("GetContractStateByID: %w", err)
	}

	nnsReader := rpcnns.NewReader(p, nnsHash)
	addr, err := nnsReader.ResolveFSContract(contractName)
	if err != nil {
		return util.Uint160{}, fmt.Errorf("ResolveFSContract [%s]: %w", contractName, err)
	}

	return addr, nil
}
