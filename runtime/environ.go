package runtime

import (
	"mvdan.cc/sh/expand"
)

// Environ implements expand.Environ on a map
type Environ map[string]expand.Variable

// Get retrieves a variable by its name
func (e Environ) Get(name string) expand.Variable {
	if v, ok := e[name]; ok {
		return v
	}
	return expand.Variable{}
}

// Each iterates over all the currently set variables
func (e Environ) Each(fn func(string, expand.Variable) bool) {
	for k, v := range e {
		fn(k, v)
	}
}

// MultiEnviron implements environment variable lookup in multiple environments
type MultiEnviron []expand.Environ

// Get retrieves a variable by its name from the first environ providing it
func (e MultiEnviron) Get(name string) expand.Variable {
	z := expand.Variable{}
	for _, env := range e {
		if vr := env.Get(name); vr != z {
			return vr
		}
	}
	return expand.Variable{}
}

// Each iterates over all the currently set variables, in every environ
func (e MultiEnviron) Each(fn func(string, expand.Variable) bool) {
	for _, env := range e {
		env.Each(fn)
	}
}
