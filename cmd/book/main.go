package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/Bruoche/book-compiler/internal/nodes"
	"github.com/Bruoche/book-compiler/internal/obj"
	"github.com/Bruoche/book-compiler/pkg/comp"
)

func main() {
	fmt.Println("Compiling...")
	nodes.Init()
	err := Read("./test.bk")
	if err != nil {
		fmt.Printf("Error encountered during compilation: %s\n", err)
	}
}

func Read(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	r := bufio.NewReader(f)
	c, err := process(path, r)
	if err != nil {
		return err
	}
	c.Print()
	return nil
}

func process(path string, r *bufio.Reader) (comp.Context, error) {
	var ignored = comp.Lookup{}
	ignored['\r'] = true
	var c comp.Context = &comp.Compiler{}
	var node comp.Processor = &nodes.Root
	var object any = obj.Empty{}
	tracker := PositionalTracker{
		path: path,
		line: 0,
		char: 0,
	}
	char := byte(0)
	for true {
		var err error
		if !ignored[char] {
			err = node.Try(c, object, char)
			if err != nil {
				return c, tracker.Throw(err)
			}
			object = node.Get()
		}
		char, err = r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return c, nil
			}
			return c, tracker.Throw(err)
		}
		tracker.Increment(char)
		if !ignored[char] {
			node, err = node.Next(char)
			if err != nil {
				return c, tracker.Throw(err)
			}
		}
	}
	return c, nil
}

type CompilationError string

var _ error = CompilationError("")

func (err CompilationError) Error() string {
	return string(err)
}

type PositionalTracker struct {
	path string
	line int
	char int
}

func (t PositionalTracker) Throw(err error) CompilationError {
	return CompilationError(fmt.Sprintf("%s:%d:%d: %s", t.path, t.line+1, t.char+1, err))
}

func (t *PositionalTracker) Increment(char byte) {
	if char == '\n' {
		t.char = 0
		t.line += 1
	} else {
		t.char += 1
	}
}
