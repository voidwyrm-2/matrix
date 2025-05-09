package proc

import (
	"fmt"
	"testing"
)

func TestProcApplication(t *testing.T) {
	cases := [][3]string{
		{"one-two", "'-|0[", "one"},
		{"one-two-three", "4⌊", "two-three"},
		{"one-two-three", "3⌋2⌊", "e"},
		{"hello", "2]s", "l"},
		{"wow ok sure", "' |⎲", "sure"},
		{"wow ok sure", "⎳s", "e"},
	}

	for i, c := range cases {
		input, process, expect := c[0], c[1], c[2]

		result, err := Apply(fmt.Sprintf("TEST %d", i), input, process)
		if err != nil {
			t.Fatal(err.Error())
		} else if result != expect {
			t.Fatalf("expected result of `%s` + `%s` to be `%s`, but got `%s` instead", input, process, expect, result)
		}
	}
}
