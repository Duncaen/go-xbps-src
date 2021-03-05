package runtime

import (
	"mvdan.cc/sh/expand"
)

// Env returns the xbps-src environment
func (r *Runtime) Env(arch, cross string) Environ {
	var m, t expand.Variable
	m = expand.Variable{
		Exported: true,
		ReadOnly: true,
		Value:    arch,
	}
	t = m

	if cross != "" {
		t = expand.Variable{
			Exported: true,
			ReadOnly: true,
			Value:    cross,
		}
	}

	return Environ{
		"XBPS_UHELPER_CMD": expand.Variable{
			Exported: true,
			ReadOnly: true,
			Value:    "xbps-uhelper",
		},
		"XBPS_MACHINE":        m,
		"XBPS_TARGET_MACHINE": t,
	}
}
