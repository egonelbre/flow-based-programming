package main

import (
	"fmt"
	"strings"

	"fbp.example/flow"
)

/*

Now a complete change of the approach.

In principle we can take ideas from the previous and combine them into
a easier to use and more minimal usage.

For example, this approach uses reflection heavily and has a basic DSL
for defining the networks.

This was one of the ideas we discussed with Samuel Lampa a long while ago
in the thread https://groups.google.com/forum/#!msg/golang-nuts/vgj_d-MjUHA/T9sE64Yrcq0J.

I do not recommend this approach due to the heavy use of reflection, however,
it does demonstrate some of the possible ways that can be implemented.
*/

type Comm struct{ In, Out chan string }

func main() {
	comm := &Comm{}
	graph := flow.New(comm)

	graph.Registry = flow.Registry{
		"Split": NewSplit,
		"Lower": NewLower,
		"Upper": NewUpper,
	}

	graph.Setup(`
		: s Split
		: l Lower
		: u Upper

		$.In    -> s.In
		s.Left  -> l.In
		s.Right -> u.In

		l.Out -> $.Out
		u.Out -> $.Out
	`)
	graph.Start()

	for i := range []int{1, 2, 3, 4, 5} {
		comm.In <- fmt.Sprintf("Hello %v", i)
	}
	close(comm.In)

	for v := range comm.Out {
		fmt.Printf("%v\n", v)
	}
}

type Split struct{ In, Left, Right chan string }

func NewSplit() flow.Node { return &Split{} }

func (node *Split) Run() error {
	defer close(node.Left)
	defer close(node.Right)
	for v := range node.In {
		m := len(v) / 2
		node.Left <- v[:m]
		node.Right <- v[m:]
	}
	return nil
}

type Lower struct{ In, Out chan string }

func NewLower() flow.Node { return &Lower{} }

func (node *Lower) Run() error {
	defer close(node.Out)
	for v := range node.In {
		node.Out <- strings.ToLower(v)
	}
	return nil
}

type Upper struct{ In, Out chan string }

func NewUpper() flow.Node { return &Upper{} }

func (node *Upper) Run() error {
	defer close(node.Out)
	for v := range node.In {
		node.Out <- strings.ToUpper(v)
	}
	return nil
}
