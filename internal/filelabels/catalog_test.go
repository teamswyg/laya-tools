package filelabels

import (
	"github.com/teamswyg/laya-tools/internal/sweaudit"
	"reflect"
	"testing"
)

func TestCatalogMatchDistinguishesKinds(t *testing.T) {
	entries := []sweaudit.TreeEntry{{Path: "dir", Type: "tree", Mode: "040000"}, {Path: "dir/a", Type: "blob", Mode: "100644"}, {Path: "exec", Type: "blob", Mode: "100755"}, {Path: "link", Type: "blob", Mode: "120000"}, {Path: "sub", Type: "commit", Mode: "160000"}}
	paths := []string{"dir", "dir/a", "exec", "link", "missing", "sub"}
	before := append([]sweaudit.TreeEntry(nil), entries...)
	m, e := MatchCatalog(paths, entries)
	if e != nil {
		t.Fatal(e)
	}
	if m.Targets != 6 || m.RegularFiles != 2 || m.Symlinks != 1 || m.Submodules != 1 || m.Directories != 1 || m.Missing != 1 || !reflect.DeepEqual(m.MissingPaths, []string{"missing"}) {
		t.Fatalf("%+v", m)
	}
	if !reflect.DeepEqual(entries, before) {
		t.Fatal("mutated catalog")
	}
	for _, bad := range [][]string{{"missing", "dir"}, {"dir", "dir"}, {"../escape"}} {
		if _, e := MatchCatalog(bad, entries); e == nil {
			t.Fatal("accepted invalid targets")
		}
	}
	for _, bad := range [][]sweaudit.TreeEntry{{entries[1], entries[0]}, {entries[0], entries[0]}, {{Path: "dir", Type: "blob", Mode: "040000"}}} {
		if _, e := MatchCatalog([]string{"dir"}, bad); e == nil {
			t.Fatal("accepted invalid catalog")
		}
	}
}
