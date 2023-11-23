package fuzz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/scanner"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

type TestTarget struct {
	Name        string
	Package     string
	RootPackage string
}

func (tt TestTarget) String() string {
	return fmt.Sprintf("%s#%s", tt.Package, tt.Name)
}

func (p *Project) ListFuzzTargets(ctx context.Context, packages ...string) ([]TestTarget, error) {
	relativePackages := make([]string, len(packages))
	for i, pkg := range packages {
		// NOTE: Go will look in GOPATH for packages that cannot be found in the current module which is slow and breaks the tool
		// TODO: find a better solution to allow calling go-ci-fuzz with fully qualified module name such as github.com/<org>/<repo>/some/package
		relativePackages[i] = fmt.Sprintf("./%s", pkg)
	}

	targets, err := p.listTestTargets(ctx, "^Fuzz*", relativePackages...)
	if err != nil {
		return nil, fmt.Errorf("discovering fuzz targets failed: %s", err)
	}

	return targets, nil
}

type Module struct {
	Path string
	Main bool
}

type Package struct {
	Dir          string
	ImportPath   string
	Name         string
	TestGoFiles  []string
	XTestGoFiles []string
	Module       Module
}

func (p *Project) listPackages(ctx context.Context, packages ...string) ([]Package, error) {
	args := append([]string{
		"list",
		"-find",
		"-json",
	}, packages...)

	cmd := exec.CommandContext(ctx, "go", args...)
	if p.Directory != "" {
		cmd.Dir = p.Directory
	}

	pkgBytes, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("cannot get package list for: %w", err)
	}

	var pkgs []Package

	decoder := json.NewDecoder(bytes.NewReader(pkgBytes))
	for decoder.More() {
		var pkg Package
		if err := decoder.Decode(&pkg); err != nil {
			return nil, err
		}
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

// We cannot use go test -list because of this bug: https://github.com/golang/go/issues/25339
// So we list all packages and test files and look for test targets ourselves by running go.Scanner
func (p *Project) listTestTargets(ctx context.Context, pattern string, packages ...string) ([]TestTarget, error) {
	pkgs, err := p.listPackages(ctx, packages...)
	if err != nil {
		return nil, fmt.Errorf("error listing packages: %w", err)
	}

	pat, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("error compiling pattern")
	}

	var targets []TestTarget

	for _, pkg := range pkgs {
		var testFiles []string
		testFiles = append(testFiles, pkg.TestGoFiles...)
		testFiles = append(testFiles, pkg.XTestGoFiles...)

		for _, testFile := range testFiles {
			path := filepath.Join(pkg.Dir, testFile)
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}

			var scan scanner.Scanner
			fs := token.NewFileSet()
			f := fs.AddFile(testFile, fs.Base(), len(content))
			scan.Init(f, content, nil, 0)

			// TODO: we can improve this so that it checks if the first param of the function is testing.F
			var previousToken token.Token
			for {
				_, tok, lit := scan.Scan()
				if tok == token.EOF {
					break
				}
				if previousToken == token.FUNC && tok.IsLiteral() {
					if pat.MatchString(lit) {
						targets = append(targets, TestTarget{
							Name:        lit,
							Package:     pkg.ImportPath,
							RootPackage: pkg.Module.Path,
						})
					}
				}
				previousToken = tok
			}
		}
	}
	return targets, nil
}
