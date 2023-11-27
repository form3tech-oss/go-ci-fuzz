package nocorpus

import "testing"

func FuzzSubTarget(f *testing.F) {
	f.Add("a")
	f.Add("z")
	f.Fuzz(func(t *testing.T, in string) {
		if in == "z" {
			t.Fail()
		}
	})
}
