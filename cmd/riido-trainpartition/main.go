// riido-trainpartition freezes identity-only roles before inspecting outcomes.
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
	"slices"
	"strings"
	"time"

	"github.com/teamswyg/laya-tools/internal/githubmeta"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

const projectionSHA = "82419343bf7a06daffe79b2145800648e776f971a0f6f873fd11277602a87f51"
const inventorySHA = "4bd5026f1f903a6d28322988e5b24e00a194b51348b52202940ba92eadc11ef5"
const poolSHA = "db3270fc963c77be985059b0514faea5821a5c42c5575b3b62f74c65a40603e8"
const planSHA = "a9d629db2374687f2ec91daf45b2a1847313e0b1c6256fa802cca9edfae597f5"

func writeJSON(path string, x any) error {
	b, e := json.MarshalIndent(x, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}
func run() error {
	out := flag.String("out", "", "new private output directory")
	online := flag.Bool("fetch-identities", false, "allow up to128 bounded public metadata requests")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("require output")
	}
	plan, e := os.ReadFile("experiments/training-partition/plan-37.json")
	if e != nil {
		return e
	}
	hash := sha256.Sum256(plan)
	if hex.EncodeToString(hash[:]) != planSHA {
		return fmt.Errorf("plan changed")
	}
	train, e := sweaudit.ReadTraining(".cache/training-identity-36.jsonl", projectionSHA)
	if e != nil {
		return e
	}
	full, e := sweaudit.Read(".cache/real-task-full-28.jsonl", sweaudit.FullProjectionSHA256)
	if e != nil {
		return e
	}
	multi, e := sweaudit.Read(".cache/real-task-multilingual-28.jsonl", sweaudit.MultilingualProjectionSHA256)
	if e != nil {
		return e
	}
	if len(train) != 19008 || len(full) != 2294 || len(multi) != 300 {
		return fmt.Errorf("row count changed")
	}
	eval := append(slices.Clone(full), multi...)
	prior := []string{"etcd-io/etcd", "kubernetes/kubernetes", "lxc/lxd"}
	previous, _, e := sweaudit.TrainingCandidates(train, eval, prior)
	if e != nil || previous.MembershipSHA256 != poolSHA {
		return fmt.Errorf("previous pool changed")
	}
	old, _, e := sweaudit.GroupEvaluation(full, multi)
	if e != nil || old.SelectedSHA256 != sweaudit.FrozenSelectionSHA256 {
		return fmt.Errorf("old evaluation changed")
	}
	names := slices.Clone(prior)
	for _, rows := range [][]sweaudit.Row{train, eval} {
		for _, r := range rows {
			names = append(names, strings.ToLower(r.Repository))
		}
	}
	slices.Sort(names)
	names = slices.Compact(names)
	calls := 0
	var last time.Time
	fetch := func(ctx context.Context, endpoint string) ([]byte, error) {
		if !*online || calls >= 128 {
			return nil, fmt.Errorf("identity fetch disabled or budget exhausted")
		}
		if wait := 300*time.Millisecond - time.Since(last); wait > 0 {
			time.Sleep(wait)
		}
		last = time.Now()
		calls++
		child, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		return githubmeta.Fetch(child, endpoint)
	}
	var identities []sweaudit.RepositoryIdentity
	for i, name := range names {
		x, e := githubmeta.Repository(context.Background(), name, ".cache/repository-identities-37", fetch)
		if e != nil {
			return fmt.Errorf("identity %d/%d unavailable: %w", i+1, len(names), e)
		}
		identities = append(identities, x)
		if (i+1)%10 == 0 {
			fmt.Fprintf(os.Stderr, "identities=%d/%d\n", i+1, len(names))
		}
	}
	inventory, e := json.Marshal(identities)
	if e != nil {
		return e
	}
	ih := sha256.Sum256(inventory)
	if hex.EncodeToString(ih[:]) != inventorySHA {
		return fmt.Errorf("frozen identity inventory changed; do not repartition")
	}
	report, tasks, e := sweaudit.PartitionTraining(train, eval, prior, identities)
	if e != nil {
		return e
	}
	result := struct {
		PlanSHA256, PreviousPoolSHA256, IdentityInventorySHA256                             string
		Identities                                                                          []sweaudit.RepositoryIdentity
		Partition                                                                           sweaudit.PartitionReport
		HistoricalLicensesReviewed, TrainingExecuted, FinalOutcomesInspected, ModelReleased bool
	}{planSHA, poolSHA, hex.EncodeToString(ih[:]), identities, report, false, false, false, false}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	if e = writeJSON(filepath.Join(*out, "results.json"), result); e != nil {
		return e
	}
	if report.Complete {
		if e = writeJSON(filepath.Join(*out, "membership.json"), tasks); e != nil {
			return e
		}
	}
	fmt.Fprintf(os.Stderr, "identity_requests=%d units=%d partition_complete=%v\n", calls, report.Units, report.Complete)
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
