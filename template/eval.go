package template

import (
	"context"
	"fmt"
	"io"
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
	case "date":
	case "xbps-uhelper":
	default:
		log.Println("trying to exec:", args[0])
		panic("exec")
	}
	return nil
}

func limitedOpen(ctx context.Context, path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	panic("open")
	return nil, nil
}


func (t *Template) findSubPackages() string {
	// if subpackages variable is defined use it.
	// second try, walk through all nodes and find functions
	// with _package suffix.
	subs := []string{}
	syntax.Walk(t.file, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.FuncDecl:
			name := strings.TrimSuffix(x.Name.Value, "_package")
			if name != x.Name.Value {
				subs = append(subs, name)	
			}
			return false
		}
		return true
	})
	return strings.Join(subs, " ")
}

func (t *Template) Eval(env *Environ) (map[string]expand.Variable, error) {
	opts := make(Options)
	r, _ := interp.New(
		interp.Env(MultiEnviron{env, opts}),
	)
	r.Exec = limitedExec
	r.Open = limitedOpen

	// pass 1 to get options
	if err := r.Run(context.TODO(), t.file); err != nil {
		return nil, fmt.Errorf("could not run: %v", err)
	}
	opts.setFromTemplateVars(
		r.Vars["build_options"].String(),
		r.Vars["build_options_default"].String(),
	)

	// pass 2 with options in environment
	ctx := context.WithValue(context.Background(), OptionsCtxKey{}, opts)
	if err := r.Run(ctx, t.file); err != nil {
		return nil, fmt.Errorf("could not run: %v", err)
	}
	delete(r.Vars, "PWD")
	delete(r.Vars, "HOME")
	delete(r.Vars, "PATH")
	delete(r.Vars, "IFS")
	delete(r.Vars, "OPTIND")

	_, ok := r.Vars["subpackages"]
	if !ok {
		r.Vars["subpackages"] = expand.Variable{
			Local: true,
			Value: t.findSubPackages(),
		}
	}

	return r.Vars, nil
}
