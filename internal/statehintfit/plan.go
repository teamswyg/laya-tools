// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintfit

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"sort"
	"strings"

	"github.com/teamswyg/laya-tools/internal/statehintcorpus"
	"github.com/teamswyg/laya-tools/pkg/statehint"
)

const parentSHA = "ae6761dc81501b39ae17f29b48f0f0a8b35305fe11e297b9a787af0f9636df05"
const rubricSHA = "b76a0e59bf046e82b0b868c56b9aba21a9d607c94bc42d5f56471a3947556215"
const rubricKO = "af440c0125ef73c456355b294d54c13d684485b128c38a334ad9508819b6adef"
const rubricEN = "7dfecbab58498fad10b6f2b3b62fe52f6dec95e8931b9fe7f9ec048c812e5b8d"

type Pin struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type Arm struct {
	Name string  `json:"name"`
	Warm bool    `json:"warm"`
	Rate float64 `json:"learning_rate"`
}

// Plan is a concrete ready-to-run record, distinct from the earlier proposal.
// Every field is required. Paths are local relative paths, never URLs.
type Plan struct {
	Schema       string  `json:"schema"`
	SourceCommit string  `json:"source_commit"`
	BinarySHA    string  `json:"binary_sha256"`
	Sources      []Pin   `json:"sources"`
	Rubric       Pin     `json:"rubric_receipt"`
	Definitions  [2]Pin  `json:"rubric_definitions_ko_en"`
	Manifest     Pin     `json:"family_manifest"`
	Audit        Pin     `json:"authoring_audit"`
	Parent       Pin     `json:"parent"`
	Partitions   [4]Pin  `json:"partitions_train_validation_calibration_test"`
	Arms         [4]Arm  `json:"arms"`
	Seed         int64   `json:"seed"`
	Batch        int     `json:"batch_size"`
	Epochs       int     `json:"epochs"`
	Decay        float64 `json:"weight_decay"`
	Confidence   float64 `json:"confidence"`
	Margin       float64 `json:"margin"`
	NLLFloor     float64 `json:"nll_probability_floor"`
	FalseCost    [3]int  `json:"false_cost_progress_completion_question"`
	Coverage     float64 `json:"correct_coverage_floor"`
	Support      int     `json:"correct_family_support_floor"`
	Lineages     int     `json:"correct_declared_lineage_support_floor"`
	Precision    float64 `json:"completion_precision_floor"`
	AbstainCost  float64 `json:"all_abstain_cost"`
	Temperature  [2]int  `json:"temperature_integer_grid_tenths_start_end"`
}

type FamilyPin struct {
	ID        string           `json:"family_id"`
	Lineage   string           `json:"leakage_group_id"`
	Partition string           `json:"partition"`
	Expected  statehint.Intent `json:"expected_intent"`
	RowIDs    [2]string        `json:"row_ids_ko_en"`
	TextSHA   [2]string        `json:"text_sha256_ko_en"`
}

type Manifest struct {
	Schema    string      `json:"schema"`
	RubricSHA string      `json:"rubric_receipt_sha256"`
	Families  []FamilyPin `json:"families"`
}

// Audit fields are recorded reviewer claims bound to this exact manifest.
// They are not authenticated external evidence or human ground truth.
type Audit struct {
	Schema                string `json:"schema"`
	ManifestSHA           string `json:"manifest_sha256"`
	RubricSHA             string `json:"rubric_receipt_sha256"`
	OriginalSynthetic     bool   `json:"original_synthetic"`
	PrivateExcluded       bool   `json:"private_data_excluded"`
	ExposedExcluded       bool   `json:"exposed_evaluation_excluded"`
	RightsReviewed        bool   `json:"rights_reviewed"`
	SemanticReviewed      bool   `json:"semantic_reviewed"`
	BilingualReviewed     bool   `json:"bilingual_reviewed"`
	AncestryReviewed      bool   `json:"ancestry_near_duplicate_reviewed"`
	DisagreementsResolved bool   `json:"included_disagreements_resolved"`
	PublicReleaseAllowed  bool   `json:"public_release_allowed"`
}

func digest(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func validHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'f' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

func localPath(value string) bool {
	return value != "" && len(value) <= 512 && !strings.ContainsAny(value, "\\:\x00") && !strings.HasPrefix(value, "/") && value != "." && path.Clean(value) == value && value != ".." && !strings.HasPrefix(value, "../")
}

func validPlan(p Plan) error {
	if p.Schema != "riido-statehint-v4-ready-plan-v1" || !validHex(p.SourceCommit, 40) || !validHex(p.BinarySHA, 64) || p.Rubric.SHA256 != rubricSHA || p.Definitions[0].SHA256 != rubricKO || p.Definitions[1].SHA256 != rubricEN || p.Parent.SHA256 != parentSHA || p.Seed != 1729 || p.Batch != 32 || p.Epochs != 40 || p.Decay != .001 || p.Confidence != .9 || p.Margin != .05 || p.NLLFloor != 1e-15 || p.FalseCost != [3]int{1, 10, 3} || p.Coverage != .2 || p.Support != 5 || p.Lineages != 2 || p.Precision != .98 || p.AbstainCost != .375 || p.Temperature != [2]int{5, 50} {
		return ErrStudy
	}
	if p.Arms != [4]Arm{{"warm_lr005", true, .005}, {"warm_lr010", true, .01}, {"warm_lr020", true, .02}, {"fresh_lr020", false, .02}} {
		return ErrStudy
	}
	pins := []Pin{p.Rubric, p.Definitions[0], p.Definitions[1], p.Manifest, p.Audit, p.Parent}
	pins = append(pins, p.Partitions[:]...)
	pins = append(pins, p.Sources...)
	paths := make([]string, len(pins))
	for i, pin := range pins {
		if !localPath(pin.Path) || !validHex(pin.SHA256, 64) {
			return ErrStudy
		}
		paths[i] = pin.Path
	}
	sort.Strings(paths)
	for i := 1; i < len(paths); i++ {
		if paths[i-1] == paths[i] {
			return ErrStudy
		}
	}
	if len(p.Sources) == 0 || len(p.Sources) > 1024 {
		return ErrStudy
	}
	return nil
}

func validateManifest(m Manifest) ([]statehintcorpus.Family, error) {
	if m.Schema != "riido-statehint-v4-family-manifest-v1" || m.RubricSHA != rubricSHA || len(m.Families) != 1200 {
		return nil, ErrStudy
	}
	families := make([]statehintcorpus.Family, len(m.Families))
	ids, hashes := make([]string, 0, 2400), make([]string, 0, 2400)
	var counts [4][statehint.IntentCount]int
	partitions := [4]string{"train", "validation", "calibration", "test"}
	for i, f := range m.Families {
		families[i] = statehintcorpus.Family{ID: f.ID, Lineage: f.Lineage, Partition: f.Partition, Expected: f.Expected}
		class, ok := statehint.IntentIndex(f.Expected)
		partition := -1
		for j, name := range partitions {
			if f.Partition == name {
				partition = j
			}
		}
		if !ok || partition < 0 {
			return nil, ErrStudy
		}
		counts[partition][class]++
		for locale, id := range f.RowIDs {
			if !opaque(id) || !validHex(f.TextSHA[locale], 64) {
				return nil, ErrStudy
			}
			ids, hashes = append(ids, id), append(hashes, f.TextSHA[locale])
		}
	}
	if _, err := statehintcorpus.ValidateMetadata(families); err != nil {
		return nil, ErrStudy
	}
	for partition, classes := range counts {
		want := 15
		if partition == 0 {
			want = 105
		}
		for _, count := range classes {
			if count != want {
				return nil, ErrStudy
			}
		}
	}
	for _, values := range [][]string{ids, hashes} {
		sort.Strings(values)
		for i := 1; i < len(values); i++ {
			if values[i-1] == values[i] {
				return nil, ErrStudy
			}
		}
	}
	return families, nil
}

func opaque(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

func matchRows(c statehintcorpus.Corpus, m Manifest, metadata []statehintcorpus.Family, partition string) error {
	if statehintcorpus.MatchMetadata(c, metadata, partition) != nil {
		return ErrStudy
	}
	entries := append([]FamilyPin(nil), m.Families...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	rows := c.Rows()
	for _, pair := range c.Pairs() {
		i := sort.Search(len(entries), func(i int) bool { return entries[i].ID >= pair.Family.ID })
		if i == len(entries) || entries[i].ID != pair.Family.ID {
			return ErrStudy
		}
		for locale, index := range pair.Rows {
			if rows[index].ID != entries[i].RowIDs[locale] || digest([]byte(rows[index].Text)) != entries[i].TextSHA[locale] {
				return ErrStudy
			}
		}
	}
	return nil
}
