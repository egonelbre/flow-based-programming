package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
)

/*

Instead of using a plain `func(context.Context) error` as the component definition.
It switches to using an interface:

type Component interface {
	Name() string
	Run(context.Context) error
}

This allows to put more constraints on the component itself.
*/

func main() {
	// Setup cancellable context.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Use a custom network to avoid needing to use sync.WaitGroup etc.
	var network Network

	// create components
	hello := &Hello{Count: 10}
	upper := &Upper{}
	printer := &Printer{}

	// create connections between components
	hello.Out, upper.In = StringConnection()
	upper.Out, printer.In = StringConnection()

	// add components
	network.Add(hello)
	network.Add(upper)
	network.Add(printer)

	// start the network
	err := network.Run(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// StringConnection creates a new channel with input and output channels.
func StringConnection() (out chan<- string, in <-chan string) {
	ch := make(chan string)
	return ch, ch
}

// Hello components generates Count hellos.
type Hello struct {
	Count int
	Out   chan<- string
}

func (*Hello) Name() string { return "Hello" }

func (hello *Hello) Run(ctx context.Context) error {
	defer close(hello.Out)
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case hello.Out <- fmt.Sprintf("Hello %d", i):
		}
	}
	return nil
}

// Upper component upper-cases the strings.
type Upper struct {
	In  <-chan string
	Out chan<- string
}

func (*Upper) Name() string { return "Upper" }

func (upper *Upper) Run(ctx context.Context) error {
	defer close(upper.Out)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case value, ok := <-upper.In:
			if !ok {
				return nil
			}

			value = strings.ToUpper(value)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case upper.Out <- value:
			}
		}
	}

	return nil
}

// Printer prints the input values.
type Printer struct {
	In <-chan string
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
