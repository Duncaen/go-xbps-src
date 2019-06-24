// Package linter provides a linter for xbps-src template files.
//
// Most of the implemented checks are originally from xtools xlint
// shell script.
package linter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"mvdan.cc/sh/syntax"
)

// Pos represents the position in the source (line, column).
type Pos struct {
	line uint
	col  uint
}

func newPos(p syntax.Pos) Pos {
	return Pos{p.Line(), p.Col()}
}

func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.line, p.col)
}

// Error represents an issue the linter found.
type Error struct {
	Pos  Pos
	File string
	Msg  string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s:%s: %s", e.File, e.Pos, e.Msg)
}

func (l *linter) emit(e Error) {
	l.errors = append(l.errors, e)
}

func (l *linter) error(p Pos, msg string) {
	l.emit(Error{
		File: l.f.Name,
		Pos:  p,
		Msg:  msg,
	})
}

func (l *linter) errorf(p Pos, format string, a ...interface{}) {
	l.error(p, fmt.Sprintf(format, a...))
}

const (
	_ = 1 << iota
	// LintHeader flag enables header linting.
	LintHeader
	// LintFunctions flag enables function linting.
	LintFunctions
	// LintVariables flag enables variable linting.
	LintVariables
	// LintSubPackages flag enables sub package linting.
	LintSubPackages
	// LintAll flag enables all available lint flags.
	LintAll = LintHeader | LintFunctions | LintVariables | LintSubPackages
)

type linter struct {
	errors []Error
	f      *syntax.File
	flags  int
}

func makeValue(node syntax.Node) string {
	buf := new(bytes.Buffer)
	printer := syntax.NewPrinter(syntax.Minify)
	printer.Print(buf, node)
	return buf.String()
}

// LintFile runs the linters specified flags on a file specified by path.
func LintFile(path string, flags ...int) ([]Error, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return Lint(r, path, flags...)
}

// Lint runs the linters specified flags on a reader.
func Lint(r io.Reader, name string, flags ...int) ([]Error, error) {
	parser := syntax.NewParser(
		syntax.KeepComments,
		syntax.Variant(syntax.LangBash),
	)
	f, err := parser.Parse(r, name)
	if err != nil {
		return nil, err
	}
	if parser.Incomplete() {
		return nil, errors.New("inclomplete")
	}
	l := &linter{
		f:      f,
		errors: []Error{},
	}
	flag := 0
	for _, f := range flags {
		flag |= f
	}
	if flag == 0 {
		flag = LintAll
	}
	if flag&LintHeader != 0 {
		l.header()
	}
	if flag&LintVariables != 0 {
		l.variables()
	}
	if flag&LintFunctions != 0 {
		l.functions()
	}
	if flag&LintSubPackages != 0 {
		l.subpackages()
	}
	return l.errors, nil
}
