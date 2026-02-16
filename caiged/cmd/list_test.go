package cmd

import "testing"

func TestNonEmptyLines(t *testing.T) {
	got := nonEmptyLines("\nfoo\n\n bar \n")
	if len(got) != 2 || got[0] != "foo" || got[1] != "bar" {
		t.Fatalf("unexpected parsed lines: %v", got)
	}
}
