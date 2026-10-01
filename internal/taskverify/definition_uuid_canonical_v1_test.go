package taskverify

import (
	"bytes"
	"context"
	"crypto/sha1" // Public Git blob identity, not an authentication primitive.
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

const uuidCanonicalCorrect = `
// ParseCanonical accepts the project-selected lowercase canonical text form.
func ParseCanonical(s string) (UUID,error) {
 if len(s)!=36 { return Nil,ErrInvalidUUIDFormat }
 for i:=0;i<len(s);i++ {
  if i==8||i==13||i==18||i==23 { if s[i]!='-' { return Nil,ErrInvalidUUIDFormat }; continue }
  b:=s[i]; if !((b>='0'&&b<='9')||(b>='a'&&b<='f')) { return Nil,ErrInvalidUUIDFormat }
 }
 value,err:=Parse(s); if err!=nil { return Nil,err }; return value,nil
}
`

func uuidCanonicalFixture(t *testing.T) []BaseFile {
	t.Helper()
	d := uuidCanonicalDefinition()
	out := make([]BaseFile, 0, len(d.Files))
	for _, f := range d.Files {
		raw, err := os.ReadFile(filepath.Join("testdata/uuid-canonical-v1", f.Path))
		if err != nil || digest(raw) != f.SHA256 {
			t.Fatal("UUID public source/attribution pin mismatch", f.Path)
		}
		wantBlob := ""
		for _, b := range uuidCanonicalPublicBlobs {
			if b.path == f.Path {
				wantBlob = b.sha
			}
		}
		blob := sha1.New()
		fmt.Fprintf(blob, "blob %d%c", len(raw), 0)
		blob.Write(raw)
		if wantBlob == "" || hex.EncodeToString(blob.Sum(nil)) != wantBlob {
			t.Fatal("UUID original Git blob mismatch", f.Path)
		}
		out = append(out, BaseFile{Path: f.Path, SHA256: f.SHA256, Data: raw})
	}
	return out
}

// uuidCanonicalOfflineCheck compares the original suite with the independently
// supplied contract during development. The integration test below separately
// checks the central Verify pipeline. Neither bypasses unsupported isolation or
// publishes raw upstream test output.
func uuidCanonicalOfflineCheck(t *testing.T, files []BaseFile, mutable string, contract bool) testOutcome {
	t.Helper()
	result := testOutcome{code: "independent_tests_failed"}
	if !sandboxSupported() {
		return testOutcome{unknown: true, code: "isolation_unavailable"}
	}
	dir := t.TempDir()
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	dir = resolved
	for _, f := range files {
		data := f.Data
		if f.Path == "uuid.go" {
			data = []byte(mutable)
		}
		if err := os.WriteFile(filepath.Join(dir, f.Path), data, 0600); err != nil {
			t.Fatal("UUID stage failed")
		}
	}
	// The original module is pinned above; this explicit trusted execution recipe
	// chooses language/toolchain 1.27.1 without a dependency or network requirement.
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+uuidCanonicalModulePath+"\n\ngo 1.27.1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if contract {
		if err := os.WriteFile(filepath.Join(dir, "taskverify_contract_test.go"), []byte(uuidCanonicalContractTests), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cache, work := filepath.Join(dir, "cache"), filepath.Join(dir, "work")
	for _, p := range []string{cache, work} {
		if err := os.Mkdir(p, 0700); err != nil {
			t.Fatal(err)
		}
	}
	toolchain, err := filepath.EvalSymlinks(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(toolchain, "bin", "go")
	ctx, cancel := context.WithTimeout(context.Background(), MaxTimeout)
	defer cancel()
	env := isolatedEnv(toolchain, dir, cache, work)
	probe, err := sandboxCommand(ctx, binary, []string{"version"}, toolchain, dir)
	if err != nil {
		return testOutcome{unknown: true, code: "isolation_unavailable"}
	}
	probe.Dir, probe.Env = dir, env
	probeOut := &boundedOutput{cancel: cancel}
	probe.Stdout, probe.Stderr = probeOut, probeOut
	probe.WaitDelay = time.Second
	if probe.Run() != nil || probeOut.overflow || strings.TrimSpace(probeOut.b.String()) != "go version go1.27.1 "+runtime.GOOS+"/"+runtime.GOARCH {
		return testOutcome{unknown: true, code: "isolation_execution_unavailable"}
	}
	cmd, err := sandboxCommand(ctx, binary, []string{"test", "-json", "-count=1", "-timeout=10s", uuidCanonicalModulePath}, toolchain, dir)
	if err != nil {
		return testOutcome{unknown: true, code: "isolation_unavailable"}
	}
	cmd.Dir, cmd.Env = dir, env
	out := &boundedOutput{cancel: cancel}
	cmd.Stdout, cmd.Stderr = out, out
	cmd.WaitDelay = time.Second
	runErr := cmd.Run()
	result.isolated = true
	if ctx.Err() != nil || out.overflow {
		return testOutcome{unknown: true, isolated: true, code: "check_resource_limit"}
	}
	if runErr != nil {
		if bytes.Contains(out.b.Bytes(), []byte("sandbox-exec:")) || bytes.Contains(out.b.Bytes(), []byte("operation not permitted")) {
			result.unknown, result.code = true, "isolation_execution_unavailable"
			return result
		}
		if contract {
			terminated := make(map[string]bool)
			failed := false
			for _, line := range bytes.Split(out.b.Bytes(), []byte{'\n'}) {
				var event struct{ Action, Package, Test string }
				if json.Unmarshal(line, &event) != nil {
					continue
				}
				if event.Package != uuidCanonicalModulePath {
					continue
				}
				for _, required := range uuidCanonicalDefinition().RequiredTests {
					if event.Test == required.Test && (event.Action == "pass" || event.Action == "fail") {
						terminated[event.Test] = true
						failed = failed || event.Action == "fail"
					}
				}
			}
			result.tests = len(terminated)
			if result.tests == 6 && failed {
				result.code = "authored_contract_failed"
			}
		}
		return result
	}
	if contract {
		return versionedTerminalPasses(uuidCanonicalDefinition(), out.b.Bytes())
	}
	packagePassed := false
	for _, line := range bytes.Split(out.b.Bytes(), []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var event struct{ Action, Package, Test string }
		if json.Unmarshal(line, &event) != nil {
			return testOutcome{isolated: true, code: "invalid_test_output"}
		}
		if event.Package == uuidCanonicalModulePath && event.Action == "pass" {
			if event.Test == "" {
				packagePassed = true
			} else {
				result.tests++
			}
		}
	}
	result.passed = packagePassed && result.tests > 0
	if result.passed {
		result.code = "upstream_tests_passed"
	}
	return result
}

func uuidCanonicalMutable(t *testing.T, files []BaseFile) string {
	t.Helper()
	for _, f := range files {
		if f.Path == "uuid.go" {
			return string(f.Data)
		}
	}
	t.Fatal("missing uuid source")
	return ""
}
func uuidCanonicalReplace(t *testing.T, body, old, new string) string {
	t.Helper()
	if strings.Count(body, old) != 1 {
		t.Fatal("authored UUID mutant anchor not unique")
	}
	return strings.Replace(body, old, new, 1)
}

func TestUUIDCanonicalPublicPinsAndClosedBaseline(t *testing.T) {
	files := uuidCanonicalFixture(t)
	d := uuidCanonicalDefinition()
	if d.BaseRevision != uuidCanonicalRevision || d.ModulePath != uuidCanonicalModulePath || d.License != "BSD-3-Clause" || len(d.Files) != 25 || definitionPin(d, "NOTICE") != "" {
		t.Fatal("UUID origin/attribution scope changed")
	}
	if len(d.RequiredTests) != 6 || d.ContractSHA256 != digest([]byte(uuidCanonicalContractTests)) || d.PromptSHA256 != digest([]byte(uuidCanonicalPrompt)) {
		t.Fatal("UUID independent contract not bound")
	}
	baseline := uuidCanonicalMutable(t, files)
	formattedBaseline, formatErr := format.Source([]byte(baseline))
	if formatErr != nil || len(formattedBaseline) != 10251 || digest(formattedBaseline) != "b19f8aad57fbe17219758f0142742fb3e1879d26506f7fce99169b3ec770c7e8" || !uuidCanonicalSupportedSource(formattedBaseline) {
		t.Fatal("trusted formatter normalization of original UUID source changed")
	}
	if !supportedDefinitionSource(d, []byte(baseline)) {
		t.Fatal("original source shape unsupported")
	}
	if !sandboxSupported() {
		t.Skip("UUID offline isolation unsupported; no outcome claimed")
	}
	upstream := uuidCanonicalOfflineCheck(t, files, baseline, false)
	if !upstream.passed || upstream.unknown || !upstream.isolated {
		t.Fatal("pinned full upstream Go suite did not pass offline", upstream.code)
	}
	missingAPI := uuidCanonicalOfflineCheck(t, files, baseline, true)
	if missingAPI.passed || missingAPI.unknown || !missingAPI.isolated {
		t.Fatal("unchanged baseline was accepted or not assessed", missingAPI.code)
	}
	t.Logf("development evidence: 25 pinned files; full upstream tests %d passed; absent new API rejects independent contract; no model attempts", upstream.tests)
}

func TestUUIDCanonicalCorrectAndSeededWrong(t *testing.T) {
	files := uuidCanonicalFixture(t)
	baseline := uuidCanonicalMutable(t, files)
	if !sandboxSupported() {
		t.Skip("UUID offline isolation unsupported; no outcome claimed")
	}
	modes := []string{"correct", "permissive-Parse", "lowercase-normalization", "uppercase-36", "skip-last-lowercase", "digits-only", "nil-max-only", "rfc-variant-only", "always-zero", "reversed-bytes", "partial-on-error", "nil-error-on-invalid", "trim-whitespace", "strip-urn", "strip-wrapper"}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			function := uuidCanonicalCorrect
			switch mode {
			case "correct":
			case "permissive-Parse":
				function = "\nfunc ParseCanonical(s string)(UUID,error){ return Parse(s) }\n"
			case "lowercase-normalization":
				function = "\nfunc ParseCanonical(s string)(UUID,error){ return Parse(strings.ToLower(s)) }\n"
			case "uppercase-36":
				function = "\nfunc ParseCanonical(s string)(UUID,error){ if len(s)!=36{return Nil,ErrInvalidUUIDFormat}; v,e:=Parse(s);if e!=nil{return Nil,e};return v,nil }\n"
			case "skip-last-lowercase":
				function = strings.Replace(function, "if !((b>='0'", "if i!=35&&!((b>='0'", 1)
			case "digits-only":
				function = strings.Replace(function, "||(b>='a'&&b<='f')", "", 1)
			case "nil-max-only":
				function = strings.Replace(function, "value,err:=Parse(s);", "if s!=Nil.String()&&s!=Max.String(){return Nil,ErrInvalidUUIDFormat}; value,err:=Parse(s);", 1)
			case "rfc-variant-only":
				function = strings.Replace(function, "return value,nil", "if value.Variant()!=RFC4122{return Nil,ErrInvalidUUIDFormat}; return value,nil", 1)
			case "always-zero":
				function = strings.Replace(function, "return value,nil", "_ = value; return Nil,nil", 1)
			case "reversed-bytes":
				function = strings.Replace(function, "return value,nil", "for i:=0;i<8;i++{value[i],value[15-i]=value[15-i],value[i]};return value,nil", 1)
			case "partial-on-error":
				function = "\nfunc ParseCanonical(s string)(UUID,error){if len(s)!=36||s!=strings.ToLower(s){return Nil,ErrInvalidUUIDFormat};return Parse(s)}\n"
			case "nil-error-on-invalid":
				function = strings.ReplaceAll(function, "return Nil,ErrInvalidUUIDFormat", "return Nil,nil")
			case "trim-whitespace":
				function = strings.Replace(function, "if len(s)!=36", "s=strings.TrimSpace(s); if len(s)!=36", 1)
			case "strip-urn":
				function = strings.Replace(function, "if len(s)!=36", "s=strings.TrimPrefix(s,\"urn:uuid:\"); if len(s)!=36", 1)
			case "strip-wrapper":
				function = strings.Replace(function, "if len(s)!=36", "s=strings.Trim(s,\"{}[]\"); if len(s)!=36", 1)
			}
			source := baseline + function
			if !supportedDefinitionSource(uuidCanonicalDefinition(), []byte(source)) {
				t.Fatal("authored control unexpectedly changed supported source shape", mode)
			}
			outcome := uuidCanonicalOfflineCheck(t, files, source, true)
			if mode == "correct" {
				if !outcome.passed || outcome.unknown || !outcome.isolated || outcome.tests < 6 {
					t.Fatal("independent UUID positive control failed", outcome.code)
				}
			} else if outcome.passed || outcome.unknown || !outcome.isolated || outcome.tests != 6 || outcome.code != "authored_contract_failed" {
				t.Fatal("seeded UUID wrong control accepted or unassessed", mode, outcome.code)
			}
			t.Logf("development control=%s isolated=%t accepted=%t tests=%d no_model=true", mode, outcome.isolated, outcome.passed, outcome.tests)
		})
	}
}

func TestUUIDCanonicalUnsupportedSourceShape(t *testing.T) {
	files := uuidCanonicalFixture(t)
	baseline := uuidCanonicalMutable(t, files)
	source := baseline + uuidCanonicalCorrect
	d := uuidCanonicalDefinition()
	modes := []string{"import", "alias", "init", "directive", "package", "delete-attribution", "legacy-uppercase-regression", "legacy-raw-regression", "entropy-read", "direct-crypto-rand", "reset-reader", "disable-pool", "global-write", "global-slice-alias", "global-copy-write", "pure-call-alias", "correct-extra-helper", "local-function", "scope-shadow-pool", "scope-shadow-pool-enabled", "scope-shadow-pool-pos"}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			changed := source
			add := func(statement string) string {
				return uuidCanonicalReplace(t, source, "if len(s)!=36", statement+"; if len(s)!=36")
			}
			switch mode {
			case "import":
				changed = uuidCanonicalReplace(t, source, "\"bytes\"", "\"bytes\"\n\"os\"")
			case "alias":
				changed = uuidCanonicalReplace(t, source, "\"bytes\"", "bytesAlias \"bytes\"")
			case "init":
				changed += "\nfunc init(){}\n"
			case "directive":
				changed += "\n//go:generate unused\n"
			case "package":
				changed = uuidCanonicalReplace(t, source, "package uuid", "package other")
			case "delete-attribution":
				changed = strings.TrimPrefix(source, "// Copyright 2018 Google Inc.  All rights reserved.\n")
			case "legacy-uppercase-regression":
				changed = uuidCanonicalReplace(t, source, "func Parse(s string) (UUID, error) {", "func Parse(s string) (UUID, error) { if s!=strings.ToLower(s){return Nil,ErrInvalidUUIDFormat}")
			case "legacy-raw-regression":
				changed = uuidCanonicalReplace(t, source, "func ParseBytes(b []byte) (UUID, error) {", "func ParseBytes(b []byte) (UUID, error) { if len(b)==32{return Nil,ErrInvalidUUIDFormat}")
			case "entropy-read":
				changed = add("_,_=NewRandom()")
			case "direct-crypto-rand":
				changed = add("buf:=make([]byte,1); _,_=rand.Read(buf)")
			case "reset-reader":
				changed = add("SetRand(nil)")
			case "disable-pool":
				changed = add("DisableRandPool()")
			case "global-write":
				changed = add("Nil=UUID{}")
			case "global-slice-alias":
				changed = add("v:=Nil[:]; _=v")
			case "global-copy-write":
				changed = add("copy(Nil[:],make([]byte,16))")
			case "pure-call-alias":
				changed = uuidCanonicalReplace(t, source, "value,err:=Parse(s);", "parseAlias:=Parse; value,err:=parseAlias(s);")
			case "correct-extra-helper":
				changed += "\nfunc canonicalUnused(s string)(UUID,error){return Parse(s)}\n"
			case "local-function":
				changed = add("f:=func(){};f()")
			case "scope-shadow-pool":
				changed = add("if false { pool := [256]byte{}; _=pool }; pool[0]=77")
			case "scope-shadow-pool-enabled":
				changed = add("if false { poolEnabled := false; _=poolEnabled }; poolEnabled=false")
			case "scope-shadow-pool-pos":
				changed = add("if false { poolPos := 0; _=poolPos }; poolPos=0")
			}
			if changed == source {
				t.Fatal("authored unsupported control did not change source", mode)
			}
			if uuidCanonicalSupportedSource([]byte(changed)) || supportedDefinitionSource(d, []byte(changed)) {
				t.Fatal("unsupported UUID shape should become verifier_unknown", mode)
			}
			t.Logf("development shape=%s supported=false no_model=true", mode)
		})
	}
}

func TestUUIDCanonicalVerifierIntegration(t *testing.T) {
	files := uuidCanonicalFixture(t)
	var base, attribution []BaseFile
	for _, f := range files {
		if f.Path == "LICENSE" {
			attribution = append(attribution, f)
		} else {
			base = append(base, f)
		}
	}
	if _, err := validateBase(uuidCanonicalTask, uuidCanonicalRevision, base); err != nil {
		t.Fatal("public UUID closure unavailable", err)
	}
	if _, err := validateAttribution(uuidCanonicalTask, attribution); err != nil {
		t.Fatal("public UUID LICENSE unavailable", err)
	}
	source := uuidCanonicalMutable(t, files) + uuidCanonicalCorrect
	formatted, err := format.Source([]byte(source))
	if err != nil || !uuidCanonicalSupportedSource(formatted) {
		t.Fatal("formatted positive control changed original prefix")
	}
	for _, mode := range []string{"correct-candidate-tests-ignored", "wrong-permissive-Parse", "missing-api", "direct-crypto-rand", "delete-attribution", "scope-shadow-global-write", "correct-extra-helper"} {
		t.Run(mode, func(t *testing.T) {
			changed := string(formatted)
			switch mode {
			case "missing-api":
				changed = uuidCanonicalMutable(t, files)
			case "wrong-permissive-Parse":
				changed = uuidCanonicalMutable(t, files) + "\nfunc ParseCanonical(s string)(UUID,error){return Parse(s)}\n"
			case "direct-crypto-rand":
				changed = strings.Replace(changed, "if len(s) != 36", "b:=make([]byte,1);_,_=rand.Read(b); if len(s) != 36", 1)
			case "delete-attribution":
				changed = strings.TrimPrefix(changed, "// Copyright 2018 Google Inc.  All rights reserved.\n")
			case "scope-shadow-global-write":
				changed = strings.Replace(changed, "if len(s) != 36", "if false { pool := [256]byte{}; _=pool }; pool[0]=77; if len(s) != 36", 1)
			case "correct-extra-helper":
				changed += "\nfunc canonicalUnused(s string)(UUID,error){return Parse(s)}\n"
			}
			formatted, err := format.Source([]byte(changed))
			if err != nil {
				t.Fatal("authored shape control must compile syntactically")
			}
			dir := candidateFixture(t, files)
			if err := os.WriteFile(filepath.Join(dir, "uuid.go"), formatted, 0600); err != nil {
				t.Fatal(err)
			}
			// Candidate assertions are deliberately broken and must never be executed.
			if err := os.WriteFile(filepath.Join(dir, "uuid_test.go"), []byte("package uuid\ninvalid authored candidate assertions\n"), 0600); err != nil {
				t.Fatal(err)
			}
			req := Request{TaskID: uuidCanonicalTask, BaseRevision: uuidCanonicalRevision, BaseFiles: base, AttributionFiles: attribution, CandidateDir: dir, Timeout: MaxTimeout}
			r, err := Verify(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			if r.CandidateTestsUsed || r.AttributionSHA256 == "" {
				t.Fatal("UUID attribution/independent checks were not bound")
			}
			switch mode {
			case "correct-candidate-tests-ignored":
				if !sandboxSupported() {
					if r.Accepted || r.Status != "verifier_unknown" {
						t.Fatal("unsupported platform claimed outcome")
					}
					return
				}
				if !r.Accepted || r.Status != "accepted" || !r.ExecutionIsolated || r.IndependentTests < 6 {
					t.Fatal("UUID positive pipeline did not pass", r)
				}
			case "missing-api":
				// Formatting the pinned prefix changes its bytes, so this control
				// reaches the compiler to establish the missing API. Platforms
				// without an enforced sandbox cannot establish that outcome.
				if !sandboxSupported() {
					if r.Accepted || r.Status != "verifier_unknown" || r.ExecutionIsolated || r.Checks[len(r.Checks)-1].Code != "isolation_unavailable" {
						t.Fatal("unsupported platform claimed missing-API outcome")
					}
					return
				}
				if r.Accepted || r.Status != "rejected" || !r.ExecutionIsolated {
					t.Fatal("normalized UUID without API accepted or untested", r)
				}
			case "wrong-permissive-Parse":
				if !sandboxSupported() {
					if r.Accepted || r.Status != "verifier_unknown" {
						t.Fatal("unsupported platform claimed outcome")
					}
					return
				}
				if r.Accepted || r.Status != "rejected" || !r.ExecutionIsolated {
					t.Fatal("wrong UUID supported pipeline accepted or untested", r)
				}
			default:
				if r.Accepted || r.Status != "verifier_unknown" || r.ExecutionIsolated {
					t.Fatal("UUID unsupported shape should not become capability label", mode, r)
				}
			}
		})
	}
	dir := candidateFixture(t, files)
	for _, mode := range []string{"wrong-revision", "wrong-source", "wrong-license", "invented-notice"} {
		t.Run(mode, func(t *testing.T) {
			req := Request{TaskID: uuidCanonicalTask, BaseRevision: uuidCanonicalRevision, BaseFiles: slices.Clone(base), AttributionFiles: slices.Clone(attribution), CandidateDir: dir}
			switch mode {
			case "wrong-revision":
				req.BaseRevision = eventKeyBoundsRevision
			case "wrong-source":
				req.BaseFiles[0].Data = []byte("wrong original source")
			case "wrong-license":
				req.AttributionFiles[0].Data = []byte("wrong original attribution")
			case "invented-notice":
				req.AttributionFiles = append(req.AttributionFiles, BaseFile{Path: "NOTICE", Data: []byte("not present upstream")})
			}
			r, err := Verify(context.Background(), req)
			if err == nil || r.Accepted {
				t.Fatal("tampered UUID origin accepted")
			}
		})
	}
}

func TestUUIDCanonicalFrozenIdentity(t *testing.T) {
	d, ok := TaskDefinition(uuidCanonicalTask)
	if !ok {
		t.Fatal("UUID definition missing")
	}
	s, err := TaskSpec(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	db, de := json.Marshal(d)
	sb, se := json.Marshal(s)
	if de != nil || se != nil || s.DefinitionSHA256 != digest(db) || d.ContractSHA256 != digest([]byte(uuidCanonicalContractTests)) || d.PromptSHA256 != digest([]byte(uuidCanonicalPrompt)) {
		t.Fatal("UUID identities not bound")
	}
	if digest(sb) != "c8e3a2edcc2024ce68ba7ea7b5fed00554b2e1c36a12174f942f32b5a5a5eec3" || digest(db) != "9413ed54a10ee9a1576f71a68cf664134570cf84e735f8daeb88c66c1f4ad7c8" || d.ContractSHA256 != "259f3fcd56ac9bf0d4d307cc1b8ed77968b6b3bf21327160e09efd440f71ffd5" || d.PromptSHA256 != "5c6456db8da12a5e9510cf9f39a448298df4dd73bead5706a41288f567f30cb2" {
		t.Fatal("UUID version1 identity changed in place")
	}
	t.Logf("UUID development pins spec=%s definition=%s contract=%s prompt=%s", digest(sb), digest(db), d.ContractSHA256, d.PromptSHA256)
}
