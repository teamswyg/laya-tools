package semanticframe

import (
	"strconv"
	"unicode/utf8"
)

const maxPointerBytes = len("reply_quote_adoption_retraction_negation_constraints") + 5

// jsonSize counts the exact compact encoding/json size (including HTML escaping)
// without serializing caller input. Scalar and collection limits are checked
// before scanning strings; aggregate overflow stops the bounded walk immediately.
type jsonSize struct {
	n  int
	ok bool
}

func (s *jsonSize) add(n int) {
	if !s.ok {
		return
	}
	if n < 0 || n > MaxFrameBytes-s.n {
		s.ok = false
		return
	}
	s.n += n
}
func (s *jsonSize) str(v string, max int) {
	if !s.ok {
		return
	}
	if len(v) > max || !utf8.ValidString(v) {
		s.ok = false
		return
	}
	s.add(2 + len(v))
	for i := 0; s.ok && i < len(v); i++ {
		switch v[i] {
		case '"', '\\', '\b', '\f', '\n', '\r', '\t':
			s.add(1)
		case '<', '>', '&':
			s.add(5)
		default:
			if v[i] < 0x20 {
				s.add(5)
			} else if v[i] == 0xe2 && i+2 < len(v) && v[i+1] == 0x80 && (v[i+2] == 0xa8 || v[i+2] == 0xa9) {
				s.add(3)
				i += 2
			}
		}
	}
}
func (s *jsonSize) num(v int64) { var b [20]byte; s.add(len(strconv.AppendInt(b[:0], v, 10))) }
func (s *jsonSize) array(n int, nilArray bool) {
	if nilArray {
		s.add(4)
		return
	}
	s.add(2)
	if n > 0 {
		s.add(n - 1)
	}
}
func (s *jsonSize) file(f File) {
	s.add(len(`{"path":,"sha256":,"bytes":}`))
	s.str(f.Path, 4096)
	s.str(f.SHA256, 64)
	s.num(f.Bytes)
}
func (s *jsonSize) span(p Span) {
	s.add(len(`{"kind":,"start":,"end":}`))
	s.str(p.Kind, 32)
	s.num(int64(p.Start))
	s.num(int64(p.End))
}
func (s *jsonSize) spans(ps []Span) {
	if len(ps) > MaxItems {
		s.ok = false
		return
	}
	s.array(len(ps), ps == nil)
	for _, p := range ps {
		if !s.ok {
			return
		}
		s.span(p)
	}
}

func typedFrameSize(f Frame) (int, bool) {
	s := jsonSize{ok: true}
	s.add(len(`{"schema":,"family_id":,"source_id":,"family_slot":,"source":,"accepted_inventory":,"accepted_inventory_binding":,"accepted_version_evidence":,"accepted_observation":,"accepted_source_review":,"definitions":,"inventory_schema_source":,"source_boundary":,"communication_plan":,"preserved_ambiguities":,"desired_labels_in_packet":,"target_input_mode":}`))
	s.str(f.Schema, 64)
	s.str(f.FamilyID, 128)
	s.str(f.SourceID, 128)
	s.num(int64(f.Slot))
	s.str(f.TargetMode, 64)
	for _, p := range [...]File{f.Source, f.Inventory, f.Producer, f.VersionEvidence, f.Observation, f.SourceReview, f.Definitions, f.InventorySchemaSource} {
		s.file(p)
	}
	s.span(f.Boundary)
	if f.DesiredLabels {
		s.add(4)
	} else {
		s.add(5)
	}
	s.add(len(`{"bindings":,"source_evidence":,"description_private":,"wording_freedom":}`))
	s.str(f.Plan.Description, 1024)
	s.str(f.Plan.WordingFreedom, 64)
	s.spans(f.Plan.Evidence)
	if len(f.Plan.Bindings) > MaxItems || len(f.Ambiguities) > MaxItems {
		return s.n, false
	}
	s.array(len(f.Plan.Bindings), f.Plan.Bindings == nil)
	for _, b := range f.Plan.Bindings {
		if !s.ok {
			return s.n, false
		}
		s.add(len(`{"inventory_pointer":,"role":}`))
		s.str(b.Pointer, maxPointerBytes)
		s.str(b.Role, 32)
	}
	s.array(len(f.Ambiguities), f.Ambiguities == nil)
	for _, a := range f.Ambiguities {
		if !s.ok || len(a.Pointers) > MaxItems {
			return s.n, false
		}
		s.add(len(`{"inventory_pointers":,"source_evidence":,"description_private":,"operation":}`))
		s.str(a.Description, 512)
		s.str(a.Operation, 64)
		s.spans(a.Evidence)
		s.array(len(a.Pointers), a.Pointers == nil)
		for _, p := range a.Pointers {
			if !s.ok {
				return s.n, false
			}
			s.str(p, maxPointerBytes)
		}
	}
	return s.n, s.ok
}
