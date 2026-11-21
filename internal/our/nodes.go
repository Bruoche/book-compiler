package our

import (
	"github.com/Bruoche/book-compiler/pkg/chars"
	"github.com/Bruoche/book-compiler/pkg/parser"
)

var Root = parser.Make(
	"root",
	parser.Pass[Empty](),
)

type Subject struct {
	name string
}

var startInstruction = parser.Make(
	"start_instruction",
	func(_ Empty, char byte) (Subject, any) {
		return Subject{
			name: string(char),
		}, nil
	},
)
var subject = parser.Make(
	"start_instruction",
	func(s Subject, char byte) (Subject, any) {
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
	func(s Subject, char byte) (Assignable, any) {
		return &VarDef{
			Name: s.name,
			val:  Val{},
		}, nil
	},
)
var assignVal = parser.Make(
	"assign_val",
	func(s Subject, char byte) (Assignable, any) {
		return &Assignation{
			targetName: s.name,
			val:        Val{},
		}, nil
	},
)

func DefValStart(val func(char byte) Val) func(Assignable, byte) (Assignable, any) {
	return func(a Assignable, char byte) (Assignable, any) {
		a.Assign(val(char))
		return a, nil
	}
}

var DefValAppend = func(a Assignable, char byte) (Assignable, any) {
	v := a.GetVal()
	v.def += string(char)
	a.Assign(v)
	return a, nil
}

var DefinedVarIntStart = parser.Make(
	"define_int_val",
	DefValStart(func(char byte) Val {
		return Val{
			def: string(char),
			T:   Int,
		}
	}),
)
var DefinedVarInt = parser.Make(
	"define_int_val",
	DefValAppend,
)
var DefinedVarStrStart = parser.Make(
	"define_str_val",
	DefValStart(func(char byte) Val {
		return Val{
			def: "",
			T:   String,
		}
	}),
)
var DefinedVarStr = parser.Make(
	"define_str_val",
	DefValAppend,
)
var DefinedVarReady = parser.Make(
	"defined_str_val",
	parser.Pass[Assignable](),
)
var DefinedVarDone = parser.Make(
	"end",
	func(a Assignable, char byte) (Empty, any) {
		return Empty{}, a
	},
)
var root = parser.Ignoring(' ').Has(
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
	subjects()
	vars()
	vals()
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

func val(start *parser.Node[Assignable, Assignable], assign *parser.Node[Assignable, Assignable], links ...parser.Linkable[Assignable]) {
	start.Has(
		links...,
	)
	assign.Has(
		links...,
	)
}
