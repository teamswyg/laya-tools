//go:build darwin && arm64

// SPDX-License-Identifier: Apache-2.0
// Author-only owned controls. No executable, child process, JWT or crypto key.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestOwnedBoundedWriterExactLimitAndIncompleteRecord(t *testing.T) {
	alert := make(chan struct{}, 1)
	w := boundedWriter{Buffer: make([]byte, 0, 4), Limit: 4, Alert: alert}
	source := []byte("ABCD")
	n, err := w.Write(source)
	if n != 4 || err != nil || w.Observed != 4 || w.Overflow {
		t.Fatalf("exact limit: n=%d err=%v observed=%d overflow=%v", n, err, w.Observed, w.Overflow)
	}
	source[0] = 'z'
	if string(w.Buffer) != "ABCD" || len(alert) != 0 {
		t.Fatal("writer must own copied bytes and remain below overflow alert")
	}
	complete := w.record(true)
	if !complete.FullResponsePreserved || complete.RetainedBytes != 4 || complete.WriteObservedBytes != 4 || complete.RetainedSHA != digest([]byte("ABCD")) || complete.Overflow {
		t.Fatalf("complete exact-limit receipt: %+v", complete)
	}
	incomplete := w.record(false)
	if incomplete.FullResponsePreserved || !bytes.Equal(incomplete.Raw, []byte("ABCD")) || incomplete.RetainedSHA != complete.RetainedSHA {
		t.Fatal("incomplete lifecycle must retain identical bytes/hash without complete claim")
	}
}

func TestOwnedBoundedWriterOverflowKeepsFirstPrefixAndCounts(t *testing.T) {
	alert := make(chan struct{}, 1)
	w := boundedWriter{Buffer: make([]byte, 0, 4), Limit: 4, Alert: alert}
	if n, err := w.Write([]byte("AB")); n != 2 || err != nil {
		t.Fatalf("first chunk: %d %v", n, err)
	}
	n, err := w.Write([]byte("CDE"))
	if n != 2 || !errors.Is(err, io.ErrShortBuffer) || string(w.Buffer) != "ABCD" || w.Observed != 5 || !w.Overflow || len(alert) != 1 {
		t.Fatalf("overflow did not preserve first prefix/count: n=%d err=%v prefix=%q observed=%d alert=%d", n, err, w.Buffer, w.Observed, len(alert))
	}
	r := w.record(true)
	if r.FullResponsePreserved || !r.Overflow || r.RetainedBytes != 4 || r.WriteObservedBytes != 5 {
		t.Fatalf("overflow must remain incomplete even for clean lifecycle: %+v", r)
	}
	// An already full bounded alert channel must not block another overflow.
	n, err = w.Write([]byte("FG"))
	if n != 0 || !errors.Is(err, io.ErrShortBuffer) || string(w.Buffer) != "ABCD" || w.Observed != 7 || len(alert) != 1 {
		t.Fatal("full alert channel must preserve cap and observed count without blocking")
	}
}

func TestOwnedIndependentWriterOwnershipAfterJoin(t *testing.T) {
	alert := make(chan struct{}, 2)
	writers := [2]boundedWriter{
		{Buffer: make([]byte, 0, 4), Limit: 4, Alert: alert},
		{Buffer: make([]byte, 0, 3), Limit: 3, Alert: alert},
	}
	inputs := [2][]byte{[]byte("alpha"), []byte("BETA")}
	done := make(chan int, 2)
	for i := range writers {
		go func(index int) {
			_, _ = writers[index].Write(inputs[index])
			done <- index
		}(i)
	}
	var joined [2]bool
	for range writers {
		index := <-done
		if joined[index] {
			t.Fatal("duplicate owner completion")
		}
		joined[index] = true
	}
	// Main observes shared writer state only after every owner has signalled.
	inputs[0][0], inputs[1][0] = 'x', 'y'
	if !joined[0] || !joined[1] || string(writers[0].Buffer) != "alph" || string(writers[1].Buffer) != "BET" || len(alert) != 2 {
		t.Fatal("independent owners must retain disjoint copied prefixes after join")
	}
	if writers[0].Observed != 5 || writers[1].Observed != 4 || !writers[0].Overflow || !writers[1].Overflow {
		t.Fatal("each owned stream must retain its own count and overflow")
	}
}

func ownedControlDir(t *testing.T) string {
	t.Helper()
	// Darwin's ordinary temp prefix can have a symlink ancestor. Resolve only
	// this owned test directory; the runtime API must still reject aliases.
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	return dir
}

func ownedConfigFixture(t *testing.T) config {
	t.Helper()
	dir := ownedControlDir(t)
	return config{
		Schema: "riido-one-shot-private-process-config-v1", Profile: "observer",
		TargetPath: filepath.Join(dir, "owned-not-an-executable"), TargetSHA: strings.Repeat("a", 64),
		InputPath: filepath.Join(dir, "owned-fictional-input"), InputSHA: strings.Repeat("b", 64),
		CapturePath: filepath.Join(dir, "owned-first-capture.json"),
	}
}

func ownedConfigBytes(t *testing.T, cfg config) []byte {
	t.Helper()
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestOwnedConfigExactStringFieldsAndProfile(t *testing.T) {
	fixture := ownedConfigFixture(t)
	for _, name := range [2]string{"signer", "observer"} {
		fixture.Profile = name
		got, code := decodeConfig(ownedConfigBytes(t, fixture))
		if code != "" || got != fixture {
			t.Fatalf("fixed profile exact config: code=%q got=%+v", code, got)
		}
	}
}

func TestOwnedConfigDuplicateMissingUnknownAndTrailing(t *testing.T) {
	fixture := ownedConfigFixture(t)
	valid := string(ownedConfigBytes(t, fixture))
	withoutLastBrace := valid[:len(valid)-1]
	cases := []struct{ name, raw, code string }{
		{"duplicate profile", withoutLastBrace + `,"profile":"signer"}`, "config_unknown_or_duplicate_key"},
		{"duplicate pinned hash", withoutLastBrace + `,"target_sha256":"` + strings.Repeat("c", 64) + `"}`, "config_unknown_or_duplicate_key"},
		{"unknown field", withoutLastBrace + `,"unfrozen":"owned"}`, "config_unknown_or_duplicate_key"},
		{"missing input hash", strings.Replace(valid, `,"input_sha256":"`+fixture.InputSHA+`"`, "", 1), "config_missing_key"},
		{"trailing object", valid + `{}`, "config_trailing_data"},
		{"trailing scalar", valid + ` false`, "config_trailing_data"},
		{"trailing garbage", valid + `x`, "config_trailing_data"},
		{"array root", `[]`, "config_object_invalid"},
		{"number value", strings.Replace(valid, `"profile":"observer"`, `"profile":7`, 1), "config_value_invalid"},
		{"uppercase hash", strings.Replace(valid, fixture.TargetSHA, strings.Repeat("A", 64), 1), "config_contract_invalid"},
		{"incorrect schema", strings.Replace(valid, fixture.Schema, "owned-wrong-schema", 1), "config_contract_invalid"},
		{"extra long value", strings.Replace(valid, `"profile":"observer"`, `"profile":"`+strings.Repeat("x", 1025)+`"`, 1), "config_value_invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, code := decodeConfig([]byte(tc.raw)); code != tc.code {
				t.Fatalf("want %q, got %q", tc.code, code)
			}
		})
	}
	badUTF8 := append([]byte(valid[:len(valid)-1]), 0xff, '}')
	if _, code := decodeConfig(badUTF8); code != "config_utf8_invalid" {
		t.Fatalf("invalid source UTF8: %q", code)
	}
}

func TestOwnedConfigPrivateBoundaryNullCannotSelectProfile(t *testing.T) {
	fixture := ownedConfigFixture(t)
	raw := strings.Replace(string(ownedConfigBytes(t, fixture)), `"profile":"observer"`, `"profile":null`, 1)
	got, code := decodeConfig([]byte(raw))
	// encoding/json can decode null to zero string. The pure decoder does not
	// select a profile; run's frozen profile lookup rejects the empty result.
	if code != "" || got.Profile != "" {
		t.Fatalf("document current null projection boundary: profile=%q code=%q", got.Profile, code)
	}
	var matches int
	for _, p := range profiles {
		if p.Name == got.Profile {
			matches++
		}
	}
	if matches != 0 {
		t.Fatal("null profile must never match either allowed execution profile")
	}
}

func TestOwnedConfigRejectsPathAliasesAndControlCharacters(t *testing.T) {
	fixture := ownedConfigFixture(t)
	dir := filepath.Dir(fixture.TargetPath)
	alias := filepath.Join(dir, "owned-alias")
	if err := os.Symlink(dir, alias); err != nil {
		t.Fatal(err)
	}
	cases := [4]string{filepath.Join(alias, "leaf"), dir + "/../leaf", dir + "/line\nleaf", "relative/leaf"}
	for _, path := range cases {
		changed := fixture
		changed.TargetPath = path
		if _, code := decodeConfig(ownedConfigBytes(t, changed)); code != "config_contract_invalid" {
			t.Fatalf("path alias/control must be refused: code=%q", code)
		}
	}
}

func TestOwnedResourceUnavailableStaysNullAndDarwinWireFixture(t *testing.T) {
	unavailable := resourceForState(nil)
	if unavailable.CPUAvailable || unavailable.RusageAvailable || unavailable.UserNS != nil || unavailable.SystemNS != nil || unavailable.DarwinMaxRSSBytes != nil || unavailable.Rusage != nil {
		t.Fatal("missing ProcessState must remain unavailable, never measured zero")
	}
	packed, err := json.Marshal(unavailable)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range [4][]byte{[]byte(`"user_cpu_ns":null`), []byte(`"system_cpu_ns":null`), []byte(`"darwin_maxrss_bytes":null`), []byte(`"syscall_rusage":null`)} {
		if !bytes.Contains(packed, field) {
			t.Fatalf("unknown field must be explicit null: %s", field)
		}
	}
	// A literal owned wire fixture, not a real ProcessState observation. It
	// protects integer/ns/byte/null serialization; no native RSS unit claim.
	userNS, systemNS, rss := int64(1250000000), int64(250000000), int64(1234567)
	r := resource{CPUAvailable: true, UserNS: &userNS, SystemNS: &systemNS, RusageAvailable: true, DarwinMaxRSSBytes: &rss, Rusage: &syscall.Rusage{Utime: syscall.Timeval{Sec: 1, Usec: 250000}, Stime: syscall.Timeval{Sec: 0, Usec: 250000}, Maxrss: rss}}
	packed, err = json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var round resource
	if json.Unmarshal(packed, &round) != nil || round.UserNS == nil || *round.UserNS != userNS || round.SystemNS == nil || *round.SystemNS != systemNS || round.DarwinMaxRSSBytes == nil || *round.DarwinMaxRSSBytes != rss || round.Rusage == nil || round.Rusage.Maxrss != rss || round.Rusage.Utime.Sec != 1 || round.Rusage.Utime.Usec != 250000 {
		t.Fatal("literal Darwin byte/integer CPU fixture must not acquire KiB multiplication or lose raw Rusage")
	}
}

func TestOwnedPersistJSONExclusiveImmutableAndExactRawBytes(t *testing.T) {
	dir := ownedControlDir(t)
	path := filepath.Join(dir, "owned-complete-capture.json")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	raw := []byte{'A', 0, 0xff, '\n'}
	value := struct {
		Raw            []byte `json:"raw_base64"`
		Interpretation string `json:"interpretation"`
	}{raw, "UNINTERPRETED_RAW_FIRST_RESPONSE"}
	packed, code := persistJSON(f, value, 512)
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if code != "" || packed[len(packed)-1] != '\n' {
		t.Fatalf("complete private capture persistence: %q", code)
	}
	if err := syncParent(path); err != nil {
		t.Fatal(err)
	}
	disk, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var round struct {
		Raw            []byte `json:"raw_base64"`
		Interpretation string `json:"interpretation"`
	}
	if !bytes.Equal(disk, packed) || json.Unmarshal(disk, &round) != nil || !bytes.Equal(round.Raw, raw) || round.Interpretation != value.Interpretation {
		t.Fatal("base64 must preserve every raw byte, including invalid UTF8, before interpretation")
	}
	st, err := os.Stat(path)
	if err != nil || st.Mode().Perm() != 0o600 {
		t.Fatal("private capture mode must remain0600")
	}
	second, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if second != nil {
		second.Close()
	}
	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("a second capture must not replace first bytes: %v", err)
	}
	again, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(again, packed) {
		t.Fatal("exclusive failed reopening changed first capture")
	}
}

func TestOwnedPersistOverLimitAndPartialArtifactsRemain(t *testing.T) {
	dir := ownedControlDir(t)
	empty := filepath.Join(dir, "owned-empty-failed-reservation.json")
	f, err := os.OpenFile(empty, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if packed, code := persistJSON(f, struct {
		Data string `json:"data"`
	}{strings.Repeat("x", 64)}, 16); packed != nil || code != "private_record_encoding_or_size_failed" {
		t.Fatalf("overlimit must have no success bytes: %q", code)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(empty)
	if err != nil || len(b) != 0 {
		t.Fatal("failed reservation must remain as an existing empty first artifact")
	}
	partial := filepath.Join(dir, "owned-partial-first-capture.json")
	prefix := []byte(`{"raw_base64":"owned-incomplete`)
	pf, err := os.OpenFile(partial, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := pf.Write(prefix); err != nil || n != len(prefix) {
		t.Fatal("owned synthetic partial write failed")
	}
	if pf.Sync() != nil || pf.Close() != nil {
		t.Fatal("owned partial artifact persistence failed")
	}
	for _, path := range [2]string{empty, partial} {
		next, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if next != nil {
			next.Close()
		}
		if !errors.Is(err, os.ErrExist) {
			t.Fatal("empty/partial first artifact must block replacement")
		}
		st, err := os.Stat(path)
		if err != nil || st.Mode().Perm() != 0o600 {
			t.Fatal("failed private artifact must keep0600")
		}
	}
	b, err = os.ReadFile(partial)
	if err != nil || !bytes.Equal(b, prefix) || json.Valid(b) {
		t.Fatal("partial bytes must remain incomplete and unchanged")
	}
}

type ownedFakePipe struct {
	Data       [16]byte
	Used       int
	ShortLimit int
	FailureAt  int
	CloseFail  bool
	Closed     bool
}

func (p *ownedFakePipe) Write(b []byte) (int, error) {
	if p.FailureAt >= 0 && p.Used >= p.FailureAt {
		return 0, errors.New("owned_fake_write_failed")
	}
	n := len(b)
	if p.ShortLimit > 0 && n > p.ShortLimit {
		n = p.ShortLimit
	}
	if p.FailureAt >= 0 && n > p.FailureAt-p.Used {
		n = p.FailureAt - p.Used
	}
	if n > len(p.Data)-p.Used {
		return 0, io.ErrShortBuffer
	}
	copy(p.Data[p.Used:p.Used+n], b[:n])
	p.Used += n
	return n, nil
}

func (p *ownedFakePipe) Close() error {
	p.Closed = true
	if p.CloseFail {
		return errors.New("owned_fake_close_failed")
	}
	return nil
}

func TestOwnedStdinAcknowledgedWritesAndFailureAreDistinct(t *testing.T) {
	input := []byte("ABCDEFG")
	complete := ownedFakePipe{ShortLimit: 3, FailureAt: -1}
	r := supplyStdin(&complete, input)
	if r.Bytes != 7 || !r.Complete || r.Code != "stdin_full_write_and_close" || !complete.Closed || !bytes.Equal(complete.Data[:complete.Used], input) {
		t.Fatalf("owned short writes must complete with exact acknowledgement: %+v", r)
	}
	partial := ownedFakePipe{ShortLimit: 2, FailureAt: 3}
	r = supplyStdin(&partial, input)
	if r.Bytes != 3 || r.Complete || r.Code != "stdin_write_failed" || !partial.Closed || string(partial.Data[:partial.Used]) != "ABC" {
		t.Fatalf("partial acknowledgement cannot be called complete: %+v", r)
	}
	closeFailure := ownedFakePipe{FailureAt: -1, CloseFail: true}
	r = supplyStdin(&closeFailure, input)
	if r.Bytes != 7 || r.Complete || r.Code != "stdin_close_failed" || !closeFailure.Closed {
		t.Fatalf("full writes with failed close remain incomplete: %+v", r)
	}
}
