package seed

import "testing"

func FuzzTarget(f *testing.F) {
	f.Add("z")
	f.Fuzz(func(t *testing.T, in string) {
		if in == "z" {
			t.Fail()
		}
	})
}
