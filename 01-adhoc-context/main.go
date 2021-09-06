package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"golang.org/x/sync/errgroup"
)

/*

This is pretty similar to 00-adhoc, however has added proper cancellation.
This means you can tear-down the network with Ctrl-C and handle each component
stopping gracefully.

Otherwise, it has pretty similar problems as previous... and also, the code
isn't as clear as before.

*/

func main() {
	// Setup cancellable context.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Use a errgroup for managing goroutines to avoid needing to use sync.WaitGroup etc.
	var processes errgroup.Group
	defer processes.Wait()

	// Setup connections between components.
	upcaseIn := make(chan string)
	upcaseOut := make(chan string)
	printerIn := upcaseOut

	// Start input component.
	processes.Go(func() error {
		defer close(upcaseIn)
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case upcaseIn <- fmt.Sprintf("Hello %d", i):
			}
		}
		return nil
	})

	// Start upcase component.
	processes.Go(func() error {
		defer close(upcaseOut)

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case value, ok := <-upcaseIn:
				if !ok {
					return nil
				}

				value = strings.ToUpper(value)

				select {
				case <-ctx.Done():
					return ctx.Err()
				case upcaseOut <- value:
				}
			}
		}
	})

	// Start output component
	processes.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case value, ok := <-printerIn:
				if !ok {
					return nil
				}
				fmt.Println(value)
			}
		}
	})
}
