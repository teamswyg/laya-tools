// Package pathclaim defines the bounded numeric FP32 experiment from plan 46.
// It does not establish source permissions, model usefulness or release approval.
package pathclaim

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/teamswyg/laya-tools/internal/pairlearn"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

const (
	PlanSHA256       = "4f01c6e4d84cf1f369930f65ef53e5ae7fc4688c0dee49c4245b832d6e842b30"
	MembershipSHA256 = "5ceb4aa4681784431606f2ff4410c9225b9639d88a18f618b9de70fa2c5da1fd"
	MaxRows          = 32768
	MaxPages         = 5000
	WeightScale      = 5016.0
	MinimumEligible  = 2400
	ReadinessSchema  = "riido-path-cost-claim-input-v1"
)

// Row contains only inference features and paired numeric targets. Repository
// and Group are private bookkeeping, never coefficient inputs. Group must be a
// global membership-component ordinal, not a separately numbered per-role ID.
// Only source-eligible, scoreable rows belong here; pending/excluded tasks remain
// in the runner's full coverage report, not fabricated negative training rows.
type Row struct {
	Role, Repository           string
	Group                      int
	Features                   [searchclaim.PathDimension]float64
	BaselinePages, HelperPages int
}

type RepositoryCount struct {
	Repository string
	Eligible   int
}

// Readiness pins a separately committed input seal. Digests are references to
// evidence verified by the runner; their syntax alone does not prove legal
// clearance. NumericRowsSHA256 is additionally recomputed by Fit from its actual
// numeric rows. No boolean field can bypass the coverage/readiness gates.
type Readiness struct {
	Schema, PlanSHA256, MembershipSHA256                                  string
	NumericRowsSHA256, AvailabilityCheckpointSHA256, SourceEvidenceSHA256 string
	InputSealSHA256, RunnerRevision                                       string
	TrainingCoverage, ValidationCoverage, ProtectedFinal                  int
	TrainingEligible, ValidationEligible                                  int
	ValidationRepositories                                                [5]RepositoryCount
}

// ValidationRepositories returns the original frozen membership spellings.
// The returned array is owned by the caller and cannot mutate a shared catalog.
func ValidationRepositories() [5]string {
	return [5]string{"Lightning-AI/lightning", "googleapis/google-cloud-python", "numpy/numpy", "pandas-dev/pandas", "pyca/cryptography"}
}

func trainingRepositories() [20]string {
	return [20]string{"DataDog/integrations-core", "PrefectHQ/prefect", "Qiskit/qiskit", "apache/airflow", "conda/conda", "dagster-io/dagster", "explosion/spaCy", "gitpython-developers/GitPython", "google/jax", "huggingface/transformers", "ipython/ipython", "jupyterlab/jupyterlab", "kubeflow/pipelines", "pantsbuild/pants", "pypa/pip", "python/typeshed", "ray-project/ray", "twisted/twisted", "wagtail/wagtail", "ytdl-org/youtube-dl"}
}

func Penalties() [5]float64 { return [5]float64{0, .25, 1, 4, 16} }
func Seeds() [2]uint64      { return [2]uint64{1729, 2718} }

func validPenalty(lambda float64) bool {
	for _, allowed := range Penalties() {
		if lambda == allowed {
			return true
		}
	}
	return false
}

func digest(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func repositoryName(s string) bool {
	if len(s) > 256 || strings.Count(s, "/") != 1 {
		return false
	}
	parts := strings.Split(s, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return false
		}
		for _, c := range part {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.') {
				return false
			}
		}
	}
	return true
}

func validateFeatures(features [searchclaim.PathDimension]float64) error {
	for _, v := range features {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return fmt.Errorf("invalid normalized path feature")
		}
	}
	return nil
}

func validateRows(rows []Row) error {
	if len(rows) == 0 || len(rows) > MaxRows {
		return fmt.Errorf("numeric row bound")
	}
	groups := make([]int, len(rows))
	type repoRole struct{ repo, role string }
	repos := make([]repoRole, len(rows))
	validationRepos := ValidationRepositories()
	trainingRepos := trainingRepositories()
	for i, row := range rows {
		if row.Role != "train" && row.Role != "validation" {
			return fmt.Errorf("nondevelopment role")
		}
		if !repositoryName(row.Repository) || row.Group < 0 || row.Group >= MaxRows {
			return fmt.Errorf("invalid numeric bookkeeping")
		}
		if (row.Role == "validation" && !slices.Contains(validationRepos[:], row.Repository)) || (row.Role == "train" && !slices.Contains(trainingRepos[:], row.Repository)) {
			return fmt.Errorf("repository role mismatch")
		}
		if row.BaselinePages < 1 || row.BaselinePages > MaxPages || row.HelperPages < 1 || row.HelperPages > MaxPages {
			return fmt.Errorf("paired page bound")
		}
		if err := validateFeatures(row.Features); err != nil {
			return err
		}
		groups[i], repos[i] = row.Group, repoRole{row.Repository, row.Role}
	}
	slices.Sort(groups)
	if len(slices.Compact(groups)) != len(rows) {
		return fmt.Errorf("duplicate membership component")
	}
	slices.SortFunc(repos, func(a, b repoRole) int { return strings.Compare(a.repo, b.repo) })
	for i := 1; i < len(repos); i++ {
		if repos[i].repo == repos[i-1].repo && repos[i].role != repos[i-1].role {
			return fmt.Errorf("repository role leakage")
		}
	}
	return nil
}

// RowsSHA256 hashes this typed projection in input order. It is distinct from a
// raw private source-file hash; the input seal must also pin that source evidence.
func RowsSHA256(rows []Row) (string, error) {
	if err := validateRows(rows); err != nil {
		return "", err
	}
	b, err := json.Marshal(rows)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func (r Readiness) Validate() error {
	if r.Schema != ReadinessSchema || r.PlanSHA256 != PlanSHA256 || r.MembershipSHA256 != MembershipSHA256 {
		return fmt.Errorf("input seal schema or frozen plan mismatch")
	}
	for _, s := range []string{r.NumericRowsSHA256, r.AvailabilityCheckpointSHA256, r.SourceEvidenceSHA256, r.InputSealSHA256} {
		if !digest(s, 64) {
			return fmt.Errorf("missing or invalid input evidence digest")
		}
	}
	if !digest(r.RunnerRevision, 40) {
		return fmt.Errorf("invalid runner revision")
	}
	if r.TrainingCoverage != 7335 || r.ValidationCoverage != 5686 || r.ProtectedFinal != 2402 {
		return fmt.Errorf("full cohort denominator mismatch")
	}
	if r.TrainingEligible < MinimumEligible || r.TrainingEligible > r.TrainingCoverage || r.ValidationEligible < MinimumEligible || r.ValidationEligible > r.ValidationCoverage {
		return fmt.Errorf("insufficient eligible training or validation coverage")
	}
	var total int
	coverage := [5]int{289, 722, 706, 3639, 330}
	for i, repo := range ValidationRepositories() {
		got := r.ValidationRepositories[i]
		if got.Repository != repo || got.Eligible < 100 || got.Eligible > coverage[i] {
			return fmt.Errorf("insufficient validation repository coverage")
		}
		total += got.Eligible
	}
	if total != r.ValidationEligible {
		return fmt.Errorf("validation repository count mismatch")
	}
	return nil
}

func VerifyReadiness(rows []Row, readiness Readiness) error {
	if err := readiness.Validate(); err != nil {
		return err
	}
	h, err := RowsSHA256(rows)
	if err != nil {
		return err
	}
	if h != readiness.NumericRowsSHA256 {
		return fmt.Errorf("numeric input digest mismatch")
	}
	var train, validation int
	var counts [5]int
	repos := ValidationRepositories()
	for _, row := range rows {
		if row.Role == "train" {
			train++
			continue
		}
		validation++
		counts[slices.Index(repos[:], row.Repository)]++
	}
	if train != readiness.TrainingEligible || validation != readiness.ValidationEligible {
		return fmt.Errorf("eligible row count differs from input seal")
	}
	for i, got := range counts {
		if got != readiness.ValidationRepositories[i].Eligible {
			return fmt.Errorf("repository rows differ from input seal")
		}
	}
	return nil
}

type DatasetCounts struct{ Rows, ZeroWeightRows int }

// Prepare builds owned contiguous columns without fitting. All rows are checked,
// including roles not selected for this Dataset. It preserves zero-weight rows,
// normalizes by the constant 5016 (never a data-dependent clipping operation),
// and requires positive total loss weight for the selected role.
func Prepare(rows []Row, role string, lambda float64) (pairlearn.Dataset, DatasetCounts, error) {
	var counts DatasetCounts
	var out pairlearn.Dataset
	if role != "train" && role != "validation" {
		return out, counts, fmt.Errorf("invalid preparation role")
	}
	if !validPenalty(lambda) {
		return out, counts, fmt.Errorf("unregistered penalty")
	}
	if err := validateRows(rows); err != nil {
		return out, counts, err
	}
	out.Split = "validation"
	if role == "train" {
		out.Split = "development"
	}
	out.Offsets = []int{0}
	var total float64
	for _, row := range rows {
		if row.Role != role {
			continue
		}
		for j, value := range row.Features {
			if value != 0 {
				out.Indices = append(out.Indices, uint16(j))
				out.Values = append(out.Values, value)
			}
		}
		out.Offsets = append(out.Offsets, len(out.Values))
		delta := float64(row.BaselinePages-row.HelperPages) - lambda
		weight := math.Abs(delta) / WeightScale
		if !(weight >= 0 && weight < 1) {
			return pairlearn.Dataset{}, DatasetCounts{}, fmt.Errorf("cost weight outside mathematical bound")
		}
		label := 0.0
		if delta > 0 {
			label = 1
		}
		out.Labels = append(out.Labels, label)
		out.Groups = append(out.Groups, row.Group)
		out.SampleWeights = append(out.SampleWeights, weight)
		counts.Rows++
		if weight == 0 {
			counts.ZeroWeightRows++
		}
		total += weight
	}
	if counts.Rows == 0 || !(total > 0) {
		return pairlearn.Dataset{}, counts, fmt.Errorf("empty or zero-weight selected dataset")
	}
	return out, counts, nil
}
