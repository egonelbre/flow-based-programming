package main

import (
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
)

/*

This demonstrates an ad-hoc FBP like system in Go. This approach is useful when
you want a fixed pipeline in an existing system.

Obviously there are a bunch of problems with this system:

1. There's no explicit notion of a network, process nor component.
2. It's easy to mess up channel usage,
	e.g. when one of the components never closes it's output,
	it leaves all other components running.
3. There's no way to change connections live.
4. There's no way to start/stop components.
5. There's no way to reconfigure components live.
6. It doesn't respond to signals (e.g. CTRL-C gracefully)

Nevertheless, it demonstrates a minimal FBP approach in Go.

*/

func main() {
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
			upcaseIn <- fmt.Sprintf("Hello %d", i)
		}
		return nil
	})

	// Start upcase component.
	processes.Go(func() error {
		defer close(upcaseOut)
		for value := range upcaseIn {
			upcaseOut <- strings.ToUpper(value)
		}
		return nil
	})

	// Start output component
	processes.Go(func() error {
		for value := range printerIn {
			fmt.Println(value)
		}
		return nil
	})
}
