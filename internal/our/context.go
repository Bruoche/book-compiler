package our

import (
	"bufio"
	"errors"
	"fmt"

	"github.com/Bruoche/book-compiler/pkg/parser"
)

func Line(format string, vals ...any) string {
	return fmt.Sprintf(format+";\n", vals...)
}

type intermediate struct {
	parsed []parser.Compilable
}

func IR() *intermediate {
	return &intermediate{}
}

var _ parser.IR = &intermediate{}

func (ir *intermediate) Add(obj parser.Compilable) {
	ir.parsed = append(ir.parsed, obj)
}

const TabSize = 2

func (ir intermediate) Parsed() []parser.Compilable {
	return ir.parsed
}

func (ir *intermediate) Compile(w *bufio.Writer) error {
	errs := []error{}
	globals, main, err := play(ir)
	if err != nil {
		return err
	}
	indent := 0
	compile(w, indent, "#include <stdlib.h>\n\n")
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

func play(ir *intermediate) (globals []string, main []string, err error) {
	return parser.Compile(ir, BaseState())
}

type state struct {
	types []parser.Type
	vars  []parser.Var
}

var _ parser.State = &state{}

func (s state) GetTypes() []parser.Type {
	return s.types
}
func (s state) GetVars() []parser.Var {
	return s.vars
}
func (s *state) NewType(T parser.Type) error {
	err := s.vetType(T)
	if err != nil {
		return err
	}
	s.types = append(s.types, T)
	return nil
}
func (s *state) NewVar(v parser.Var) error {
	err := s.vetVar(v)
	if err != nil {
		return err
	}
	s.vars = append(s.vars, v)
	return nil
}
func (s state) vetType(T parser.Type) error {
	if isDefined(T, s.types) {
		return fmt.Errorf("A type named \"%s\" already exist. Cannot define two types of the same name.", T.GetName())
	}
	return nil
}
func (s state) vetVar(v parser.Var) error {
	if !isDefined(v.Type(), s.GetTypes()) {
		return fmt.Errorf("Unknown type %s used in assignation of variable %s", v.Type().GetName(), v.GetName())
	}
	if isDefined(v, s.GetVars()) {
		return fmt.Errorf("Variable \"%s\" is already defined.", v.GetName())
	}
	return nil
}

func BaseState() *state {
	return &state{
		types: []parser.Type{
			Int,
			String,
		},
	}
}

func VarIn(s parser.State, name string) (parser.Var, bool) {
	for _, v := range s.GetVars() {
		if v.GetName() == name {
			return v, true
		}
	}
	return Var{}, false
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
