package main

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Component = func(context.Context) error

type Network struct {
	list []Component
}

func (net *Network) Add(proc Component) {
	net.list = append(net.list, proc)
}

func (net *Network) Run(ctx context.Context) error {
	var group errgroup.Group

	for _, proc := range net.list {
		proc := proc
		group.Go(func() error {
			return proc(ctx)
		})
	}

	return group.Wait()
}
