package githubmeta

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFetchClassifiesResponseBoundWithoutRawOutput(t *testing.T) {
	dir := t.TempDir()
	if e := os.WriteFile(filepath.Join(dir, "gh"), []byte("#!/bin/sh\nprintf '12345'\n"), 0700); e != nil {
		t.Fatal(e)
	}
	t.Setenv("PATH", dir)
	if _, e := fetchBounded(context.Background(), "original-fixture", 3); !errors.Is(e, ErrResponseLimit) {
		t.Fatalf("wrong error classification: %v", e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := fetchBounded(ctx, "original-fixture", 3); !errors.Is(e, context.Canceled) {
		t.Fatalf("wrong cancellation classification: %v", e)
	}
}
