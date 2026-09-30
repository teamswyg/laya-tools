// Package sweaudit audits public task identity without scoring or executing tasks.
package sweaudit

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
)

const PlanSHA256 = "ad6c84f0b34525721807d39a5508b0a8ed0a69cd5d893b99be1cd72089ed4ec3"
const FullProjectionSHA256 = "abbe1b7ce1904ece24e267da0a89401e8817d48ad95a76d2954e5c14ba202601"
const MultilingualProjectionSHA256 = "425ca04943f415e3d65b223d4ab6676a95e85a24fd46727ea29096f86e376973"
const FullSourceSHA256 = "d4f5a245c75319fa8240c540674958c4d491e82edf274b144d43836bdcbc4567"
const MultilingualSourceSHA256 = "92abca7cb527b41a9f66d03a26ce441ff7319e3a49f985998fd56be4bb9b08b2"

type Row struct {
	ID         string `json:"instance_id"`
	Repository string `json:"repo"`
	BaseCommit string `json:"base_commit"`
	Request    string `json:"problem_statement"`
}
type RepositoryCount struct {
	Repository string
	Rows       int
}
type Counts struct {
	Rows, DistinctIDs, DistinctRequests, DistinctSnapshots           int
	EmptyIDs, EmptyRequests, InvalidRepositories, InvalidBaseCommits int
	Repositories                                                     []RepositoryCount
}
type Report struct {
	Schema, PlanSHA256, FullRevision, MultilingualRevision                                                           string
	FullSourceSHA256, MultilingualSourceSHA256, FullProjectionSHA256, MultilingualProjectionSHA256, MembershipSHA256 string
	Full, Multilingual, Combined                                                                                     Counts
	SharedSourceIDs, SharedSourceRequests                                                                            int
	PriorDistinctRequests, SharedPriorRequests                                                                       int
	SharedPriorRepositories                                                                                          []string
	Meets2400DistinctStrings, IndependentFinalEvaluationEstablished, ProductionReady                                 bool
}

var repoPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func normalized(s string) string    { return strings.Join(strings.Fields(strings.ToLower(s)), " ") }
func digest(b []byte) string        { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func distinct(xs []string) []string { slices.Sort(xs); return slices.Compact(xs) }
func intersect(a, b []string) []string {
	var out []string
	for i, j := 0, 0; i < len(a) && j < len(b); {
		if a[i] < b[j] {
			i++
		} else if b[j] < a[i] {
			j++
		} else {
			out = append(out, a[i])
			i++
			j++
		}
	}
	return out
}
func keys(rows []Row, field int) []string {
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		var s string
		switch field {
		case 0:
			s = r.ID
			if strings.TrimSpace(s) == "" {
				s = ""
			}
		case 1:
			s = normalized(r.Request)
			if s != "" {
				s = digest([]byte(s))
			}
		case 2:
			s = r.Repository
		case 3:
			if repoPattern.MatchString(r.Repository) && commitPattern.MatchString(r.BaseCommit) {
				s = r.Repository + "\x00" + r.BaseCommit
			}
		}
		if s != "" {
			out = append(out, s)
		}
	}
	return distinct(out)
}
func count(rows []Row) Counts {
	c := Counts{Rows: len(rows), DistinctIDs: len(keys(rows, 0)), DistinctRequests: len(keys(rows, 1)), DistinctSnapshots: len(keys(rows, 3))}
	var repos []string
	for _, r := range rows {
		if strings.TrimSpace(r.ID) == "" {
			c.EmptyIDs++
		}
		if normalized(r.Request) == "" {
			c.EmptyRequests++
		}
		if !repoPattern.MatchString(r.Repository) {
			c.InvalidRepositories++
		} else {
			repos = append(repos, r.Repository)
		}
		if !commitPattern.MatchString(r.BaseCommit) {
			c.InvalidBaseCommits++
		}
	}
	slices.Sort(repos)
	for i := 0; i < len(repos); {
		j := i + 1
		for j < len(repos) && repos[j] == repos[i] {
			j++
		}
		c.Repositories = append(c.Repositories, RepositoryCount{repos[i], j - i})
		i = j
	}
	return c
}
func Read(path, expected string) ([]Row, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, (32<<20)+1))
	if e != nil {
		return nil, e
	}
	if len(b) > 32<<20 || digest(b) != expected {
		return nil, fmt.Errorf("projection size or hash mismatch")
	}
	return parse(b)
}
func parse(b []byte) ([]Row, error) {
	s := bufio.NewScanner(bytes.NewReader(b))
	s.Buffer(make([]byte, 4096), 1<<20)
	var out []Row
	for s.Scan() {
		var r Row
		d := json.NewDecoder(bytes.NewReader(s.Bytes()))
		d.DisallowUnknownFields()
		if e := d.Decode(&r); e != nil {
			return nil, fmt.Errorf("invalid projected row %d", len(out)+1)
		}
		if e := d.Decode(new(any)); e != io.EOF {
			return nil, fmt.Errorf("extra row content")
		}
		out = append(out, r)
		if len(out) > 4096 {
			return nil, fmt.Errorf("too many rows")
		}
	}
	if e := s.Err(); e != nil {
		return nil, e
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty projection")
	}
	return out, nil
}

// Analyze checks exact identities and normalized strings, not semantic
// independence, benchmark contamination or fitness for a particular task.
func Analyze(full, multi []Row, priorRequests, priorRepos []string) Report {
	all := append(slices.Clone(full), multi...)
	r := Report{Schema: "riido-real-task-audit-v1", PlanSHA256: PlanSHA256, FullRevision: "c6fe717fd7a4c3ac1daa4055a4fd082c6a1d28a2", MultilingualRevision: "846e647b9f33c0b51b739d005d13d85493c9af09", FullSourceSHA256: FullSourceSHA256, MultilingualSourceSHA256: MultilingualSourceSHA256, FullProjectionSHA256: FullProjectionSHA256, MultilingualProjectionSHA256: MultilingualProjectionSHA256, Full: count(full), Multilingual: count(multi), Combined: count(all)}
	r.SharedSourceIDs = len(intersect(keys(full, 0), keys(multi, 0)))
	r.SharedSourceRequests = len(intersect(keys(full, 1), keys(multi, 1)))
	var prior []string
	for _, s := range priorRequests {
		if n := normalized(s); n != "" {
			prior = append(prior, digest([]byte(n)))
		}
	}
	prior = distinct(prior)
	r.PriorDistinctRequests = len(prior)
	r.SharedPriorRequests = len(intersect(keys(all, 1), prior))
	r.SharedPriorRepositories = intersect(keys(all, 2), distinct(slices.Clone(priorRepos)))
	var members []string
	for i, rows := range [][]Row{full, multi} {
		tag := []string{"full", "multilingual"}[i]
		for _, x := range rows {
			members = append(members, tag+"\x00"+x.ID+"\x00"+digest([]byte(normalized(x.Request))))
		}
	}
	slices.Sort(members)
	h := sha256.New()
	for _, m := range members {
		fmt.Fprintf(h, "%d:%s\n", len(m), m)
	}
	r.MembershipSHA256 = hex.EncodeToString(h.Sum(nil))
	r.Meets2400DistinctStrings = r.Combined.DistinctRequests >= 2400
	return r
}
