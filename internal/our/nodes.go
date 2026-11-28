package our

import (
	"github.com/Bruoche/book-compiler/pkg/chars"
	"github.com/Bruoche/book-compiler/pkg/parser"
)

var Root = parser.Make(
	"root",
	parser.Pass[Empty](),
)

var ProcedureCall = parser.Make(
	"call_func",
	parser.Pass[Empty](),
)

var startCallFunc = parser.Make(
	"call_func",
	func(_ Empty, char byte) (ProcCall, parser.Compilable) {
		return ProcCall{
			args:    []Arg{},
			name:    string(char),
			valUsed: false,
		}, nil
	},
)

var callFunc = parser.Make(
	"call_func",
	func(f ProcCall, char byte) (ProcCall, parser.Compilable) {
		f.name += string(char)
		return f, nil
	},
)

var readyCallArg = parser.Make(
	"call_func_arg",
	parser.Pass[ProcCall](),
)

var wantCallArg = parser.Make(
	"call_func_new_arg",
	parser.Pass[ProcCall](),
)

var addStrArg = parser.Make(
	"call_func_arg_str",
	func(f ProcCall, char byte) (ProcCall, parser.Compilable) {
		f.args = append(f.args, &rawVal{
			def: "",
			T:   RawString,
		})
		return f, nil
	},
)

var addVarArg = parser.Make(
	"call_func_arg",
	func(f ProcCall, char byte) (ProcCall, parser.Compilable) {
		f.args = append(f.args, &varVal{
			string(char),
		})
		return f, nil
	},
)

func addCharToArg(f ProcCall, char byte) (ProcCall, parser.Compilable) {
	last := len(f.args) - 1
	f.args[last].Add(char)
	return f, nil
}

var writeArg = parser.Make(
	"call_func_arg",
	addCharToArg,
)

var writeStrArg = parser.Make(
	"call_func_arg_str",
	addCharToArg,
)

var endStrArg = parser.Make(
	"call_func_arg",
	parser.Pass[ProcCall](),
)

var endCall = parser.Make(
	"end_func_call",
	func(f ProcCall, char byte) (Empty, parser.Compilable) {
		return Empty{}, f
	},
)

type Subject struct {
	name string
}

var startInstruction = parser.Make(
	"start_instruction",
	func(_ Empty, char byte) (Subject, parser.Compilable) {
		return Subject{
			name: string(char),
		}, nil
	},
)
var subject = parser.Make(
	"start_instruction",
	func(s Subject, char byte) (Subject, parser.Compilable) {
		s.name = s.name + string(char)
		return s, nil
	},
)
var subjectDone = parser.Make(
	"action",
	parser.Pass[Subject](),
)
var startDef = parser.Make(
	"define_var",
	parser.Pass[Subject](),
)
var startDefinedVar = parser.Make(
	"define_with_val",
	func(s Subject, char byte) (AssignationReciever, parser.Compilable) {
		return &VarDef{
			Name: s.name,
		}, nil
	},
)
var assignVal = parser.Make(
	"assign_val",
	func(s Subject, char byte) (AssignationReciever, parser.Compilable) {
		return &Assignation{
			targetName: s.name,
		}, nil
	},
)

func DefValStart(initialize func(char byte) (ValDeclaration, parser.Type)) func(AssignationReciever, byte) (AssignationReciever, parser.Compilable) {
	return func(a AssignationReciever, char byte) (AssignationReciever, parser.Compilable) {
		val, T := initialize(char)
		a.Assign(val)
		a.SetType(T)
		return a, nil
	}
}

var DefValAppend = func(a AssignationReciever, char byte) (AssignationReciever, parser.Compilable) {
	v := a.GetVal()
	v.Add(char)
	a.Assign(v)
	return a, nil
}

var DefinedVarIntStart = parser.Make(
	"define_int_val",
	DefValStart(func(char byte) (ValDeclaration, parser.Type) {
		return &rawVal{
			def: string(char),
			T:   Int,
		}, Int
	}),
)
var DefinedVarInt = parser.Make(
	"define_int_val",
	DefValAppend,
)
var DefinedVarStrStart = parser.Make(
	"define_str_val",
	DefValStart(func(char byte) (ValDeclaration, parser.Type) {
		return &rawVal{
			def: "",
			T:   RawString,
		}, String
	}),
)
var DefinedVarStr = parser.Make(
	"define_str_val",
	DefValAppend,
)
var DefinedVarReady = parser.Make(
	"defined_str_val",
	parser.Pass[AssignationReciever](),
)
var DefinedVarDone = parser.Make(
	"end",
	func(a AssignationReciever, char byte) (Empty, parser.Compilable) {
		return Empty{}, a
	},
)
var root = parser.Ignoring(' ').Has(
	parser.Bind(&ProcedureCall, '-'),
	parser.Bind(&startInstruction, chars.Uncapped...),
	parser.Bind(&Root, '\n'),
)

func Init() {
	Root.Has(
		root...,
	)
	instructions()
}
func instructions() {
	funcCall()
	subjects()
	vars()
	vals()
}
func funcCall() {
	ProcedureCall.Has(parser.Ignoring(' ').Has(
		parser.Bind(&startCallFunc, chars.Uncapped...),
	)...)
	startCallFunc.Has(
		parser.Bind(&callFunc, chars.Letters...),
	)
	callFunc.Has(
		parser.Bind(&endCall, '\n'),
		parser.Bind(&readyCallArg, ' '),
		parser.Bind(&callFunc, chars.Letters...),
	)
	newArg := parser.Prep(
		parser.Bind(&addStrArg, '"'),
		parser.Bind(&addVarArg, chars.Uncapped...),
	)
	readyCallArg.Has(parser.Ignoring(' ').Has(
		append([]parser.Linkable[ProcCall]{parser.Bind(&endCall, '\n')},
			newArg...)...,
	)...)
	argCall := parser.Prep(
		parser.Bind(&writeArg, chars.Letters...),
		parser.Bind(&wantCallArg, ','),
		parser.Bind(&endCall, '\n'),
	)
	addVarArg.Has(
		argCall...,
	)
	writeArg.Has(
		argCall...,
	)
	argStr := parser.Prep(
		parser.Bind(&endStrArg, '"'),
		parser.Bind(&writeStrArg, chars.All...),
	)
	addStrArg.Has(
		argStr...,
	)
	writeStrArg.Has(
		argStr...,
	)
	endStrArg.Has(
		parser.Bind(&wantCallArg, ','),
		parser.Bind(&endCall, '\n'),
	)
	wantCallArg.Has(parser.Ignoring(' ', '\n').Has(
		newArg...,
	)...)
	endCall.Has(
		root...,
	)
}
func subjects() {
	subjectAction := parser.Prep(
		parser.Bind(&startDef, ':'),
		parser.Bind(&assignVal, '='),
	)
	subjectDef := parser.Prep(append(
		subjectAction,
		parser.Bind(&subject, chars.Letters...),
		parser.Bind(&subjectDone, ' '),
	)...)
	startInstruction.Has(
		subjectDef...,
	)
	subject.Has(
		subjectDef...,
	)
	subjectDone.Has(
		parser.Ignoring(' ').Has(
			subjectAction...,
		)...,
	)
}
func vars() {
	startDef.Has(
		parser.Ignoring(' ').Has(
			parser.Bind(&startDefinedVar, '='),
		)...,
	)
	assign := parser.Prep(
		parser.Bind(&DefinedVarIntStart, chars.Nums...),
		parser.Bind(&DefinedVarStrStart, '"'),
	)
	startDefinedVar.Has(
		parser.Ignoring(' ').Has(
			assign...,
		)...,
	)
	assignVal.Has(
		parser.Ignoring(' ').Has(
			assign...,
		)...,
	)

}
func vals() {
	val(&DefinedVarIntStart, &DefinedVarInt,
		parser.Bind(&DefinedVarInt, chars.Nums...),
		parser.Bind(&DefinedVarDone, '\n'),
	)
	val(&DefinedVarStrStart, &DefinedVarStr,
		parser.Bind(&DefinedVarReady, '"'),
		parser.Bind(&DefinedVarStr, chars.All...),
	)
	DefinedVarReady.Has(
		parser.Bind(&DefinedVarDone, '\n'),
	)
	DefinedVarDone.Has(
		root...,
	)
}

func val(start *parser.Node[AssignationReciever, AssignationReciever], assign *parser.Node[AssignationReciever, AssignationReciever], links ...parser.Linkable[AssignationReciever]) {
	start.Has(
		links...,
	)
	assign.Has(
		links...,
	)
}
