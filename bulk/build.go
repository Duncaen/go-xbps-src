package bulk

import (
	"fmt"
	"strings"

	"github.com/Duncaen/go-xbps/pkgver"
)

type Build struct {
	Pkgname string
	Deps    []string
}

var cycle map[string]bool

func init() {
	cycle = make(map[string]bool)
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
	bulk   *Bulk
	config Config
	cache  map[string][]string
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

func (e *explicit) Deps(pkgname string, stack []string) []string {
	stack = append(stack, pkgname)
	if cycle[pkgname] {
		panic(fmt.Sprintf("cycle: %s %t", pkgname, stack))
	}
	cycle[pkgname] = true
	var res []string
	if deps, ok := e.cache[pkgname]; ok {
		cycle[pkgname] = false
		return deps
	}
	vars, ok := e.bulk.variables[e.config][pkgname]
	if !ok {
		cycle[pkgname] = false
		return res
	}
	var subpkgs []string
	if s, ok := vars["subpackages"]; ok {
		subpkgs = strings.Fields(s)
	}

	uniq := make(map[string]interface{})
	for _, k := range []string{"hostmakedepends", "makedepends"} {
		deps, ok := vars[k]
		if !ok {
			continue
		}
		for _, dep := range strings.Fields(deps) {
			issub := false
			for _, sub := range subpkgs {
				if sub == dep {
					issub = true
					break
				}
			}
			if issub {
				continue
			}
			mainpkg := e.Mainpkg(dep)
			if e.IsEdge(mainpkg) {
				uniq[mainpkg] = nil
			}
			for _, dep := range e.Deps(mainpkg, stack) {
				uniq[dep] = nil
			}
		}
	}

	deps, ok := vars["depends"]
	if ok {
		for _, dep := range strings.Fields(deps) {
			if strings.HasPrefix(dep, "virtual?") {
				var err error
				dep = strings.TrimPrefix(dep, "virtual?")
				pkg, err := pkgver.Parse(dep)
				if err != nil {
					panic(err)
				}
				dep, err = e.bulk.runtime.GetVirtual(pkg.Name)
				if err != nil {
					panic(err)
				}
			}
			pkg, err := pkgver.Parse(dep)
			if err != nil {
				panic(err)
			}
			issub := false
			mainpkg := e.Mainpkg(pkg.Name)
			for _, sub := range subpkgs {
				if sub == pkg.Name {
					issub = true
					break
				}
			}
			if issub {
				continue
			}
			if e.IsEdge(mainpkg) {
				uniq[mainpkg] = nil
			}
			for _, dep := range e.Deps(mainpkg, stack) {
				uniq[dep] = nil
			}
		}
	}
	for k, _ := range uniq {
		res = append(res, k)
	}
	e.cache[pkgname] = res
	cycle[pkgname] = false
	return res
}

func (b *Bulk) Edges() []Build {
	var res []Build
	for _, c := range b.Configs {
		e := &explicit{b, c, make(map[string][]string)}
		for _, pkgname := range b.edges {
			b := Build{
				Pkgname: pkgname,
				Deps:    e.Deps(pkgname, []string{}),
			}
			res = append(res, b)
		}
	}
	return res
}
