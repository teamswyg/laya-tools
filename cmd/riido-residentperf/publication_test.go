package main

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// Added after the one official replay. This guard reads the frozen public
// observation; it never reruns a child or retroactively enters the31-file
// pre-collection source manifest. Historical platform bytes are not regenerated.
func TestPublishedResident56dRetainsFrozenPlanSourcesAndAllObservations(t *testing.T) {
	repo := residentRepoRoot(t)
	read := func(name string) []byte {
		t.Helper()
		raw, e := os.ReadFile(filepath.Join(repo, name))
		if e != nil {
			t.Fatal("public resident record unavailable")
		}
		return raw
	}
	const base = "experiments/short-claim/"
	planRaw := read(base + "execution-plan-56d.json")
	if hashBytes(planRaw) != "8c0b7a5b697e52d4afffbb5e4ca2f9b5d3c992a31d4eae4f9767bf57e2dbfed2" {
		t.Fatal("frozen resource plan changed")
	}
	var p Plan
	if strictJSON(planRaw, &p) != nil || p.Rows != plannedRows() || len(p.ImplementationFiles) != 31 || p.SourceCommit != "60b03500fab417470d48d1a096be4ca1f43fafc2" {
		t.Fatal("published resource policy differs")
	}
	for _, pin := range p.ImplementationFiles {
		if hashBytes(read(pin.Path)) != pin.SHA256 {
			t.Fatal("pre-collection source changed")
		}
	}
	for _, pin := range []FilePin{{base + "wire-corpus-56d.json", p.CorpusSHA256}, {base + "build-recipe-56d.json", p.BuildRecipeSHA256}, {legacyPath, p.LegacyInputSHA256}, {typedPath, p.TypedInputSHA256}} {
		if hashBytes(read(pin.Path)) != pin.SHA256 {
			t.Fatal("public input or recipe changed")
		}
	}
	var freeze struct {
		State    string `json:"state"`
		Source   string `json:"preparation_source_commit"`
		GitBlobs int    `json:"source_commit_git_blob_matches"`
		Replays  int    `json:"official_resource_replays_at_freeze"`
		Public   []struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
			Bytes  int    `json:"bytes"`
		} `json:"public_artifacts"`
		Preserved []struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
			Bytes  int    `json:"bytes"`
		} `json:"preserved_artifacts"`
		Docs []struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
			Bytes  int    `json:"bytes"`
		} `json:"preparation_documents"`
	}
	if json.Unmarshal(read(base+"freeze-56d.json"), &freeze) != nil || freeze.State != "frozen_before_official_resource_replay" || freeze.Source != p.SourceCommit || freeze.GitBlobs != 31 || freeze.Replays != 0 || len(freeze.Public) != 4 || len(freeze.Preserved) != 9 || len(freeze.Docs) != 4 {
		t.Fatal("before-data freeze record differs")
	}
	for _, pins := range [][]struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
		Bytes  int    `json:"bytes"`
	}{freeze.Public, freeze.Preserved, freeze.Docs} {
		for _, pin := range pins {
			b := read(pin.Path)
			if len(b) != pin.Bytes || hashBytes(b) != pin.SHA256 {
				t.Fatal("historical or frozen artifact changed")
			}
		}
	}
	raw := read(base + "results-56d.json")
	if hashBytes(raw) != "d24a50cf71968bb0336f58812ce02a647facff28cb40595e6cebbcd9587c4eaf" || len(raw) != 81471 {
		t.Fatal("actual observation bytes changed")
	}
	var r Report
	if strictJSON(raw, &r) != nil || r.SourceCommit != p.SourceCommit || r.PlanSHA256 != hashBytes(planRaw) || r.CorpusSHA256 != p.CorpusSHA256 || r.ChildBinarySHA256 != p.ChildBinarySHA256 || r.DriverBinarySHA256 != p.DriverBinarySHA256 || r.PlannedRows != 24 || r.PlannedRequests != 25080 || r.UniqueFeaturePayloads != 48 || r.Fits != 0 || r.Models != 0 || r.ProtectedFinalReads != 0 || r.CacheHits != 0 || r.SavingsClaim || r.GlobalDeadlineExceeded {
		t.Fatal("published observation scope differs")
	}
	for i, row := range r.Rows {
		if row.Index != i || row.Plan != p.Rows[i] || row.Status != "complete" || row.Error != "" || row.CleanupError != "" || row.WaitError != "" || row.StdoutError != "" || row.StderrError != "" || !row.Started || !row.EOFObserved || !row.Reaped || row.WaitCalls != 1 || !row.StdoutJoined || !row.StderrJoined || !row.WatcherJoined || !row.CleanupComplete || row.StderrBytes != 0 || row.StderrLimitExceeded || !row.ExitAvailable || row.ExitCode != 0 || row.Cancelled || row.ChildDeadlineExceeded || row.ContextDeadlineExceeded || row.CleanupWallNS > 1e9 || !row.Resource.CPUAvailable || !row.Resource.RSSAvailable || row.Resource.LifetimePeakRSSUnit != "bytes" || row.Resource.LifetimePeakRSSRaw != row.Resource.LifetimePeakRSSBytes || row.ActualInputStreamSHA256 != row.ExpectedInputStreamSHA256 || row.ActualOutputStreamSHA256 != row.ExpectedOutputStreamSHA256 || !row.CPUPerRequestAvailable || row.CPURequestDenominator != 1045 {
			t.Fatal("actual row or failure accounting changed")
		}
		for phase, c := range row.Phases {
			want := [3]int{1, 20, 1024}[phase]
			if c.Planned != want || c.Attempted != want || c.FullyWritten != want || c.Received != want || c.Validated != want || c.Failed != 0 || c.Incomplete != 0 || c.NotAttempted != 0 {
				t.Fatal("repeated request denominator changed")
			}
		}
		if row.TimedRTT.Count != 1024 || math.Abs(row.CPUSecondsPerRequest-(row.Resource.UserCPUSeconds+row.Resource.SystemCPUSeconds)/1045) > 1e-12 {
			t.Fatal("resource denominator changed")
		}
	}
	if r.Rows[0].StartupAndFirstResponseNS != 577130000 || r.ReplayWallNS != 1482910833 || !r.Controller.Available || r.Controller.LifetimePeakRSSBytes != 34717696 || r.PureStartup != "unobserved" || r.PureChildProcessing != "unobserved" || r.PurePipe != "unobserved" {
		t.Fatal("first observation or resource scope changed")
	}
	// Canonical emitted report must not hide fixed-array elements either.
	canonical, e := encodePublic(r, maxReportBytes)
	if e != nil || !bytes.Equal(canonical, raw) {
		t.Fatal("published report is not canonical")
	}
	total := 0
	for _, name := range [7]string{"wire-corpus-56d.json", "preparation-56d.json", "execution-plan-56d.json", "build-recipe-56d.json", "freeze-56d.json", "results-56d.json", "summary-56d.json"} {
		total += len(read(base + name))
	}
	if total != 252423 || total > resultLimit {
		t.Fatal("public JSON aggregate changed")
	}
}
