package main

import (
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
)

/*

This is pretty much the previous implementation, but with a helper to setup connections.

*/

func main() {
	// Use a errgroup for managing goroutines to avoid needing to use sync.WaitGroup etc.
	var processes errgroup.Group
	defer processes.Wait()

	// create components
	hello := &Hello{Count: 10}
	upper := &Upper{}
	printer := &Printer{}

	// create connections between components
	helloToUpper := make(chan string)
	hello.Out = helloToUpper
	upper.In = helloToUpper

	upperToPrinter := make(chan string)
	upper.Out = upperToPrinter
	printer.In = upperToPrinter

	// start components
	processes.Go(hello.Run)
	processes.Go(upper.Run)
	processes.Go(printer.Run)
}

// Hello components generates Count hellos.
type Hello struct {
	Count int
	Out   chan<- string
}

func (hello *Hello) Run() error {
	defer close(hello.Out)
	for i := 0; i < 10; i++ {
		hello.Out <- fmt.Sprintf("Hello %d", i)
	}
	return nil
}

// Upper component upper-cases the strings.
type Upper struct {
	In  <-chan string
	Out chan<- string
}

func (upper *Upper) Run() error {
	defer close(upper.Out)
	for value := range upper.In {
		upper.Out <- strings.ToUpper(value)
	}
	return nil
}

// Printer prints the input values.
type Printer struct {
	In <-chan string
}

func (printer *Printer) Run() error {
	for value := range printer.In {
		fmt.Println(value)
	}
	return nil
}
