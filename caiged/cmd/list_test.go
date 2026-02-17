package cmd

import "testing"

func TestFilterNonEmpty(t *testing.T) {
	got := filterNonEmpty([]string{"", "foo", "", " bar ", ""})
	if len(got) != 2 || got[0] != "foo" || got[1] != "bar" {
		t.Fatalf("unexpected parsed lines: %v", got)
	}
}
