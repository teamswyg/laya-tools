package sourcecohort

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func freezeMetadataPlan(t *testing.T) (fixture, Allocation) {
	t.Helper()
	f := allocationFixture(t)
	out := filepath.Join(f.root, "allocation-out")
	if _, err := Run(f.root, f.planFile, out, true); err != nil {
		t.Fatal(err)
	}
	var allocation Allocation
	ref := fileAt(t, f.root, "allocation-out/ALLOCATION.private.json")
	if err := Decode(readFixture(t, f.root, ref.Path), &allocation); err != nil {
		t.Fatal(err)
	}
	// These future files deliberately do not exist: provenance must fail first.
	unopened := File{"unopened.json", Hash([]byte("{}")), 2}
	f.plan.Allocation, f.plan.Creation, f.plan.Reviews, f.plan.SourceSchema = &ref, &unopened, &unopened, &unopened
	return f, allocation
}

func readFixture(t *testing.T, root, name string) []byte {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(root, name))
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func assertNoFutureManifestRead(t *testing.T, f fixture, out string) {
	t.Helper()
	starts, err := filepath.Glob(filepath.Join(out, "receipts", "*.start.json"))
	if err != nil || len(starts) == 0 {
		t.Fatal("missing retained metadata read receipts", err)
	}
	for _, path := range starts {
		var receipt start
		if err := Decode(readFixture(t, filepath.Dir(path), filepath.Base(path)), &receipt); err != nil {
			t.Fatal(err)
		}
		for _, future := range []*File{f.plan.Creation, f.plan.Reviews, f.plan.SourceSchema} {
			if receipt.Expected == future.SHA256 {
				t.Fatal("provenance failure reached a future manifest read")
			}
		}
	}
}

func TestFreezeRequiresPinnedProspectiveRegistry(t *testing.T) {
	f, allocation := freezeMetadataPlan(t)
	allocation.Registry.SHA256 = strings.Repeat("0", 64)
	a := jsonFile(t, f.root, "bad-allocation.json", allocation)
	f.plan.Allocation = &a
	p := jsonFile(t, f.root, "bad-freeze-plan.json", f.plan)
	out := filepath.Join(f.root, "blocked")
	s, err := Run(f.root, p, out, false)
	if err == nil || s.Expected != 400 || s.Verified != 0 || s.State != "blocked" {
		t.Fatal("unverified prospective registry reached material")
	}
	assertNoFutureManifestRead(t, f, out)
}

func TestFreezeRequiresNullSourcesInAllocationArtifact(t *testing.T) {
	f, allocation := freezeMetadataPlan(t)
	allocation.Rows[0].Source = &File{"unopened-source.txt", Hash([]byte("x")), 1}
	a := jsonFile(t, f.root, "bad-allocation.json", allocation)
	f.plan.Allocation = &a
	p := jsonFile(t, f.root, "bad-freeze-plan.json", f.plan)
	out := filepath.Join(f.root, "blocked")
	s, err := Run(f.root, p, out, false)
	if err == nil || s.Expected != 400 || s.Verified != 0 || s.State != "blocked" {
		t.Fatal("already-realized allocation artifact reached material")
	}
	if s.Code != "allocation_registry_binding" {
		t.Fatal("wrong provenance failure", s.Code)
	}
	assertNoFutureManifestRead(t, f, out)
}
