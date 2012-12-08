package c

import (
	"circuit/c/errors"
	"circuit/c/types"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
)

type Build struct {
	layout    *Layout
	jail      *Jail

	fileSet   *token.FileSet
	pkgs      map[string]*ast.Package  // pkgPath to package AST
	depTable  *DepTable
	typeTable *types.TypeTable
}

func NewBuild(layout *Layout, jaildir string) (b *Build, err error) {

	b = &Build{layout: layout}

	// Create a new compilation jail
	if b.jail, err = NewJail(jaildir); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Build) Build(pkgs ...string) error {

	var err error

	// Calculate package dependencies
	b.fileSet = token.NewFileSet()
	b.pkgs = make(map[string]*ast.Package)
	if err = b.compileDep(pkgs...); err != nil {
		return err
	}

	// Parse types
	b.typeTable = types.NewTypeTable()
	if err = b.parseTypes(); err != nil {
		return err
	}

	for _, typ := range b.typeTable.List() {
		println(typ)
	}

	return nil
}

// ParsePkg parses the requested package path and saves the resulting package
// AST node into the pkgs field
func (b *Build) ParsePkg(pkgPath string) (map[string]*ast.Package, error) {
	pkgs, err := ParsePkg(b.layout, b.fileSet, pkgPath, false, parser.ParseComments)
	if err != nil {
		Log("- %s skipping", pkgPath)
		// This is intended for Go's packages itself, which we don't want to parse for now
		return nil, nil
	}
	Log("+ %s", pkgPath)
	// Save package AST into global map
	for pkgName, pkg := range pkgs {
		// Note that only one package is expected in pkgs
		_, pkgDirName := path.Split(pkgPath)
		if pkgName != pkgDirName {
			// Package source directories will often contain files with main or xxx_test package clauses.
			// We ignore those, by guessing they are not part of the program.
			// The correct way to handle those is to recognize the comment directive: // +build ignore
			continue
		}
		if _, present := b.pkgs[pkgPath]; present {
			return nil, errors.New("package %s already parsed", pkgPath)
		}
		b.pkgs[pkgPath] = pkg
	}
	return pkgs, nil
}

// compileDep causes all packages that pkgs depend on to be parsed
func (b *Build) compileDep(pkgPaths ...string) error {
	Log("Calculating dependencies ...")
	Indent()
	defer Unindent()

	b.depTable = NewDepTable(b)
	for _, pkgPath := range pkgPaths {
		if err := b.depTable.Add(pkgPath); err != nil {
			return err
		}
	}
	return nil
}

// parseTypes finds all type declarations and registers them with a global map
func (b *Build) parseTypes() error {
	for pkgPath, pkg := range b.pkgs {
		if err := b.typeTable.AddPackage(b.fileSet, pkgPath, pkg); err != nil {
			return err
		}
	}
	return nil
}
