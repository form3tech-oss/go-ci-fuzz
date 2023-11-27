package new

import "testing"

func FuzzTarget(f *testing.F) {
	f.Add("a")
	f.Fuzz(func(t *testing.T, in string) {
		if in == "z" {
			t.Fail()
		}
	})
}
