package template

import (
	"errors"
	"io"
	"os"

	"mvdan.cc/sh/syntax"
)

type Template struct {
	file *syntax.File
}

// ParseFile
func ParseFile(path string) (*Template, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return Parse(r, path)
}

// Parse
func Parse(r io.Reader, name string) (*Template, error) {
	parser := syntax.NewParser(syntax.Variant(syntax.LangBash))
	f, err := parser.Parse(r, name)
	if err != nil {
		return nil, err
	}
	if parser.Incomplete() {
		return nil, errors.New("inclomplete")
	}
	return &Template{file: f}, nil
}
