package statehintcatalog

import (
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

func validRef(ref string) bool {
	if ref == "" || len(ref) > 256 || !utf8.ValidString(ref) || strings.TrimSpace(ref) != ref {
		return false
	}
	for _, r := range ref {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
func validState(state statehint.State) bool {
	return state == statehint.Todo || state == statehint.Active || state == statehint.Done || state == statehint.Cancelled
}
func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func canonicalVersion(text string) (uint64, bool) {
	if text == "" || len(text) > 20 {
		return 0, false
	}
	version, err := strconv.ParseUint(text, 10, 64)
	return version, err == nil && version > 0 && version != ^uint64(0) && strconv.FormatUint(version, 10) == text
}

func validPrediction(p statehint.Prediction) bool {
	index, ok := statehint.IntentIndex(p.Intent)
	if !ok || !finite(p.Confidence) || !finite(p.Margin) || p.Margin < 0 || p.Margin > 1 {
		return false
	}
	if p.Source != statehint.Learned && p.Source != statehint.Untrained && p.Source != statehint.RuleSource {
		return false
	}
	if (p.Source == statehint.Learned) != (p.TrainingSteps > 0) {
		return false
	}
	var sum, runner float64
	for class, probability := range p.Probabilities {
		if !finite(probability) || probability < 0 || probability > 1 || probability > p.Probabilities[index]+1e-12 {
			return false
		}
		sum += probability
		if class != index && probability > runner {
			runner = probability
		}
	}
	return math.Abs(sum-1) < 1e-9 && math.Abs(p.Confidence-p.Probabilities[index]) < 1e-9 && math.Abs(p.Margin-(p.Confidence-runner)) < 1e-9
}

func validateSnapshot(snapshot Snapshot, scope Scope, requested map[string]bool) error {
	if snapshot.Scope != scope || len(snapshot.Works) > MaxWorks || len(snapshot.Catalog.Labels) > MaxLabels ||
		len(snapshot.Catalog.Statuses) > MaxStatuses || len(snapshot.Catalog.Bindings) > statehint.IntentCount ||
		len(snapshot.Catalog.Emojis) > MaxEmoji || len(snapshot.Catalog.PreferredStatusRefs) > 4 {
		return ErrSnapshot
	}
	seen := map[string]bool{}
	for _, work := range snapshot.Works {
		if !validRef(work.Ref) || !requested[work.Ref] || seen[work.Ref] || len(work.ExistingOpaqueLabels) > MaxLabels || len(work.CanonicalVersion) > 256 ||
			(work.State != "" && !validState(work.State)) || (work.CurrentEmoji != "" && !statehint.ValidEmojiCode(work.CurrentEmoji)) {
			return ErrSnapshot
		}
		seen[work.Ref] = true
		for _, ref := range work.ExistingOpaqueLabels {
			if !validRef(ref) {
				return ErrSnapshot
			}
		}
	}
	for _, label := range snapshot.Catalog.Labels {
		if !validRef(label.Ref) {
			return ErrSnapshot
		}
	}
	refs := 0
	for _, binding := range snapshot.Catalog.Bindings {
		if _, ok := statehint.IntentIndex(binding.Intent); !ok {
			return ErrSnapshot
		}
		refs += len(binding.LabelRefs)
		if refs > MaxLabels {
			return ErrSnapshot
		}
		for _, ref := range binding.LabelRefs {
			if !validRef(ref) {
				return ErrSnapshot
			}
		}
	}
	for _, emoji := range snapshot.Catalog.Emojis {
		if _, ok := statehint.IntentIndex(emoji.Intent); !ok || !statehint.ValidEmojiCode(emoji.Code) {
			return ErrSnapshot
		}
	}
	for _, status := range snapshot.Catalog.Statuses {
		if !validRef(status.Ref) || !validState(status.StateType) {
			return ErrSnapshot
		}
	}
	for state, ref := range snapshot.Catalog.PreferredStatusRefs {
		if !validState(state) || !validRef(ref) {
			return ErrSnapshot
		}
	}
	for _, window := range [3]ReadWindow{snapshot.ReadTimes.Works, snapshot.ReadTimes.Labels, snapshot.ReadTimes.Statuses} {
		if window.Started.IsZero() != window.Completed.IsZero() || !window.Started.IsZero() && window.Completed.Before(window.Started) {
			return ErrSnapshot
		}
	}
	return nil
}

func resolveLabels(catalog Catalog) ([]statehint.CatalogLabel, []MappingIssue) {
	labels := map[string]Label{}
	duplicates := map[string]bool{}
	var issues []MappingIssue
	for _, label := range catalog.Labels {
		if _, exists := labels[label.Ref]; exists {
			duplicates[label.Ref] = true
		}
		labels[label.Ref] = label
	}
	intentCounts := map[statehint.Intent]int{}
	refCounts := map[string]int{}
	for _, binding := range catalog.Bindings {
		intentCounts[binding.Intent]++
		for _, ref := range binding.LabelRefs {
			refCounts[ref]++
		}
	}
	resolved := make([]statehint.CatalogLabel, 0)
	for _, binding := range catalog.Bindings {
		if intentCounts[binding.Intent] != 1 {
			issues = append(issues, MappingIssue{Reason: "duplicate_intent_binding", Intent: binding.Intent})
			continue
		}
		for _, ref := range binding.LabelRefs {
			label, exists := labels[ref]
			reason := ""
			switch {
			case refCounts[ref] != 1 || duplicates[ref]:
				reason = "duplicate_label_ref"
			case !exists:
				reason = "missing_label_ref"
			case label.GroupContainer:
				reason = "group_is_not_label"
			case !label.Active:
				reason = "inactive_label_ref"
			}
			if reason != "" {
				issues = append(issues, MappingIssue{Reason: reason, Ref: ref, Intent: binding.Intent})
				continue
			}
			resolved = append(resolved, statehint.CatalogLabel{ID: ref, Intent: binding.Intent, Active: true})
		}
	}
	return resolved, issues
}

func resolveEmojis(catalog Catalog) ([]statehint.EmojiCandidate, []MappingIssue) {
	counts := map[statehint.Intent]int{}
	for _, emoji := range catalog.Emojis {
		counts[emoji.Intent]++
	}
	resolved := make([]statehint.EmojiCandidate, 0)
	var issues []MappingIssue
	for _, emoji := range catalog.Emojis {
		if counts[emoji.Intent] != 1 {
			issues = append(issues, MappingIssue{Reason: "ambiguous_emoji_binding", Intent: emoji.Intent})
			continue
		}
		resolved = append(resolved, emoji)
	}
	return resolved, issues
}

func targetState(intent statehint.Intent) statehint.State {
	switch intent {
	case statehint.Progress:
		return statehint.Active
	case statehint.CompletionReport:
		return statehint.Done
	case statehint.CancelRequest:
		return statehint.Cancelled
	default:
		return ""
	}
}

func resolveStatus(catalog Catalog, target statehint.State) (string, string) {
	if target == "" {
		return "", "not_a_state_intent"
	}
	counts := map[string]int{}
	for _, status := range catalog.Statuses {
		counts[status.Ref]++
	}
	var candidates []string
	for _, status := range catalog.Statuses {
		if status.Active && status.StateType == target && counts[status.Ref] == 1 {
			candidates = append(candidates, status.Ref)
		}
	}
	if preferred, configured := catalog.PreferredStatusRefs[target]; configured {
		for _, candidate := range candidates {
			if candidate == preferred {
				return candidate, "preferred_current_status"
			}
		}
		return "", "preferred_status_unavailable"
	}
	if len(candidates) == 1 {
		return candidates[0], "unique_current_status"
	}
	if len(candidates) > 1 {
		return "", "ambiguous_current_status"
	}
	return "", "current_status_unavailable"
}
