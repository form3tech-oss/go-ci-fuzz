package fuzz

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFuzz(t *testing.T) {
	t.Run("Add() seed entry", func(t *testing.T) {
		ctx := context.Background()
		p := Project{Directory: "./testdata/fuzzing/seed", Quiet: true}

		err := p.Fuzz(ctx, Target{
			Name:        "FuzzTarget",
			Package:     "seed",
			RootPackage: "seed",
		}, 1*time.Minute)

		assert.ErrorIs(t, err, FailingInputError{ID: "seed#0", File: "", Seed: true})
	})

	t.Run("file seed entry", func(t *testing.T) {
		ctx := context.Background()
		p := Project{Directory: "./testdata/fuzzing/seedfile", Quiet: true}

		err := p.Fuzz(ctx, Target{
			Name:        "FuzzTarget",
			Package:     "seedfile",
			RootPackage: "seedfile",
		}, 1*time.Minute)

		assert.ErrorIs(t, err, FailingInputError{ID: "0a7e5e215d8c088d4b9c4993d0189a07e81603fbdf64f2ca44738aa27159acef", File: "testdata/fuzz/FuzzTarget/0a7e5e215d8c088d4b9c4993d0189a07e81603fbdf64f2ca44738aa27159acef", Seed: true})
	})

	t.Run("new entry discovered", func(t *testing.T) {
		ctx := context.Background()
		p := Project{Directory: "./testdata/fuzzing/new", Quiet: true}

		removeTestData := func() {
			err := os.RemoveAll(filepath.Join(p.Directory, "testdata"))
			if err != nil {
				t.Fatal("removing old testdata failed", err)
			}
		}
		removeTestData()
		t.Cleanup(removeTestData)

		err := p.Fuzz(ctx, Target{
			Name:        "FuzzTarget",
			Package:     "github.com/form3tech-oss/new",
			RootPackage: "github.com/form3tech-oss/new",
		}, 1*time.Minute)

		var inputErr FailingInputError

		assert.ErrorAs(t, err, &inputErr)
		assert.True(t, strings.HasPrefix(inputErr.File, "testdata/fuzz/FuzzTarget/"), "error.File must begin with testdata/fuzz/FuzzTarget/ it's %s instead", inputErr.File)
		assert.False(t, inputErr.Seed, "newly found failing input must not be marked as Seed")
	})

	t.Run("no findings", func(t *testing.T) {
		ctx := context.Background()
		p := Project{Directory: "./testdata/fuzzing/nofindings", Quiet: true}

		err := p.Fuzz(ctx, Target{
			Name:        "FuzzTarget",
			Package:     "nofindings",
			RootPackage: "nofindings",
		}, 5*time.Second)

		assert.NoError(t, err)
	})
}
