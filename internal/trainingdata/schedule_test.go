package trainingdata

import (
	"reflect"
	"slices"
	"testing"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

func TestFairOrderUnevenQueuesAndExhaustion(t *testing.T) {
	rows := []sweaudit.TaskRole{member("a2", "train", "a/r"), member("a1", "train", "a/r"), member("b1", "train", "b/r"), member("v2", "validation", "v/r"), member("v1", "validation", "v/r"), member("w1", "validation", "w/r"), member("a3", "train", "a/r")}
	before := slices.Clone(rows)
	got, err := FairOrder(rows)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, r := range got {
		ids = append(ids, r.Task.ID)
	}
	if !slices.Equal(ids, []string{"a1", "v1", "b1", "w1", "a2", "v2", "a3"}) {
		t.Fatal(ids)
	}
	if !reflect.DeepEqual(rows, before) {
		t.Fatal("mutated source")
	}
	slices.Reverse(rows)
	again, err := FairOrder(rows)
	if err != nil || !reflect.DeepEqual(got, again) {
		t.Fatal("input order affects schedule", err)
	}
	for _, role := range []string{"train", "validation"} {
		one := []sweaudit.TaskRole{member("b", role, "b/r"), member("a", role, "a/r"), member("c", role, "a/r")}
		out, err := FairOrder(one)
		if err != nil || len(out) != 3 || out[0].Task.ID != "a" || out[1].Task.ID != "b" || out[2].Task.ID != "c" {
			t.Fatal("single role drain", out, err)
		}
	}
}
func TestFairOrderRejectsInvalidMembership(t *testing.T) {
	for _, rows := range [][]sweaudit.TaskRole{nil, {member("f", "final", "a/r")}, {member("a", "train", "a/r"), member("a", "train", "b/r")}, {member("a", "train", "a/r"), member("v", "validation", "a/r")}, {member("", "train", "a/r")}} {
		if _, err := FairOrder(rows); err == nil {
			t.Fatal("invalid schedule accepted")
		}
	}
}
