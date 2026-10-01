package taskrun

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	ParentBudgetSchema    = "riido-parent-attempt-budget-v1"
	MaxParentReservations = 4
	maxParentRecordBytes  = 4 << 20
)

// ChildAttemptPins identify the one permitted native CLI invocation in a slot.
// PromptSHA256 hashes actual stdin; it may differ from the original task prompt.
// RecipeSHA256 is the owned EvaluationRecipe identity, not its full JSON hash.
type ChildAttemptPins struct {
	PlanSHA256     string `json:"plan_sha256"`
	AttemptOrdinal int    `json:"attempt_ordinal"`
	TaskID         string `json:"task_id"`
	SpecSHA256     string `json:"spec_sha256"`
	PromptSHA256   string `json:"prompt_sha256"`
	RecipeSHA256   string `json:"recipe_sha256"`
	ProfileID      string `json:"profile_id"`
	Model          string `json:"model"`
	Reasoning      string `json:"reasoning"`
	CLIHash        string `json:"cli_sha256"`
	GoHash         string `json:"go_sha256"`
}

type ParentChildPlan struct {
	PlanSHA256   string `json:"plan_sha256"`
	BaseRevision string `json:"base_revision"`
}

type ParentBudgetAttempt struct {
	GlobalOrdinal int `json:"global_ordinal"`
	ChildAttemptPins
}

// ParentBudgetManifest pins two child plans without a circular parent hash.
// Each child supplies two slots; the four slots have a fixed global order.
type ParentBudgetManifest struct {
	Schema          string                `json:"schema"`
	MaxReservations int                   `json:"max_reservations"`
	Concurrency     int                   `json:"concurrency"`
	NoRefund        bool                  `json:"no_refund"`
	ChildPlans      []ParentChildPlan     `json:"child_plans"`
	OrderedAttempts []ParentBudgetAttempt `json:"ordered_attempts"`
}

type ParentBudgetRequest struct {
	ParentFile    string
	ParentSHA256  string
	LedgerDir     string
	GlobalOrdinal int
}

// ParentBudgetStatus counts distinct durable proofs. Started is a ledger
// marker count, not a claim about backend requests. A reservation without a
// marker has unknown launch state; it is never refunded or automatically retried.
type ParentBudgetStatus struct {
	ParentSHA256       string `json:"parent_sha256"`
	MaxReservations    int    `json:"max_reservations"`
	Reservations       int    `json:"reservations"`
	Started            int    `json:"durable_start_markers"`
	Completed          int    `json:"terminal_receipts"`
	LaunchStateUnknown int    `json:"launch_state_unknown"`
	Active             bool   `json:"active"`
	Stopped            bool   `json:"stopped"`
	Exhausted          bool   `json:"exhausted"`
}

type parentReservation struct {
	Schema        string `json:"schema"`
	ParentSHA256  string `json:"parent_sha256"`
	GlobalOrdinal int    `json:"global_ordinal"`
	ChildAttemptPins
	Owner string `json:"owner"`
}

type parentStarted struct {
	Schema            string `json:"schema"`
	ParentSHA256      string `json:"parent_sha256"`
	GlobalOrdinal     int    `json:"global_ordinal"`
	ReservationSHA256 string `json:"reservation_sha256"`
	RecordSHA256      string `json:"started_record_sha256"`
}

type parentTerminal struct {
	Schema              string `json:"schema"`
	ParentSHA256        string `json:"parent_sha256"`
	GlobalOrdinal       int    `json:"global_ordinal"`
	ReservationSHA256   string `json:"reservation_sha256"`
	StartedRecordSHA256 string `json:"started_record_sha256"`
	RecordSHA256        string `json:"terminal_record_sha256"`
	ProcessStatus       string `json:"process_status"`
	VerificationStatus  string `json:"verification_status"`
	GroupCleanup        string `json:"process_group_cleanup"`
	AuthCleanup         string `json:"auth_cleanup"`
}

type parentOwner struct {
	Schema        string `json:"schema"`
	ParentSHA256  string `json:"parent_sha256"`
	GlobalOrdinal int    `json:"global_ordinal"`
	Owner         string `json:"owner"`
}

type parentStop struct {
	Schema        string `json:"schema"`
	ParentSHA256  string `json:"parent_sha256"`
	GlobalOrdinal int    `json:"global_ordinal"`
	Reason        string `json:"reason"`
}

// ParentBudgetLease owns the cross-process active directory. A process crash
// leaves it in place. No API guesses whether a stale owner is safe to replace.
type ParentBudgetLease struct {
	mu             sync.Mutex
	req            ParentBudgetRequest
	manifest       ParentBudgetManifest
	pins           ChildAttemptPins
	owner          string
	reservationSHA string
	startedSHA     string
	done           bool
	stopped        bool
}

func parentHex(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			if c < 'a' || c > 'f' {
				return false
			}
		}
	}
	return true
}

func parentLabel(s string) bool {
	if s == "" || len(s) > 128 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func validChildPins(p ChildAttemptPins) bool {
	return parentHex(p.PlanSHA256, 64) && p.AttemptOrdinal >= 1 && p.AttemptOrdinal <= 2 &&
		parentLabel(p.TaskID) && parentHex(p.SpecSHA256, 64) && parentHex(p.PromptSHA256, 64) &&
		parentHex(p.RecipeSHA256, 64) && parentLabel(p.ProfileID) && parentLabel(p.Model) &&
		parentLabel(p.Reasoning) && parentHex(p.CLIHash, 64) && parentHex(p.GoHash, 64)
}

// Reject duplicate decoded keys at every level as well as unknown fields. The
// parent is small, immutable configuration; ambiguous JSON has no valid meaning.
func parentJSON(b []byte, v any) error {
	// Go's struct decoder accepts case-folded field aliases. Owned schema keys
	// use lowercase ASCII only, so reject those aliases before reflection can
	// silently merge two different decoded keys. The generic boundary controls
	// use *any solely to exercise byte/depth limits without schema semantics.
	_, arbitraryKeys := v.(*any)
	d := json.NewDecoder(bytes.NewReader(b))
	var walk func(int) error
	walk = func(depth int) error {
		if depth > 8 {
			return Error("parent_invalid_json")
		}
		t, e := d.Token()
		if e != nil {
			return e
		}
		open, ok := t.(json.Delim)
		if !ok {
			return nil
		}
		if open == '{' {
			var seen [32]string
			count := 0
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return e
				}
				key, ok := k.(string)
				if !ok || len(key) > 128 || count == len(seen) {
					return Error("parent_invalid_json")
				}
				if !arbitraryKeys {
					for _, c := range key {
						if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
							return Error("parent_invalid_json")
						}
					}
				}
				for _, prior := range seen[:count] {
					if prior == key {
						return Error("parent_invalid_json")
					}
				}
				seen[count] = key
				count++
				if e := walk(depth + 1); e != nil {
					return e
				}
			}
		} else if open == '[' {
			count := 0
			for d.More() {
				if count == 32 {
					return Error("parent_invalid_json")
				}
				count++
				if e := walk(depth + 1); e != nil {
					return e
				}
			}
		} else {
			return Error("parent_invalid_json")
		}
		_, e = d.Token()
		return e
	}
	if walk(0) != nil {
		return Error("parent_invalid_json")
	}
	if _, e := d.Token(); e != io.EOF {
		return Error("parent_invalid_json")
	}
	d = json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		return Error("parent_invalid_json")
	}
	return nil
}

func loadParent(req ParentBudgetRequest) (ParentBudgetManifest, []byte, error) {
	var m ParentBudgetManifest
	if !parentHex(req.ParentSHA256, 64) || !filepath.IsAbs(req.ParentFile) || !filepath.IsAbs(req.LedgerDir) {
		return m, nil, Error("parent_invalid_request")
	}
	r, e := os.OpenRoot(filepath.Dir(req.ParentFile))
	if e != nil {
		return m, nil, Error("parent_unavailable")
	}
	defer r.Close()
	b, e := readRootFile(r, filepath.Base(req.ParentFile), MaxPlanBytes)
	if e != nil {
		return m, nil, Error("parent_unavailable")
	}
	if digest(b) != req.ParentSHA256 {
		return m, nil, Error("parent_pin_mismatch")
	}
	if parentJSON(b, &m) != nil || m.Schema != ParentBudgetSchema || m.MaxReservations != MaxParentReservations || m.Concurrency != 1 || !m.NoRefund || len(m.ChildPlans) != 2 || len(m.OrderedAttempts) != MaxParentReservations {
		return m, nil, Error("parent_invalid_manifest")
	}
	if !parentHex(m.ChildPlans[0].PlanSHA256, 64) || !parentHex(m.ChildPlans[1].PlanSHA256, 64) || m.ChildPlans[0].PlanSHA256 == m.ChildPlans[1].PlanSHA256 || !parentHex(m.ChildPlans[0].BaseRevision, 40) || !parentHex(m.ChildPlans[1].BaseRevision, 40) || m.ChildPlans[0].BaseRevision == m.ChildPlans[1].BaseRevision {
		return m, nil, Error("parent_invalid_child_plans")
	}
	counts := [2]int{}
	for i, a := range m.OrderedAttempts {
		if a.GlobalOrdinal != i+1 || !validChildPins(a.ChildAttemptPins) {
			return m, nil, Error("parent_invalid_order")
		}
		child := -1
		for j, p := range m.ChildPlans {
			if p.PlanSHA256 == a.PlanSHA256 {
				child = j
			}
		}
		if child < 0 {
			return m, nil, Error("parent_invalid_child_reference")
		}
		counts[child]++
		if a.AttemptOrdinal != counts[child] {
			return m, nil, Error("parent_invalid_child_order")
		}
		for _, prev := range m.OrderedAttempts[:i] {
			if prev.CLIHash != a.CLIHash || prev.GoHash != a.GoHash {
				return m, nil, Error("parent_toolchain_mismatch")
			}
			if prev.PlanSHA256 == a.PlanSHA256 && (prev.TaskID != a.TaskID || prev.SpecSHA256 != a.SpecSHA256 || prev.PromptSHA256 != a.PromptSHA256 || prev.RecipeSHA256 != a.RecipeSHA256) {
				return m, nil, Error("parent_child_task_mismatch")
			}
		}
	}
	if counts != [2]int{2, 2} {
		return m, nil, Error("parent_invalid_child_counts")
	}
	return m, b, nil
}

func parentWrite(r *os.Root, name string, value any) error {
	b, e := json.Marshal(value)
	if e != nil {
		return Error("parent_write_failed")
	}
	return parentWriteBytes(r, name, append(b, '\n'))
}

func parentWriteBytes(r *os.Root, name string, b []byte) error {
	f, e := r.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return Error("parent_write_failed")
	}
	n, e := f.Write(b)
	if e == nil && n != len(b) {
		e = io.ErrShortWrite
	}
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e != nil || ce != nil {
		return Error("parent_write_failed")
	}
	return nil
}

func parentSyncDir(r *os.Root, name string) error {
	f, e := r.Open(name)
	if e != nil {
		return Error("parent_sync_failed")
	}
	e = f.Sync()
	ce := f.Close()
	if e != nil || ce != nil {
		return Error("parent_sync_failed")
	}
	return nil
}

func parentLedgerPath(req ParentBudgetRequest) (string, error) {
	p, e := isolatedPrivatePath(req.LedgerDir)
	if e != nil {
		return "", Error("parent_unsafe_ledger")
	}
	if p == filepath.Dir(p) || filepath.Base(p) == "." {
		return "", Error("parent_unsafe_ledger")
	}
	return p, nil
}

func openParentLedger(req ParentBudgetRequest, raw []byte, create bool) (*os.Root, error) {
	path, e := parentLedgerPath(req)
	if e != nil {
		return nil, e
	}
	created := false
	st, e := os.Lstat(path)
	if errors.Is(e, os.ErrNotExist) && create {
		if os.Mkdir(path, 0700) != nil {
			return nil, Error("parent_ledger_unavailable")
		}
		created = true
	} else if e != nil {
		return nil, Error("parent_ledger_unavailable")
	} else if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 || st.Mode().Perm() != 0700 {
		return nil, Error("parent_unsafe_ledger")
	}
	r, e := os.OpenRoot(path)
	if e != nil {
		return nil, Error("parent_ledger_unavailable")
	}
	fail := func(err error) (*os.Root, error) { r.Close(); return nil, err }
	if created {
		if r.Mkdir("initializing", 0700) != nil {
			return fail(Error("parent_ledger_initialization_failed"))
		}
		if parentWriteBytes(r, "manifest.json", raw) != nil || parentSyncDir(r, ".") != nil || parentWriteBytes(r, "ready", []byte(req.ParentSHA256+"\n")) != nil || parentSyncDir(r, ".") != nil {
			return fail(Error("parent_ledger_initialization_failed"))
		}
		// The ledger directory's own name must survive a host crash too.
		p, e := os.Open(filepath.Dir(path))
		if e != nil {
			return fail(Error("parent_sync_failed"))
		}
		e = p.Sync()
		ce := p.Close()
		if e != nil || ce != nil {
			return fail(Error("parent_sync_failed"))
		}
		if r.Remove("initializing") != nil || parentSyncDir(r, ".") != nil {
			// Refuse reuse if the initialization release was not durable. This
			// best-effort seal never invents recovery from a filesystem failure.
			_ = r.Mkdir("initialization-failed", 0700)
			_ = r.Remove("ready")
			_ = parentSyncDir(r, ".")
			return fail(Error("parent_ledger_initialization_failed"))
		}
	}
	b, e := readRootFile(r, "manifest.json", MaxPlanBytes)
	if e != nil || !bytes.Equal(b, raw) {
		return fail(Error("parent_ledger_pin_mismatch"))
	}
	b, e = readRootFile(r, "ready", 128)
	if e != nil || string(b) != req.ParentSHA256+"\n" {
		return fail(Error("parent_ledger_not_ready"))
	}
	for _, name := range []string{"manifest.json", "ready"} {
		st, e := r.Lstat(name)
		if e != nil || st.Mode().Perm() != 0600 {
			return fail(Error("parent_ledger_integrity"))
		}
	}
	return r, nil
}

func parentSlot(i int) string { return fmt.Sprintf("slot-%02d", i) }

func parentEntries(r *os.Root, name string) ([]os.DirEntry, error) {
	st, e := r.Lstat(name)
	if e != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 || st.Mode().Perm() != 0700 {
		return nil, Error("parent_ledger_integrity")
	}
	f, e := r.Open(name)
	if e != nil {
		return nil, Error("parent_ledger_integrity")
	}
	entries, e := f.ReadDir(-1)
	ce := f.Close()
	if e != nil || ce != nil {
		return nil, Error("parent_ledger_integrity")
	}
	return entries, nil
}

func parentRead(r *os.Root, name string, v any) ([]byte, error) {
	b, e := readRootFile(r, name, MaxPlanBytes)
	if e != nil || parentJSON(b, v) != nil {
		return nil, Error("parent_ledger_integrity")
	}
	st, e := r.Lstat(name)
	if e != nil || st.Mode().Perm() != 0600 {
		return nil, Error("parent_ledger_integrity")
	}
	expected, e := json.Marshal(v)
	if e != nil || !bytes.Equal(b, append(expected, '\n')) {
		return nil, Error("parent_ledger_integrity")
	}
	return b, nil
}

func parentTerminalStatus(s string) bool {
	switch s {
	case "exited_zero", "nonzero_exit", "timeout_or_canceled", "capture_limit", "inherited_pipe_timeout", "capture_write_failed":
		return true
	}
	return false
}

func inspectParentRoot(r *os.Root, req ParentBudgetRequest, m ParentBudgetManifest, pendingOwner bool) (ParentBudgetStatus, error) {
	s := ParentBudgetStatus{ParentSHA256: req.ParentSHA256, MaxReservations: MaxParentReservations}
	entries, e := parentEntries(r, ".")
	if e != nil {
		return s, e
	}
	var present [MaxParentReservations]bool
	for _, ent := range entries {
		switch ent.Name() {
		case "manifest.json", "ready":
		case "active":
			s.Active = true
		case "stop.json":
			s.Stopped = true
		default:
			found := false
			for i := range present {
				if ent.Name() == parentSlot(i+1) {
					present[i] = true
					s.Reservations++
					found = true
				}
			}
			if !found {
				return s, Error("parent_ledger_integrity")
			}
		}
	}
	s.Exhausted = s.Reservations >= MaxParentReservations
	s.LaunchStateUnknown = s.Reservations
	// Count reservations before examining their content, including partial writes.
	for i, exists := range present {
		if !exists {
			continue
		}
		if i > 0 && !present[i-1] {
			return s, Error("parent_ledger_integrity")
		}
		slot := parentSlot(i + 1)
		es, e := parentEntries(r, slot)
		if e != nil {
			return s, e
		}
		hasReservation, hasStart, hasTerminal := false, false, false
		for _, ent := range es {
			switch ent.Name() {
			case "reservation.json":
				hasReservation = true
			case "started.json":
				hasStart = true
			case "terminal.json":
				hasTerminal = true
			default:
				return s, Error("parent_ledger_integrity")
			}
		}
		if !hasReservation {
			return s, Error("parent_ledger_integrity")
		}
		var v parentReservation
		rb, e := parentRead(r, slot+"/reservation.json", &v)
		if e != nil || v.Schema != "riido-parent-reservation-v1" || v.ParentSHA256 != req.ParentSHA256 || v.GlobalOrdinal != i+1 || v.ChildAttemptPins != m.OrderedAttempts[i].ChildAttemptPins || !parentHex(v.Owner, 32) {
			return s, Error("parent_ledger_integrity")
		}
		var start parentStarted
		if hasStart {
			_, e := parentRead(r, slot+"/started.json", &start)
			if e != nil || start.Schema != "riido-parent-started-v1" || start.ParentSHA256 != req.ParentSHA256 || start.GlobalOrdinal != i+1 || start.ReservationSHA256 != digest(rb) || !parentHex(start.RecordSHA256, 64) {
				return s, Error("parent_ledger_integrity")
			}
			s.Started++
			s.LaunchStateUnknown--
		}
		if hasTerminal {
			var terminal parentTerminal
			_, e := parentRead(r, slot+"/terminal.json", &terminal)
			if e != nil || !hasStart || terminal.Schema != "riido-parent-terminal-v1" || terminal.ParentSHA256 != req.ParentSHA256 || terminal.GlobalOrdinal != i+1 || terminal.ReservationSHA256 != digest(rb) || terminal.StartedRecordSHA256 != start.RecordSHA256 || !parentHex(terminal.RecordSHA256, 64) || !parentTerminalStatus(terminal.ProcessStatus) || terminal.GroupCleanup != "group_terminated" || terminal.AuthCleanup != "removed_or_not_present" {
				return s, Error("parent_ledger_integrity")
			}
			s.Completed++
		} else if i != s.Reservations-1 || !s.Active {
			return s, Error("parent_ledger_integrity")
		}
	}
	if s.Active && !pendingOwner {
		var owner parentOwner
		_, e := parentRead(r, "active/owner.json", &owner)
		if e != nil || owner.Schema != "riido-parent-owner-v1" || owner.ParentSHA256 != req.ParentSHA256 || owner.GlobalOrdinal != s.Reservations || owner.GlobalOrdinal < 1 || !parentHex(owner.Owner, 32) {
			return s, Error("parent_ledger_integrity")
		}
		var reservation parentReservation
		_, e = parentRead(r, parentSlot(owner.GlobalOrdinal)+"/reservation.json", &reservation)
		if e != nil || reservation.Owner != owner.Owner {
			return s, Error("parent_ledger_integrity")
		}
		es, e := parentEntries(r, "active")
		if e != nil || len(es) != 1 || es[0].Name() != "owner.json" {
			return s, Error("parent_ledger_integrity")
		}
	}
	if s.Stopped {
		var stop parentStop
		_, e := parentRead(r, "stop.json", &stop)
		if e != nil || stop.Schema != "riido-parent-stop-v1" || stop.ParentSHA256 != req.ParentSHA256 || stop.GlobalOrdinal < 1 || stop.GlobalOrdinal > MaxParentReservations || !parentStopReason(stop.Reason) {
			return s, Error("parent_ledger_integrity")
		}
	}
	return s, nil
}

// InspectParentBudget is read-only and never repairs incomplete state.
func InspectParentBudget(req ParentBudgetRequest) (ParentBudgetStatus, error) {
	m, b, e := loadParent(req)
	if e != nil {
		return ParentBudgetStatus{}, e
	}
	r, e := openParentLedger(req, b, false)
	if e != nil {
		return ParentBudgetStatus{}, e
	}
	defer r.Close()
	return inspectParentRoot(r, req, m, false)
}

// ReserveParentBudget atomically acquires the sole controller lease, then
// consumes an ordered slot durably before the caller may start Codex. Invalid
// request pins are rejected before reservation; a partially created slot is
// consumed and blocks further execution. There is intentionally no recovery API.
func ReserveParentBudget(req ParentBudgetRequest, pins ChildAttemptPins) (*ParentBudgetLease, error) {
	m, b, e := loadParent(req)
	if e != nil {
		return nil, e
	}
	if req.GlobalOrdinal < 1 || req.GlobalOrdinal > MaxParentReservations {
		return nil, Error("parent_global_limit")
	}
	if m.OrderedAttempts[req.GlobalOrdinal-1].ChildAttemptPins != pins {
		return nil, Error("parent_attempt_pin_mismatch")
	}
	r, e := openParentLedger(req, b, true)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	s, e := inspectParentRoot(r, req, m, false)
	if e != nil {
		return nil, e
	}
	if s.Active || s.Stopped {
		return nil, Error("parent_budget_stopped")
	}
	if s.Exhausted {
		return nil, Error("parent_global_limit")
	}
	if req.GlobalOrdinal != s.Reservations+1 || s.Completed != s.Reservations {
		return nil, Error("parent_order_mismatch")
	}
	if r.Mkdir("active", 0700) != nil {
		return nil, Error("parent_controller_active")
	}
	// Once acquired, every failure preserves active. Another controller cannot
	// infer that this owner failed before launching, even if no marker is present.
	l := &ParentBudgetLease{req: req, manifest: m, pins: pins}
	current, e := inspectParentRoot(r, req, m, true)
	if e != nil || current.Stopped || current.Reservations+1 != req.GlobalOrdinal || current.Completed != current.Reservations {
		l.stopRoot(r, "reservation_failed")
		return nil, Error("parent_reservation_failed")
	}
	var nonce [16]byte
	if _, e := rand.Read(nonce[:]); e != nil {
		l.stopRoot(r, "reservation_failed")
		return nil, Error("parent_reservation_failed")
	}
	l.owner = hex.EncodeToString(nonce[:])
	// Recheck prior slots under the exclusive lease. The newly empty active dir
	// is expected here, so validate the prior receipts without that transient name.
	if parentSyncDir(r, ".") != nil {
		l.stopRoot(r, "reservation_failed")
		return nil, Error("parent_reservation_failed")
	}
	slot := parentSlot(req.GlobalOrdinal)
	if r.Mkdir(slot, 0700) != nil {
		l.stopRoot(r, "reservation_failed")
		return nil, Error("parent_reservation_failed")
	}
	v := parentReservation{Schema: "riido-parent-reservation-v1", ParentSHA256: req.ParentSHA256, GlobalOrdinal: req.GlobalOrdinal, ChildAttemptPins: pins, Owner: l.owner}
	if parentWrite(r, slot+"/reservation.json", v) != nil || parentSyncDir(r, slot) != nil || parentWrite(r, "active/owner.json", parentOwner{Schema: "riido-parent-owner-v1", ParentSHA256: req.ParentSHA256, GlobalOrdinal: req.GlobalOrdinal, Owner: l.owner}) != nil || parentSyncDir(r, "active") != nil || parentSyncDir(r, ".") != nil {
		l.stopRoot(r, "reservation_failed")
		return nil, Error("parent_reservation_failed")
	}
	raw, e := json.Marshal(v)
	if e != nil {
		l.stopRoot(r, "reservation_failed")
		return nil, Error("parent_reservation_failed")
	}
	l.reservationSHA = digest(append(raw, '\n'))
	return l, nil
}

func parentStopReason(s string) bool {
	return s == "reservation_failed" || s == "start_proof_failed" || s == "terminal_proof_failed" || s == "terminal_release_failed" || s == "owner_stopped"
}

func (l *ParentBudgetLease) stopRoot(r *os.Root, reason string) {
	l.stopped = true
	// An existing stop remains immutable, even after a repeated cleanup failure.
	_ = parentWrite(r, "stop.json", parentStop{Schema: "riido-parent-stop-v1", ParentSHA256: l.req.ParentSHA256, GlobalOrdinal: l.req.GlobalOrdinal, Reason: reason})
	_ = parentSyncDir(r, ".")
}

func (l *ParentBudgetLease) ownedRoot() (*os.Root, error) {
	_, b, e := loadParent(l.req)
	if e != nil {
		return nil, e
	}
	r, e := openParentLedger(l.req, b, false)
	if e != nil {
		return nil, e
	}
	var o parentOwner
	_, e = parentRead(r, "active/owner.json", &o)
	if e != nil || o.Schema != "riido-parent-owner-v1" || o.ParentSHA256 != l.req.ParentSHA256 || o.GlobalOrdinal != l.req.GlobalOrdinal || o.Owner != l.owner {
		r.Close()
		return nil, Error("parent_owner_mismatch")
	}
	var v parentReservation
	rb, e := parentRead(r, parentSlot(l.req.GlobalOrdinal)+"/reservation.json", &v)
	if e != nil || digest(rb) != l.reservationSHA || v.Owner != l.owner {
		r.Close()
		return nil, Error("parent_owner_mismatch")
	}
	return r, nil
}

func (l *ParentBudgetLease) validRecord(r Record) bool {
	base := ""
	for _, p := range l.manifest.ChildPlans {
		if p.PlanSHA256 == l.pins.PlanSHA256 {
			base = p.BaseRevision
		}
	}
	if r.Schema != "riido-owned-task-attempt-v1" || r.ParentSHA256 != l.req.ParentSHA256 || r.GlobalOrdinal != l.req.GlobalOrdinal || r.PlanSHA256 != l.pins.PlanSHA256 || r.AttemptOrdinal != l.pins.AttemptOrdinal || r.TaskID != l.pins.TaskID || r.TaskSpecSHA256 != l.pins.SpecSHA256 || r.PromptSHA256 != l.pins.PromptSHA256 || r.RecipeSHA256 != l.pins.RecipeSHA256 || r.ProfileID != l.pins.ProfileID || r.Applied.Model != l.pins.Model || r.Applied.Reasoning != l.pins.Reasoning || r.ExecutableSHA256 != l.pins.CLIHash || r.GoBinarySHA256 != l.pins.GoHash || r.BaseRevision != base || !r.Started || !r.DurableStart || r.EvaluationRecipe == nil || r.EvaluationRecipe.RecipeSHA256 != l.pins.RecipeSHA256 {
		return false
	}
	recipe := *r.EvaluationRecipe
	recipe.RecipeSHA256 = ""
	return digestJSON(recipe) == l.pins.RecipeSHA256
}

func parentRecordBytes(record Record) ([]byte, error) {
	b, e := json.Marshal(record)
	if e != nil || len(b) >= maxParentRecordBytes {
		return nil, Error("parent_record_limit")
	}
	return append(b, '\n'), nil
}

// MarkStarted records the canonical bytes of the already durable executor start
// record. Complete later reads the actual file independently. Failure does not
// release the reservation, so even a launch without a valid marker stays spent.
func (l *ParentBudgetLease) MarkStarted(record Record) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.done || l.stopped || l.startedSHA != "" {
		return Error("parent_invalid_lease_state")
	}
	r, e := l.ownedRoot()
	if e != nil {
		l.stopped = true
		return e
	}
	defer r.Close()
	b, e := parentRecordBytes(record)
	if e != nil || !l.validRecord(record) || record.ProcessStatus != "running" || record.ExitCode != nil || record.GroupCleanup != "pending" {
		l.stopRoot(r, "start_proof_failed")
		return Error("parent_start_proof_failed")
	}
	sha := digest(b)
	if parentWrite(r, parentSlot(l.req.GlobalOrdinal)+"/started.json", parentStarted{Schema: "riido-parent-started-v1", ParentSHA256: l.req.ParentSHA256, GlobalOrdinal: l.req.GlobalOrdinal, ReservationSHA256: l.reservationSHA, RecordSHA256: sha}) != nil || parentSyncDir(r, parentSlot(l.req.GlobalOrdinal)) != nil {
		l.stopRoot(r, "start_proof_failed")
		return Error("parent_start_proof_failed")
	}
	l.startedSHA = sha
	return nil
}

// Complete releases concurrency only after byte-identical durable start/final
// files and affirmative cleanup facts. Failed model tasks and unknown usage
// metrics may finish safely; unknown process/auth cleanup cannot. This is a
// host-owned evidence boundary, not protection against a malicious local host.
func (l *ParentBudgetLease) Complete(privateDir string, record Record) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.done || l.stopped {
		return Error("parent_invalid_lease_state")
	}
	r, e := l.ownedRoot()
	if e != nil {
		l.stopped = true
		return e
	}
	defer r.Close()
	fail := func() error { l.stopRoot(r, "terminal_proof_failed"); return Error("parent_terminal_proof_failed") }
	if l.startedSHA == "" || !l.validRecord(record) || !parentTerminalStatus(record.ProcessStatus) || record.ExitCode == nil || record.GroupCleanup != "group_terminated" || record.AuthCleanup != "removed_or_not_present" {
		return fail()
	}
	if record.ProcessStatus == "exited_zero" && *record.ExitCode != 0 || record.ProcessStatus == "nonzero_exit" && *record.ExitCode == 0 {
		return fail()
	}
	p, e := isolatedPrivatePath(privateDir)
	if e != nil || !filepath.IsAbs(privateDir) {
		return fail()
	}
	st, e := os.Lstat(p)
	if e != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 || st.Mode().Perm() != 0700 {
		return fail()
	}
	pr, e := os.OpenRoot(p)
	if e != nil {
		return fail()
	}
	defer pr.Close()
	actual, e := readRootFile(pr, "record.json", maxParentRecordBytes)
	expected, ee := parentRecordBytes(record)
	if e != nil || ee != nil || !bytes.Equal(actual, expected) {
		return fail()
	}
	st, e = pr.Lstat("record.json")
	if e != nil || st.Mode().Perm() != 0600 {
		return fail()
	}
	start, e := readRootFile(pr, "started-record.json", maxParentRecordBytes)
	if e != nil || digest(start) != l.startedSHA {
		return fail()
	}
	st, e = pr.Lstat("started-record.json")
	if e != nil || st.Mode().Perm() != 0600 {
		return fail()
	}
	// Reject a replaced start file even if its JSON superficially reports success.
	var sr Record
	if json.Unmarshal(start, &sr) != nil || !l.validRecord(sr) || sr.ProcessStatus != "running" || sr.ExitCode != nil || sr.GroupCleanup != "pending" {
		return fail()
	}
	st, e = pr.Lstat("codex-home")
	if e != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return fail()
	}
	if _, e = pr.Lstat("codex-home/auth.json"); !errors.Is(e, os.ErrNotExist) {
		return fail()
	}
	var sm parentStarted
	_, e = parentRead(r, parentSlot(l.req.GlobalOrdinal)+"/started.json", &sm)
	if e != nil || sm.RecordSHA256 != l.startedSHA || sm.ReservationSHA256 != l.reservationSHA || sm.ParentSHA256 != l.req.ParentSHA256 || sm.GlobalOrdinal != l.req.GlobalOrdinal {
		return fail()
	}
	v := parentTerminal{Schema: "riido-parent-terminal-v1", ParentSHA256: l.req.ParentSHA256, GlobalOrdinal: l.req.GlobalOrdinal, ReservationSHA256: l.reservationSHA, StartedRecordSHA256: l.startedSHA, RecordSHA256: digest(actual), ProcessStatus: record.ProcessStatus, VerificationStatus: record.VerificationStatus, GroupCleanup: record.GroupCleanup, AuthCleanup: record.AuthCleanup}
	if parentWrite(r, parentSlot(l.req.GlobalOrdinal)+"/terminal.json", v) != nil || parentSyncDir(r, parentSlot(l.req.GlobalOrdinal)) != nil {
		return fail()
	}
	// Remove only this immutable ownership marker. Never recursively delete an
	// unknown directory or another controller's files.
	if r.Remove("active/owner.json") != nil || parentSyncDir(r, "active") != nil || r.Remove("active") != nil || parentSyncDir(r, ".") != nil {
		l.stopRoot(r, "terminal_release_failed")
		return Error("parent_terminal_release_failed")
	}
	l.done = true
	return nil
}

// Stop is suitable for defer immediately after reservation. It is a no-op only
// after successful completion. It never removes a lease or refunds a slot.
func (l *ParentBudgetLease) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.done {
		return nil
	}
	r, e := l.ownedRoot()
	if e != nil {
		l.stopped = true
		return e
	}
	defer r.Close()
	l.stopRoot(r, "owner_stopped")
	var s parentStop
	if _, e = parentRead(r, "stop.json", &s); e != nil || s.ParentSHA256 != l.req.ParentSHA256 {
		return Error("parent_stop_failed")
	}
	return nil
}
