package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/Bruoche/book-compiler/internal/our"
	"github.com/Bruoche/book-compiler/pkg/parser"
)

func main() {
	fmt.Println("Compiling...")
	our.Init()
	err := compile("test")
	if err != nil {
		fmt.Println(err)
	}
}

func compile(filename string) error {
	c, err := Read(filename)
	if err != nil {
		return fmt.Errorf("Syntax error: %s\n", err)
	}
	err = Write(c, filename)
	if err != nil {
		return fmt.Errorf("Error during compilation: %s\n", err)
	}
	return nil
}

func Read(filename string) (parser.Context, error) {
	path := fmt.Sprintf("./%s.bk", filename)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	c, err := process(path, r)
	if err != nil {
		return c, err
	}
	return c, nil
}

func Write(c parser.Context, filename string) error {
	path := fmt.Sprintf("./%s.c", filename)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		err = f.Close()
	}()
	w := bufio.NewWriter(f)
	err = c.Compile(w)
	if err != nil {
		return err
	}
	err = w.Flush()
	return err
}

func process(path string, r *bufio.Reader) (parser.Context, error) {
	var ignored = parser.Lookup{}
	ignored['\r'] = true
	var c parser.Context = our.BaseContext()
	var node parser.Processor = &our.Root
	var object any = our.Empty{}
	tracker := PositionalTracker{
		path: path,
		line: 0,
		char: 0,
	}
	char := byte(0)
	for true {
		var err error
		if !ignored[char] {
			defined, err := node.Try(object, char)
			if err != nil {
				return c, tracker.Throw(err)
			}
			if defined != nil {
				err := c.Add(defined)
				if err != nil {
					return c, err
				}
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
