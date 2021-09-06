package main

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Component interface {
	Name() string
	Run(context.Context) error
}

type Network struct {
	list []Component
}

func (net *Network) Add(com Component) {
	net.list = append(net.list, com)
}

func (net *Network) Run(ctx context.Context) error {
	var group errgroup.Group

	for _, com := range net.list {
		com := com
		group.Go(func() error {
			return com.Run(ctx)
		})
	}

	return group.Wait()
}
