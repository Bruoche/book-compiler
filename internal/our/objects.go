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
	types        []Type
	definables   []Declarable
	vars         []Var
	instructions []Instruction
}

var _ parser.Context = &context{}

func BaseContext() *context {
	return &context{
		types: []Type{
			Int,
		},
	}
}

func (c *context) Add(obj any) error {
	fmt.Printf("Added %T\n", obj)
	switch o := obj.(type) {
	case Declarable:
		c.definables = append(c.definables, o)
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
	indent := 0
	compile(w, indent, "int main() {\n")
	err := indented(w, indent, func(write func(string), indent int) error {
		for _, declaration := range c.definables {
			v := declaration.Declared()
			if !isDefined(v.T, c.types) {
				errs = append(errs, fmt.Errorf("Unkown type %s used in assignation of variable %s", v.T.Name, v.Name))
			}
			if isDefined(v, c.vars) {
				errs = append(errs, fmt.Errorf("Variable \"%s\" is already defined.", v.Name))
			}
			compiled, err := declaration.Declare()
			if err != nil {
				errs = append(errs, fmt.Errorf("Error compiling variable declaration for \"%s\" : %s", declaration.GetName(), err))
			} else {
				c.vars = append(c.vars, v)
				write(Line(compiled))
			}
		}
		return errors.Join(errs...)
	})
	compile(w, indent, "}\n")
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
	Declare() (string, error)
	Declared() Var
	GetName() string
}

type Var struct {
	Name string
	T    Type
}

var _ Declarable = Var{}

func (v Var) GetName() string {
	return v.Name
}

func (v Var) Declare() (string, error) {
	return Line("%s %s", v.T.Signature, v.T.Name), nil
}

func (v Var) Declared() Var {
	return v
}

type DefVal struct {
	name string
	val  Val
}

var _ Declarable = DefVal{}

func (v DefVal) Declare() (string, error) {
	val, err := v.val.T.Compile(v.val.def)
	return fmt.Sprintf("%s %s = %s", v.val.T.Signature, v.name, val), err
}

func (v DefVal) Declared() Var {
	return Var{
		Name: v.name,
		T:    v.val.T,
	}
}

func (v DefVal) GetName() string {
	return v.name
}

type Instruction interface {
	Vet(context) error
	Compile() string
}

type Val struct {
	def string
	T   Type
}

type Empty struct{}
