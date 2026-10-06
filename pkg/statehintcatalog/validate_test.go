package statehintcatalog

import (
	"errors"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

func TestValidateBeforeNarrowingCatalog(t *testing.T) {
	scope := Scope{WorkspaceRef: "synthetic-workspace", OwnerRef: "synthetic-owner"}
	snapshot := Snapshot{Scope: scope, Works: []Work{{Ref: "synthetic-work", CanonicalVersion: "1"}}, Catalog: Catalog{Emojis: []statehint.EmojiCandidate{{Intent: statehint.Blocker, Code: "not-an-emoji-code"}}}}
	// An invalid binding outside the three-display scope must remain invalid
	// before any scope filter has an opportunity to erase it.
	if err := ValidateSnapshot(snapshot, scope, []string{"synthetic-work"}); !errors.Is(err, ErrSnapshot) {
		t.Fatalf("hidden invalid metadata: %v", err)
	}
	snapshot.Catalog.Emojis[0].Code = "1f6a7"
	if err := ValidateSnapshot(snapshot, scope, []string{"synthetic-work"}); err != nil {
		t.Fatal(err)
	}
	snapshot.Works[0].Ref = "different-work"
	if err := ValidateSnapshot(snapshot, scope, []string{"synthetic-work"}); !errors.Is(err, ErrSnapshot) {
		t.Fatalf("unrequested work accepted: %v", err)
	}
}

func TestSnapshotRequestMustBeBoundedAndUnique(t *testing.T) {
	scope := Scope{WorkspaceRef: "synthetic-workspace", OwnerRef: "synthetic-owner"}
	for _, refs := range [][]string{nil, {"duplicate", "duplicate"}, {""}} {
		if err := ValidateSnapshot(Snapshot{Scope: scope}, scope, refs); !errors.Is(err, ErrInput) {
			t.Fatalf("request admitted: %v", err)
		}
	}
}
