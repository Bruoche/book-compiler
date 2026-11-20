package our

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"

	"github.com/Bruoche/book-compiler/pkg/parser"
)

func Line(format string, vals ...any) string {
	return fmt.Sprintf(format+";\n", vals...)
}

type context struct {
	instructions []Instruction
}

func Context() *context {
	return &context{}
}

type DefinedVar struct {
	Name string
	val  Val
}

var _ Declarable = DefinedVar{}
var _ Assignable = &DefinedVar{}

func (v DefinedVar) Compile(state) (string, error) {
	return v.Declare()
}

func (v DefinedVar) Declare() (string, error) {
	val, err := v.val.T.Compile(v.val.def)
	return fmt.Sprintf("%s %s = %s", v.val.T.Signature, v.Name, val), err
}

func (v DefinedVar) Declared() DefinedVar {
	return DefinedVar{
		Name: v.Name,
		val:  v.val,
	}
}

func (v DefinedVar) GetName() string {
	return v.Name
}

func (v *DefinedVar) Assign(new Val) {
	v.val = new
}

func (v DefinedVar) GetVal() Val {
	return v.val
}

type state struct {
	types []Type
	vars  []DefinedVar
}

var _ parser.Context = &context{}

func BaseState() state {
	return state{
		types: []Type{
			Int,
		},
	}
}

func (s state) GetVar(name string) (DefinedVar, bool) {
	for _, v := range s.vars {
		if v.Name == name {
			return v, true
		}
	}
	return DefinedVar{}, false
}

func (c *context) Add(obj any) error {
	switch o := obj.(type) {
	case Instruction:
		c.instructions = append(c.instructions, o)
	default:
		return fmt.Errorf("Unhandled object type %T recieved during compilation.", obj)
	}
	return nil
}

const TabSize = 2

func (c context) Compile(w *bufio.Writer) error {
	fmt.Println("Writing to file...")
	errs := []error{}
	globals, main, err := play(c)
	if err != nil {
		return err
	}
	indent := 0
	for _, compiled := range globals {
		compile(w, indent, Line(compiled))
	}
	compile(w, indent, "\nint main() {\n")
	indented(w, indent, func(write func(string), indent int) {
		for _, compiled := range main {
			write(Line(compiled))
		}
	})
	compile(w, indent, "}\n")
	return errors.Join(errs...)
}

func play(c context) (globals []string, main []string, err error) {
	state := BaseState()
	for _, instruction := range c.instructions {
		declaration, ok := instruction.(Declarable)
		if ok {
			v := declaration.Declared()
			if !isDefined(v.val.T, state.types) {
				return globals, main, fmt.Errorf("Unkown type %s used in assignation of variable %s", v.val.T.Name, v.Name)
			}
			if isDefined(v, state.vars) {
				return globals, main, fmt.Errorf("Variable \"%s\" is already defined.", v.Name)
			}
			compiled, err := declaration.Declare()
			if err != nil {
				return globals, main, fmt.Errorf("Error compiling variable declaration for \"%s\" : %s", declaration.GetName(), err)
			}
			state.vars = append(state.vars, v)
			globals = append(globals, compiled)
		} else {
			compiled, err := instruction.Compile(state)
			if err != nil {
				return globals, main, err
			}
			main = append(main, compiled)
		}
	}
	return globals, main, nil
}

func indented(w *bufio.Writer, current_indent int, compilation func(write func(string), indent int)) {
	indent := current_indent + 1
	writeFunc := func(s string) {
		compile(w, indent, s)
	}
	compilation(writeFunc, indent)
}

func compile(w *bufio.Writer, depth int, line string) {
	for range depth * TabSize {
		w.WriteString(" ")
	}
	w.WriteString(line)
}

type Namable interface {
	GetName() string
}

func isDefined[T Namable](o Namable, defined []T) bool {
	for _, def := range defined {
		if def.GetName() == o.GetName() {
			return true
		}
	}
	return false
}

type Type struct {
	Name      string
	Signature string
	Compile   func(val string) (string, error)
}

func (T Type) SatisfiedBy(other Type) bool {
	return T.Name == other.Name
}

func (T Type) GetName() string {
	return T.Name
}

var Int = Type{
	Name:      "Int",
	Signature: "int",
	Compile: func(val string) (string, error) {
		n, err := strconv.Atoi(val)
		return strconv.Itoa(n), err
	},
}

type Declarable interface {
	Instruction
	Declare() (string, error)
	Declared() DefinedVar
	GetName() string
}

type Var struct {
	Name string
	T    Type
}

func (v Var) GetName() string {
	return v.Name
}

type Assignable interface {
	Assign(Val)
	GetVal() Val
}

type Instruction interface {
	Compile(state) (string, error)
}

type Assignation struct {
	targetName string
	val        Val
}

var _ Instruction = Assignation{}
var _ Assignable = &Assignation{}

func (a Assignation) Compile(s state) (string, error) {
	target, ok := s.GetVar(a.targetName)
	if !ok {
		return "", fmt.Errorf("Variable \"%s\" is undefined. Cannot assign a value to it. Please use \":=\" for implicit declaration.", a.targetName)
	}
	if !target.val.T.SatisfiedBy(a.val.T) {
		return "", fmt.Errorf("Cannot assign value \"%s\" of type %s to variable %s of type %s", a.val.def, a.val.T.Name, target.Name, target.val.T.Name)
	}
	compiled, err := a.val.Compile()
	return fmt.Sprintf("%s = %s", target.Name, compiled), err
}

func (a *Assignation) Assign(new Val) {
	a.val = new
}

func (a Assignation) GetVal() Val {
	return a.val
}

type Val struct {
	def string
	T   Type
}

func (v Val) Compile() (string, error) {
	return v.T.Compile(v.def)
}

type Empty struct{}
