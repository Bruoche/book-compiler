package our

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Bruoche/book-compiler/pkg/parser"
)

type VarDef struct {
	VarSignature
	val ValDeclaration
	T   parser.Type
}

var _ AssignationReciever = &VarDef{}

func (v *VarDef) Declare(s *parser.Stack) (declare []string, assign []string, err error) {
	signature, err := v.val.Signature()
	if err != nil {
		return declare, assign, err
	}
	declare, assign, vars, err := v.T.DefineVar(v.GetName(), signature)
	for _, v := range vars {
		s.AddVar(v)
	}
	return declare, assign, err
}
func (v VarDef) GetName() string {
	return v.Name
}
func (v *VarDef) Assign(new ValDeclaration) {
	v.val = new
}
func (v VarDef) GetVal() ValDeclaration {
	return v.val
}
func (v *VarDef) SetType(T parser.Type) {
	v.T = T
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
func (T SimpleType) Assign(s *parser.Stack, signature string, value parser.Val) ([]string, error) {
	val, err := T.Recieve(value)
	return []string{parser.Line("%s = %s", signature, val)}, err
}
func (T SimpleType) Recieve(value parser.Val) (string, error) {
	if value.Type().Signature() == T.Signature() {
		val, err := value.Signature()
		if err != nil {
			return "", err
		}
		return val, nil
	}
	return "", fmt.Errorf("Type value of type %s cannot be assigned to variable of type %s", value.Type().GetName(), T.GetName())
}

type Ptr[T parser.Type] struct {
	ref T
}

func Ref[T parser.Type](t T) Ptr[T] {
	return Ptr[T]{t}
}

var _ parser.Type = &Ptr[parser.Type]{}

func (ptr Ptr[T]) GetName() string {
	return "Pointer to " + ptr.ref.GetName()
}
func (ptr Ptr[T]) Signature() string {
	return ptr.ref.Signature() + "*"
}
func (ptr Ptr[T]) DefineVar(name string, val string) (declare []string, assign []string, result []parser.Var, err error) {
	return declare, assign, result, fmt.Errorf("Cannot define a Ptr var through this generic type, do it explicitly.")
}
func (ptr Ptr[T]) CompileVal(val string) (string, error) {
	return "*" + ptr.ref.Signature(), nil
}
func (ptr Ptr[T]) Assign(s *parser.Stack, signature string, v parser.Val) ([]string, error) {
	return []string{}, fmt.Errorf("Cannot assign to a pointer through this generic type, do it explicitly.")
}
func (ptr Ptr[T]) Recieve(parser.Val) (string, error) {
	return "", fmt.Errorf("Cannot assign to a pointer through this generic type, do it explicitly.")
}

type Type struct {
	SimpleType
	declaration func(name string, val string) (declare []string, assign []string, alloc bool, err error)
	assigation  func(name string, val parser.Val) ([]string, error)
	reception   func(val parser.Val) (string, error)
}

var _ parser.Type = Type{}

func SimpleAssign(recieve func(parser.Val) (string, error)) func(signature string, val parser.Val) ([]string, error) {
	return func(signature string, val parser.Val) ([]string, error) {
		recieved, err := recieve(val)
		return []string{fmt.Sprintf("%s = %s", signature, recieved)}, err
	}
}

func Object(
	name string,
	signature string,
	valCompilation func(string) (string, error),
	declaration func(signature string, val string) (declare []string, assign []string, alloc bool, err error),
	assignation func(signature string, val parser.Val) ([]string, error),
	reception func(val parser.Val) (string, error),
) parser.Type {
	return Type{
		SimpleType:  Primitive(name, signature, valCompilation),
		declaration: declaration,
		assigation:  assignation,
		reception:   reception,
	}
}

func (T Type) DefineVar(name string, val string) (declare []string, assign []string, result []parser.Var, err error) {
	v := New(name, T, false)
	declare, assign, allocated, err := T.declaration(v.Signature(), val)
	v.allocated = allocated
	return declare, assign, []parser.Var{v}, err
}
func (T Type) Assign(s *parser.Stack, signature string, val parser.Val) ([]string, error) {
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

const strName = "String"

var String = Object(
	strName,
	"char *",
	func(val string) (string, error) {
		return val, nil
	},
	func(signature, val string) ([]string, []string, bool, error) {
		declare := []string{
			parser.Line("char *%s", signature),
			parser.Line("int len_%s = %d", signature, len(val)),
		}
		assign := []string{
			parser.Line("%s = strdup(\"%s\")", signature, val),
			fmt.Sprintf("if (%s == NULL) {", signature),
			fmt.Sprintf("  fprintf(stderr, \"ERROR OOM: Out of memory, string \\\"%s\\\" could not be duplicated into %s.\\n\"); ", val, signature),
			"  exit(1);",
			"}\n",
		}
		return declare, assign, true, nil
	},
	func(name string, val parser.Val) ([]string, error) {
		signature, err := val.Signature()
		if err != nil {
			return []string{}, err
		}
		switch val.Type().GetName() {
		case strName:
			return []string{
				parser.Line("f__realloc(&%s, &len_%s, %s, len_%s)", name, name, signature, signature),
			}, nil
		case RawString.GetName():
			return []string{
				parser.Line("f__realloc(&%s, &len_%s, \"%s\", %d)", name, name, signature, len(signature)),
			}, nil
		}
		return []string{}, fmt.Errorf("Cannot assign value of type %s to %s", val.Type().GetName(), strName)
	},
	func(val parser.Val) (string, error) {
		switch val.Type().GetName() {
		case strName:
			return val.Signature()
		case RawString.GetName():
			signature, err := val.Signature()
			return fmt.Sprintf("strdup(%s)", signature), err
		default:
			return "", fmt.Errorf("Type %s is not assignable to string", val)
		}
	},
)

const rawStrName = "Raw String"

var RawString = Object(
	"Raw String",
	"char *",
	func(val string) (string, error) {
		return fmt.Sprintf("\"%s\"", val), nil
	},
	func(name, val string) ([]string, []string, bool, error) {
		return []string{}, []string{}, false, fmt.Errorf("Cannot initialize a raw string variable (use normal String type).")
	},
	func(name string, val parser.Val) ([]string, error) {
		return []string{}, fmt.Errorf("Cannot assign a value to a raw string value.")
	},
	func(val parser.Val) (string, error) {
		switch val.Type().GetName() {
		case strName, rawStrName:
			return val.Signature()
		default:
			return "", fmt.Errorf("Cannot assign type %s to %s", val.Type().GetName(), rawStrName)
		}
	},
)

type RawInst string

var _ parser.Compilable = RawInst("")

func BuiltIn(instructions ...string) []parser.Blamable {
	inst := []parser.Blamable{}
	for _, raw := range instructions {
		inst = append(inst, parser.Blame(
			"(builtin)",
			0,
			RawInst(raw),
		))
	}
	return inst
}

func (raw RawInst) Declare(*parser.Stack) (declared []string, main []string, err error) {
	return declared, []string{string(raw)}, nil
}

type Function struct {
	name         string
	returnType   parser.Type
	args         []parser.Signature
	instructions []parser.Blamable
}

var _ parser.Compilable = Function{}
var _ parser.Func = Function{}

func Func(name string, ret parser.Type, args []parser.Signature, instructions []parser.Blamable) Function {
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
	declared = append(declared, fmt.Sprintf("%s %s (%s) {\n", ret, f.Signature(), strings.Join(args, ", ")))
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
		VarSignature{"_str1", Ref(String)},
		VarSignature{"_len1", Ref(Int)},
		VarSignature{"_str2", String},
		VarSignature{"_len2", Int},
	},
	BuiltIn(
		"if (*v__len1 != v__len2) {\n",
		"  char *_tmp_realloc = (char *) realloc(*v__str1, v__len2 * sizeof(char) + 1);\n",
		"  if (_tmp_realloc == NULL) { ",
		"    fprintf(stderr, \"ERROR OOM: Out of memory, string could not be resized.\\n\"); ",
		"    exit(1); ",
		"  }\n",
		"  *v__str1 = _tmp_realloc;\n",
		"  *v__len1 = v__len2;\n",
		"}\n",
		"memcpy(*v__str1, v__str2, v__len2+1);\n",
	),
)

var Print = Func(
	"print",
	nil,
	[]parser.Signature{
		VarSignature{"_str", String},
	},
	BuiltIn(
		"printf(\"%s\\n\", v__str);\n",
	),
)

var Put = Func(
	"put",
	nil,
	[]parser.Signature{
		VarSignature{"_str", String},
	},
	BuiltIn(
		"printf(\"%s\", v__str);\n",
	),
)

type Arg interface {
	Add(byte)
	ToVal(*parser.Stack) (parser.Val, error)
}

type varVal struct {
	name string
}

var _ Arg = &varVal{}

func (arg varVal) GetName() string {
	return arg.name
}
func (arg *varVal) Add(char byte) {
	arg.name += string(char)
}
func (arg varVal) ToVal(s *parser.Stack) (parser.Val, error) {
	v, ok := get(arg, s.GetVars())
	if !ok {
		return VarVal{}, fmt.Errorf("Cannot assign value of unknown variable \"%s\".", arg.name)
	}
	return VarVal{rawVal{v.Signature(), v.Type()}}, nil
}

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
		fmt.Printf("WARNING: Value returned by \"%s\" is unused.", f.GetName())
	}
	args := []string{}
	for i, arg := range call.args {
		val, err := arg.ToVal(s)
		if err != nil {
			return declared, assign, fmt.Errorf("Error getting value from argument n°%d: %s", i, err)
		}
		if i >= len(f.Args()) {
			return declared, assign, fmt.Errorf("Too many arguments given to %s: Expected %d, got %d", f.GetName(), len(f.Args()), len(call.args))
		}
		assigned, err := f.Args()[i].Type().Recieve(val)
		if err != nil {
			return declared, assign, err
		}
		args = append(args, assigned)
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

type ValDeclaration interface {
	parser.Val
	Add(byte)
}

type AssignationReciever interface {
	parser.Compilable
	Assign(ValDeclaration)
	GetVal() ValDeclaration
	SetType(parser.Type)
}

type Assignation struct {
	targetName string
	val        ValDeclaration
	T          parser.Type
}

var _ AssignationReciever = &Assignation{}

func (a Assignation) Declare(s *parser.Stack) (declare []string, main []string, err error) {
	target, ok := VarIn(s.GetVars(), a.targetName)
	if !ok {
		return declare, main, fmt.Errorf("Variable \"%s\" is undefined. Cannot assign a value to it. Please use \":=\" for implicit declaration.", a.targetName)
	}
	assignation, err := target.Type().Assign(s, target.Signature(), a.GetVal())
	if err != nil {
		signature, _ := a.val.Signature()
		return declare, main, fmt.Errorf("Cannot assign value \"%s\" of type %s to variable %s of type %s: %s", signature, a.val.Type().GetName(), target.GetName(), target.Type().GetName(), err)
	}
	return declare, assignation, err
}
func (a *Assignation) Assign(new ValDeclaration) {
	a.val = new
}
func (a Assignation) GetVal() ValDeclaration {
	return a.val
}
func (a *Assignation) SetType(T parser.Type) {
	a.T = T
}

type rawVal struct {
	def string
	T   parser.Type
}

var _ Arg = &rawVal{}
var _ parser.Val = &rawVal{}

func (v rawVal) Signature() (string, error) {
	return v.def, nil
}
func (v rawVal) Type() parser.Type {
	return v.T
}
func (v *rawVal) Add(char byte) {
	v.def += string(char)
}
func (v *rawVal) ToVal(s *parser.Stack) (parser.Val, error) {
	return v, nil
}

type VarVal struct {
	rawVal
}

func (v VarVal) Signature() (string, error) {
	return v.def, nil
}

type Empty struct{}
