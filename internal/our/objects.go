package our

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/Bruoche/book-compiler/pkg/parser"
)

func Line(format string, vals ...any) string {
	return fmt.Sprintf(format+";\n", vals...)
}

type Context struct {
	types        []Type
	variables    []Declarable
	instructions []Instruction
}

var _ parser.Context = &Context{}

func (c *Context) Add(obj any) error {
	switch o := obj.(type) {
	case Declarable:
		c.variables = append(c.variables, o)
	case Instruction:
		c.instructions = append(c.instructions, o)
	default:
		return fmt.Errorf("Unhandled object type %T recieved during compilation.", obj)
	}
	return nil
}

const TabSize = 2

func (c Context) Compile(w *bufio.Writer) error {
	errs := []error{}
	indent := 0
	compile(w, indent, "int main() {")
	err := indented(w, indent, func(write func(string), indent int) error {
		for _, declaration := range c.variables {
			v := declaration.Declared()
			if !isDefined(v.T, c.types) {
				errs = append(errs, fmt.Errorf("Unkown type %s used in assignation of variable %s", v.T.Name, v.Name))
			}
			if isDefined(v, c.variables) {
				errs = append(errs, fmt.Errorf("Variable \"%s\" is already defined.", v.Name))
			}
			write(declaration.Declare())
		}
		return errors.Join(errs...)
	})
	compile(w, indent, "}")
	return err
}

func indented(w *bufio.Writer, current_indent int, compilation func(write func(string), indent int) error) error {
	indent := current_indent + 1
	writeFunc := func(s string) {
		compile(w, indent, s)
	}
	return compilation(writeFunc, indent)
}

func compile(w *bufio.Writer, depth int, line string) {
	w.WriteString(Line(
		"%s%s",
		strings.Repeat(" ", (depth*TabSize)),
		line,
	))
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
	Compile   func(val string) string
}

func (T Type) GetName() string {
	return T.Name
}

type Declarable interface {
	Declare() string
	Declared() Var
	GetName() string
}

type Var struct {
	Name string
	T    Type
}

func (v Var) GetName() string {
	return v.Name
}

func (v Var) Declare() string {
	return Line("%s %s", v.T.Signature, v.T.Name)
}

type Instruction interface {
	Vet(Context) error
	Compile() string
}

type Val struct {
	def string
	T   Type
}

type Empty struct{}
