package runtime

import (
	"context"
	"fmt"
	"io"
	// "io/fs"
	"log"
	"os"
	"strings"

	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"
)

func limitedExec(ctx context.Context, path string, args []string) error {
	switch args[0] {
	case "vopt_if":
		return shVoptIf(ctx, args)
	case "vopt_with":
	case "vopt_enable":
	case "vopt_conflict":
	case "vopt_bool":
	case "vopt_feature":
	case "date":
	case "xbps-uhelper":
	case "seq":
	case "cut":
	case "grep":
	default:
		panic(args[0])
	}
	return nil
}

type null struct{}

func (n null) Read(b []byte) (int, error) {
	return 0, io.EOF
}

func (n null) Write(b []byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func (n null) Close() error {
	return nil
}

func limitedOpen(ctx context.Context, path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	log.Printf("warn: open: %s: skipped\n", path)
	// return nil, fs.ErrPermission
	return null{}, nil
}

// evalSubPkgs
func (r *Runtime) evalSubPkgs(
	run *interp.Runner,
	ctx context.Context,
	subpkgs ...string,
) ([]map[string]string, error) {
	// Clean the environment using common/environment/setup-subpkg/*.sh
	for _, f := range r.setupSubpkg {
		if err := run.Run(context.TODO(), f); err != nil {
			// XXX: ignore exit status?
			if _, ok := err.(interp.ExitStatus); !ok {
				return nil, fmt.Errorf("could not run: %v", err)
			}
		}
	}

	sourcepkg := run.Vars["pkgname"]
	run.Vars["sourcepkg"] = sourcepkg

	vars := make(map[string]expand.Variable)
	for k, v := range run.Vars {
		vars[k] = v
	}

	res := make([]map[string]string, len(subpkgs))
	for i, subpkgname := range subpkgs {
		fnname := fmt.Sprintf("%s_package", subpkgname)
		fn, ok := run.Funcs[fnname]
		if !ok {
			return nil, fmt.Errorf("missing sub package function: %s", fnname)
		}

		run.Vars["pkgname"] = expand.Variable{Value: subpkgname}

		if err := run.Run(ctx, fn); err != nil {
			// XXX: ignore exit status?
			if _, ok := err.(interp.ExitStatus); !ok {
				return nil, fmt.Errorf("could not run: %v", err)
			}
		}

		res[i] = templateVars(run.Vars)

		// reset variables
		if i < len(subpkgs) {
			run.Vars = make(map[string]expand.Variable)
			for k, v := range vars {
				run.Vars[k] = v
			}
		}
	}

	return res, nil
}

func templateVars(vars map[string]expand.Variable) map[string]string {
	res := make(map[string]string)
	for k, v := range vars {
		// ignore variables starting with uppercase or _
		if k[0] == '_' || (k[0] >= 'A' && k[0] <= 'Z') {
			continue
		}
		res[k] = strings.Join(strings.Fields(v.String()), " ")
	}
	return res
}

func getSubPackages(run *interp.Runner) []string {
	if subs, ok := run.Vars["subpackages"]; ok {
		return strings.Fields(subs.String())
	}
	var res []string
	for fn, _ := range run.Funcs {
		if len(fn) < len("_package") {
			continue
		}
		if s := fn[len(fn)-len("_package"):]; s == "_package" {
			res = append(res, fn[:len(fn)-len("_package")])
		}
	}
	return res
}

// Eval evaluates a template
func (r *Runtime) Eval(
	file *syntax.File,
	arch, cross string,
) ([]map[string]string, error) {

	env := r.Env(arch, cross)
	opts := make(Options)

	run, err := interp.New(interp.Env(MultiEnviron{env, opts}))
	if err != nil {
		return nil, err
	}
	run.Exec = limitedExec
	run.Open = limitedOpen

	// pass 1 to get options
	if err := run.Run(context.TODO(), file); err != nil {
		// XXX: ignore exit status?
		if _, ok := err.(interp.ExitStatus); !ok {
			return nil, fmt.Errorf("could not run: %s", err)
		}
	}
	opts.Add(run.Vars["build_options"].String())
	opts.Defaults(run.Vars["build_options_default"].String())

	// Add the build_style environment if set
	if name := run.Vars["build_style"].String(); name != "" {
		// check if the buildstyle has an environment script
		if file, ok := r.buildStyleEnv[name]; ok {
			if err := run.Run(context.TODO(), file); err != nil {
				// XXX: ignore exit status?
				if _, ok := err.(interp.ExitStatus); !ok {
					return nil, fmt.Errorf("could not run: %v", err)
				}
			}
		}
	}

	// pass 2 with options and buildstyle environment
	ctx := context.WithValue(context.Background(), "options", opts)
	if err := run.Run(ctx, file); err != nil {
		// XXX: ignore exit status?
		if _, ok := err.(interp.ExitStatus); !ok {
			return nil, fmt.Errorf("could not run: %v", err)
		}
	}

	var vars []map[string]string
	mainvars := templateVars(run.Vars)
	vars = append(vars, mainvars)

	subpkgs := getSubPackages(run)
	if len(subpkgs) == 0 {
		return vars, nil
	}

	subvars, err := r.evalSubPkgs(run, ctx, subpkgs...)
	if err != nil {
		return nil, err
	}
	mainvars["subpackages"] = strings.Join(subpkgs, " ")
	return append(vars, subvars...), nil

}
