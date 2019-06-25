package template

import (
	"mvdan.cc/sh/expand"
)

type Environ struct {
	env map[string]expand.Variable
}

func (e *Environ) Get(name string) expand.Variable {
	vr, ok := e.env[name]
	if !ok {
		return expand.Variable{}
	}
	return vr
}

func (e *Environ) Each(fn func (string, expand.Variable) bool) {
}

func (e *Environ) Set(name string, vr expand.Variable) error {
	return nil
}

func Environment(machine string, targetMachine string) *Environ {
	e := Environ{env: make(map[string]expand.Variable)}
	e.env["XBPS_MACHINE"] = expand.Variable{
		Exported: true,
		ReadOnly: true,
		Value: machine,
	}
	if targetMachine != "" {
		e.env["XBPS_TARGET_MACHINE"] = expand.Variable{
			Exported: true,
			ReadOnly: true,
			Value: targetMachine,
		}
	} else {
		e.env["XBPS_TARGET_MACHINE"] = e.env["XBPS_MACHINE"]
	}
	e.env["XBPS_UHELPER_CMD"] = expand.Variable{
		Exported: true,
		ReadOnly: true,
		Value: "xbps-uhelper",
	}
	return &e
}

type MultiEnviron []expand.Environ

func (e MultiEnviron) Get(name string) expand.Variable {
	z := expand.Variable{}
	for _, env := range e {
		if vr := env.Get(name); vr != z {
			return vr
		}
	}
	return expand.Variable{}
}

func (e MultiEnviron) Each(fn func (string, expand.Variable) bool) {
	for _, env := range e {
		env.Each(fn)
	}
}
