package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

/*

So far we haven't had an explicit idea of a Connection.

We'll start one which we can reconfigure while the network is running.

We'll use a separate goroutine to pump messages from one channel to another.

One of the concerns which such an approach is that hwat happens when a
connection is cut while a message is in that "channel". For example the other
node has stalled.
*/

func main() {
	// Setup cancellable context.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Use a custom network to avoid needing to use sync.WaitGroup etc.
	var network Network

	// create components
	hello := NewHello(time.Second)
	upper := NewUpper()
	lower := NewLower()
	printer := NewPrinter()

	// add components to the network
	network.Add(hello)
	network.Add(upper)
	network.Add(lower)
	network.Add(printer)

	var group errgroup.Group
	group.Go(func() error { return network.Run(ctx) })

	// configure the network live:
	group.Go(func() error {
		// connect both upper and lower to the printer
		ConnectString(upper.Out, printer.In)
		ConnectString(lower.Out, printer.In)

		// This starts changing the Hello connection between upper and lower.
		for {
			helloToUpper := ConnectString(hello.Out, upper.In)

			select {
			case <-time.After(3 * time.Second):
			case <-ctx.Done():
				return ctx.Err()
			}

			helloToUpper.Cut()

			helloToLower := ConnectString(hello.Out, lower.In)
			select {
			case <-time.After(3 * time.Second):
			case <-ctx.Done():
				return ctx.Err()
			}
			helloToLower.Cut()
		}
		return nil
	})

	// start the network
	err := group.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// Connection pumps message from one port to another.
func ConnectString(out chan string, in chan string) *StringConnection {
	var conn StringConnection
	conn.pump(out, in)
	return &conn
}

type StringConnection struct {
	cut    sync.Once
	stop   chan struct{}
	exited chan struct{}
}

func (conn *StringConnection) pump(out, in chan string) {
	conn.stop = make(chan struct{})
	conn.exited = make(chan struct{})

	go func() {
		defer close(conn.exited)

		for {
			select {
			case <-conn.stop:
				return
			case val := <-out:
				select {
				case <-conn.stop:
					// TODO: what should happen when the receiving component
					// is not able to receive and you cut the connection?
					return
				case in <- val:
				}
			}
		}
	}()
}

func (conn *StringConnection) Cut() {
	conn.cut.Do(func() { close(conn.stop) })
	<-conn.exited
}

// Hello components generates Count hellos.
type Hello struct {
	Interval time.Duration
	Out      chan string
}

func NewHello(interval time.Duration) *Hello {
	return &Hello{
		Interval: interval,
		Out:      make(chan string),
	}
}

func (*Hello) Name() string { return "Hello" }

func (hello *Hello) Run(ctx context.Context) error {
	for count := 0; ; count++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case hello.Out <- fmt.Sprintf("Hello %d", count):
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(hello.Interval):
		}
	}
	return nil
}

// Upper component upper-cases the strings.
func NewUpper() *StringProcessor { return NewStringProcessor("Upper", strings.ToUpper) }

// Lower component lower-cases the strings.
func NewLower() *StringProcessor { return NewStringProcessor("Lower", strings.ToLower) }

// StringProcessor implements a generic component that can be used to process strings.
type StringProcessor struct {
	In  chan string
	Out chan string

	name    string
	process func(string) string
}

func NewStringProcessor(name string, process func(string) string) *StringProcessor {
	return &StringProcessor{
		In:  make(chan string),
		Out: make(chan string),

		name:    name,
		process: process,
	}
}

func (p *StringProcessor) Name() string { return p.name }

func (p *StringProcessor) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case value, ok := <-p.In:
			if !ok {
				return nil
			}

			if p.process != nil {
				value = p.process(value)
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case p.Out <- value:
			}
		}
	}

	return nil
}

// Printer prints the input values.
type Printer struct {
	In chan string
}

func NewPrinter() *Printer {
	return &Printer{
		In: make(chan string),
	}
}

func (*Printer) Name() string { return "Printer" }

func (printer *Printer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case value, ok := <-printer.In:
			if !ok {
				return nil
			}
			fmt.Println(value)
		}
	}
	return nil
}
