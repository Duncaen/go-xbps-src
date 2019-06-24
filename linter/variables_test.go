package linter

import (
	"strings"
	"testing"
)

var tests = []struct {
	in     string
	errors []string
}{
	{`short_desc="foo."`, []string{errShortDescDot, errShortDescUpper}},
	{`short_desc="Foo."`, []string{errShortDescDot}},
	{`short_desc="Foo"`, nil},
	{`short_desc="An foo"`, []string{errShortDescArticle}},
	{`short_desc="The foo"`, []string{errShortDescArticle}},
	{`short_desc="And this is ok"`, nil},
	{`short_desc=" Foo"`, []string{errShortDescWhitespace}},
	{`short_desc="Foo "`, []string{errShortDescWhitespace}},
	{`short_desc=" Foo "`, []string{errShortDescWhitespace}},
	{`short_desc="Foo Bar Foo Bar Foo Bar Foo Bar Foo Bar Foo Bar Foo Bar Foo Bar Foo Bar"`, []string{errShortDescLength}},
	{`license="GPL"`, []string{errLicenseGPLVersion}},
	{`license="GPL-2.0-or-later"`, nil},
	{`license="LGPL"`, []string{errLicenseLGPLVersion}},
	{`license="LGPL-2.0-or-later"`, nil},
	{`license="SSPL"`, []string{errLicenseSSPL}},
	{`revision=0`, []string{errRevisionZero}},
	{`revision=1`, nil},
	{`version=1-1`, []string{errVersionInvalid}},
	{`version=1:1`, []string{errVersionInvalid}},
	{`version=1.1`, nil},
	{`version=1`, nil},
	{`maintainer="foo bar <foo@users.noreply.github.com>"`, []string{errMaintainerMail}},
	{`maintainer="foo bar"`, []string{errMaintainerMail}},
	{`maintainer="foo bar <foo@example.com>"`, nil},
	{`replaces="foo"`, []string{errReplacesVersion}},
	{`replaces="foo bar"`, []string{errReplacesVersion}},
	{`replaces="foo>=0 foo"`, []string{errReplacesVersion}},
	{`replaces="foo>=0"`, nil},
	{`replaces="foo>=0 foo>=1"`, nil},
	{`foo=bar`, []string{errVariableName("foo")}},
	{`_foo=bar`, nil},
}

func TestVariables(t *testing.T) {
	for _, test := range tests {
		i := 0
		errs, err := Lint(strings.NewReader(test.in), "buffer", LintVariables)
		if err != nil {
			t.Fatal(err)
		}
		for _, err := range errs {
			if i < len(test.errors) && err.Msg != test.errors[i] {
				t.Errorf("%q returned %q, want %q", test.in, err.Msg, test.errors[i])
			}
			i++
		}
		if i != len(test.errors) {
			t.Errorf("%q returned %d errors, want %d", test.in, i, len(test.errors))
		}
	}
}
