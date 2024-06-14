package pool

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-exporter/pkg/monitor"
	"github.com/nspcc-dev/neo-go/pkg/core/native/noderoles"
	"github.com/nspcc-dev/neo-go/pkg/core/state"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/neorpc/result"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/invoker"
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
	clients []*rpcclient.Client
	opts    rpcclient.Options

	lastHealthyTimestamp int64
	recheckInterval      time.Duration

	current, next int
	endpoints     []string
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
		clients:         make([]*rpcclient.Client, len(prm.Endpoints)),
	}

	if err := pool.dial(ctx); err != nil {
		return nil, err
	}

	go func() {
		tick := time.NewTicker(recheck)

		for {
			select {
			case <-tick.C:
				pool.recheck(ctx)
			case <-ctx.Done():
				tick.Stop()
				return
			}
		}
	}()

	return pool, nil
}

func (p *Pool) dial(ctx context.Context) error {
	opts := rpcclient.Options{DialTimeout: p.opts.DialTimeout}

	for i, ep := range p.endpoints {
		neoClient, err := neoGoClient(ctx, ep, opts)
		if err != nil {
			return err
		}

		p.clients[i] = neoClient
	}

	return nil
}

func (p *Pool) recheck(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < len(p.endpoints); i++ {
		var (
			cl  = p.clients[i]
			err error
		)

		if cl != nil {
			_, err = cl.GetBlockCount()
		}
		if cl == nil || err != nil {
			p.clients[i], err = neoGoClient(ctx, p.endpoints[i], p.opts)
			if err != nil {
				log.Printf("reconnect to Neo node %s failed: %v", cl.Endpoint(), err)
			}

			continue
		}
	}
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
	return p.clients[p.current]
}

func (p *Pool) invokerConn() *invoker.Invoker {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return invoker.New(p.clients[p.current], nil)
}

func (p *Pool) establishNewConnection() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var err error

	for i := p.next; i < p.next+len(p.endpoints); i++ {
		index := i % len(p.endpoints)
		if p.clients[p.current], err = neoGoClient(p.ctx, p.endpoints[index], p.opts); err == nil {
			p.current = index % len(p.endpoints)
			p.next = (index + 1) % len(p.endpoints)
			return nil
		}
	}

	return fmt.Errorf("no healthy client")
}

func neoGoClient(ctx context.Context, endpoint string, opts rpcclient.Options) (*rpcclient.Client, error) {
	cli, err := rpcclient.New(ctx, endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("create Neo RPC client: %w", err)
	}

	err = cli.Init()
	if err != nil {
		return nil, fmt.Errorf("init Neo RPC client: %w", err)
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

func (p *Pool) FetchHeight() []monitor.HeightData {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var (
		heights    []monitor.HeightData
		wg         sync.WaitGroup
		heightChan = make(chan monitor.HeightData, len(p.clients))
	)

	for _, cl := range p.clients {
		wg.Add(1)

		go func(cl *rpcclient.Client) {
			defer wg.Done()

			stHeight, err := cl.GetStateHeight()
			if err != nil {
				log.Printf("read state height of Neo node %s: %v", cl.Endpoint(), err)
				return
			}

			heightChan <- monitor.HeightData{
				Host:  cl.Endpoint(),
				Value: stHeight.Local,
			}
		}(cl)
	}

	wg.Wait()
	close(heightChan)

	for height := range heightChan {
		heights = append(heights, height)
	}

	return heights
}

func (p *Pool) FetchState(height uint32) []monitor.StateData {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var (
		states    []monitor.StateData
		wg        sync.WaitGroup
		stateChan = make(chan monitor.StateData, len(p.clients))
	)

	for _, cl := range p.clients {
		wg.Add(1)

		go func(cl *rpcclient.Client) {
			defer wg.Done()

			stHeight, err := cl.GetStateRootByHeight(height)
			if err != nil {
				log.Printf("read state root at height #%d from Neo node %s: %v", height, cl.Endpoint(), err)
				return
			}

			stateChan <- monitor.StateData{
				Host:  cl.Endpoint(),
				Value: stHeight.Hash().String(),
			}
		}(cl)
	}

	wg.Wait()
	close(stateChan)

	for st := range stateChan {
		states = append(states, st)
	}

	return states
}
