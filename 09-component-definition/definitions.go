package main

import "fmt"

/*

	Here we'll look at different ways of writing components.

	We'll use `chan string` as the port, we'll look the connections
	in a separate folder.

*/

/*
	First of all the usual verison, you have a struct per component,
	where fields may define configuration.

	The `In` could be hooked up manually or via reflection.
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
	Then we could do the port lookup inside the component constructor:
*/

type Printer2 struct {
	in <-chan string
}

func NewPrinter2(p *Process) *Printer2 {
	return &Printer2{
		in: p.In("IN"),
	}
}

func (printer *Printer2) Execute(p *Process) {
	for value := range printer.in {
		fmt.Println(value)
	}
}

/*
	Alternatively, it could be done as part of Execute:
*/

type Printer3 struct {
	in <-chan string
}

func (printer *Printer3) Execute(p *Process) {
	printer.in = p.In("IN")

	for value := range printer.in {
		fmt.Println(value)
	}
}

/*
	One common approach is to use closures to define functionality.

	This return the execute function.
*/

func Printer4(p *Process) (execute func()) {
	in := p.In("IN")

	return func() {
		for value := range in {
			fmt.Println(value)
		}
	}
}

/*
	Now via reflection it would also be possible to define components as functions
	and the ports as arguments.

	One of the issues is that with reflection it's not possible to figure out the
	argument names.
*/

func Printer5(in <-chan string) {
	for value := range in {
		fmt.Println(value)
	}
}

/*
	One option to capture the names is to use a struct instead.

	Of course, both versions using reflection will have some overhead.

	To gain persitence across runs it would either need to persist the arguments.
*/

func Printer6(port *struct {
	In <-chan string
}) {
	for value := range port.In {
		fmt.Println(value)
	}
}

/*
	Although in principle it doesn't differ much from this definition
	that uses reflection to fill in the ports.

	This version would be preferred over the previous ones, because it's slightly
	clearer how it works.
*/

type Printer7 struct {
	In <-chan string
}

func (p *Printer7) Execute() {
	for value := range p.In {
		fmt.Println(value)
	}
}

/*
	It's also possible to treat components as just functionality and no state at all.

	The first example uses a `map[string]string` to hang data to the process.
	This is quite similar to
*/

func Printer8(p *Process) {
	in := p.In("IN")
	data := p.Data()
	for value := range in {
		data[value] = "found"
	}
}

/*
	It's also possible to treat components as just functionality and no state at all.

	We could also use an approach similar to sync.Pool to persist data.
*/

func Printer9(p *Process) {
	in := p.In("IN")

	type Data struct {
		Counter int
	}

	data := p.Data2("default", func() interface{} {
		return &Data{}
	}).(*Data)

	for value := range in {
		fmt.Println(value)
		data.Counter++
	}
}

/*
	To prevent name collisions between components a tag type can be used instead
	of a string.
*/

func Printer10(p *Process) {
	in := p.In("IN")

	type Tag struct{}
	type Data struct {
		Counter int
	}
	data := p.Data2(Tag{}, func() interface{} {
		return &Data{}
	}).(*Data)

	for value := range in {
		fmt.Println(value)
		data.Counter++
	}
}

/* stub to make compilation work */

type Process struct{}

func (p *Process) In(name string) <-chan string {
	// TODO:
	return nil
}

func (p *Process) Data() map[string]string {
	// TODO:
	return nil
}

func (p *Process) Data2(tag interface{}, create func() interface{}) interface{} {
	// TODO:
	return nil
}
