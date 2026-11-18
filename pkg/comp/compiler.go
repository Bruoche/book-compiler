package comp

import (
	"fmt"
	"strings"
)

type Context interface {
	Add(sentence []string)
	Print()
}

type Compiler struct {
	sentences [][]string
}

var _ Context = &Compiler{}

func (c *Compiler) Add(sentence []string) {
	c.sentences = append(c.sentences, sentence)
}

func (c Compiler) Print() {
	for _, sentence := range c.sentences {
		fmt.Printf("%s;\n", strings.Join(sentence, " "))
	}
}
