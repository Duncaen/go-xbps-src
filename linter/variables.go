package linter

import (
	"fmt"
	"regexp"
	"strings"

	"mvdan.cc/sh/syntax"
)

var variables = []string{
	"_.*",
	".*_descr",
	".*_groups",
	".*_homedir",
	".*_pgroup",
	".*_shell",
	"desc_option_.*",
	"AR",
	"AS",
	"CC",
	"CFLAGS",
	"CPP",
	"CPPFLAGS",
	"CXX",
	"CXXFLAGS",
	"GCC",
	"LD",
	"LDFLAGS",
	"LD_LIBRARY_PATH",
	"NM",
	"OBJCOPY",
	"OBJDUMP",
	"RANLIB",
	"READELF",
	"STRIP",
	"XBPS_FETCH_CMD",
	"allow_unknown_shlibs",
	"alternatives",
	"archs",
	"binfmts",
	"bootstrap",
	"broken",
	"build_options",
	"build_options_default",
	"build_style",
	"build_helper",
	"build_wrksrc",
	"changelog",
	"checkdepends",
	"checksum",
	"cmake_builddir",
	"conf_files",
	"configure_args",
	"configure_script",
	"conflicts",
	"create_wrksrc",
	"depends",
	"disable_parallel_build",
	"distfiles",
	"dkms_modules",
	"fetch_cmd",
	"font_dirs",
	"force_debug_pkgs",
	"go_build_tags",
	"go_get",
	"go_import_path",
	"go_ldflags",
	"go_package",
	"go_mod_mode",
	"homepage",
	"hostmakedepends",
	"keep_libtool_archives",
	"kernel_hooks_version",
	"lib32depends",
	"lib32disabled",
	"lib32files",
	"lib32mode",
	"lib32symlinks",
	"license",
	"maintainer",
	"make_build_args",
	"make_build_target",
	"make_check_args",
	"make_check_target",
	"make_cmd",
	"make_dirs",
	"make_install_args",
	"make_install_target",
	"make_use_env",
	"makedepends",
	"mutable_files",
	"nocross",
	"nodebug",
	"nopie",
	"nopie_files",
	"noshlibprovides",
	"nostrip",
	"nostrip_files",
	"noverifyrdeps",
	"only_for_archs",
	"patch_args",
	"pkgname",
	"preserve",
	"provides",
	"pycompile_dirs",
	"pycompile_module",
	"pycompile_version",
	"python_version",
	"register_shell",
	"replaces",
	"repository",
	"restricted",
	"reverts",
	"revision",
	"run_depends",
	"sgml_catalogs",
	"sgml_entries",
	"shlib_provides",
	"shlib_requires",
	"short_desc",
	"skip_extraction",
	"skiprdeps",
	"stackage",
	"subpackages",
	"system_accounts",
	"system_groups",
	"tags",
	"triggers",
	"version",
	"wrksrc",
	"xml_catalogs",
	"xml_entries",
}

var patVariables = regexp.MustCompile(fmt.Sprintf("^(%s)$", strings.Join(variables, "|")))

const (
	errShortDescDot        = "unwanted trailing dot in short_desc"
	errShortDescUpper      = "short_desc should start uppercase"
	errShortDescArticle    = "short_desc should not start with an article"
	errShortDescWhitespace = "short_desc should not start or end with whitespace"
	errShortDescLength     = "short_desc should be less than 72 chars"
	errLicenseGPLVersion   = "license GPL without version"
	errLicenseLGPLVersion  = "license LGPL without version"
	errLicenseSSPL         = "uses the SSPL license, which is not packageable"
	errRevisionZero        = "revision must not be zero"
	errVersionInvalid      = "version must not contain the characters : or -"
	errMaintainerMail      = "maintainer needs a valid mail address"
	errReplacesVersion     = "replaces needs depname with version"
)

func errVariableName(s string) string {
	return fmt.Sprintf("custom variables should use _ prefix: %s", s)
}

var checks = []struct {
	name  string
	match bool
	pat   *regexp.Regexp
	err   string
}{
	{"short_desc", true, regexp.MustCompile(`\.["']?$`), errShortDescDot},
	{"short_desc", true, regexp.MustCompile(`^["']?[a-z]`), errShortDescUpper},
	{"short_desc", true, regexp.MustCompile(`^["']?(An?|The) `), errShortDescArticle},
	{"short_desc", true, regexp.MustCompile(`(^["']?[\t ]|[\t ]["']?$)`), errShortDescWhitespace},
	{"short_desc", true, regexp.MustCompile(`^["']?.{73}`), errShortDescLength},
	{"license", true, regexp.MustCompile(`[^NL]GPL[^-]`), errLicenseGPLVersion},
	{"license", true, regexp.MustCompile(`LGPL[^-]`), errLicenseLGPLVersion},
	{"license", true, regexp.MustCompile(`SSPL`), errLicenseSSPL},
	{"revision", true, regexp.MustCompile(`^0$`), errRevisionZero},
	{"version", true, regexp.MustCompile(`[:-]`), errVersionInvalid},
	{"maintainer", true, regexp.MustCompile(`@users.noreply.github.com`), errMaintainerMail},
	{"maintainer", false, regexp.MustCompile(`<[^>@]+@[^>]+>`), errMaintainerMail},
	{"replaces", false, regexp.MustCompile(`^['"]?([\w_-]+[<>=]+[^ \t]+[ \t]*)+['"]$`), errReplacesVersion},
}

func (l *linter) variable(a *syntax.Assign) {
	if !patVariables.MatchString(a.Name.Value) {
		l.error(newPos(a.Pos()), errVariableName(a.Name.Value))
		return
	}
	for _, vc := range checks {
		if vc.name != a.Name.Value {
			continue
		}
		if vc.pat.MatchString(makeValue(a.Value)) == vc.match {
			l.error(newPos(a.Pos()), vc.err)
		}
	}
	switch a.Name.Value {
	case "nonfree":
		l.error(newPos(a.Pos()), "use repository=nonfree")
	}
}

var vars = map[string]bool{
	"pkgname":  true,
	"version":  true,
	"revision": true,
}

func (l *linter) quotes(a *syntax.Assign) {
	nam := a.Name.Value
	if _, ok := vars[nam]; !ok {
		return
	}
	if len(a.Value.Parts) != 1 {
		goto error
	}
	switch a.Value.Parts[0].(type) {
	case *syntax.Lit:
	default:
		goto error
	}
	return
error:
	l.errorf(newPos(a.Pos()), `%s must not be quoted`, nam)
}

func (l *linter) variables() {
	syntax.Walk(l.f, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.FuncDecl:
			return false
		case *syntax.Assign:
			l.variable(x)
			l.quotes(x)
		}
		return true
	})
}
