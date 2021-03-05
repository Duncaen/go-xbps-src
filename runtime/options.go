package runtime

import (
	"fmt"
	"strings"

	"mvdan.cc/sh/expand"
)

type Options map[string]bool

// Defaults adds all options from str and sets them to true
func (o Options) Defaults(str string) {
	for _, opt := range strings.Fields(str) {
		o[opt] = true
	}
}

// Add adds all options from str without overwriting previous set options
func (o Options) Add(str string) {
	for _, opt := range strings.Fields(str) {
		if _, ok := o[opt]; ok {
			continue
		}
		o[opt] = false
	}
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
		Local:    true,
		Value:    val,
	}
}

// Each implements the expand.Environ interface for build options.
func (o Options) Each(fn func(string, expand.Variable) bool) {
	for k, v := range o {
		val := ""
		if v {
			val = "1"
		}
		fn(fmt.Sprintf("build_option_%s", k), expand.Variable{
			Exported: false,
			ReadOnly: true,
			Local:    true,
			Value:    val,
		})
	}
}
