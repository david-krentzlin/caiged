package cmd

import "testing"

func TestNormalizeOptionsNoMountGHWins(t *testing.T) {
	opts := RunOptions{MountGH: true, MountGHRW: true, NoMountGH: true}
	normalized := normalizeOptions(opts)

	if normalized.MountGH {
		t.Fatalf("expected MountGH to be false when NoMountGH is set")
	}
	if normalized.MountGHRW {
		t.Fatalf("expected MountGHRW to be false when NoMountGH is set")
	}
}

func TestNormalizeOptionsMountGHRWEnablesMountGH(t *testing.T) {
	opts := RunOptions{MountGH: false, MountGHRW: true}
	normalized := normalizeOptions(opts)

	if !normalized.MountGH {
		t.Fatalf("expected MountGH to be true when MountGHRW is set")
	}
}
