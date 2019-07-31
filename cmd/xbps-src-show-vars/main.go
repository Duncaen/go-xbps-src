package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/Duncaen/go-xbps-src/runtime"
	"github.com/Duncaen/go-xbps-src/template"
)

var (
	distdir = flag.String("distdir", "/home/duncan/void-packages", "distdir")
	arch = flag.String("arch", "x86_64", "architecture")
	cross = flag.String("cross", "", "cross architecture")
	variables = []string{
		"version",
		"revision",
		"archs",
		"arch",
		"broken",
		"nocross",
		"hostmakedepends",
		"makedepends",
		"depends",
	}
)

func main() {
	flag.Parse()
	run, err := runtime.New(*distdir)
	if err != nil {
		log.Fatal(err)
	}
	for _, path := range flag.Args() {
		t, err := template.ParseFile(path)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Evaluating: %s\n", path)
		vs, err := t.Eval(run, *arch, *cross)
		if err != nil {
			log.Fatal(err)
		}
		for _, kv := range vs {
			pkgname := kv["pkgname"]
			for _, s := range variables {
				val := strings.Join(strings.Fields(kv[s]), " ")
				if val == "" {
					continue
				}
				fmt.Printf("%s: %s=%q\n", pkgname, s, val)
			}
		}
		// for _, sv := range vars.SubPackages {
		// 	fmt.Printf("%s: depends=%v\n", sv.PkgName, sv.Depends)
		// }
	}
}
