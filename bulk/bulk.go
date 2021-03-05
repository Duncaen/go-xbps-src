package bulk

import (
	"log"
	"path"
	"strings"

	"github.com/Duncaen/go-xbps-src/runtime"
	"github.com/Duncaen/go-xbps-src/template"

	"github.com/Duncaen/go-xbps/pkgver"
)

// Config represents a build configuration
type Config struct {
	Arch      string
	Cross     string
	Hostdir   string
	Masterdir string
}

func (c Config) String() string {
	if c.Cross != "" {
		return strings.Join([]string{c.Cross, c.Arch}, "@")
	}
	return c.Arch
}

type Bulk struct {
	Distdir string
	Configs []Config

	// parsed template cache
	templates map[string]*template.Template
	// variable cache
	variables map[Config]map[string]map[string]string

	runtime *runtime.Runtime

	edges []string
}

// New creates and initializes a new Bulk instance
func New(distdir string, configs ...Config) (*Bulk, error) {
	runtime, err := runtime.New(distdir)
	if err != nil {
		return nil, err
	}
	variables := make(map[Config]map[string]map[string]string)
	for _, c := range configs {
		variables[c] = make(map[string]map[string]string)
	}
	return &Bulk{
		Distdir:   distdir,
		Configs:   configs,
		templates: make(map[string]*template.Template),
		variables: variables,
		runtime:   runtime,
	}, nil
}

func (b *Bulk) loadDeps(c Config, vars map[string]string) error {
	for _, k := range []string{"hostmakedepends", "makedepends"} {
		deps, ok := vars[k]
		if !ok {
			continue
		}
		for _, dep := range strings.Fields(deps) {
			err := b.load(c, dep)
			if err != nil {
				return err
			}
		}
	}
	if deps, ok := vars["depends"]; ok {
		for _, dep := range strings.Fields(deps) {
			if strings.HasPrefix(dep, "virtual?") {
				dep = strings.TrimPrefix(dep, "virtual?")
				pkg, err := pkgver.Parse(dep)
				if err != nil {
					return err
				}
				dep, err = b.runtime.GetVirtual(pkg.Name)
				if err != nil {
					return err
				}
			}
			pkg, err := pkgver.Parse(dep)
			if err != nil {
				return err
			}
			err = b.load(c, pkg.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// load evaluates a template with the given config and returns its variables
func (b *Bulk) load(c Config, pkgname string) error {
	// check if template with given configuration is already loaded
	if _, ok := b.variables[c][pkgname]; ok {
		return nil
	}
	// XXX: handle -32bit and -dbg correct
	if strings.HasSuffix(pkgname, "-32bit") {
		b.variables[c][pkgname] = make(map[string]string)
		b.variables[c][pkgname]["sourcepkg"] = strings.TrimSuffix(pkgname, "-32bit")
		return nil
	}
	if strings.HasSuffix(pkgname, "-dbg") {
		b.variables[c][pkgname] = make(map[string]string)
		b.variables[c][pkgname]["sourcepkg"] = strings.TrimSuffix(pkgname, "-dbg")
		return nil
	}
	// check if we parsed the template already
	t, ok := b.templates[pkgname]
	if !ok {
		var err error
		path := path.Join(b.Distdir, "srcpkgs", pkgname, "template")
		t, err = template.ParseFile(path)
		if err != nil {
			log.Print(err)
			return nil
		}
		b.templates[pkgname] = t
		log.Println("parsed template", path)
	}
	log.Printf("evaluating %q for %s\n", pkgname, c)
	vs, err := t.Eval(b.runtime, c.Arch, c.Cross)
	if err != nil {
		log.Print(err)
		return nil
	}
	for _, vars := range vs {
		pkgname := vars["pkgname"]
		b.variables[c][pkgname] = vars
	}

	// load main packages dependencies
	err = b.loadDeps(c, vs[0])
	if err != nil {
		return err
	}

	return nil
}

func (b *Bulk) Add(pkgname string) error {
	for _, c := range b.Configs {
		if err := b.load(c, pkgname); err != nil {
			return err
		}
	}
	b.edges = append(b.edges, pkgname)
	return nil
}
