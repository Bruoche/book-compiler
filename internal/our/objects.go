package our

import (
	"fmt"
	"strconv"

	"github.com/Bruoche/book-compiler/pkg/parser"
)

type VarDef struct {
	VarSignature
	val Val
}

var _ parser.Compilable = &VarDef{}
var _ AssignationReciever = &VarDef{}

func (v *VarDef) Declare(parser.State) (declare []string, assign []string, vars []parser.Var, types []parser.Type, err error) {
	declare, assign, vars, err = v.val.T.DefineVar(v.GetName(), v.val.def)
	return declare, assign, vars, []parser.Type{}, err
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

func isDefined[T parser.Namable](o parser.Namable, defined []T) bool {
	for _, def := range defined {
		if def.GetName() == o.GetName() {
			return true
		}
	}
	return false
}

type SimpleType struct {
	name           string
	signature      string
	valCompilation func(string) (string, error)
}

var _ parser.Type = SimpleType{}

func Primitive(name, signature string, valCompilation func(string) (string, error)) SimpleType {
	return SimpleType{
		name:           name,
		signature:      signature,
		valCompilation: valCompilation,
	}
}

func (T SimpleType) GetName() string {
	return T.name
}
func (T SimpleType) Signature() string {
	return T.signature
}
func (T SimpleType) SatisfiedBy(other parser.Type) bool {
	return T.GetName() == other.GetName()
}
func (T SimpleType) DefineVar(name string, val string) (declare []string, assign []string, result []parser.Var, err error) {
	v := New(name, T, false)
	return []string{
			fmt.Sprintf("%s %s = %s", T.signature, v.Signature(), val),
		},
		[]string{},
		[]parser.Var{v},
		nil
}
func (T SimpleType) CompileVal(val string) (string, error) {
	return T.valCompilation(val)
}
func (T SimpleType) CompileAssign(signature string, val string) ([]string, error) {
	return []string{fmt.Sprintf("%s = %s", signature, val)}, nil
}

type Type struct {
	SimpleType
	declaration func(name string, val string) (declare []string, assign []string, alloc bool, err error)
	assigation  func(name string, val string) ([]string, error)
}

var _ parser.Type = Type{}

func Object(
	name string,
	signature string,
	valCompilation func(string) (string, error),
	declaration func(signature string, val string) (declare []string, assign []string, alloc bool, err error),
	assignation func(signature string, val string) ([]string, error),
) parser.Type {
	return Type{
		SimpleType:  Primitive(name, signature, valCompilation),
		declaration: declaration,
		assigation:  assignation,
	}
}

func (T Type) DefineVar(name string, val string) (declare []string, assign []string, result []parser.Var, err error) {
	v := New(name, T, false)
	declare, assign, allocated, err := T.declaration(v.Signature(), val)
	v.allocated = allocated
	return declare, assign, []parser.Var{v}, err
}
func (T Type) CompileAssign(signature string, val string) ([]string, error) {
	return T.assigation(signature, val)
}

var Int = Primitive(
	"Int",
	"int",
	func(val string) (string, error) {
		n, err := strconv.Atoi(val)
		return strconv.Itoa(n), err
	},
)

var String = Object(
	"String",
	"str",
	func(val string) (string, error) {
		return fmt.Sprintf("\"%s\"", val), nil
	},
	func(name, val string) ([]string, []string, bool, error) {
		declare := []string{}
		assign := []string{fmt.Sprintf("%s = (char *) malloc(%d * sizeof(char))", name, len(val))}
		for i, char := range val {
			assign = append(assign, fmt.Sprintf("%s[%d] = '%s'", name, i, string(char)))
		}
		return declare, assign, true, nil
	},
	func(name, val string) ([]string, error) {
		return []string{}, nil
	},
)

type Function struct {
	Name         string
	Return       SimpleType
	Args         []VarSignature
	Instructions []Instructable
}

type VarSignature struct {
	Name string
	T    parser.Type
}

func (v VarSignature) GetName() string {
	return v.Name
}
func (v VarSignature) Signature() string {
	return "v_" + v.Name
}

type Var struct {
	VarSignature
	allocated bool
}

var _ parser.Var = Var{}

func New(name string, T parser.Type, allocated bool) Var {
	return Var{
		Name:      name,
		T:         T,
		allocated: allocated,
	}
}

func (v Var) Allocated() bool {
	return v.allocated
}
func (v Var) Type() parser.Type {
	return v.T
}

type Assignable interface {
	Assigned() string
}

type AssignationReciever interface {
	parser.Compilable
	Assign(Val)
	GetVal() Val
}

type Instruction struct {
	action Instructable
}

var _ parser.Compilable = Instruction{}

func (i Instruction) Declare(state parser.State) (declare []string, execute []string, vars []parser.Var, types []parser.Type, err error) {
	execute, err = i.action.Compile(state)
	return
}

type Instructable interface {
	Compile(parser.State) ([]string, error)
}

type Assignation struct {
	targetName string
	val        Val
}

var _ AssignationReciever = &Assignation{}

func (a Assignation) Declare(s parser.State) (declare []string, main []string, result []parser.Var, types []parser.Type, err error) {
	target, ok := VarIn(s, a.targetName)
	if !ok {
		return declare, main, result, types, fmt.Errorf("Variable \"%s\" is undefined. Cannot assign a value to it. Please use \":=\" for implicit declaration.", a.targetName)
	}
	if !target.Type().SatisfiedBy(a.val.T) {
		return declare, main, result, types, fmt.Errorf("Cannot assign value \"%s\" of type %s to variable %s of type %s", a.val.def, a.val.T.GetName(), target.GetName(), target.Type().GetName())
	}
	main, err = a.val.T.CompileAssign(target.Signature(), a.val.def)
	return []string{}, main, []parser.Var{}, []parser.Type{}, err
}

func (a *Assignation) Assign(new Val) {
	a.val = new
}

func (a Assignation) GetVal() Val {
	return a.val
}

type Val struct {
	def string
	T   parser.Type
}

func (v Val) Compile() (string, error) {
	return v.T.CompileVal(v.def)
}

type Empty struct{}
