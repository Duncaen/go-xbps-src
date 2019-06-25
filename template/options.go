package template

import (
	"strings"
	"fmt"

	"mvdan.cc/sh/expand"
)

type OptionsCtxKey struct{}

type Options map[string]bool

func (o Options) setFromTemplateVars (opts, defs string) Options {
	for _, opt := range strings.Split(defs, " ") {
		o[opt] = true
	}
	for _, opt := range strings.Split(opts, " ") {
		if _, ok := o[opt]; ok {
			continue
		}
		o[opt] = false
	}
	return o
}

func (o Options) String() string {
	strs := make([]string, len(o))
	for k, v := range o {
		if v {
			strs = append(strs, k)
		} else {
			strs = append(strs, fmt.Sprintf("~%s", k))
		}
	}
	return strings.Join(strs, " ")
}

// Get implements the expand.Environ interface for build options.
func (o Options) Get(name string) expand.Variable {
	opt := strings.TrimPrefix(name, "build_option_")
	if opt == name {
		return expand.Variable{}
	}
	v, ok := o[opt]
	if !ok {
		return expand.Variable{}
	}
	val := ""
	if v {
		val = "1"
	}
	return expand.Variable{
		Exported: false,
		ReadOnly: true,
		Local: true,
		Value: val,
	}
}

// Each implements the expand.Environ interface for build options.
func (o Options) Each(fn func (string, expand.Variable) bool) {
	for k, v := range o {
		val := ""
		if v {
			val = "1"
		}
		fn(fmt.Sprintf("build_option_%s", k), expand.Variable{
			Exported: false,
			ReadOnly: true,
			Local: true,
			Value: val,
		})
	}
}
