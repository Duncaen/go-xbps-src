package runtime

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"mvdan.cc/sh/syntax"
)

const envFilesSubPkg = "common/environment/setup-subpkg/*.sh"
const envFilesBuildStyle = "common/environment/build-style/*.sh"

type Runtime struct {
	setupSubpkg   []*syntax.File
	buildStyleEnv map[string]*syntax.File
	parser        *syntax.Parser
	distdir       string
	env           Environ
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

	return r, nil
}
