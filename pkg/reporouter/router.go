// Package reporouter provides a local, preview-only repository selector.
// It never fetches repositories, checks permissions, or executes tasks.
package reporouter

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type Repository struct {
	Name     string   `json:"name"`
	Summary  string   `json:"summary"`
	Keywords []string `json:"keywords,omitempty"`
	Disabled bool     `json:"disabled,omitempty"`
}

type Config struct {
	Candidates int     `json:"candidates"`
	Threshold  float64 `json:"threshold"`
	Margin     float64 `json:"margin"`
}

func DefaultConfig() Config { return Config{4, .9, .15} }

type Judgment struct {
	// One probability per candidate, followed by a none/ambiguous option.
	Probabilities []float64 `json:"probabilities"`
	Truncated     bool      `json:"truncated"`
	Tokens        int       `json:"tokens"`
	MS            float64   `json:"ms"`
}
type Judge interface {
	Choose(query string, candidates []Repository) (Judgment, error)
}

type Candidate struct {
	Name         string  `json:"name"`
	LexicalScore float64 `json:"lexical_score"`
}
type Result struct {
	Preview    bool        `json:"preview"`
	Status     string      `json:"status"` // candidate, suggest, abstain; never execution authority
	Suggested  string      `json:"suggested,omitempty"`
	Reason     string      `json:"reason"`
	Candidates []Candidate `json:"candidates"`
	Judgment   *Judgment   `json:"judgment,omitempty"`
	Confidence float64     `json:"confidence,omitempty"`
	Margin     float64     `json:"margin,omitempty"`
}

type document struct {
	repo   Repository
	terms  map[string]float64
	length float64
}
type Index struct {
	docs    []document
	df      map[string]int
	average float64
}

var namePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,38}/[A-Za-z0-9_.-]{1,100}$`)

func terms(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

// New snapshots only caller-supplied, enabled repositories. The caller must
// filter this registry by access permissions before creating the index.
func New(repos []Repository) (*Index, error) {
	if len(repos) == 0 || len(repos) > 4096 {
		return nil, fmt.Errorf("require 1..4096 repositories")
	}
	idx := &Index{df: map[string]int{}}
	seen := map[string]bool{}
	total := 0
	for _, r := range repos {
		key := strings.ToLower(r.Name)
		if !namePattern.MatchString(r.Name) || seen[key] || len(r.Summary) > 2048 || len(r.Keywords) > 32 {
			return nil, fmt.Errorf("invalid or duplicate repository metadata")
		}
		seen[key] = true
		total += len(r.Name) + len(r.Summary)
		for _, s := range r.Keywords {
			if len(s) > 128 {
				return nil, fmt.Errorf("keyword exceeds 128 bytes")
			}
			total += len(s)
		}
		if total > 4<<20 {
			return nil, fmt.Errorf("catalog exceeds 4 MiB")
		}
		if r.Disabled {
			continue
		}
		r.Keywords = append([]string(nil), r.Keywords...)
		d := document{repo: r, terms: map[string]float64{}}
		for _, part := range []struct {
			text   string
			weight float64
		}{{r.Name, 3}, {strings.Join(r.Keywords, " "), 2}, {r.Summary, 1}} {
			for _, term := range terms(part.text) {
				d.terms[term] += part.weight
				d.length += part.weight
			}
		}
		for term := range d.terms {
			idx.df[term]++
		}
		idx.average += d.length
		idx.docs = append(idx.docs, d)
	}
	if len(idx.docs) == 0 {
		return nil, fmt.Errorf("catalog has no enabled repositories")
	}
	idx.average /= float64(len(idx.docs))
	return idx, nil
}

func (idx *Index) Preview(query string, c Config, j Judge) (Result, error) {
	r := Result{Preview: true, Status: "abstain", Candidates: []Candidate{}}
	if strings.TrimSpace(query) == "" || len(query) > 8192 || c.Candidates < 1 || c.Candidates > 8 || math.IsNaN(c.Threshold) || c.Threshold < .5 || c.Threshold > 1 || math.IsNaN(c.Margin) || c.Margin < 0 || c.Margin > 1 {
		return r, fmt.Errorf("invalid query or preview policy")
	}
	queryTerms := map[string]bool{}
	for _, term := range terms(query) {
		queryTerms[term] = true
	}
	sortedTerms := make([]string, 0, len(queryTerms))
	for term := range queryTerms {
		sortedTerms = append(sortedTerms, term)
	}
	sort.Strings(sortedTerms)
	type scored struct {
		position int
		score    float64
	}
	hits := make([]scored, 0, len(idx.docs))
	for i, d := range idx.docs {
		score := 0.0
		for _, term := range sortedTerms {
			tf := d.terms[term]
			if tf == 0 {
				continue
			}
			idf := math.Log(1 + (float64(len(idx.docs)-idx.df[term])+.5)/(float64(idx.df[term])+.5))
			score += idf * tf * 2.2 / (tf + 1.2*(.25+.75*d.length/idx.average))
		}
		if score > 0 {
			hits = append(hits, scored{i, score})
		}
	}
	sort.Slice(hits, func(i, k int) bool {
		if hits[i].score == hits[k].score {
			return idx.docs[hits[i].position].repo.Name < idx.docs[hits[k].position].repo.Name
		}
		return hits[i].score > hits[k].score
	})
	if len(hits) == 0 {
		r.Reason = "no_lexical_evidence"
		return r, nil
	}
	hits = hits[:min(c.Candidates, len(hits))]
	choices := make([]Repository, len(hits))
	for i, h := range hits {
		choices[i] = idx.docs[h.position].repo
		r.Candidates = append(r.Candidates, Candidate{choices[i].Name, h.score})
	}
	r.Suggested = choices[0].Name
	r.Status = "candidate"
	r.Reason = "lexical_only_not_calibrated"
	if j == nil {
		return r, nil
	}
	// This preview uses the installed English checkpoint. Confidence cannot
	// protect against unsupported languages; never silently translate metadata.
	if nonLatin(query) {
		r.Reason = "language_unvalidated"
		return r, nil
	}
	for _, repo := range choices {
		if nonLatin(repo.Summary) {
			r.Reason = "metadata_language_unvalidated"
			return r, nil
		}
	}
	p, err := j.Choose(query, choices)
	if err != nil {
		r.Reason = "inference_unavailable"
		return r, nil
	}
	r.Suggested = ""
	r.Status = "abstain"
	if len(p.Probabilities) != len(choices)+1 {
		r.Reason = "invalid_judgment"
		return r, nil
	}
	sum := 0.0
	best, second := -1.0, -1.0
	winner := 0
	for i, v := range p.Probabilities {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			r.Reason = "invalid_judgment"
			return r, nil
		}
		sum += v
		if v > best {
			second = best
			best = v
			winner = i
		} else if v > second {
			second = v
		}
	}
	if math.Abs(sum-1) > 1e-5 || p.Tokens < 0 || math.IsNaN(p.MS) || math.IsInf(p.MS, 0) || p.MS < 0 {
		r.Reason = "invalid_judgment"
		return r, nil
	}
	r.Judgment = &p
	r.Confidence = best
	r.Margin = best - second
	if p.Truncated {
		r.Reason = "input_truncated"
		return r, nil
	}
	if winner == len(choices) {
		r.Reason = "none_or_ambiguous"
		return r, nil
	}
	if best < c.Threshold || r.Margin < c.Margin {
		r.Reason = "confidence_or_margin"
		return r, nil
	}
	r.Suggested = choices[winner].Name
	r.Status = "suggest"
	r.Reason = "experimental_laya_choice"
	return r, nil
}
func nonLatin(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.In(r, unicode.Latin) {
			return true
		}
	}
	return false
}
