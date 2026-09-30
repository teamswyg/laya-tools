package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/teamswyg/laya-tools/internal/githubmeta"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

func run() error {
	cache := flag.String("cache", ".cache/source-preflight-29", "private root cache (reuse experiment29)")
	plan := flag.String("plan", "experiments/historical-roots/plan-30.json", "fixed plan")
	out := flag.String("out", "", "new local output directory; candidates.json is private")
	offline := flag.Bool("offline", false, "use existing validated cache only; never call GitHub")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("require new output directory")
	}
	b, e := os.ReadFile(*plan)
	if e != nil {
		return e
	}
	h := sha256.Sum256(b)
	if hex.EncodeToString(h[:]) != sweaudit.RootCoveragePlanSHA256 {
		return fmt.Errorf("plan hash mismatch")
	}
	full, e := sweaudit.Read(".cache/real-task-full-28.jsonl", sweaudit.FullProjectionSHA256)
	if e != nil {
		return e
	}
	multi, e := sweaudit.Read(".cache/real-task-multilingual-28.jsonl", sweaudit.MultilingualProjectionSHA256)
	if e != nil {
		return e
	}
	group, selected, e := sweaudit.GroupEvaluation(full, multi)
	if e != nil {
		return e
	}
	if len(selected) != 2400 || group.SelectedSHA256 != sweaudit.FrozenSelectionSHA256 {
		return fmt.Errorf("frozen selection mismatch")
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	type result struct {
		observation sweaudit.RootObservation
		candidates  []sweaudit.LicenseCandidate
	}
	slots := make([]result, len(selected))
	jobs := make(chan int)
	completed := make(chan int)
	for worker := 0; worker < 2; worker++ {
		go func() {
			var last time.Time
			fetch := func(ctx context.Context, endpoint string) ([]byte, error) {
				if *offline {
					return nil, fmt.Errorf("root not in offline cache")
				}
				delay := 250*time.Millisecond - time.Since(last)
				if delay > 0 {
					timer := time.NewTimer(delay)
					select {
					case <-timer.C:
					case <-ctx.Done():
						timer.Stop()
						return nil, ctx.Err()
					}
				}
				last = time.Now()
				return githubmeta.Fetch(ctx, endpoint)
			}
			for i := range jobs {
				s := selected[i]
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				b, e := githubmeta.Root(ctx, s.Repository, s.BaseCommit, *cache, fetch)
				cancel()
				slots[i].observation = sweaudit.RootObservation{Repository: s.Repository}
				if e == nil {
					observation, oe := sweaudit.InspectRoot(s.Repository, b)
					candidates, ce := sweaudit.RootCandidates(s.Repository, b)
					if oe == nil && ce == nil {
						slots[i] = result{observation, candidates}
					}
				}
				completed <- i
			}
		}()
	}
	go func() {
		for i := range selected {
			jobs <- i
		}
		close(jobs)
	}()
	for n := 1; n <= len(selected); n++ {
		<-completed
		if n%100 == 0 {
			fmt.Fprintf(os.Stderr, "checked=%d/%d\n", n, len(selected))
		}
	}
	var observations []sweaudit.RootObservation
	var candidates []sweaudit.LicenseCandidate
	for _, slot := range slots {
		observations = append(observations, slot.observation)
		candidates = append(candidates, slot.candidates...)
	}
	report := sweaudit.SummarizeRoots(observations, candidates)
	for i, obj := range []any{report, sweaudit.CompactCandidates(candidates)} {
		b, e = json.MarshalIndent(obj, "", "  ")
		if e != nil {
			return e
		}
		if e = os.WriteFile(filepath.Join(*out, []string{"results.json", "candidates.json"}[i]), append(b, '\n'), 0600); e != nil {
			return e
		}
	}
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
