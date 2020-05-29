package bulk

import (
	"strings"
)

type Build struct {
	Pkgname string
	Deps []string
}

// func (b *Bulk) edgeDeps(c Config, pkgname string) []string {
// 	print("edges of: ")
// 	println(pkgname)
// 	var res []string
// 	vars, ok := b.variables[c][pkgname]
// 	if !ok {
// 		return res
// 	}
// 	for _, k := range []string{"hostmakedepends", "makedepends"} {
// 		deps, ok := vars[k]
// 		if !ok {
// 			continue
// 		}
// 		for _, dep := range strings.Fields(deps) {
// 			if b.isEdge(c, dep) {
// 				res = append(res, dep)
// 			}
// 			res = append(res, b.edgeDeps(c, dep)...)
// 		}
// 	}
// 	return res
// }

type explicit struct {
	bulk *Bulk
	config Config
	cache map[string][]string
}

func (e *explicit) Mainpkg(pkgname string) string {
	vars := e.bulk.variables[e.config][pkgname]
	if sourcepkg, ok := vars["sourcepkg"]; ok {
		return sourcepkg
	}
	return vars["pkgname"]
}

func (e *explicit) IsEdge(pkgname string) bool {
	for _, e := range e.bulk.edges {
		if e == pkgname {
			return true
		}
	}
	return false
}

func (e *explicit) Deps(pkgname string) []string {
	var res []string
	if deps, ok := e.cache[pkgname]; ok {
		return deps
	}
	vars, ok := e.bulk.variables[e.config][pkgname]
	if !ok {
		return res
	}
	uniq := make(map[string]interface{})
	for _, k := range []string{"hostmakedepends", "makedepends", "depends"} {
		deps, ok := vars[k]
		if !ok {
			continue
		}
		for _, dep := range strings.Fields(deps) {
			mainpkg := e.Mainpkg(dep)
			if e.IsEdge(mainpkg) {
				uniq[mainpkg] = nil
			}
			for _, dep := range e.Deps(mainpkg) {
				uniq[dep] = nil
			}
		}
	}
	for k, _ := range uniq {
		res = append(res, k)
	}
	e.cache[pkgname] = res
	return res
}

func (b *Bulk) Edges() []Build {
	var res []Build
	for _, c := range b.Configs {
		e := &explicit{b, c, make(map[string][]string)}
		for _, pkgname := range b.edges {
			b := Build{
				Pkgname: pkgname,
				Deps: e.Deps(pkgname),
			}
			res = append(res, b)
		}
	}
	return res
}
