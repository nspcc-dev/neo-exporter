package multinodepool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/rpcclient"
	"github.com/nspcc-dev/neofs-net-monitor/pkg/monitor"
)

// Pool collects data from each node.
type Pool struct {
	endpoints   []string
	dialTimeout time.Duration
	clients     []*rpcclient.Client
}

func NewPool(endpoints []string, dialTimeout time.Duration) *Pool {
	return &Pool{
		endpoints:   endpoints,
		dialTimeout: dialTimeout,
		clients:     make([]*rpcclient.Client, len(endpoints)),
	}
}

func (c *Pool) Dial(ctx context.Context) error {
	opts := rpcclient.Options{DialTimeout: c.dialTimeout}

	for i, ep := range c.endpoints {
		neoClient, err := neoGoClient(ctx, ep, opts)
		if err != nil {
			return fmt.Errorf("neoGoClient: %w", err)
		}

		c.clients[i] = neoClient
	}

	return nil
}

func (c *Pool) FetchHeight() []monitor.HeightData {
	var (
		heights    []monitor.HeightData
		wg         sync.WaitGroup
		heightChan = make(chan monitor.HeightData, len(c.clients))
	)

	for _, cl := range c.clients {
		wg.Add(1)

		go func(cl *rpcclient.Client) {
			defer wg.Done()

			stHeight, err := cl.GetStateHeight()
			if err != nil {
				log.Printf("GetStateHeight for %s: %v", cl.Endpoint(), err)
				return
			}

			heightChan <- monitor.HeightData{
				Host:  cl.Endpoint(),
				Value: stHeight.Local,
			}
		}(cl)
	}

	go func() {
		wg.Wait()
		close(heightChan)
	}()

	for height := range heightChan {
		heights = append(heights, height)
	}

	return heights
}

func (c *Pool) FetchState(height uint32) []monitor.StateData {
	var (
		states    []monitor.StateData
		wg        sync.WaitGroup
		stateChan = make(chan monitor.StateData, len(c.clients))
	)

	for _, cl := range c.clients {
		wg.Add(1)

		go func(cl *rpcclient.Client) {
			defer wg.Done()

			stHeight, err := cl.GetStateRootByHeight(height)
			if err != nil {
				log.Printf("GetStateRootByHeight for %s: %v", cl.Endpoint(), err)
				return
			}

			stateChan <- monitor.StateData{
				Host:  cl.Endpoint(),
				Value: stHeight.Hash().String(),
			}
		}(cl)
	}

	go func() {
		wg.Wait()
		close(stateChan)
	}()

	for state := range stateChan {
		states = append(states, state)
	}

	return states
}

func neoGoClient(ctx context.Context, endpoint string, opts rpcclient.Options) (*rpcclient.Client, error) {
	cli, err := rpcclient.New(ctx, endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("can't create neo-go client: %w", err)
	}

	return cli, nil
}
