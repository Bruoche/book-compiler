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
var startDefVal = parser.Make(
	"define_with_val",
	func(s Subject, char byte) (DefVal, any) {
		return DefVal{
			name: s.name,
			val:  Val{},
		}, nil
	},
)
var defValIntStart = parser.Make(
	"define_int_val",
	func(s DefVal, char byte) (DefVal, any) {
		s.val = Val{
			def: string(char),
			T:   Int,
		}
		return s, nil
	},
)
var defValInt = parser.Make(
	"define_int_val",
	func(s DefVal, char byte) (DefVal, any) {
		s.val.def += string(char)
		return s, nil
	},
)
var defValDone = parser.Make(
	"end",
	func(s DefVal, char byte) (Empty, any) {
		return Empty{}, s
	},
)

func Init() {
	root := parser.Ignoring(' ').Has(
		parser.Bind(&startInstruction, chars.Uncapped...),
	)
	Root.Has(
		root...,
	)
	wordHandles := parser.Prep(
		parser.Bind(&subject, chars.Letters...),
		parser.Bind(&subjectDone, ' '),
		parser.Bind(&startDef, ':'),
	)
	startInstruction.Has(
		wordHandles...,
	)
	subject.Has(
		wordHandles...,
	)
	subjectDone.Has(
		parser.Ignoring(' ').Has(
			parser.Bind(&startDef, ':'),
		)...,
	)
	startDef.Has(
		parser.Ignoring(' ').Has(
			parser.Bind(&startDefVal, '='),
		)...,
	)
	startDefVal.Has(
		parser.Ignoring(' ').Has(
			parser.Bind(&defValIntStart, chars.Nums...),
		)...,
	)
	valInt := parser.Prep(
		parser.Bind(&defValInt, chars.Nums...),
		parser.Bind(&defValDone, '\n'),
	)
	defValIntStart.Has(
		valInt...,
	)
	defValInt.Has(
		valInt...,
	)
	defValDone.Has(
		root...,
	)
}
