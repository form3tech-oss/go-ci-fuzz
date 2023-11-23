package fuzz

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func (p *Project) CorpusExtract(ctx context.Context, destination string, packages ...string) error {
	targets, err := p.ListFuzzTargets(ctx, packages...)
	if err != nil {
		return err
	}

	if len(targets) == 0 {
		return nil
	}

	if err := os.MkdirAll(destination, 0700); err != nil {
		return fmt.Errorf("cannot create destination corpus directory: %w", err)
	}

	for _, target := range targets {
		corpusDir := p.relCorpusDir(target)

		srcCorpusDir := filepath.Join(p.Directory, corpusDir)
		if _, err := os.Stat(srcCorpusDir); os.IsNotExist(err) {
			continue
		}

		destCorpusDir := filepath.Join(destination, corpusDir)
		if err := os.MkdirAll(destCorpusDir, 0700); err != nil {
			return fmt.Errorf("cannot create target destination corpus for %s: %w", destCorpusDir, err)
		}

		if err := copyDirectory(destCorpusDir, srcCorpusDir); err != nil {
			return fmt.Errorf("copying %q to %q failed: %w", srcCorpusDir, destCorpusDir, err)
		}
	}

	return nil
}

func (p *Project) CorpusDelete(ctx context.Context, packages ...string) error {
	targets, err := p.ListFuzzTargets(ctx, packages...)
	if err != nil {
		return err
	}

	for _, target := range targets {
		relDir := p.relCorpusDir(target)

		currentCorpusDir := filepath.Join(p.Directory, relDir)
		err := os.RemoveAll(currentCorpusDir)
		if err != nil {
			return fmt.Errorf("error deleting corpus entries for %q located at %s", target, relDir)
		}
	}

	return nil
}

func (p *Project) CorpusReplace(ctx context.Context, external string, packages ...string) error {
	err := p.CorpusDelete(ctx, packages...)
	if err != nil {
		return err
	}
	return p.CorpusMerge(ctx, external, packages...)
}

func (p *Project) CorpusMerge(ctx context.Context, external string, packages ...string) error {
	targets, err := p.ListFuzzTargets(ctx, packages...)
	if err != nil {
		return err
	}

	for _, target := range targets {
		corpusDir := p.relCorpusDir(target)

		currentCorpusDir := filepath.Join(p.Directory, corpusDir)
		externalCorpusDir := filepath.Join(external, corpusDir)

		if err := copyDirectory(currentCorpusDir, externalCorpusDir); err != nil {
			return fmt.Errorf("copying %q to %q failed: %w", externalCorpusDir, currentCorpusDir, err)
		}
	}

	return nil
}
