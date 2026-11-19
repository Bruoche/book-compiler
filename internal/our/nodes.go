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

func (v DefVal) Compile() string {
	return fmt.Sprintf("%s = %s", v.name, v.val.T.Compile(v.val.def))
}

var startDefAssign = parser.Make(
	"define_with_val",
	func(c parser.Context, s Subject, char byte) DefVal {
		return DefVal{
			name: s.name,
			val:  Val{},
		}
	},
)
var assigned = parser.Make(
	"end",
	func(c parser.Context, s DefVal, char byte) Empty {
		c.Add(s)
		return Empty{}
	},
)

func Init() {
	Root.Has(
		parser.Ignoring(' ').Has(
			parser.Bind(&startInstruction, chars.Uncapped...),
		)...,
	)
	wordHandles := parser.Has(
		parser.Bind(&subject, chars.Letters...),
		parser.Bind(&subjectDone, ' '),
		parser.Bind(&startDefAssign, '='),
	)
	startInstruction.Has(
		wordHandles...,
	)
	subject.Has(
		wordHandles...,
	)
	subjectDone.Has(
		parser.Ignoring(' ').Has(
			parser.Bind(&startDefAssign, '='),
		)...,
	)
	startDefAssign.Has()
	assigned.Has(
		parser.Bind(&Root, '\n'),
	)
}
