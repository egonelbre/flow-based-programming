package flow

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

type Network struct {
	components []Component
}

func (net *Network) Add(components ...Component) {
	net.components = append(net.components, components...)
}

func (net *Network) Run(ctx context.Context) error {
	var g errgroup.Group
	for _, c := range net.components {
		c := c
		g.Go(func() error {
			return c.Run(ctx)
		})
	}
	return g.Wait()
}

type Component interface {
	Run(ctx context.Context) error
}

type Conn[T any] struct {
	from *Out[T]
	to   *In[T]
}

func Connect[T any](from *Out[T], to *In[T]) *Conn[T] {
	conn := Conn[T]{}
	conn.from = from
	conn.to = to

	data := make(chan T)
	conn.from.swap(data)
	conn.to.swap(data)

	return &conn
}

func (conn *Conn[T]) Disconnect() {
	conn.from.swap(nil)
	conn.to.swap(nil)
}

type In[T any] struct {
	// TODO: support multiple inbound channels

	mu   sync.Mutex
	data chan T
	ping chan struct{}

	create sync.Once
}

func (in *In[T]) init() { in.create.Do(func() { in.ping = make(chan struct{}) })}

func (in *In[T]) swap(data chan T) {
	in.init()

	in.mu.Lock()
	in.data = data
	in.mu.Unlock()

	select{
	case in.ping<-struct{}{}:
	default:
	}
}

func (in *In[T]) current() chan T {
	in.mu.Lock()
	defer in.mu.Unlock()
	return in.data
}

func (in *In[T]) Recv(ctx context.Context) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	in.init()

	for {
		select {
		case <-in.ping:
		default:
		}

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case v := <-in.current():
			return v, nil
		case <-in.ping:
		}
	}
}

type Out[T any] struct {
	mu   sync.Mutex
	data chan T
	ping chan struct{}

	create sync.Once
}

func (out *Out[T]) init() { out.create.Do(func() { out.ping = make(chan struct{}) })}

func (out *Out[T]) swap(data chan T) {
	out.init()

	out.mu.Lock()
	out.data = data
	out.mu.Unlock()

	select{
	case out.ping<-struct{}{}:
	default:
	}
}

func (out *Out[T]) current() chan T {
	out.mu.Lock()
	defer out.mu.Unlock()
	return out.data
}

func (out *Out[T]) Send(ctx context.Context, v T) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	out.init()

	for {
		select {
		case <-out.ping:
		default:
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case out.current() <- v:
			return nil
		case <-out.ping:
		}
	}
}
