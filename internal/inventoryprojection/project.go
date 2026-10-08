package inventoryprojection

import (
	"encoding/json"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/semanticframe"
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
func relative(s string) bool {
	return len(s) > 0 && len(s) <= 4096 && utf8.ValidString(s) && !strings.ContainsAny(s, "\\:\x00") && !path.IsAbs(s) && path.Clean(s) == s && s != "." && s != ".." && !strings.HasPrefix(s, "../")
}
func validFile(f File) bool {
	if len(f.SHA256) != 64 || !relative(f.Path) || f.Bytes <= 0 || f.Bytes > MaxBytes {
		return false
	}
	for _, c := range f.SHA256 {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func verify(p PinnedBytes, max int) error {
	if len(p.Bytes) == 0 || len(p.Bytes) > max || !validFile(p.File) || p.File.Bytes != int64(len(p.Bytes)) || sourcecohort.Hash(p.Bytes) != p.File.SHA256 {
		return Error("byte_pin")
	}
	return nil
}
func sameList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func sortedIDs(ids []string) ([]string, error) {
	if len(ids) > MaxItems {
		return nil, Error("id_count")
	}
	for _, id := range ids {
		if !identifier(id) {
			return nil, Error("identifier")
		}
	}
	out := append([]string{}, ids...)
	sort.Strings(out)
	for i := 1; i < len(out); i++ {
		if out[i] == out[i-1] {
			return nil, Error("duplicate_id")
		}
	}
	return out, nil
}
func present(ids []string, id string) bool {
	i := sort.SearchStrings(ids, id)
	return i < len(ids) && ids[i] == id
}
func refs(ids, allowed []string) error {
	sorted, err := sortedIDs(ids)
	if err != nil {
		return err
	}
	for _, id := range sorted {
		if !present(allowed, id) {
			return Error("endpoint")
		}
	}
	return nil
}
func checkPins(pins []File) error {
	for _, p := range pins {
		if !validFile(p) {
			return Error("reference_pin")
		}
	}
	pins = append([]File{}, pins...)
	sort.Slice(pins, func(i, j int) bool { return pins[i].Path < pins[j].Path })
	for i := 1; i < len(pins); i++ {
		if pins[i].Path == pins[i-1].Path && pins[i] != pins[i-1] {
			return Error("conflicting_reference_pin")
		}
	}
	return nil
}

// Project verifies supplied bytes/config and derives a graph. It performs no
// opens/writes, acceptance generation, historical adaptation or semantic QA.
func Project(in Inputs) (Result, error) {
	var out Result
	all := [...]PinnedBytes{in.Config, in.Full, in.Native, in.Mapping, in.Adapter}
	total := 0
	pins := make([]File, 0, 9)
	for _, p := range all {
		if err := verify(p, MaxBytes); err != nil {
			return out, err
		}
		total += len(p.Bytes)
		pins = append(pins, p.File)
	}
	if in.Source != nil {
		if err := verify(*in.Source, MaxSourceBytes); err != nil {
			return out, err
		}
		total += len(in.Source.Bytes)
		pins = append(pins, in.Source.File)
	}
	if total > MaxTotalBytes {
		return out, Error("total_bytes")
	}
	var c Config
	if err := closed(in.Config.Bytes, &c); err != nil {
		return out, err
	}
	if c.Full != in.Full.File || c.Adapter != in.Adapter.File || c.Mapping != in.Mapping.File || c.Native != in.Native.File || c.Version != Version || !relative(c.GraphPath) {
		return out, Error("config_join")
	}
	var native NativeContract
	if err := closed(in.Native.Bytes, &native); err != nil {
		return out, err
	}
	if len(native.Schema) == 0 || len(native.Schema) > 128 || strings.TrimSpace(native.Schema) != native.Schema {
		return out, Error("native_schema")
	}
	keys, err := sortedIDs(native.Fields)
	if err != nil {
		return out, err
	}
	var mapping MappingContract
	if err := closed(in.Mapping.Bytes, &mapping); err != nil {
		return out, err
	}
	supported := SupportedMapping()
	if mapping.GraphSchema != supported.GraphSchema || !sameList(mapping.CopiedFields, supported.CopiedFields) || !sameList(mapping.FullOnlyFields, supported.FullOnlyFields) {
		return out, Error("mapping_contract")
	}
	f, err := decodeFull(in.Full.Bytes)
	if err != nil {
		return out, err
	}
	if f.Schema != native.Schema || !identifier(f.ID) || f.DesiredLabels {
		return out, Error("full_identity")
	}
	actual := make([]string, len(f.OfficialFields))
	for i, x := range f.OfficialFields {
		actual[i] = x.Name
	}
	sort.Strings(actual)
	if !sameList(actual, keys) {
		return out, Error("official_keyset")
	}
	pins = append(pins, f.Source, f.Creation, f.ReadBinding)
	for _, r := range f.AdditionalReads {
		if r.Source != f.Source || len(r.Actor) == 0 || len(r.Actor) > 128 {
			return out, Error("additional_read_join")
		}
		pins = append(pins, r.Source, r.Start, r.Result)
	}
	if err = checkPins(pins); err != nil {
		return out, err
	}
	if f.Source.Bytes > MaxSourceBytes {
		return out, Error("source_bytes")
	}
	if in.Source != nil && (in.Source.File != f.Source || !utf8.Valid(in.Source.Bytes)) {
		return out, Error("source_join")
	}
	if err = validate(f, in.Source); err != nil {
		return out, err
	}
	g := semanticframe.InventoryGraph{Schema: semanticframe.InventorySchema, SourceID: f.ID, Source: f.Source, Boundary: f.Boundary, Components: f.Components, Propositions: f.Propositions, Oppositions: f.Oppositions, Temporal: f.Temporal, Referents: f.Referents, Constraints: f.Constraints, Dependencies: f.Dependencies}
	// Count compact JSON bytes before encoding: HTML/control escaping may expand
	// a native document that originally used literal UTF-8 text.
	if !graphFits(g) {
		return out, Error("graph_bytes")
	}
	raw, err := json.Marshal(g)
	if err != nil || len(raw) > MaxBytes {
		return out, Error("graph_bytes")
	}
	graphPin := File{Path: c.GraphPath, SHA256: sourcecohort.Hash(raw), Bytes: int64(len(raw))}
	if err = checkPins(append(pins, graphPin)); err != nil {
		return out, err
	}
	out = Result{RetainedFullBytes: append([]byte{}, in.Full.Bytes...), Full: f, Graph: g, GraphBytes: raw, Binding: Binding{Full: in.Full.File, Graph: graphPin, Adapter: c.Adapter, Version: c.Version, Mapping: c.Mapping, Native: c.Native, Config: in.Config.File}, UTF8SpansVerified: in.Source != nil}
	return out, nil
}
