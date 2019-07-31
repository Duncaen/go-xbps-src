package template

import (
	"errors"
	"io"
	"os"

	"github.com/Duncaen/go-xbps-src/runtime"

	"mvdan.cc/sh/syntax"
)

type Template struct {
	file *syntax.File
}

// ParseFile parses the template at path
func ParseFile(path string) (*Template, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return Parse(r, path)
}

// Parse parses a template from r
func Parse(r io.Reader, name string) (*Template, error) {
	parser := syntax.NewParser(syntax.Variant(syntax.LangBash))
	f, err := parser.Parse(r, name)
	if err != nil {
		return nil, err
	}
	if parser.Incomplete() {
		return nil, errors.New("inclomplete")
	}
	t := &Template{file: f}
	return t, nil
}

// Eval evaluates a template
func (t *Template) Eval(
	runtime *runtime.Runtime,
	arch, cross string,
) ([]map[string]string, error) {
	return runtime.Eval(t.file, arch, cross)
}
