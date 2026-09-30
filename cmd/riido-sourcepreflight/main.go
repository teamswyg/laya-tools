package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

// boundedBuffer caps captured API output. Error output is kept separate and is
// never copied into a public report (it may contain local configuration details).
type boundedBuffer struct {
	b     []byte
	limit int
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(b.b)+len(p) > b.limit {
		return 0, fmt.Errorf("API output limit")
	}
	b.b = append(b.b, p...)
	return len(p), nil
}
func run() error {
	selection := flag.String("selected", ".cache/evaluation-groups-29/selected.json", "private frozen selection")
	cache := flag.String("cache", ".cache/source-preflight-29", "private root metadata cache")
	out := flag.String("out", "", "new aggregate report file")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("require new report path")
	}
	// Recompute frozen selection from pinned projections; do not trust a modified
	// selection file to choose easier or more convenient source samples.
	full, e := sweaudit.Read(".cache/real-task-full-28.jsonl", sweaudit.FullProjectionSHA256)
	if e != nil {
		return e
	}
	multi, e := sweaudit.Read(".cache/real-task-multilingual-28.jsonl", sweaudit.MultilingualProjectionSHA256)
	if e != nil {
		return e
	}
	groupReport, selected, e := sweaudit.GroupEvaluation(full, multi)
	if e != nil {
		return e
	}
	expected, e := json.MarshalIndent(selected, "", "  ")
	if e != nil {
		return e
	}
	b, e := os.ReadFile(*selection)
	if e != nil {
		return e
	}
	if string(b) != string(append(expected, '\n')) {
		return fmt.Errorf("selection differs from pinned grouping")
	}
	samples, e := sweaudit.PreflightSamples(selected)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(*cache, 0700); e != nil {
		return e
	}
	type result struct {
		observation sweaudit.RootObservation
		err         error
	}
	slots := make([]result, len(samples))
	jobs := make(chan int)
	done := make(chan struct{}, 2)
	for worker := 0; worker < 2; worker++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for i := range jobs {
				s := samples[i]
				h := sha256.Sum256([]byte(s.Repository + "\x00" + s.BaseCommit))
				path := filepath.Join(*cache, hex.EncodeToString(h[:])+".json")
				b, e := os.ReadFile(path)
				if os.IsNotExist(e) {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					cmd := exec.CommandContext(ctx, "gh", "api", "repos/"+s.Repository+"/git/trees/"+s.BaseCommit)
					stdout, stderr := &boundedBuffer{limit: 2 << 20}, &boundedBuffer{limit: 4096}
					cmd.Stdout = stdout
					cmd.Stderr = stderr
					e = cmd.Run()
					cancel()
					if e == nil {
						b = stdout.b
						if _, e = sweaudit.InspectRoot(s.Repository, b); e == nil {
							e = os.WriteFile(path, b, 0600)
						}
					}
				}
				if e != nil {
					slots[i] = result{sweaudit.RootObservation{Repository: s.Repository}, fmt.Errorf("root metadata unavailable")}
					continue
				}
				observation, e := sweaudit.InspectRoot(s.Repository, b)
				slots[i] = result{observation, e}
			}
		}()
	}
	for i := range samples {
		jobs <- i
	}
	close(jobs)
	<-done
	<-done
	report := struct {
		Schema, PlanSHA256, SelectedSHA256                                           string
		SnapshotsSampled, Available, ResponseBytes, RootLicenseCandidateRepositories int
		Observations                                                                 []sweaudit.RootObservation
		AllSelectedSnapshotsChecked, LicensesReviewed, ProductionReady               bool
	}{Schema: "riido-source-preflight-v1", PlanSHA256: sweaudit.GroupsPlanSHA256, SelectedSHA256: groupReport.SelectedSHA256, SnapshotsSampled: len(samples)}
	for _, slot := range slots {
		if slot.err == nil {
			report.Available++
			report.ResponseBytes += slot.observation.ResponseBytes
			if len(slot.observation.LicenseCandidates) > 0 {
				report.RootLicenseCandidateRepositories++
			}
		}
		report.Observations = append(report.Observations, slot.observation)
	}
	b, e = json.MarshalIndent(report, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	_, e = f.Write(append(b, '\n'))
	ce := f.Close()
	if e != nil {
		return e
	}
	return ce
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
