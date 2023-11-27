package fuzz

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"path/filepath"
	"testing"
)

func listFilesRecursively(dir string) ([]string, error) {
	var files []string

	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

func TestCorpusExtract(t *testing.T) {
	t.Run("extracts fuzzing targets", func(t *testing.T) {
		project := Project{Directory: "./testdata/corpus/multiple"}
		ctx := context.Background()
		tempDir := t.TempDir()

		err := project.CorpusExtract(ctx, tempDir, "...")
		if !assert.NoError(t, err, "corpus copying should not fail") {
			return
		}

		files, err := listFilesRecursively(tempDir)
		if !assert.NoError(t, err, "listing tempDir should not fail") {
			return
		}

		assert.ElementsMatch(t, files, []string{
			"testdata/fuzz/FuzzTarget/0a7e5e215d8c088d4b9c4993d0189a07e81603fbdf64f2ca44738aa27159acef",
			"sub/testdata/fuzz/FuzzSubTarget/0a7e5e215d8c088d4b9c4993d0189a07e81603fbdf64f2ca44738aa27159acef",
		})
	})
}

func TestCorpusDelete(t *testing.T) {
	t.Run("deletes corpora that have matching target", func(t *testing.T) {
		tempDir := t.TempDir()

		if err := copyDirectory(tempDir, "./testdata/corpus/multiple"); err != nil {
			t.Fatal(err)
		}

		files, err := listFilesRecursively(tempDir)
		if !assert.NoError(t, err, "listing tempDir should not fail") {
			return
		}

		if !assert.ElementsMatch(t, files, []string{
			"go.mod",
			"main_test.go",
			"nocorpus/main_test.go",
			"sub/main_test.go",
			"testdata/fuzz/FuzzTarget/0a7e5e215d8c088d4b9c4993d0189a07e81603fbdf64f2ca44738aa27159acef",
			"sub/testdata/fuzz/FuzzSubTarget/0a7e5e215d8c088d4b9c4993d0189a07e81603fbdf64f2ca44738aa27159acef",
			"sub/testdata/fuzz/FuzzNonExistingTarget/0a7e5e215d8c088d4b9c4993d0189a07e81603fbdf64f2ca44738aa27159acef",
		}) {
			return
		}

		project := Project{Directory: tempDir}
		ctx := context.Background()
		err = project.CorpusDelete(ctx, "...")
		assert.NoError(t, err, "corpus deletion")

		files, err = listFilesRecursively(tempDir)
		if !assert.NoError(t, err, "listing tempDir should not fail") {
			return
		}

		if !assert.ElementsMatch(t, files, []string{
			"go.mod",
			"main_test.go",
			"nocorpus/main_test.go",
			"sub/main_test.go",
			"sub/testdata/fuzz/FuzzNonExistingTarget/0a7e5e215d8c088d4b9c4993d0189a07e81603fbdf64f2ca44738aa27159acef",
		}) {
			return
		}

	})
}
