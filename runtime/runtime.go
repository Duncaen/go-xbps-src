package runtime

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"mvdan.cc/sh/syntax"
)

const envFilesSubPkg = "common/environment/setup-subpkg/*.sh"
const envFilesBuildStyle = "common/environment/build-style/*.sh"
const virtualPkgDefaults = "etc/defaults.virtual"

type Runtime struct {
	setupSubpkg   []*syntax.File
	buildStyleEnv map[string]*syntax.File
	parser        *syntax.Parser
	distdir       string
	env           Environ
	virtdefs      map[string]string
}

// Parse parses a bash script
func (r *Runtime) Parse(path string) (*syntax.File, error) {
	rd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer rd.Close()
	f, err := r.parser.Parse(rd, path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// New creates and initializes the runtime
func New(distdir string) (*Runtime, error) {
	r := &Runtime{
		distdir:       distdir,
		parser:        syntax.NewParser(syntax.Variant(syntax.LangBash)),
		buildStyleEnv: make(map[string]*syntax.File),
	}

	pat := path.Join(r.distdir, envFilesSubPkg)
	files, err := filepath.Glob(pat)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("No files found for pattern %q", pat)
	}
	for _, path := range files {
		f, err := r.Parse(path)
		if err != nil {
			return nil, err
		}
		r.setupSubpkg = append(r.setupSubpkg, f)
	}

	pat = path.Join(r.distdir, envFilesBuildStyle)
	files, err = filepath.Glob(pat)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("No files found for pattern %q", pat)
	}
	for _, path := range files {
		f, err := r.Parse(path)
		if err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(filepath.Base(path), ".sh")
		r.buildStyleEnv[name] = f
	}

	r.virtdefs = make(map[string]string)
	pat = path.Join(r.distdir, virtualPkgDefaults)
	file, err := os.Open(pat)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			panic("invalid default virtual dependency")
		}
		r.virtdefs[fields[0]] = fields[1]
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return r, nil
}

// GetVirtual
func (r *Runtime) GetVirtual(pkgname string) (string, error) {
	if pkg, ok := r.virtdefs[pkgname]; ok {
		return pkg, nil
	}
	return "", fmt.Errorf("virtual package not in defaults: %s", pkgname)
}
