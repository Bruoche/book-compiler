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

type VarDef struct {
	Name string
	val  Val
}

var _ Declarable = &VarDef{}
var _ Assignable = &VarDef{}

func (v VarDef) Compile(state) ([]string, error) {
	declare, assign, _, err := v.Declare()
	return append(declare, assign...), err
}

func (v *VarDef) Declare() (declare []string, assign []string, alloc bool, err error) {
	return v.val.T.DefineVar(v.Name, v.val)
}

func (v VarDef) Declared() VarDef {
	return VarDef{
		Name: v.Name,
		val:  v.val,
	}
}

func (v VarDef) GetName() string {
	return v.Name
}

func (v *VarDef) Assign(new Val) {
	v.val = new
}

func (v VarDef) GetVal() Val {
	return v.val
}

type DefinedVar struct {
	Name      string
	Val       Val
	Allocated bool
}

var _ Namable = DefinedVar{}

func (v DefinedVar) GetName() string {
	return v.Name
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
			String,
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
	errs := []error{}
	globals, main, err := play(c)
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

func play(c context) (globals []string, main []string, err error) {
	state := BaseState()
	for _, instruction := range c.instructions {
		declaration, ok := instruction.(Declarable)
		if ok {
			v, err := varFrom(state, declaration)
			if err != nil {
				return globals, main, err
			}
			declared, assign, alloc, err := declare(state, declaration)
			if err != nil {
				return globals, main, err
			}
			state.vars = append(state.vars, DefinedVar{
				Name:      v.Name,
				Val:       v.val,
				Allocated: alloc,
			})
			globals = append(globals, declared...)
			main = append(main, assign...)
		} else {
			compiled, err := instruction.Compile(state)
			if err != nil {
				return globals, main, err
			}
			main = append(main, compiled...)
		}
	}
	for _, v := range state.vars {
		if v.Allocated {
			main = append(main, fmt.Sprintf("free(%s)", v.Name))
		}
	}
	return globals, main, nil
}

func varFrom(state state, declaration Declarable) (VarDef, error) {
	v := declaration.Declared()
	if !isDefined(v.val.T, state.types) {
		return v, fmt.Errorf("Unknown type %s used in assignation of variable %s", v.val.T.Name, v.Name)
	}
	if isDefined(v, state.vars) {
		return v, fmt.Errorf("Variable \"%s\" is already defined.", v.Name)
	}
	return v, nil
}

func declare(state state, declaration Declarable) (declared []string, assign []string, alloc bool, err error) {
	declared, assign, alloc, err = declaration.Declare()
	if err != nil {
		return declared, assign, alloc, fmt.Errorf("Error compiling variable declaration for \"%s\" : %s", declaration.GetName(), err)
	}
	return declared, assign, alloc, nil
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
	Name       string
	DefineVar  func(name string, val Val) (declare []string, assign []string, alloc bool, err error)
	CompileVal func(val string) (string, error)
}

func SimpleType(name, signature string, compile func(val string) (string, error)) Type {
	return ComplexType(
		name,
		func(varName, val string) ([]string, []string, bool, error) {
			return []string{fmt.Sprintf("%s %s = %s", signature, varName, val)}, []string{}, false, nil
		},
		compile,
	)
}
func ComplexType(
	name string,
	defineVar func(varName string, compiledVal string) (declare []string, assign []string, needFree bool, err error),
	compileVal func(val string) (string, error),
) Type {
	return Type{
		Name: name,
		DefineVar: func(n string, val Val) ([]string, []string, bool, error) {
			v, err := compileVal(val.def)
			if err != nil {
				return []string{}, []string{}, false, err
			}
			return defineVar(n, v)
		},
		CompileVal: compileVal,
	}
}

func (T Type) SatisfiedBy(other Type) bool {
	return T.Name == other.Name
}
func (T Type) GetName() string {
	return T.Name
}

var Int = SimpleType(
	"Int",
	"int",
	func(val string) (string, error) {
		n, err := strconv.Atoi(val)
		return strconv.Itoa(n), err
	},
)

var String = ComplexType(
	"String",
	func(name, val string) ([]string, []string, bool, error) {
		declare := []string{"char *" + name}
		assign := []string{fmt.Sprintf("%s = (char *) malloc(%d * sizeof(char))", name, len(val))}
		for i, char := range val {
			assign = append(assign, fmt.Sprintf("%s[%d] = '%s'", name, i, string(char)))
		}
		return declare, assign, true, nil
	},
	func(val string) (string, error) {
		return val, nil
	},
)

type Declarable interface {
	Instruction
	Declare() (declared []string, assign []string, alloc bool, err error)
	Declared() VarDef
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
	Compile(state) ([]string, error)
}

type Assignation struct {
	targetName string
	val        Val
}

var _ Instruction = Assignation{}
var _ Assignable = &Assignation{}

func (a Assignation) Compile(s state) ([]string, error) {
	target, ok := s.GetVar(a.targetName)
	if !ok {
		return []string{}, fmt.Errorf("Variable \"%s\" is undefined. Cannot assign a value to it. Please use \":=\" for implicit declaration.", a.targetName)
	}
	if !target.Val.T.SatisfiedBy(a.val.T) {
		return []string{}, fmt.Errorf("Cannot assign value \"%s\" of type %s to variable %s of type %s", a.val.def, a.val.T.Name, target.Name, target.Val.T.Name)
	}
	compiled, err := a.val.Compile()
	return []string{fmt.Sprintf("%s = %s", target.Name, compiled)}, err
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
	return v.T.CompileVal(v.def)
}

type Empty struct{}
