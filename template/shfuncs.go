package template

import (
	"context"
	"errors"
	"io"

	"mvdan.cc/sh/interp"
)

// write writes a string to a module context's stdout (linked in ctx).
func write(ctx context.Context, s string) error {
	mod, ok := interp.FromModuleContext(ctx)
	if !ok {
		return errors.New("unable to acquire module context")
	}
	_, err := io.WriteString(mod.Stdout, s)
	return err
}

func shVoptIf(ctx context.Context, args []string) error {
	var opt string
	v := false
	ifTrue, ifFalse := "", ""

	switch len(args) {
	case 4:
		ifFalse = args[3]
		fallthrough
	case 3:
		opt = args[1]
		ifTrue = args[2]
	default:
		return errors.New("missing argument")
	}

	switch x := ctx.Value(OptionsCtxKey{}).(type) {
	case Options:
		var ok bool
		if v, ok = x[opt]; !ok {
			return errors.New("invalid option")
		}
	}

	if v {
		write(ctx, ifTrue)
	} else {
		write(ctx, ifFalse)
	}
	return nil
}
