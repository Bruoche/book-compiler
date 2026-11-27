package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/Bruoche/book-compiler/internal/our"
	"github.com/Bruoche/book-compiler/pkg/parser"
)

func main() {
	fmt.Println("Compiling...")
	our.Init()
	name := "test"
	sources, err := compile(name)
	handle(err)
	err = run(sources, name)
	handle(err)
}

func handle(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(path, output string) error {
	build := exec.Command("wsl", "gcc", path, "-o", output)
	out, err := build.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		return fmt.Errorf("C Compilation error: %s", err)
	}
	run := exec.Command("wsl", "./"+output)
	out, err = run.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		return fmt.Errorf("Runtime error: %s", err)
	}
	return nil
}

func compile(filename string) (string, error) {
	c, err := Read(filename)
	if err != nil {
		return "", fmt.Errorf("Syntax error: %s\n", err)
	}
	result, err := Write(c, filename)
	if err != nil {
		return "", fmt.Errorf("Error during compilation: %s\n", err)
	}
	return result, nil
}

func Read(filename string) (parser.IR, error) {
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

func Write(c parser.IR, filename string) (string, error) {
	path := fmt.Sprintf("./%s.c", filename)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer func() {
		err = f.Close()
	}()
	w := bufio.NewWriter(f)
	err = c.Compile(w)
	if err != nil {
		return "", err
	}
	err = w.Flush()
	return path, err
}

func process(path string, r *bufio.Reader) (parser.IR, error) {
	var ignored = parser.Lookup{}
	ignored['\r'] = true
	var c parser.IR = our.IR()
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
				c.Add(defined)
			}
			object = node.Get()
		}
		char, err = r.ReadByte()
		if err != nil {
			if err == io.EOF {
				last, done := object.(parser.Compilable)
				if done {
					c.Add(last)
					return c, nil
				}
				_, done = object.(our.Empty)
				if done {
					return c, nil
				}
				return c, err
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
