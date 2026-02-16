package cmd

import "testing"

func TestIsBenignTmuxNoServerError(t *testing.T) {
	if !isBenignTmuxNoServerError(assertErr("tmux: exit status 1 (no server running on /tmp/tmux-1000/default)")) {
		t.Fatalf("expected no-server tmux error to be benign")
	}
	if isBenignTmuxNoServerError(assertErr("tmux: permission denied")) {
		t.Fatalf("expected non no-server tmux error to be non-benign")
	}
}

type staticErr string

func (e staticErr) Error() string {
	return string(e)
}

func assertErr(msg string) error {
	return staticErr(msg)
}
