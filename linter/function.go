package linter

import (
	"mvdan.cc/sh/syntax"
)

var defaultFns = map[string]bool{
	"pre_fetch":      true,
	"do_fetch":       true,
	"post_fetch":     true,
	"pre_extract":    true,
	"do_extract":     true,
	"post_extract":   true,
	"pre_patch":      true,
	"do_patch":       true,
	"post_patch":     true,
	"pre_configure":  true,
	"do_configure":   true,
	"post_configure": true,
	"pre_build":      true,
	"do_build":       true,
	"post_build":     true,
	"pre_check":      true,
	"do_check":       true,
	"post_check":     true,
	"pre_install":    true,
	"do_install":     true,
	"post_install":   true,
	"do_clean":       true,
	"pkg_install":    true,
}

func (l *linter) function(fn *syntax.FuncDecl) {
	nam := fn.Name.Value
	if fn.RsrvWord {
		l.errorf(newPos(fn.Pos()), `%s: must use posix style function declaration`, nam)
	}
	if _, ok := defaultFns[nam]; ok {
		return
	}
	switch {
	case patSubPkg.MatchString(nam):
	case nam[0] == '_':
	default:
		l.errorf(newPos(fn.Pos()), `%s: custom function should start with '_'`, nam)
	}
}

func (l *linter) functions() {
	syntax.Walk(l.f, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.FuncDecl:
			l.function(x)
		}
		return true
	})
}
