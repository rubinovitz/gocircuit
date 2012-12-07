package c

import (
	"circuit/c/errors"
	"os"
	"path"
	"strings"
)

// Layout describes a Go compilation environment
type Layout struct {
	goRoot        string    // GOROOT directory
	goPaths       GoPaths   // All GOPATH paths
	workingGoPath string    // A distinct GOPATH
}

func NewLayout(goroot string, gopaths GoPaths, working string) *Layout {
	return &Layout{
		goRoot:        goroot,
		goPaths:       gopaths,
		workingGoPath: working,
	}
}

// NewWorkingLayout creates a new build environment, where the working
// gopath is derived from the current working directory.
func NewWorkingLayout() (*Layout, error) {
	gopath, err := FindWorkingGoPath()
	if err != nil {
		return nil, err
	}
	return &Layout{
		goRoot:        os.Getenv("GOROOT"),
		workingGoPath: gopath,
		goPaths:       GetGoPaths(),
	}, nil
}

// FindPkg returns the first gopath that contains package pkg.
// If includeGoRoot is set, goroot is checked first.
func (l *Layout) FindPkg(pkgPath string, includeGoRoot bool) (pkgAbs string, err error) {
	if includeGoRoot {
		pkgAbs, err = GoRootExistPkg(l.goRoot, pkgPath)
		if err == nil {
			return pkgAbs, nil
		}
		if err != errors.ErrNotFound {
			return "", err
		}
	}
	_, pkgAbs, err = l.goPaths.FindPkg(pkgPath)
	return pkgAbs, err
}

// FindWorkingPath returns the first gopath that parents the absolute directory dir.
// If includeGoRoot is set, goroot is checked first.
func (l *Layout) FindWorkingPath(dir string, includeGoRoot bool) (gopath string, err error) {
	if includeGoRoot {
		if strings.HasPrefix(dir, path.Join(l.goRoot, "src")) {
			return l.goRoot, nil
		}
	}
	return l.goPaths.FindWorkingPath(dir)
}
