package nodes

import (
	"github.com/Bruoche/book-compiler/internal/obj"
	"github.com/Bruoche/book-compiler/pkg/chars"
	"github.com/Bruoche/book-compiler/pkg/comp"
)

var Root = comp.Make(
	"root",
	comp.Pass[obj.Empty](),
)
var startWord = comp.Make(
	"word_start",
	func(c comp.Context, i obj.Empty, char byte) obj.Sentence {
		return obj.Sentence{
			Words: []string{string(char)},
		}
	},
)
var Word = comp.Make(
	"word",
	func(c comp.Context, i obj.Sentence, char byte) obj.Sentence {
		last := len(i.Words) - 1
		i.Words[last] = i.Words[last] + string(char)
		return i
	},
)
var Space = comp.Make(
	"space",
	func(c comp.Context, i obj.Sentence, char byte) obj.Sentence {
		i.Words = append(i.Words, "")
		return i
	},
)
var End = comp.Make(
	"end",
	func(c comp.Context, i obj.Sentence, char byte) obj.Empty {
		c.Add(i.Words)
		return obj.Empty{}
	},
)
var Validated = comp.Make(
	"validated",
	comp.Pass[obj.Empty](),
)

func Init() {
	Root.Has(
		comp.Ignoring(' ', comp.Has(
			comp.Bind(&startWord, chars.Capitalized...),
		))...,
	)
	wordHandles := comp.Has(
		comp.Bind(&Space, ' '),
		comp.Bind(&Word, chars.Uncapped...),
		comp.Bind(&End, '.'),
	)
	startWord.Has(
		wordHandles...,
	)
	Word.Has(
		wordHandles...,
	)
	Space.Has(
		comp.Ignoring(' ', comp.Has(
			comp.Bind(&Word, chars.Uncapped...),
		))...,
	)
	End.Has(
		comp.Bind(&Root, '\n'),
		comp.Bind(&Validated, '.'),
	)
	Validated.Has(
		comp.Ignoring('.', comp.Ignoring(' ', comp.Has[obj.Empty]()))...,
	)
}
