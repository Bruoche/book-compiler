package our

import (
	"bufio"
	"errors"
	"fmt"

	"github.com/Bruoche/book-compiler/pkg/parser"
)

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
		compile(w, indent, compiled)
	}
	compile(w, indent, "\nint main() {\n")
	indented(w, indent, func(write func(string), indent int) {
		for _, compiled := range main {
			write(compiled)
		}
	})
	compile(w, indent, "}\n")
	return errors.Join(errs...)
}

func play(ir *intermediate) (globals []string, main []string, err error) {
	return parser.Compile(ir.Parsed(), &parser.Stack{BaseState()})
}

type state struct {
	types []parser.Type
	vars  []parser.Var
	funcs []parser.Func
}

var _ parser.State = &state{}

func (s state) GetTypes() []parser.Type {
	return s.types
}
func (s state) GetVars() []parser.Var {
	return s.vars
}
func (s state) GetFuncs() []parser.Func {
	return s.funcs
}
func (s *state) NewType(T parser.Type) {
	s.types = append(s.types, T)
}
func (s *state) NewVar(v parser.Var) {
	s.vars = append(s.vars, v)
}
func (s *state) NewFunc(f parser.Func) {
	s.funcs = append(s.funcs, f)
}
func (s state) VetType(T parser.Type) error {
	if isDefined(T, s.types) {
		return fmt.Errorf("A type named \"%s\" already exist. Cannot define two types of the same name.", T.GetName())
	}
	return nil
}
func (s state) VetVar(v parser.Var) error {
	if isDefined(v, s.GetVars()) {
		return fmt.Errorf("Variable \"%s\" is already defined.", v.GetName())
	}
	return nil
}
func (s state) VetFunc(f parser.Func) error {
	if isDefined(f, s.funcs) {
		return fmt.Errorf("A function named \"%s\" already exists.", f.GetName())
	}
	return nil
}
func (s *state) Remove(v parser.Var) error {
	for i, current := range s.vars {
		if v.GetName() == current.GetName() {
			s.vars = append(s.vars[:i], s.vars[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("The variable %s doesn't exist. It cannot be removed from state.", v.GetName())
}

func BaseState() *state {
	return &state{
		types: []parser.Type{
			Int,
			String,
		},
	}
}

func VarIn(vars []parser.Var, name string) (parser.Var, bool) {
	for _, v := range vars {
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
