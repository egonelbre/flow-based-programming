package main

import "fmt"

/*
	Here we'll look at different ways of writing connections.
*/

/*

	A really good starting point for Go is using a channel.

	However it does have a few limitations, it doesn't support prioritization
	similarly it's difficult to manipulate or observe what's in the current
	channel at all.

	Similarly, channels require to have a specific length which cannot be dynamically
	changed. Or at least not trivially.

	It also doesn't keep track to which ports it's being connected to.

	The sends on such a channel can also become a bottleneck. A good estimate is that
	a send/recv pair will take ~300ns.
*/

func ExampleChannel() {
	var in <-chan string
	var out chan<- string

	connection := make(chan string)
	in, out = connection, connection

	_, _ = in, out
}

/*
	An extension to basic channel would be to add tracking the Ports.

	This would give us capability to cut the connection by referencing
	the connection. Of course, the ports would need to support it.
*/

type InPort1 struct {
	Data <-chan string
}

type OutPort1 struct {
	Data chan<- string
}

type Connection1 struct {
	From *OutPort1
	Data chan string
	To   *InPort1
}

/*
	A custom SPSC / MPSC queue could be used that has better performance characteristics.

	This can reduce the send/recv overhead to ~40ns.

	For example https://github.com/loov/queue/tree/master/extqueue contains several implementations.

	Of course, using a custom queue wouldn't mesh well with Go channels, which is
	a more common way to handle communication.
*/

/*
	We could also avoid concurrency altogether or make it optional by using a
	more event / callback based approach.

	This would of course require writing components in a completely different
	manner.
*/

type Printer3 struct{}

func (*Printer3) Setup(p *Process3) {
	p.On("IN", func(message string) {
		fmt.Println(message)
	})
}

type Process3 struct{} // stub to make things work

func (*Process3) On(msg string, callback func(message string)) {}

/*
	A similar version would be to implement a Re-Actor approach instead.
	This approach can be much more performant compared to the concurrent version.
	See "Development and Deployment of Multiplayer Online Games" for more details.

	Of course, since it's not going to be concurrent it's going further from
	the ideals of FBP.
*/

type Printer4 struct{}
type Outbox4 struct{}

func (*Outbox4) Send(port string, message string) {}

func (*Printer4) Handle(port, msg string, out Outbox4) {
	switch port {
	case "IN":
		fmt.Println(msg)
		out.Send("OUT", msg) // we'll repeat it to out
	}
}
