package main

import "fmt"

/*
	Here we'll look at different ways of writing a process.

	The implementation will depend highly on how components and ports have been defined.

	Overall we need to make decisions:

	1. who owns the ports: process or component or connection
	2. who owns the data: process or component
	3. is it concurrent
	4. how do we stop/start the process
	5. do we combine the process and component (with/without embedding)
*/

/*
	Who owns the ports, is mainly a syntactic issue.
*/

/*
	Who owns the data is a question also on how to write components.

	By having component own the data, it becomes obvious what data
	it is working on, however the component definition becomes longer.

	Having process own the data makes it slightly easier to inspect
	and the components end up being shorter to write (in some cases).
*/

/*
	Is it concurrent is a question about the performance of the system.
	It might seem counter-intuitive, but a concurrent system is not necessarily
	faster, but it can be slower due to the communication overhead.

	If the components are large-grained, such that they don't have to
	process that many information packets, then it probably doesn't make
	a significant difference.

	However, let's say you need to process 1e6 information packets per second
	then a communication cost (assuming 50ns per packet) would end up
	as 0.05second. Which would be a significant portion of the second.
	This isn't accounting the thread/goroutine scheduling and cache trashing
	that may happen with large graphs.

	Similarly, since much of the servers handle requests from many
	users. It might make sense to run a single network in a separate
	thread/goroutine rather than each component. It would still get the
	benefit of parallelism, without the communication overhead.
*/

/*
	Stopping and starting the process is a question on whether it should
	handle things with context.Context or some other mechanism.

	Similarly, how do you introduce exit points for the components.
	One approach would be to handle exiting with in and out ports.

	This makes the implementation easy, however, it creates a question on
	what do you do with inflight messages that's being currently processed.

	Alternatively a component could have a "hard-stop" and "graceful-stop"
	distinction, where the graceful-stop is specially handled.

	Using context.Context would allow nicer integration with Go, for example
	the components could use it to make http requests and hence have cancellable
	requests as well.

	This context.Context could be an explicit parameter or could be
	integrated into Process itself -- i.e. Process itself is a context.Context.

	However, trying to fit context.Context into the system can introduce additional
	complexity. If the system is short-lived then there might not be any significant
	benefit to it.
*/

/*
	We could flip the dependency and make component embed a process.

	This would mean that the network has to deal with the generic structure
	and interface. Similarly, it could make handling the process level control
	from the network more difficult.

	The difficulty and complexity arises, because the network needs to control processes
	not components. By pushing the process a level deeper, means there needs to be a way
	to access the internal process. Or it would need to expose all the behavior and control
	parts.

	Although this approach could allow for interesting varations where some processes
	are concurrent and some are reactive.
*/

type Process struct {
	In map[string]chan string
}

type Printer struct {
	Process
}

func (printer *Printer) Execute() {
	for value := range printer.In["in"] {
		fmt.Println(value)
	}
}
