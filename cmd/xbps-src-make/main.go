package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/Duncaen/go-xbps-src/bulk"
)

var rules = `
TOBUILD = $(patsubst %,tobuild/%,$(PKGS))
BUILT = $(patsubst tobuild/%,built/%,$(TOBUILD))
SORT :=

.PHONY: all sort print_pkgs clean

all: $(BUILT)
	@echo "[Done]"

sort:
	@$(MAKE) SORT=1 all
	@[ -n "$(PKGS)" ] && mv built/* tobuild || :
	@echo "[Sorted]"

print_pkgs:
	@( [ -f pkgs-sorted.txt ] && cat pkgs-sorted.txt | xargs || echo $(PKGS) )

clean:
	@rm -f built/* pkgs.txt pkgs-sorted.txt repo-checkvers.txt
	@echo "[Clean]"

built/%: tobuild/%
	@echo "[xbps-src]       ${@F}"
ifdef SORT
	@echo ${@F} >> pkgs-sorted.txt
else
	@( $(XSC) pkg ${@F}; rval=$$?; [ $$rval -eq 2 ] && exit 0 || exit $$rval )
endif
	@touch $@
	@rm tobuild/${@F}

`

var (
	distdir   = flag.String("distdir", os.ExpandEnv("${HOME}/void-packages"), "void-packages repository path")
	masterdir = flag.String("masterdir", "masterdir", "hostdir")
	hostdir   = flag.String("hostdir", "hostdir", "masterdir")
	arch      = flag.String("arch", "x86_64", "build architecture")
	cross     = flag.String("cross", "", "cross architecture")
	flags     = flag.String("flags", "-N -t -L -E", "xbps-src flags")
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
		Hostdir:   *hostdir,
		Arch:      *arch,
		Cross:     *cross,
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
	fmt.Fprintf(os.Stdout, "XSC = %s\n", xsc)
	io.WriteString(os.Stdout, rules)

	for _, bu := range b.Edges() {
		fmt.Fprintf(os.Stdout, "built/%s:", bu.Pkgname)
		for _, deppkg := range bu.Deps {
			fmt.Fprintf(os.Stdout, " built/%s", deppkg)
		}
		fmt.Fprint(os.Stdout, "\n")
	}
}
