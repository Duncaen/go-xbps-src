package main

import (
	"flag"
	"log"
	"os"
	"io/ioutil"
	"path"
	"fmt"
	"io"
	"strings"

	"github.com/Duncaen/go-xbps-src/bulk"
)

var rules = `
TOBUILD = $(patsubst %,tobuild/%,$(PKGS))
BUILT = $(patsubst tobuild/%,built/%,$(TOBUILD))

all: $(BUILT)
	@echo "[Done]"

print_pkgs:
	@echo $(PKGS)

clean:
	@rm -f built/*
	@echo "[Clean]"

.PHONY: all print_pkgs clean
`


var (
	distdir = flag.String("distdir", os.ExpandEnv("${HOME}/void-packages"), "void-packages repository path")
	masterdir = flag.String("masterdir", "masterdir", "hostdir")
	hostdir = flag.String("hostdir", "hostdir", "masterdir")
	arch = flag.String("arch", "x86_64", "build architecture")
	cross = flag.String("cross", "", "cross architecture")
	flags = flag.String("flags", "-N -t -L -E", "xbps-src flags")
)

func all(b *bulk.Bulk) error {
	files, err := ioutil.ReadDir(path.Join(b.Distdir, "srcpkgs"))
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.Mode()&os.ModeSymlink != 0 {
			continue
		}
		err := b.Add(f.Name())
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	flag.Parse()
	b, err := bulk.New(*distdir, bulk.Config{
		Masterdir: *masterdir,
		Hostdir: *hostdir,
		Arch: *arch,
		Cross: *cross,
	})
	if err != nil {
		log.Fatal(err)
	}
	if flag.NArg() == 0 {
		if err := all(b); err != nil {
			log.Fatal(err)
		}
	} else {
		for _, pkg := range flag.Args() {
			if err := b.Add(pkg); err != nil {
				log.Fatal(err)
			}
		}
	}
	xsc := fmt.Sprintf("%s/xbps-src %s", *distdir, *flags)
	if *cross != "" {
		xsc += " -a " 
		xsc += *cross
	}
	if *masterdir != "" {
		xsc += " -m " 
		xsc += *masterdir
	}
	if *hostdir != "" {
		xsc += " -H " 
		xsc += *hostdir
	}
	io.WriteString(os.Stdout, "# generated with xbps-src-make\n")
	fmt.Fprintf(os.Stdout, "PKGS = %s\n", strings.Join(flag.Args(), " "))
	io.WriteString(os.Stdout, rules)
	io.WriteString(os.Stdout, "built/%: tobuild/%\n")
	io.WriteString(os.Stdout, "\t@echo \"[xbps-src]       ${@F}\"\n")
	fmt.Fprintf(os.Stdout, "\t@( %s pkg ${@F}; rval=$$?; [ $$rval -eq 2 ] && exit 0 || exit $$rval )\n", xsc)
	io.WriteString(os.Stdout, "\t@touch $@\n")
	io.WriteString(os.Stdout, "\t@rm tobuild/${@F}\n\n")
	
	for _, bu := range b.Edges() {
		fmt.Fprintf(os.Stdout, "built/%s:", bu.Pkgname)
		for _, deppkg := range bu.Deps {
			fmt.Fprintf(os.Stdout, " built/%s", deppkg)
		}
		fmt.Fprint(os.Stdout, "\n")
	}
}
