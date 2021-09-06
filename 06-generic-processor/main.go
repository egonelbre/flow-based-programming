package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
)

/*

If we wish to add more string processor components, each of them would end up
much longer. However we can write a type such as StringProcessor for simplifying our code.

This would mean that the upper implementation can be reduced to:

    func NewUpper() *StringProcessor { return NewStringProcessor("Upper", strings.ToUpper) }

*/

func main() {
	// Setup cancellable context.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Use a custom network to avoid needing to use sync.WaitGroup etc.
	var network Network

	// create components
	hello := &Hello{Count: 10}
	upper := NewUpper()
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
	for i := 0; i < hello.Count; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case hello.Out <- fmt.Sprintf("Hello %d", i):
		}
	}
	return nil
}

func NewUpper() *StringProcessor { return NewStringProcessor("Upper", strings.ToUpper) }
func NewLower() *StringProcessor { return NewStringProcessor("Lower", strings.ToLower) }

// StringProcessor implements a generic component that can be used to process strings.
type StringProcessor struct {
	In  <-chan string
	Out chan<- string

	name    string
	process func(string) string
}

func NewStringProcessor(name string, process func(string) string) *StringProcessor {
	return &StringProcessor{
		name:    name,
		process: process,
	}
}

func (p *StringProcessor) Name() string { return p.name }

func (p *StringProcessor) Run(ctx context.Context) error {
	defer close(p.Out)

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
