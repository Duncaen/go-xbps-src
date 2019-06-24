package linter

import (
	"regexp"

	"mvdan.cc/sh/syntax"
)

var patSubPkg = regexp.MustCompile(`_package$`)

func (l *linter) subPackage(fn *syntax.FuncDecl) {
	shortDesc, pkgInstall, meta := false, false, false
	syntax.Walk(fn, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.Assign:
			switch x.Name.Value {
			case "short_desc":
				shortDesc = true
			case "build_style":
				meta = makeValue(x.Value) == "meta"
			}
		case *syntax.FuncDecl:
			switch x.Name.Value {
			case "pkg_install":
				pkgInstall = true
			default:
			}
		}
		return true
	})
	if !shortDesc {
		l.errorf(newPos(fn.Pos()), `sub-package '%s' missing short_desc assignment`, fn.Name.Value)
	}
	if !pkgInstall && !meta {
		l.errorf(newPos(fn.Pos()), `sub-package '%s' missing pkgInstall function pr build_style=meta`, fn.Name.Value)
	} else if meta && pkgInstall {
		l.errorf(newPos(fn.Pos()), `sub package '%s' has build_style=meta and a pkg_install function`, fn.Name.Value)
	}
}

func (l *linter) subpackages() {
	syntax.Walk(l.f, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.FuncDecl:
			if patSubPkg.MatchString(x.Name.Value) {
				l.subPackage(x)
				return true
			}
			return false
		}
		return true
	})
}
