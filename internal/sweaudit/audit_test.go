package sweaudit

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestIdentityAndOverlap(t *testing.T) {
	commit := strings.Repeat("a", 40)
	a := []Row{{"a", "one/repo", commit, "Fix\t THIS"}, {"b", "one/repo", commit, "different"}}
	b := []Row{{"a", "two/repo", commit, " fix this "}, {"c", "two/repo", commit, "새로운 요청"}}
	r := Analyze(a, b, []string{"FIX THIS", "fix this", "unrelated"}, []string{"two/repo"})
	if r.Combined.Rows != 4 || r.Combined.DistinctIDs != 3 || r.Combined.DistinctRequests != 3 || r.Combined.DistinctSnapshots != 2 || r.SharedSourceIDs != 1 || r.SharedSourceRequests != 1 || r.PriorDistinctRequests != 2 || r.SharedPriorRequests != 1 || !reflect.DeepEqual(r.SharedPriorRepositories, []string{"two/repo"}) {
		t.Fatalf("bad identity accounting: %+v", r)
	}
	if r.IndependentFinalEvaluationEstablished || r.ProductionReady || r.Meets2400DistinctStrings {
		t.Fatal("false eligibility")
	}
	a[0], a[1] = a[1], a[0]
	r2 := Analyze(a, b, nil, nil)
	if r.MembershipSHA256 != r2.MembershipSHA256 {
		t.Fatal("membership depends on row order")
	}
	a[0].Request += "!"
	if r.MembershipSHA256 == Analyze(a, b, nil, nil).MembershipSHA256 {
		t.Fatal("membership ignores text change")
	}
	encoded, _ := json.Marshal(r)
	if strings.Contains(string(encoded), "Fix") || strings.Contains(string(encoded), "새로운") {
		t.Fatal("raw request exported")
	}
}
func TestMalformedFieldsAndProjection(t *testing.T) {
	c := count([]Row{{}, {"id", "bad", "xyz", " \t"}})
	if c.EmptyIDs != 1 || c.EmptyRequests != 2 || c.InvalidRepositories != 2 || c.InvalidBaseCommits != 2 || c.DistinctRequests != 0 || c.DistinctSnapshots != 0 {
		t.Fatalf("bad malformed accounting: %+v", c)
	}
	for _, b := range []string{"", `{} {}`, `{"instance_id":3}`, `{"instance_id":"x","patch":"answer"}`, `{"eval_script":"never run"}`} {
		if _, e := parse([]byte(b)); e == nil {
			t.Fatalf("accepted invalid projection %q", b)
		}
	}
	if _, e := parse([]byte(`{"instance_id":"x","repo":"a/b","base_commit":"","problem_statement":"hello"}`)); e != nil {
		t.Fatal(e)
	}
}
