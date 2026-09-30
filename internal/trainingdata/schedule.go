package trainingdata

import (
	"fmt"
	"slices"
	"strings"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

// FairOrder alternates development roles and rotates repositories within each
// role. It changes acquisition order only, never membership or task identities.
// It has no input for text, targets, model scores or acquisition outcomes.
func FairOrder(rows []sweaudit.TaskRole) ([]sweaudit.TaskRole, error) {
	if len(rows) == 0 || len(rows) > 32768 {
		return nil, fmt.Errorf("schedule size bound")
	}
	ordered := slices.Clone(rows)
	ids := make([]string, len(rows))
	for i, r := range rows {
		if (r.Role != "train" && r.Role != "validation") || r.Task.Source != "train" || strings.TrimSpace(r.Task.ID) == "" || len(r.Task.ID) > 4096 || !repoPattern.MatchString(r.Task.Repository) || !commitPattern.MatchString(r.Task.BaseCommit) || !groupPattern.MatchString(r.Task.ComponentSHA256) {
			return nil, fmt.Errorf("invalid schedule member")
		}
		ids[i] = r.Task.ID
	}
	slices.Sort(ids)
	if len(slices.Compact(ids)) != len(rows) {
		return nil, fmt.Errorf("duplicate schedule member")
	}
	slices.SortFunc(ordered, func(a, b sweaudit.TaskRole) int {
		if a.Task.Repository != b.Task.Repository {
			return strings.Compare(a.Task.Repository, b.Task.Repository)
		}
		return strings.Compare(a.Task.ID, b.Task.ID)
	})
	type queue struct{ next, end int }
	var queues [2][]queue
	for i := 0; i < len(ordered); {
		end := i + 1
		for end < len(ordered) && ordered[end].Task.Repository == ordered[i].Task.Repository {
			if ordered[end].Role != ordered[i].Role {
				return nil, fmt.Errorf("repository spans roles")
			}
			end++
		}
		role := 0
		if ordered[i].Role == "validation" {
			role = 1
		}
		queues[role] = append(queues[role], queue{i, end})
		i = end
	}
	var cursor [2]int
	out := make([]sweaudit.TaskRole, 0, len(rows))
	for len(out) < len(rows) {
		for role := range 2 {
			for checked := 0; checked < len(queues[role]); checked++ {
				i := cursor[role]
				cursor[role] = (i + 1) % len(queues[role])
				q := &queues[role][i]
				if q.next == q.end {
					continue
				}
				out = append(out, ordered[q.next])
				q.next++
				break
			}
		}
	}
	return out, nil
}
