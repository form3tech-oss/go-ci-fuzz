package fuzz

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var failingInputRegex = regexp.MustCompile(`^\s*go test -run=Fuzz([a-zA-Z0-9_]+)/([a-zA-Z0-9#]+)`)
var failingSeedInputRegex = regexp.MustCompile(`^\s*failure while testing seed corpus entry: Fuzz([a-zA-Z0-9_]+)/([a-zA-Z0-9#]+)`)

type FailingInputError struct {
	ID   string
	File string
	Seed bool
}

func (f FailingInputError) Error() string {
	newOrSeed := "new"
	if f.Seed {
		newOrSeed = "seed"
	}

	if f.File != "" {
		return fmt.Sprintf("failing %s input, saved at %s", newOrSeed, f.File)
	}
	return fmt.Sprintf("failing %s input: %s", newOrSeed, f.ID)
}

func (p *Project) relCorpusDir(target TestTarget) string {
	// target.Package contains the root package as well
	// we need to strip it because it refers to the current working directory .
	pkg := strings.TrimPrefix(strings.TrimPrefix(target.Package, target.RootPackage), "/")
	return filepath.Join(pkg, "testdata/fuzz", target.Name)
}

func (p *Project) Fuzz(ctx context.Context, target TestTarget, d time.Duration) error {
	args := []string{
		"test",
		"-test.run=^$",
		"-test.fuzz=^" + target.Name + "$",
		"-test.fuzztime=" + d.String(),
		target.Package,
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	if p.Directory != "" {
		cmd.Dir = p.Directory
	}

	var stdout bytes.Buffer
	if !p.Quiet {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	} else {
		cmd.Stdout = &stdout
	}
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err == nil {
		return nil
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return fmt.Errorf("fuzzing failed with an unexpected error: %w", err)
	}

	scanner := bufio.NewScanner(&stdout)
	corpusDirectory := p.relCorpusDir(target)
	for scanner.Scan() {
		line := scanner.Text()

		// For newly discovered inputs the CLI outputs the following:
		// > Failing input written to testdata/fuzz/FuzzTarget/0a7e5e215d8c088d4b9c4993d0189a07e81603fbdf64f2ca44738aa27159acef
		// > To re-run:
		// > go test -run=FuzzTarget/0a7e5e215d8c088d4b9c4993d0189a07e81603fbdf64f2ca44738aa27159acef
		// we match against the last line and extract the Test ID from it
		if matches := failingInputRegex.FindStringSubmatch(line); matches != nil {
			if len(matches) != 3 {
				return fmt.Errorf("parsing fuzzing output failed, matched %q, but found %d submatches, expected 2", line, len(matches))
			}

			id := matches[2]
			return FailingInputError{ID: id, File: filepath.Join(corpusDirectory, id)}
		}

		// For inputs already in the corpus we get
		// > failure while testing seed corpus entry: FuzzTarget/seed#0
		// for seed corpus entries added by f.Add() OR
		// > failure while testing seed corpus entry: FuzzTarget/0a7e5e215d8c088d4b9c4993d0189a07e81603fbdf64f2ca44738aa27159acef
		// for seed corpus stored in files in ./testdata directory
		if matches := failingSeedInputRegex.FindStringSubmatch(line); matches != nil {
			if len(matches) != 3 {
				return fmt.Errorf("parsing seed corpus fuzzing output failed, matched %q, but found %d submatches, expected 2", line, len(matches))
			}
			id := matches[2]
			if strings.HasPrefix(id, "seed#") {
				return FailingInputError{ID: id, Seed: true}
			} else {
				return FailingInputError{ID: id, File: filepath.Join(corpusDirectory, id), Seed: true}
			}
		}

	}

	return fmt.Errorf("fuzzing failed with an unexpected exit error: %w", err)
}
