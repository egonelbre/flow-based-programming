package main

import (
	"context"
	"fmt"
)

/*

Here we'll look at different ways of writing ports.
*/

/*
	The version which most Go users would lean towrads is
	using channels for the configuration.

	There are few problems with this approach though.
	It doesn't allow dynamically changing the connections.

	Similarly handling context cancellations and component
	stopping is going to introduce a bunch of code.
*/

type Printer struct {
	In <-chan string
}

func (printer *Printer) Execute() {
	for value := range printer.In {
		fmt.Println(value)
	}
}

/*
	The other question about using ports is, which party is
	responsible for creating the ports.

	First, we could push such logic into the Process and
	let the Process track the ports.

	As a benefit, the component doesn't have to make the port
	public.
*/

type Printer2 struct {
	in <-chan string
}

func NewPrinter2(p *Process) *Printer2 {
	return &Printer2{
		in: p.In("IN"),
	}
}

/*
	We could have the component itself do it.
	However, in this case it would end up introducing a lot of
	duplication in components.

	Similarly, you wouldn't be able to easily change the connections
	dynamically.

	It's going to be also confusing on who is responsible for closing
	the channels. This of course isn't a problem with a fixed-form network.
*/

type Printer3 struct {
	In chan string
}

func NewPrinter3() *Printer3 {
	return &Printer3{
		In: make(chan string),
	}
}

/*
	Alternatively, the setup could be done as part of Execute,
	either pulling it from the process or creating one yourself.
*/

type Printer4 struct {
	in <-chan string
}

func (printer *Printer4) Execute(p *Process) {
	printer.in = p.In("IN")

	for value := range printer.in {
		fmt.Println(value)
	}
}

/*
	Then it's possible to do it via reflection.
	The process would walk over the ports and fill them in.
*/

type Printer5 struct {
	In <-chan string
}

func (printer *Printer5) Execute() {
	for value := range printer.In {
		fmt.Println(value)
	}
}

/*
	The port filling could be done via descriptors.

	This approach is somewhat similar to the "reflection" version,
	however it avoids using reflect package.

	This approach however allows to easily create systems where
	the component has dynamic number of ports.
*/

type Printer6 struct {
	in chan string
}

type Port6 struct {
	name string
	in   *chan string
}

func (printer *Printer6) Ports() []Port6 {
	return []Port6{
		{"in", &printer.in},
	}
}

func (printer *Printer6) Execute() {
	for value := range printer.in {
		fmt.Println(value)
	}
}

/*
	The component could keep track ports in a separate struct.

	It probably isn't significantly better than previous,
	however it might have some specific use-cases.
*/

type Ports7 struct {
	In  map[string]chan string
	Out map[string]chan string
}

type Printer7 struct {
	Ports7
}

func (printer *Printer7) Execute() {
	for value := range printer.In["in"] {
		fmt.Println(value)
	}
}

/*
	There are few variations on how to component stopping.

	First, using the context approach.
	While it does work, it's rather verbose.
*/

type Printer8 struct {
	In <-chan string
}

func (printer *Printer8) Execute(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case value := <-printer.In:
			fmt.Println(value)
		}
	}
}

/*
	We can help the situation by adding a wrapper type
	that provides context support.
*/

type Port9 struct {
	Data chan string
}

func (port *Port9) Recv(ctx context.Context) (value string, ok bool) {
	select {
	case <-ctx.Done():
		return "", false
	case value, ok := <-port.Data:
		return value, ok
	}
}

type Printer9 struct {
	In *Port9
}

func (printer *Printer9) Execute(ctx context.Context) {
	for {
		value, ok := printer.In.Recv(ctx)
		if !ok {
			return
		}

		fmt.Println(value)
	}
}

/*
	Instead of using context.Context, we can
	provide a custom cancellation inside the process.

	This of course wouldn't mesh as well with the rest of Go,
	however, it might be beneficial for writing a framework.
*/

type Port10 struct {
	Process *Process
	Data    chan string
}

func (port *Port10) Recv() (value string, ok bool) {
	select {
	case <-port.Process.Stop:
		return "", false
	case value, ok := <-port.Data:
		return value, ok
	}
}

/*
	By providing a custom port type it would be possible to allow
	swapping out the port.
*/

type Port11ChangeRequest struct {
	New     chan string
	Swapped func() bool
}

type Port11 struct {
	Process *Process
	Data    chan string
	Swap    chan *Port11ChangeRequest
}

func (port *Port11) Recv() (value string, ok bool) {
retry:
	select {
	case swap := <-port.Swap:
		port.Data = swap.New
		swap.Swapped()
		goto retry
	case <-port.Process.Stop:
		return "", false
	case value, ok := <-port.Data:
		return value, ok
	}
}

/*
	With all these wrapper ports, one of the problem is having properly
	typed messages. Currently the main approaches how to avoid the problems
	would be to make the channels use interface{} (or something based on it).

	Alternatively, implement different port types, e.g. StringInPort.

	However, with Go 1.18 it would be possible to write a generic port.

	type Port[T any] struct {
		Process *Process[T]
		Data    chan T
	}
*/

/* stub to make compilation work */

type Process struct {
	Stop chan struct{}
}

func (p *Process) In(name string) <-chan string {
	// TODO:
	return nil
}
