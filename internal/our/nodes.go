package our

import (
	"fmt"

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
	func(c parser.Context, _ Empty, char byte) Subject {
		return Subject{
			name: string(char),
		}
	},
)
var subject = parser.Make(
	"start_instruction",
	func(c parser.Context, s Subject, char byte) Subject {
		s.name = s.name + string(char)
		return s
	},
)
var subjectDone = parser.Make(
	"action",
	parser.Pass[Subject](),
)

type DefVal struct {
	name string
	val  Val
}

var _ Declarable

func (v DefVal) Compile() (string, error) {
	val, err := v.val.T.Compile(v.val.def)
	return fmt.Sprintf("%s = %s", v.name, val), err
}

var startDefVal = parser.Make(
	"define_with_val",
	func(c parser.Context, s Subject, char byte) DefVal {
		return DefVal{
			name: s.name,
			val:  Val{},
		}
	},
)
var defValIntStart = parser.Make(
	"define_int_val",
	func(c parser.Context, s DefVal, char byte) DefVal {
		s.val = Val{
			def: string(char),
			T:   Int,
		}
		return s
	},
)
var defValInt = parser.Make(
	"define_int_val",
	func(c parser.Context, s DefVal, char byte) DefVal {
		s.val.def += string(char)
		return s
	},
)
var defValDone = parser.Make(
	"end",
	func(c parser.Context, s DefVal, char byte) Empty {
		c.Add(s)
		return Empty{}
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
		parser.Bind(&startDefVal, '='),
	)
	startInstruction.Has(
		wordHandles...,
	)
	subject.Has(
		wordHandles...,
	)
	subjectDone.Has(
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
