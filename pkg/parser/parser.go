package parser

import (
	"bufio"
	"fmt"
	"strings"
)

type IR interface {
	Add(object Compilable)
	Compile(*bufio.Writer) error
	Parsed() []Compilable
}

type Compilable interface {
	Declare(State) (declared []string, assign []string, vars []Var, types []Type, err error)
}

type State interface {
	GetTypes() []Type
	GetVars() []Var
	NewType(Type) error
	NewVar(Var) error
}

func Compile(ir IR, state State) (globals []string, main []string, err error) {
	for _, compilable := range ir.Parsed() {
		declared, assign, vars, types, err := compilable.Declare(state)
		if err != nil {
			return globals, main, fmt.Errorf("Error compiling: %s", err)
		}
		for _, T := range types {
			err := state.NewType(T)
			if err != nil {
				return globals, main, err
			}
		}
		for _, v := range vars {
			err := state.NewVar(v)
			if err != nil {
				return globals, main, err
			}
		}
		globals = append(globals, declared...)
		main = append(main, assign...)
	}
	for _, v := range state.GetVars() {
		if v.Allocated() {
			main = append(main, fmt.Sprintf("free(%s)", v.Signature()))
		}
	}
	return globals, main, nil
}

type Type interface {
	Translatable
	DefineVar(name string, val string) (declare []string, assign []string, result []Var, err error)
	CompileVal(val string) (string, error)
	CompileAssign(name string, val string) ([]string, error)
	SatisfiedBy(Type) bool
}

type Var interface {
	Translatable
	Type() Type
	Allocated() bool
}

type Translatable interface {
	Namable
	Signature() string
}

type Namable interface {
	GetName() string
}

type Lookup [256]bool

type Processor interface {
	Try(any, byte) (defined Compilable, err error)
	Get() any
	Next(byte) (Processor, error)
}

type Node[I, O any] struct {
	name    string
	links   []Linkable[O]
	process func(I, byte) (O, Compilable)
	result  O
}

func Make[I, O any](name string, process func(I, byte) (O, Compilable)) Node[I, O] {
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

func (n *Node[I, O]) Process(in I, char byte) Compilable {
	result, returned := n.process(in, char)
	n.result = result
	return returned
}

func (n *Node[I, O]) Try(in any, char byte) (Compilable, error) {
	valid, ok := in.(I)
	if !ok {
		return nil, fmt.Errorf("Compiler error: got type %T inside %s instead of expected type %T", in, n.name, valid)
	}
	defined := n.Process(valid, char)
	return defined, nil
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
	Process(I, byte) Compilable
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
	return Prep(bind...)
}

func Prep[I any](links ...Linkable[I]) []Linkable[I] {
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

func Pass[IO any]() func(IO, byte) (IO, Compilable) {
	return func(io IO, char byte) (IO, Compilable) {
		return io, nil
	}
}
