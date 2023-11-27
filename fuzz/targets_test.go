package fuzz

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiscoverTargets(t *testing.T) {
	p := Project{Directory: "./testdata/discover", Quiet: true}
	t.Run("single fuzz target", func(t *testing.T) {
		ctx := context.Background()
		targets, err := p.ListFuzzTargets(ctx, ".")
		assert.NoError(t, err)
		assert.EqualValues(t, []Target{{
			Name:        "FuzzTarget",
			Package:     "discover",
			RootPackage: "discover",
		}}, targets)
	})

	t.Run("all packages", func(t *testing.T) {
		ctx := context.Background()
		targets, err := p.ListFuzzTargets(ctx, "...")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []Target{{
			Name:        "FuzzTarget",
			Package:     "discover",
			RootPackage: "discover",
		}, {
			Name:        "FuzzSubTarget",
			Package:     "discover/subpackage",
			RootPackage: "discover",
		}, {
			Name:        "FuzzMain",
			Package:     "discover/submain",
			RootPackage: "discover",
		}}, targets)
	})

	t.Run("non-existent package", func(t *testing.T) {
		ctx := context.Background()
		_, err := p.ListFuzzTargets(ctx, "doesnotexist")
		assert.Error(t, err, "discovering non existent package should fail")
	})

	t.Run("subpackage", func(t *testing.T) {
		ctx := context.Background()
		targets, err := p.ListFuzzTargets(ctx, "subpackage")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []Target{{
			Name:        "FuzzSubTarget",
			Package:     "discover/subpackage",
			RootPackage: "discover",
		}}, targets)
	})

	t.Run("main does not run", func(t *testing.T) {
		p := Project{Directory: "./testdata/discovermain", Quiet: true}
		ctx := context.Background()
		targets, err := p.ListFuzzTargets(ctx, "...")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []Target{{
			Name:        "FuzzTarget",
			Package:     "discovermain",
			RootPackage: "discovermain",
		}}, targets)
	})
}
