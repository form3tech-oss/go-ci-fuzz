package discovermain_test

import (
	"log"
	"testing"
)

func TestMain(m *testing.M) {
	log.Fatalln("this shall not be called")
}

func FuzzTarget(f *testing.F) {
	f.Add("a")
	f.Fuzz(func(t *testing.T, in string) {
		if in == "z" {
			t.Fail()
		}
	})
}
