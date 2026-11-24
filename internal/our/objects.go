package our

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Bruoche/book-compiler/pkg/parser"
)

type VarDef struct {
	VarSignature
	val Val
}

var _ AssignationReciever = &VarDef{}

func (v *VarDef) Declare(s *parser.Stack) (declare []string, assign []string, err error) {
	declare, assign, vars, err := v.val.T.DefineVar(v.GetName(), v.val.def)
	for _, v := range vars {
		s.AddVar(v)
	}
	return declare, assign, err
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
	_, found := get(o, defined)
	return found
}

func get[T parser.Namable](o parser.Namable, defined []T) (T, bool) {
	for _, def := range defined {
		if def.GetName() == o.GetName() {
			return def, true
		}
	}
	var none T
	return none, false
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
			parser.Line("%s %s = %s", T.signature, v.Signature(), val),
		},
		[]string{},
		[]parser.Var{v},
		nil
}
func (T SimpleType) CompileVal(val string) (string, error) {
	return T.valCompilation(val)
}
func (T SimpleType) CompileAssign(signature string, val string) ([]string, error) {
	return []string{parser.Line("%s = %s", signature, val)}, nil
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
	"char *",
	func(val string) (string, error) {
		return fmt.Sprintf("\"%s\"", val), nil
	},
	func(name, val string) ([]string, []string, bool, error) {
		declare := []string{
			parser.Line("char *%s", name),
			parser.Line("int len_%s", name),
		}
		assign := []string{
			parser.Line("%s = strdup(\"%s\")", name, val),
		}
		return declare, assign, true, nil
	},
	func(name, val string) ([]string, error) {
		return []string{
			parser.Line("f__realloc(%s, len_%s, \"%s\", %d)", name, name, val, len(val)),
			parser.Line("len_%s = %d", name, len(val)),
		}, nil
	},
)

type RawInst []string

var _ parser.Compilable = RawInst{}

func (raw RawInst) Declare(*parser.Stack) (declared []string, main []string, err error) {
	return declared, raw, nil
}

type Function struct {
	name         string
	returnType   parser.Type
	args         []parser.Signature
	instructions []parser.Compilable
}

var _ parser.Compilable = Function{}
var _ parser.Func = Function{}

func Func(name string, ret parser.Type, args []parser.Signature, instructions []parser.Compilable) Function {
	return Function{
		name:         name,
		returnType:   ret,
		args:         args,
		instructions: instructions,
	}
}

func (f Function) GetName() string {
	return f.name
}
func (f Function) Return() parser.Type {
	return f.returnType
}
func (f Function) Args() []parser.Signature {
	return f.args
}

func (f Function) Declare(parent *parser.Stack) (declared []string, assign []string, err error) {
	funcState := state{}
	stack := append(*parent, &funcState)
	top, main, err := parser.Compile(f.instructions, &stack)
	if err != nil {
		return
	}
	args := []string{}
	for _, arg := range f.args {
		args = append(args, fmt.Sprintf("%s %s", arg.Type().Signature(), arg.Signature()))
	}
	ret := "void"
	if f.Return() != nil {
		ret = f.Return().Signature()
	}
	declared = append(declared, fmt.Sprintf("\n%s %s (%s) {\n", ret, f.Signature(), strings.Join(args, ", ")))
	indent := strings.Repeat(" ", TabSize)
	for _, inst := range append(top, main...) {
		declared = append(declared, indent+inst)
	}
	declared = append(declared, "}\n\n")
	return
}

func (f Function) Signature() string {
	return "f_" + f.name
}

var Realloc = Func(
	"_realloc",
	nil,
	[]parser.Signature{
		VarSignature{"_str1", String},
		VarSignature{"_len1", Int},
		VarSignature{"_str2", String},
		VarSignature{"_len2", Int},
	},
	[]parser.Compilable{
		RawInst{
			"if (v__len1 != v__len2) {\n",
			"  v__len1 = v__len2;\n",
			"  char *_tmp_realloc = (char *) realloc(v__str1, v__len1 * sizeof(char) + 1);\n",
			"  if (_tmp_realloc == NULL) {\n",
			"    printf(\"WARNING: Out of memory error, string could not be resized.\\n\");\n",
			"  } else {\n",
			"    v__str1 = _tmp_realloc;\n",
			"  }\n",
			"}\n",
			"strcpy(v__str1, v__str2);\n",
		},
	},
)

var Print = Func(
	"print",
	nil,
	[]parser.Signature{
		VarSignature{"_str", String},
	},
	[]parser.Compilable{
		RawInst{
			"printf(\"%s\\n\", v__str);\n",
		},
	},
)

var Put = Func(
	"put",
	nil,
	[]parser.Signature{
		VarSignature{"_str", String},
	},
	[]parser.Compilable{
		RawInst{
			"printf(\"%s\", v__str);\n",
		},
	},
)

type Arg interface {
	Add(byte)
	ToVal(*parser.Stack) (Val, error)
}

type varVal struct {
	name string
}

func (arg varVal) GetName() string {
	return arg.name
}
func (arg *varVal) Add(char byte) {
	arg.name += string(char)
}
func (arg varVal) ToVal(s *parser.Stack) (Val, error) {
	v, ok := get(arg, s.GetVars())
	if !ok {
		return Val{}, fmt.Errorf("Cannot assign value of unknown variable \"%s\".", arg.name)
	}
	return Val{v.Signature(), v.Type()}, nil
}

var _ Arg = &varVal{}

type ProcCall struct {
	name    string
	args    []Arg
	valUsed bool
}

var _ parser.Compilable = ProcCall{}

func (call ProcCall) Declare(s *parser.Stack) (declared []string, assign []string, err error) {
	f, ok := get(call, s.GetFuncs())
	if !ok {
		return declared, assign, fmt.Errorf("Call to unknown function \"%s\"", call.GetName())
	}
	if f.Return() != nil {
		// TODO warn value unused if pure function
	}
	args := []string{}
	for _, arg := range call.args {
		val, err := arg.ToVal(s)
		if err != nil {
			return declared, assign, fmt.Errorf("Error getting value from argument: %s", err)
		}
		args = append(args, val.def)
	}
	return []string{}, []string{
		parser.Line("%s(%s)", f.Signature(), strings.Join(args, ", ")),
	}, nil
}

func (fc ProcCall) GetName() string {
	return fc.name
}

type VarSignature struct {
	Name string
	T    parser.Type
}

var _ parser.Signature = VarSignature{}

func (v VarSignature) GetName() string {
	return v.Name
}
func (v VarSignature) Signature() string {
	return "v_" + v.Name
}
func (v VarSignature) Type() parser.Type {
	return v.T
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

type Assignation struct {
	targetName string
	val        Val
}

var _ AssignationReciever = &Assignation{}

func (a Assignation) Declare(s *parser.Stack) (declare []string, main []string, err error) {
	target, ok := VarIn(s.GetVars(), a.targetName)
	if !ok {
		return declare, main, fmt.Errorf("Variable \"%s\" is undefined. Cannot assign a value to it. Please use \":=\" for implicit declaration.", a.targetName)
	}
	if !target.Type().SatisfiedBy(a.val.T) {
		return declare, main, fmt.Errorf("Cannot assign value \"%s\" of type %s to variable %s of type %s", a.val.def, a.val.T.GetName(), target.GetName(), target.Type().GetName())
	}
	main, err = a.val.T.CompileAssign(target.Signature(), a.val.def)
	return []string{}, main, err
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
