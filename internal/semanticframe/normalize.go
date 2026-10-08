package semanticframe

import (
	"path"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

func identifier(s string) bool {
	if len(s) == 0 || len(s) > 128 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}
func validFile(f File) bool {
	if len(f.SHA256) != 64 || len(f.Path) == 0 || len(f.Path) > 4096 || f.Bytes <= 0 || f.Bytes > MaxInventoryBytes {
		return false
	}
	for _, c := range f.SHA256 {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return utf8.ValidString(f.Path) && !strings.ContainsAny(f.Path, "\\:\x00") && !path.IsAbs(f.Path) && path.Clean(f.Path) == f.Path && f.Path != "." && f.Path != ".." && !strings.HasPrefix(f.Path, "../")
}
func verify(p PinnedBytes, max int) error {
	if !validFile(p.File) || p.File.Bytes != int64(len(p.Bytes)) || len(p.Bytes) > max || sourcecohort.Hash(p.Bytes) != p.File.SHA256 {
		return Error("byte_pin")
	}
	return nil
}
func decode(data []byte, v any, max int) error {
	if len(data) == 0 || len(data) > max || sourcecohort.Decode(data, v) != nil {
		return Error("closed_json")
	}
	return nil
}
func uniquePins(pins []File) error {
	for _, p := range pins {
		if !validFile(p) {
			return Error("duplicate_pin")
		}
	}
	copyPins := append([]File{}, pins...)
	sort.Slice(copyPins, func(i, j int) bool { return copyPins[i].Path < copyPins[j].Path })
	for i, p := range copyPins {
		if !validFile(p) || i > 0 && copyPins[i-1].Path == p.Path {
			return Error("duplicate_pin")
		}
	}
	return nil
}

// referencedPins retains the complete declared reference union without opening
// metadata. Exact repeated pins are valid; one path cannot declare two contents.
func referencedPins(pins []File) ([]File, error) {
	for _, p := range pins {
		if !validFile(p) {
			return nil, Error("reference_pin")
		}
	}
	out := append([]File{}, pins...)
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	write := 0
	for _, p := range out {
		if write > 0 && out[write-1].Path == p.Path {
			if out[write-1] != p {
				return nil, Error("conflicting_reference_pin")
			}
			continue
		}
		out[write] = p
		write++
	}
	return out[:write], nil
}
func samePins(a, b []File) bool {
	if len(a) != len(b) {
		return false
	}
	a = append([]File{}, a...)
	b = append([]File{}, b...)
	sort.Slice(a, func(i, j int) bool { return a[i].Path < a[j].Path })
	sort.Slice(b, func(i, j int) bool { return b[i].Path < b[j].Path })
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func pinsOf(ps []PinnedBytes) []File {
	out := make([]File, len(ps))
	for i, p := range ps {
		out[i] = p.File
	}
	return out
}

// NormalizeAccepted verifies bytes and exact generic declared-version links.
// Historical inventory/producer/acceptance wire formats are unsupported. No
// boolean or accepted filename substitutes for separately pinned acceptance.
// Success does not authenticate declarations or prove inventory meaning.
func NormalizeAccepted(selection string, in AcceptedInputs) (NormalizedAccepted, error) {
	var out NormalizedAccepted
	if selection != "original" && selection != "amended" {
		return out, Error("selection")
	}
	if len(in.InventoryAcceptance) == 0 || len(in.InventoryAcceptance) > MaxItems || len(in.Predecessors) > MaxItems {
		return out, Error("evidence_count")
	}
	all := []PinnedBytes{in.Version, in.Source, in.Observation, in.Inventory, in.Producer, in.SourceReview}
	all = append(all, in.InventoryAcceptance...)
	all = append(all, in.Predecessors...)
	pins := make([]File, len(all))
	total := 0
	for i, p := range all {
		max := MaxInventoryBytes
		if i == 1 {
			max = MaxSourceBytes
		}
		if err := verify(p, max); err != nil {
			return out, err
		}
		total += len(p.Bytes)
		if total > 8<<20 {
			return out, Error("input_bytes")
		}
		pins[i] = p.File
	}
	if err := uniquePins(pins); err != nil {
		return out, err
	}
	if !utf8.Valid(in.Source.Bytes) || !utf8.Valid(in.Observation.Bytes) {
		return out, Error("utf8")
	}
	if strings.TrimSpace(string(in.Source.Bytes)) == "" {
		return out, Error("source_empty")
	}
	var v VersionBinding
	if err := decode(in.Version.Bytes, &v, MaxFrameBytes); err != nil {
		return out, err
	}
	if v.Schema != VersionSchema || v.Selection != selection || v.Source != in.Source.File || v.Observation != in.Observation.File || v.Inventory != in.Inventory.File || v.Producer != in.Producer.File || v.SourceReview != in.SourceReview.File || !samePins(v.InventoryAcceptance, pinsOf(in.InventoryAcceptance)) || !samePins(v.Predecessors, pinsOf(in.Predecessors)) {
		return out, Error("version_join")
	}
	if selection == "original" && len(v.Predecessors) != 0 || selection == "amended" && len(v.Predecessors) == 0 {
		return out, Error("predecessor_selection")
	}
	var producer ProducerDeclaration
	if err := decode(in.Producer.Bytes, &producer, MaxFrameBytes); err != nil {
		return out, err
	}
	producerSchema := OriginalProducerSchema
	if selection == "amended" {
		producerSchema = AmendedProducerSchema
	}
	if producer.Schema != producerSchema || !identifier(producer.ProducerID) || !identifier(producer.SourceID) || producer.Source != v.Source || producer.Observation != v.Observation || producer.Inventory != v.Inventory || !samePins(producer.Predecessors, v.Predecessors) {
		return out, Error("producer_join")
	}
	var review sourcecohort.Review
	if err := decode(in.SourceReview.Bytes, &review, MaxFrameBytes); err != nil {
		return out, err
	}
	if review.ID != producer.SourceID || review.Source != v.Source || review.Observation != v.Observation || !identifier(review.ReviewerID) || review.ReviewerID == producer.ProducerID || !review.Complete || !review.Structural || review.SourceScope != "complete_source_situation" || review.SemanticVerdict != "declared_pass" || len(review.Holds) != 0 {
		return out, Error("source_review_join")
	}
	if err := reviewMetadata(review, in.Source.Bytes); err != nil {
		return out, err
	}
	references := append([]File{}, pins...)
	references = append(references, review.SourceSchema, review.CheckBinding, review.ReadStart, review.ReadResult)
	for _, p := range in.InventoryAcceptance {
		var a InventoryAcceptance
		if err := decode(p.Bytes, &a, MaxFrameBytes); err != nil {
			return out, err
		}
		if a.Schema != AcceptanceSchema || a.SourceID != producer.SourceID || a.Selection != selection || a.Source != v.Source || a.Observation != v.Observation || a.Inventory != v.Inventory || a.Producer != v.Producer || a.SourceReview != v.SourceReview || !identifier(a.ReviewerID) || a.ReviewerID == producer.ProducerID || a.Verdict != "declared_pass" || !samePins(a.Predecessors, v.Predecessors) {
			return out, Error("inventory_acceptance_join")
		}
	}
	// Retain and join predecessors separately. They remain declared history, not
	// recursively accepted versions; a former hold is never rewritten as a pass.
	for _, p := range in.Predecessors {
		var prev VersionBinding
		if err := decode(p.Bytes, &prev, MaxFrameBytes); err != nil {
			return out, err
		}
		if prev.Schema != VersionSchema || prev.Selection != "original" && prev.Selection != "amended" || prev.Selection == "original" && len(prev.Predecessors) != 0 || prev.Selection == "amended" && len(prev.Predecessors) == 0 || prev.Source != v.Source || prev.Inventory == v.Inventory || !validFile(prev.Inventory) || !validFile(prev.Observation) || !validFile(prev.Producer) || !validFile(prev.SourceReview) || len(prev.InventoryAcceptance) == 0 || len(prev.InventoryAcceptance) > MaxItems || len(prev.Predecessors) > MaxItems {
			return out, Error("predecessor_join")
		}
		previousPins := append(append([]File{prev.Source, prev.Observation, prev.Inventory, prev.Producer, prev.SourceReview}, prev.InventoryAcceptance...), prev.Predecessors...)
		if uniquePins(previousPins) != nil {
			return out, Error("predecessor_pin")
		}
		references = append(references, previousPins...)
	}
	var inventory InventoryGraph
	if err := decode(in.Inventory.Bytes, &inventory, MaxInventoryBytes); err != nil {
		return out, err
	}
	if inventory.Schema != InventorySchema || inventory.SourceID != producer.SourceID || inventory.Source != v.Source {
		return out, Error("inventory_join")
	}
	if err := validateInventory(inventory, in.Source.Bytes); err != nil {
		return out, err
	}
	a, err := sortedIDs(review.Dependencies)
	if err != nil {
		return out, Error("review_dependencies")
	}
	b, err := sortedIDs(inventory.Dependencies)
	if err != nil || len(a) != len(b) {
		return out, Error("review_dependencies")
	}
	for i := range a {
		if a[i] != b[i] {
			return out, Error("review_dependencies")
		}
	}
	references, err = referencedPins(references)
	if err != nil {
		return out, err
	}
	out = NormalizedAccepted{version: v, versionPin: in.Version.File, source: append([]byte{}, in.Source.Bytes...), inventory: inventory, references: references}
	return out, nil
}

// Metadata checks do not open receipts, reproduce the checker, or establish
// earliest-read provenance. sourcecohort's full freeze remains a separate gate.
func reviewMetadata(r sourcecohort.Review, source []byte) error {
	t, err := time.Parse(time.RFC3339Nano, r.ReviewedUTC)
	if err != nil || !strings.HasSuffix(r.ReviewedUTC, "Z") || t.Year() <= 1970 {
		return Error("review_utc")
	}
	if uniquePins([]File{r.SourceSchema, r.CheckBinding, r.ReadStart, r.ReadResult}) != nil || r.ReadResult.Path != r.ReadStart.Path+".result" {
		return Error("review_reference_metadata")
	}
	if !evidenceOK(r.Evidence, source) {
		return Error("review_evidence")
	}
	whole := false
	for _, e := range r.Evidence {
		whole = whole || boundaryOK(e, source)
	}
	if !whole {
		return Error("review_boundary")
	}
	return nil
}
