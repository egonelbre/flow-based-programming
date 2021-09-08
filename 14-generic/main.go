package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fbp.example/flow"
)

/*
	NOTE: this code requires Go.tip (Go 1.18)

	This example shows a version that uses generics and implements several features:

	* Dynamically modifying connections.
	* Typesafe communication.

	TODO:

	* Stop/Start processes / components.
	* Connect components live.
	* Disconnect components live.
	* Multi-connect
*/

type Hello struct {
	Out flow.Out[string]

	count int
}

func (e *Hello) Run(ctx context.Context) error {
	for {
		err := e.Out.Send(ctx, "Hello " + strconv.Itoa(e.count))
		if err != nil {
			return err
		}

		e.count++
		time.Sleep(500*time.Millisecond)
	}
}

type Upper struct {
	In  flow.In[string]
	Out flow.Out[string]
}

func (u *Upper) Run(ctx context.Context) error {
	for {
		v, err := u.In.Recv(ctx)
		if err != nil {
			return err
		}

		v = strings.ToUpper(v)

		err = u.Out.Send(ctx, v)
		if err != nil {
			return err
		}
	}
}

type Lower struct {
	In  flow.In[string]
	Out flow.Out[string]
}

func (u *Lower) Run(ctx context.Context) error {
	for {
		v, err := u.In.Recv(ctx)
		if err != nil {
			return err
		}

		v = strings.ToLower(v)

		err = u.Out.Send(ctx, v)
		if err != nil {
			return err
		}
	}
}
type Printer[T any] struct {
	In flow.In[T]
}

func (p *Printer[T]) Run(ctx context.Context) error {
	for {
		v, err := p.In.Recv(ctx)
		if err != nil {
			return err
		}

		fmt.Println(v)
	}
}

func main() {
	var net flow.Network

	var (
		hello Hello
		upper Upper
		lower Lower
		printer Printer[string]
	)

	net.Add(&hello, &upper, &lower, &printer)

	go net.Run(context.Background())

	{
		first := flow.Connect(&hello.Out, &upper.In)
		second := flow.Connect(&upper.Out, &printer.In)
		time.Sleep(3 * time.Second)
		first.Disconnect()
		second.Disconnect()
	}

	{
		first := flow.Connect(&hello.Out, &lower.In)
		second := flow.Connect(&lower.Out, &printer.In)
		time.Sleep(3 * time.Second)
		first.Disconnect()
		second.Disconnect()
	}
}