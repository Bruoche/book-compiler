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
		return &DefinedVar{
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
var DefinedVarIntStart = parser.Make(
	"define_int_val",
	func(s Assignable, char byte) (Assignable, any) {
		s.Assign(Val{
			def: string(char),
			T:   Int,
		})
		return s, nil
	},
)
var DefinedVarInt = parser.Make(
	"define_int_val",
	func(s Assignable, char byte) (Assignable, any) {
		v := s.GetVal()
		v.def += string(char)
		s.Assign(v)
		return s, nil
	},
)
var DefinedVarDone = parser.Make(
	"end",
	func(s Assignable, char byte) (Empty, any) {
		return Empty{}, s
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
	valInt := parser.Prep(
		parser.Bind(&DefinedVarInt, chars.Nums...),
		parser.Bind(&DefinedVarDone, '\n'),
	)
	DefinedVarIntStart.Has(
		valInt...,
	)
	DefinedVarInt.Has(
		valInt...,
	)
	DefinedVarDone.Has(
		root...,
	)
}
