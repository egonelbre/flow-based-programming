package main

import "time"

/*
	So far we have been sending strings back-and forth.
	However we can structure messages better.
*/

/*
	First approach would be use different types for
	different messages.
*/

type User struct {
	Name string
}

/*
	We could also use a structured format, something json like.

	These messages would be trivial to send over the network.
*/

type Kind byte

const (
	StringKind = Kind(0)
	IntKind    = Kind(1)
	TreeKind   = Kind(2)
)

type Node struct {
	Kind     Kind
	String   string
	Int      int64
	Children []Node
}

/*
	To add headers to messages we can use two approaches.

	1. Wrapping
	2. Field
	3. Embedding
*/

// Wrapping apprach -- with generics it would allow to constrain the type.
type Message struct {
	SentAt  time.Time
	TTL     time.Time
	Content interface{} // arbitrary type
}

// Using a field. This approach has been used extensively in https://github.com/miekg/dns/blob/master/types.go.
type Header2 struct {
	SentAt time.Time
	TTL    time.Time
}

type User2 struct {
	Hdr  Header2
	Name string
}

// Embedding, as a varation on using a field. This can provide some convenience benefits.
type Header3 struct {
	SentAt time.Time
	TTL    time.Time
}

type User3 struct {
	Header3
	Name string
}
