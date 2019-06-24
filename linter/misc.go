package linter

import (
	"regexp"
)

var patHeader = regexp.MustCompile(` Template file for '[^']+'`)

func (l *linter) header() {
	if len(l.f.Stmts) < 1 || len(l.f.Stmts[0].Comments) < 1 {
		l.errorf(Pos{1, 1}, `missing header`)
		return
	}
	if !patHeader.MatchString(l.f.Stmts[0].Comments[0].Text) {
		l.errorf(Pos{1, 1}, `header does not match "# Template file for '<pkgname>'"`)
	}
}
