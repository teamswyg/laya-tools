package sourcecohort

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func multiReviewerFixture(t *testing.T) fixture {
	t.Helper()
	const fresh = "synthetic-checker-fresh"
	f := fullFixtureWithReviewer(t, func(i int, checker string) string {
		if i >= 200 {
			return fresh
		}
		return checker
	})
	// Deliberately unordered declarations must not reorder pinned plan inputs.
	f.plan.ReviewerAssignments = []ReviewerAssignment{{ReviewerID: fresh}, {ReviewerID: f.plan.CheckerID}}
	for i := 399; i >= 0; i-- {
		roster := 1
		if i >= 200 {
			roster = 0
		}
		f.plan.ReviewerAssignments[roster].SourceIDs = append(f.plan.ReviewerAssignments[roster].SourceIDs, syntheticID(i))
	}
	f.refresh(t)
	return f
}

func TestWhole400FreezeWithReviewerAssignments(t *testing.T) {
	f := multiReviewerFixture(t)
	before := marshal(f.plan)
	out := filepath.Join(f.root, "multi-reviewer-out")
	s, err := Run(f.root, f.planFile, out, false)
	if err != nil || s.State != "structural_provenance_freeze_ready" || s.Verified != 400 || s.Held != 0 || s.MeaningProven || s.ProviderIndependence != "unknown" {
		t.Fatalf("two declared reviewers failed: %+v %v", s, err)
	}
	var frozen Freeze
	if err := Decode(readFixture(t, out, "FREEZE.private.json"), &frozen); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, marshal(f.plan)) || !bytes.Equal(before, marshal(frozen.Inputs)) {
		t.Fatal("reviewer assignment declaration was reordered or changed")
	}
	for i, row := range frozen.Rows {
		want := f.plan.CheckerID
		if i >= 200 {
			want = "synthetic-checker-fresh"
		}
		if row.Review.ReviewerID != want {
			t.Fatalf("source %d assigned to wrong reviewer", i)
		}
	}
}

func TestReviewerAssignmentFailuresBlockWholeFreeze(t *testing.T) {
	f := multiReviewerFixture(t)
	basePlan, baseReviews := marshal(f.plan), marshal(f.reviews)
	tests := []struct {
		name string
		code string
		edit func(*fixture)
	}{
		{"missing-source", "reviewer_assignment_missing", func(f *fixture) {
			f.plan.ReviewerAssignments[0].SourceIDs = f.plan.ReviewerAssignments[0].SourceIDs[1:]
		}},
		{"duplicate-source-within", "reviewer_assignment_duplicate", func(f *fixture) {
			f.plan.ReviewerAssignments[0].SourceIDs[0] = f.plan.ReviewerAssignments[0].SourceIDs[1]
		}},
		{"duplicate-source-across", "reviewer_assignment_duplicate", func(f *fixture) {
			f.plan.ReviewerAssignments[0].SourceIDs[0] = f.plan.ReviewerAssignments[1].SourceIDs[0]
		}},
		{"unknown-source", "reviewer_assignment_source", func(f *fixture) { f.plan.ReviewerAssignments[0].SourceIDs[0] = "synthetic-unknown" }},
		{"invalid-source", "reviewer_assignment_source", func(f *fixture) { f.plan.ReviewerAssignments[0].SourceIDs[0] = "bad/source" }},
		{"reviewer-source-mismatch", "row_binding", func(f *fixture) { f.reviews.Rows[0].ReviewerID = "synthetic-checker-fresh" }},
		{"creator-as-reviewer", "reviewer_assignment_reviewer", func(f *fixture) { f.plan.ReviewerAssignments[0].ReviewerID = f.plan.AuthorID }},
		{"invalid-reviewer", "reviewer_assignment_reviewer", func(f *fixture) { f.plan.ReviewerAssignments[0].ReviewerID = "bad/reviewer" }},
		{"duplicate-reviewer", "reviewer_assignment_reviewer", func(f *fixture) { f.plan.ReviewerAssignments[0].ReviewerID = f.plan.CheckerID }},
		{"missing-primary", "reviewer_assignment_primary", func(f *fixture) { f.plan.ReviewerAssignments[1].ReviewerID = "synthetic-other-checker" }},
		{"empty-assignment", "reviewer_assignment_count", func(f *fixture) { f.plan.ReviewerAssignments[0].SourceIDs = []string{} }},
		{"oversize-assignment", "reviewer_assignment_count", func(f *fixture) {
			f.plan.ReviewerAssignments[0].SourceIDs = make([]string, 401)
		}},
		{"oversize-roster", "reviewer_assignment_count", func(f *fixture) {
			f.plan.ReviewerAssignments = make([]ReviewerAssignment, 401)
			for i := range f.plan.ReviewerAssignments {
				f.plan.ReviewerAssignments[i] = ReviewerAssignment{"synthetic-checker", []string{syntheticID(0)}}
			}
		}},
		{"wrong-receipt-actor", "receipt_binding", func(f *fixture) { replaceReadActor(t, f, 0, "synthetic-checker-fresh") }},
		{"missing-receipt-actor", "receipt_binding", func(f *fixture) { replaceReadActor(t, f, 0, "") }},
		{"legacy-mixed-reviewers", "row_binding", func(f *fixture) { f.plan.ReviewerAssignments = nil }},
		{"roster-chronology", "chronology", func(f *fixture) { f.reviews.Rows[0].ReviewedUTC = "2000-01-01T00:00:00Z" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := f
			if err := Decode(basePlan, &g.plan); err != nil {
				t.Fatal(err)
			}
			if err := Decode(baseReviews, &g.reviews); err != nil {
				t.Fatal(err)
			}
			tc.edit(&g)
			g.refresh(t)
			out := filepath.Join(g.root, "failure-"+tc.name)
			s, err := Run(g.root, g.planFile, out, false)
			if err == nil || s.State != "blocked" || s.Expected != 400 || s.Code != tc.code {
				t.Fatalf("wrong failure: %+v %v, want %s", s, err, tc.code)
			}
			if _, err := os.Stat(filepath.Join(out, "FREEZE.private.json")); !os.IsNotExist(err) {
				t.Fatal("invalid assignment exported a freeze")
			}
		})
	}
}

// Changing only the fake receipt actor also updates its pin and terminal start
// hash, so the rejection specifically exercises actor ownership.
func replaceReadActor(t *testing.T, f *fixture, i int, actor string) {
	t.Helper()
	v := &f.reviews.Rows[i]
	var a start
	var b terminal
	if err := Decode(readFixture(t, f.root, v.ReadStart.Path), &a); err != nil {
		t.Fatal(err)
	}
	if err := Decode(readFixture(t, f.root, v.ReadResult.Path), &b); err != nil {
		t.Fatal(err)
	}
	a.Actor = actor
	// Use separate fake paths so subtests keep the original positive fixture.
	path := "modified-read-" + actor + ".json"
	v.ReadStart = jsonFile(t, f.root, path, a)
	b.StartSHA = v.ReadStart.SHA256
	v.ReadResult = jsonFile(t, f.root, path+".result", b)
}

func TestAllocationRejectsReviewerAssignments(t *testing.T) {
	f := allocationFixture(t)
	f.plan.ReviewerAssignments = []ReviewerAssignment{{f.plan.CheckerID, []string{syntheticID(0)}}}
	f.planFile = jsonFile(t, f.root, "allocation-plan.json", f.plan)
	out := filepath.Join(f.root, "blocked-allocation")
	s, err := Run(f.root, f.planFile, out, true)
	if err == nil || s.Code != "allocation_reviewer_assignments" || s.State != "blocked" || s.Registered != 400 {
		t.Fatalf("allocation ignored reviewer assignments: %+v %v", s, err)
	}
	if _, err := os.Stat(filepath.Join(out, "ALLOCATION.private.json")); !os.IsNotExist(err) {
		t.Fatal("reviewer roster exported allocation")
	}
}

func TestLegacyPlanBytesAndOptionalReviewerAssignmentShape(t *testing.T) {
	// A fixed old plan provides a byte-for-byte compatibility check, independent
	// of temporary paths, dynamic receipt times, and fixture serialization.
	legacy := []byte(`{"schema":"riido-sourcecohort-plan-v1","registry":{"path":"registry.json","sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","bytes":0},"creation":null,"reviews":null,"source_schema":null,"allocation":null,"split_config":{"path":"recipe.json","sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","bytes":0},"author_id":"synthetic-author","checker_id":"synthetic-checker","input_max_bytes":4194304,"source_max_bytes":16384,"freeze_max_bytes":16777216,"output_max_bytes":33554432}` + "\n")
	var p Plan
	if err := Decode(legacy, &p); err != nil || !bytes.Equal(marshal(p), legacy) {
		t.Fatal("legacy plan bytes changed", err)
	}
	for _, optional := range []string{`"reviewer_assignments":null`, `"reviewer_assignments":[]`} {
		raw := bytes.Replace(legacy, []byte(`"schema":`), []byte(optional+`,"schema":`), 1)
		p.ReviewerAssignments = []ReviewerAssignment{{"old", []string{"old"}}}
		if err := Decode(raw, &p); err != nil || len(p.ReviewerAssignments) != 0 || !bytes.Equal(marshal(p), legacy) {
			t.Fatal("empty optional roster failed legacy reset", err)
		}
	}
	for _, assignment := range []string{
		`{"reviewer_id":"synthetic-checker","source_ids":["synthetic-000"],"extra":true}`,
		`{"Reviewer_id":"synthetic-checker","source_ids":["synthetic-000"]}`,
		`{"reviewer_id":"synthetic-checker","reviewer_id":"duplicate","source_ids":["synthetic-000"]}`,
		`{"source_ids":["synthetic-000"]}`,
		`{"reviewer_id":"synthetic-checker"}`,
		`{"reviewer_id":null,"source_ids":["synthetic-000"]}`,
		`{"reviewer_id":"synthetic-checker","source_ids":null}`,
		`{"reviewer_id":"synthetic-checker","source_ids":[null]}`,
	} {
		raw := bytes.Replace(legacy, []byte(`"schema":`), []byte(`"reviewer_assignments":[`+assignment+`],"schema":`), 1)
		if Decode(raw, &p) == nil {
			t.Fatalf("open or incomplete reviewer assignment accepted: %s", assignment)
		}
	}
}

func TestLegacyFreezeKeepsOptionalReceiptActor(t *testing.T) {
	f := fullFixture(t)
	replaceReadActor(t, &f, 0, "")
	f.refresh(t)
	s, err := Run(f.root, f.planFile, filepath.Join(f.root, "legacy-without-actor"), false)
	if err != nil || s.Verified != 400 || s.State != "structural_provenance_freeze_ready" {
		t.Fatalf("legacy actor behavior changed: %+v %v", s, err)
	}
}
