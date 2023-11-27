package nofindings

import "testing"

func FuzzTarget(f *testing.F) {
	f.Add("a")
	f.Fuzz(func(t *testing.T, in string) {
	})
}
