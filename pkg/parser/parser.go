package parser

import (
	"bufio"
	"fmt"
	"strings"
)

type IR interface {
	Add(object Blamable)
	Compile(*bufio.Writer) error
	Parsed() []Blamable
}

type Blamable struct {
	obj      Compilable
	filename string
	line     int
}

func Blame(filename string, line int, obj Compilable) Blamable {
	return Blamable{
		obj:      obj,
		filename: filename,
		line:     line,
	}
}
func (b Blamable) Get() Compilable {
	return b.obj
}
func (b Blamable) Blame() (filename string, line int) {
	return b.filename, b.line
}

type Compilable interface {
	Declare(*Stack) (declared []string, assign []string, err error)
}

type Stack []State

func (s Stack) GetTypes() []Type {
	types := []Type{}
	for _, state := range s {
		types = append(types, state.GetTypes()...)
	}
	return types
}
func (s Stack) GetVars() []Var {
	vars := []Var{}
	for _, state := range s {
		vars = append(vars, state.GetVars()...)
	}
	return vars
}
func (s Stack) GetFuncs() []Func {
	funcs := []Func{}
	for _, state := range s {
		funcs = append(funcs, state.GetFuncs()...)
	}
	return funcs
}
func (s Stack) AddVar(v Var) error {
	for _, state := range s {
		err := state.VetVar(v)
		if err != nil {
			return err
		}
	}
	if !s.hasType(v.Type()) {
		return fmt.Errorf("Unknown type %s assigned to variable %s.", v.Type().GetName(), v.GetName())
	}
	s.top().NewVar(v)
	return nil
}
func (s Stack) AddType(T Type) error {
	for _, state := range s {
		err := state.VetType(T)
		if err != nil {
			return err
		}
	}
	s.top().NewType(T)
	return nil
}
func (s Stack) AddFunc(f Func) error {
	for _, state := range s {
		err := state.VetFunc(f)
		if err != nil {
			return err
		}
		for _, arg := range f.Args() {
			err := state.VetVar(arg)
			if err != nil {
				return fmt.Errorf("Error creating argument for function \"%s\": %s", f.GetName(), err)
			}
		}
	}
	if !s.hasType(f.Return()) {
		return fmt.Errorf("Unknown type \"%s\" returned by the function %s", f.Return().GetName(), f.GetName())
	}
	for _, arg := range f.Args() {
		if !s.hasType(arg.Type()) {
			return fmt.Errorf("Error creating argument for function \"%s\": The type %s used for variable %s is undefined", f.GetName(), arg.Type().GetName(), arg.GetName())
		}
	}
	s.top().NewFunc(f)
	return nil
}
func (s Stack) hasType(T Type) bool {
	for _, state := range s {
		for _, defined := range state.GetTypes() {
			if defined.GetName() == T.GetName() {
				return true
			}
		}
	}
	return false
}
func (s Stack) Pop() (free []string, err error) {
	i := 0
	for i < len(s.top().GetVars()) {
		v := s.top().GetVars()[i]
		if v.Allocated() {
			free = append(free, Line("free(%s)", v.Signature()))
			err := s.top().Remove(v)
			if err != nil {
				return free, err
			}
		} else {
			i++
		}
	}
	return free, nil
}
func (s Stack) top() State {
	if len(s) <= 0 {
		return nil
	}
	return s[len(s)-1]
}

type State interface {
	GetTypes() []Type
	GetVars() []Var
	GetFuncs() []Func
	NewType(Type)
	NewVar(Var)
	NewFunc(Func)
	VetType(Type) error
	VetVar(Signature) error
	VetFunc(Func) error
	Remove(Signature) error
}

func Compile(instructions []Blamable, stack *Stack) (globals []string, main []string, err error) {
	for _, inst := range instructions {
		filename, line := inst.Blame()
		declared, assign, err := inst.Get().Declare(stack)
		if err != nil {
			return globals, main, fmt.Errorf("%s:%d: %s", filename, line, err)
		}
		globals = appendBlamed(globals, declared, filename, line)
		main = appendBlamed(main, assign, filename, line)
	}
	freed, err := stack.Pop()
	main = append(main, freed...)
	return globals, main, err
}

func appendBlamed(oldLines []string, newLines []string, filename string, index int) []string {
	for _, new := range newLines {
		comment := "\n"
		if filename != "" {
			comment = fmt.Sprintf(" // %s:%d\n", filename, index)
		}
		oldLines = append(oldLines, strings.ReplaceAll(new, "\n", comment))
	}
	return oldLines
}

func Line(format string, vals ...any) string {
	return fmt.Sprintf(format+";\n", vals...)
}

type Type interface {
	Translatable
	DefineVar(name string, val string) (declare []string, assign []string, result []Var, err error)
	CompileVal(val string) (string, error)
	Assign(*Stack, string, Val) ([]string, error)
	Recieve(Val) (string, error)
}

type Var interface {
	Signature
	Allocated() bool
}

type Val interface {
	Signature() (string, error)
	Type() Type
}

type Signature interface {
	Translatable
	Type() Type
}

type Func interface {
	Translatable
	Return() Type
	Args() []Signature
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
	handled := []string{}
	for _, linkable := range n.links {
		if linkable.CanHandle(char) {
			return linkable, nil
		}
		handled = append(handled, describe(linkable))
	}
	return nil, fmt.Errorf("%s", strings.ReplaceAll(fmt.Sprintf(
		"%s: Unhandled char '%s', expected: %s",
		n.name, string(char), strings.Join(handled, " or "),
	), "\n", "\\n"))
}

func describe[T any](l Linkable[T]) string {
	allowed := make([]string, 0)
	streak := make([]byte, 0)
	for _, handled := range l.Handled() {
		streak, allowed = streak_or_merge(streak, allowed, handled[0])
	}
	allowed = merge_formatted(streak, allowed)
	return fmt.Sprintf("%s (%s)", l.GetName(), strings.Join(allowed, "|"))
}

func streak_or_merge(streak []byte, allowed []string, char byte) ([]byte, []string) {
	if len(streak) <= 0 {
		return append(streak, char), allowed
	}
	last := streak[len(streak)-1]
	if char == last+1 {
		return append(streak, char), allowed
	} else {
		return []byte{char}, merge_formatted(streak, allowed)
	}
}

func merge_formatted(streak []byte, allowed []string) []string {
	if len(streak) <= 0 {
		return allowed
	}
	if len(streak) <= 1 {
		return append(allowed, fmt.Sprintf("'%s'", string(streak[0])))
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
