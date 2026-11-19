package parser

import (
	"bufio"
	"fmt"
	"strings"
)

type Context interface {
	Add(object any) error
	Compile(*bufio.Writer) error
}

type Lookup [256]bool

type Processor interface {
	Try(Context, any, byte) error
	Get() any
	Next(byte) (Processor, error)
}

type Node[I, O any] struct {
	name    string
	links   []Linkable[O]
	process func(Context, I, byte) O
	result  O
}

func Make[I, O any](name string, process func(Context, I, byte) O) Node[I, O] {
	var new O
	return Node[I, O]{
		name:    name,
		links:   []Linkable[O]{}, // Links given after to avoid circular assignation
		process: process,
		result:  new,
	}
}

var _ Processor = &Node[any, any]{}

func (n Node[I, O]) GetName() string {
	return n.name
}

func (n *Node[I, O]) Has(nodes ...Linkable[O]) {
	n.links = nodes
}

func (n *Node[I, O]) Process(c Context, in I, char byte) {
	n.result = n.process(c, in, char)
}

func (n *Node[I, O]) Try(c Context, in any, char byte) error {
	valid, ok := in.(I)
	if !ok {
		return fmt.Errorf("Compiler error: got type %T inside %s instead of expected type %T", in, n.name, valid)
	}
	n.Process(c, valid, char)
	return nil
}

func (n Node[I, O]) Get() any {
	return n.result
}

func (n Node[I, O]) Next(char byte) (Processor, error) {
	unhandled := []string{}
	for _, linkable := range n.links {
		if linkable.CanHandle(char) {
			return linkable, nil
		}
		unhandled = append(unhandled, describe(linkable))
	}
	return nil, fmt.Errorf("%s: Unhandled char %s, expected: %s", n.name, string(char), strings.Join(unhandled, " or "))
}

func describe[T any](l Linkable[T]) string {
	allowed := make([]string, 0)
	streak := make([]byte, 0)
	for _, handled := range l.Handled() {
		streak, allowed = merge_pretty(streak, allowed, handled[0])
	}
	allowed = merge(streak, allowed)
	return fmt.Sprintf("%s (%s)", l.GetName(), strings.Join(allowed, "|"))
}

func merge_pretty(streak []byte, allowed []string, char byte) ([]byte, []string) {
	if len(streak) <= 0 {
		return append(streak, char), allowed
	}
	last := streak[len(streak)-1]
	if char == last+1 {
		return append(streak, char), allowed
	} else {
		return []byte{char}, merge(streak, allowed)
	}
}

func merge(streak []byte, allowed []string) []string {
	if len(streak) <= 0 {
		return allowed
	}
	if len(streak) <= 1 {
		return append(allowed, string(streak[0]))
	}
	return append(allowed, fmt.Sprintf("[%s-%s]", string(streak[0]), string(streak[len(streak)-1])))
}

type Linkable[I any] interface {
	Processor
	Process(Context, I, byte)
	CanHandle(byte) bool
	GetName() string
	Handled() []string
}

func Bind[I, O any](n *Node[I, O], chars ...byte) Linkable[I] {
	var handled Lookup
	for _, char := range chars {
		handled[char] = true
	}
	return &Link[I, O]{
		Node:    n,
		handled: handled,
	}
}

type ignorer []byte

func Ignoring(chars ...byte) ignorer {
	return ignorer(chars)
}

func (ignored ignorer) Has[IO any](links ...Linkable[IO]) []Linkable[IO] {
	node := Make(
		"ignored",
		Pass[IO](),
	)
	bind := append(links, Bind(&node, []byte(ignored)...))
	node.Has(bind...)
	return Has(bind...)
}

func Has[I any](links ...Linkable[I]) []Linkable[I] {
	return links
}

type Link[I, O any] struct {
	*Node[I, O]
	handled Lookup
}

var _ Linkable[any] = &Link[any, any]{}

func (l Link[I, O]) CanHandle(char byte) bool {
	return l.handled[char]
}

func (l Link[I, O]) Handled() []string {
	handled := []string{}
	for i, canHandle := range l.handled {
		if canHandle {
			handled = append(handled, string(byte(i)))
		}
	}
	return handled
}

func Pass[IO any]() func(Context, IO, byte) IO {
	return func(c Context, io IO, char byte) IO {
		return io
	}
}
