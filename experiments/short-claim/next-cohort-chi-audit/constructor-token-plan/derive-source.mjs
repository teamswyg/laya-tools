// SPDX-License-Identifier: Apache-2.0
// Deterministic source preparation only. No Go, model, profile or network calls.
// Caller owns an empty temporary module outside the repository and its cleanup.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {fileURLToPath} from 'node:url';

export const pins = Object.freeze({
 baseline: [5971,'7776f6a6a4e16a20009e65c59b64604a2222980eb2fd8f787431aac24b319bdf'],
 candidate: [6344,'f359d49f7d9aa50cc75e4c6c4b3e16d22e2df06e649e4f3320b6c5f9e517ee5b'],
 probe: [18195,'e3eb062594602dae8ed59b72199defa8a3445782a6ffbd390712ba353fa3e780'],
 saved: [18373,'eb1ca246049664612bbf2fe30c73958b41afecac291524bade26076fce1db9bb'],
 fixture: [1417,'39e0aadbfe91c0b15ad3da257725de8d74c7a82c9d4222f73da8f937e72bf4c2'],
 oracle: [10429,'d0c906d1283617d0778039ff8bc4c6fff6cb0af3a813c48df088424d78bebb71'],
 scratch: [4176,'4435ac3716b16e1a6f12bac0aee475fe5492c1799a1103dde2a8b2e70c86e93e'],
 plan: [5221,'22e931e336e185b6175fadf54d86b54438a7302c4b37b682d8d5d6c5efa4167f'],
 goMod: [166,'d3dac31b277d3945ff0333c5099e5e5c47807bd8310b0c9d71d29dbdeff8e9ec'],
 goSum: [511,'b89ba980a5c517f5e9e52a67448d2db01d4b6be95a4afb54d7739a610a872132']
});
for (const pair of Object.values(pins)) Object.freeze(pair);
export const schema = 'riido-chi-constructor-token-plan-v1';
export const moduleText = 'module github.com/teamswyg/laya-tools/internal/constructor_token_plan_probe\n\ngo 1.27.1\n\nrequire github.com/teamswyg/laya-tools v0.0.0\n\nrequire golang.org/x/text v0.25.0 // indirect\n\nreplace github.com/teamswyg/laya-tools => ./repo\n';
const digest = b => crypto.createHash('sha256').update(b).digest('hex');
const check = (ok, code) => { if (!ok) throw Error(code); };
function pinned(bytes, name) {
 check(Buffer.isBuffer(bytes) && bytes.length === pins[name][0] && digest(bytes) === pins[name][1], 'source_pin_' + name);
 const text = bytes.toString('utf8');
 check(Buffer.from(text).equals(bytes), 'source_UTF8');
 return text;
}
function once(text, old, replacement) {
 const start = text.indexOf(old);
 check(start >= 0 && text.indexOf(old, start + old.length) < 0, 'derivation_anchor');
 return text.slice(0, start) + replacement + text.slice(start + old.length);
}
// Frozen, gofmt sources put top-level function closing braces at column zero.
// This is a pinned-source extractor, not a Go parser for arbitrary source text.
export function functionText(text, declaration) {
 const start = text.indexOf(declaration);
 check(start >= 0 && text.indexOf(declaration, start + declaration.length) < 0, 'function_anchor');
 const end = text.indexOf('\n}\n', start);
 check(end >= 0, 'function_end');
 return text.slice(start, end + 3);
}
function rewire(text) {
 text = text.replaceAll('riido-chi-constructor-scratch-v1', schema)
  .replace(/\bconstructor(?=[A-Z])/g, 'tokenPlan')
  .replaceAll('TestConstructor', 'TestTokenPlan')
  .replaceAll('tokenPlanLegacy', 'tokenPlanBaselineScratch');
 for (const flag of ['input','model','output','saved']) text = text.replaceAll('"constructor-' + flag + '"', '"tokenplan-' + flag + '"');
 for (const [old, next] of [['LegacyAnchor','BaselineScratchAnchor'],['ScratchAnchor','TokenPlanAnchor'],['Legacy','BaselineScratch'],['Scratch','TokenPlan']]) {
  text = text.replace(new RegExp('\\b' + old + '\\b', 'g'), next);
 }
 for (const [old, next] of [['legacy_prepare','baseline_scratch_prepare'],['scratch_prepare','token_plan_prepare'],['legacy_rank','baseline_scratch_rank'],['scratch_rank','token_plan_rank'],['legacy','baseline_scratch'],['scratch','token_plan']]) {
  text = text.replaceAll('"' + old + '"', '"' + next + '"');
 }
 text = text.replaceAll('paired scratch/legacy medians', 'paired token_plan/baseline_scratch elapsed medians');
 return text.replace(/\blegacy\b/g, 'baselineScratch').replace(/\bscratch\b/g, 'tokenOwner');
}
const gateSource = `func tokenPlanSavedGates(r tokenPlanReport) (bool, [2]float64) {
	pass := true
	var medians [2]float64
	for group := 0; group < 2; group++ {
		var ratios [6]float64
		for pair := 0; pair < 6; pair++ {
			a, b := r.Trials[group*12+pair*2], r.Trials[group*12+pair*2+1]
			if a.Method != "baseline_scratch" {
				a, b = b, a
			}
			ratios[pair] = float64(b.Cost.ElapsedNS) / float64(a.Cost.ElapsedNS)
			if b.Cost.TotalAllocBytes > a.Cost.TotalAllocBytes || b.Cost.Mallocs > a.Cost.Mallocs || a.Cost.Mallocs-b.Cost.Mallocs < 144 {
				pass = false
			}
		}
		sort.Float64s(ratios[:])
		medians[group] = (ratios[2] + ratios[3]) / 2
		if medians[group] > 1.0 {
			pass = false
		}
	}
	return pass, medians
}
`;
const extraControls = `
// Owned synthetic protocol controls; these are not measured performance/tasks.
func TestTokenPlanSavedMallocAndPairControls(t *testing.T) {
	for _, bad := range []uint64{257, 400, 401, ^uint64(0)} {
		r := tokenPlanSyntheticReport()
		r.Trials[1].Cost.Mallocs = bad
		if pass, _ := tokenPlanSavedGates(r); pass {
			t.Fatal("single pair malloc shortfall or unsigned wrap accepted")
		}
	}
	r := tokenPlanSyntheticReport()
	r.Trials[1] = r.Trials[0]
	b, e := json.Marshal(r)
	if e != nil { t.Fatal(e) }
	if _, e := tokenPlanSavedDecode(b); e == nil {
		t.Fatal("duplicated interval accepted")
	}
}
`;
export function buildSources(inputs) {
 const src = {};
 for (const name of Object.keys(pins)) src[name] = pinned(inputs[name], name);
 const baselineFn = functionText(src.baseline, 'func Prepare(');
 check(functionText(src.baseline, 'func (p *Prepared) Rank(') === functionText(src.candidate, 'func (p *Prepared) Rank('), 'Rank_changed');
 let probe = once(src.probe, functionText(src.probe, 'func constructorLegacy('), 'BASELINE_FUNCTION_PLACEHOLDER\n');
 probe = rewire(probe);
 probe = once(probe, 'BASELINE_FUNCTION_PLACEHOLDER\n', once(baselineFn, 'func Prepare(', 'func tokenPlanBaselineScratch('));
 // Both API strings are prepared before the measure/GC boundary; label lengths
 // must not introduce different allocation classes inside either timed path.
 probe = once(probe, '\tcalls := 0\n\tx.Cost = measure', '\tcalls := 0\n\tprepareAPI, rankAPI := method+"_prepare", method+"_rank"\n\tx.Cost = measure');
 probe = once(probe, 'tokenPlanCall(method+"_prepare", prepare)', 'tokenPlanCall(prepareAPI, prepare)');
 probe = once(probe, 'tokenPlanCall(method+"_rank", func()', 'tokenPlanCall(rankAPI, func()');
 const scope = 'One exposed development parent; baseline_scratch versus token_plan; all 24 intervals pay full Prepare and N Rank including token plans, prefixes and both cross-hash passes; API labels, GC/MemStats and parity conversion outside elapsed interval; Go cumulative allocation, not peak/retained heap/RSS/GPU; no profiles, labels, roles, Fit, qualification, activation or protected evaluation. Instrumented wrappers exclude accessors/runtime/hidden calls.';
 const oldScopeLine = probe.split('\n').find(l => l.startsWith('const tokenPlanScope = '));
 probe = once(probe, oldScopeLine, 'const tokenPlanScope = ' + JSON.stringify(scope));
 let saved = rewire(src.saved);
 saved = once(saved, functionText(saved, 'func tokenPlanSavedGates('), gateSource);
 saved = once(saved, '"../../experiments/short-claim/next-cohort-chi-audit/preview/INPUT.public.v3.json"', '"INPUT.public.v3.json"');
 saved = once(saved, 'Cost: tokenPlanCost{ElapsedNS: 1000, TotalAllocBytes: 1000, Mallocs: 1}', 'Cost: tokenPlanCost{ElapsedNS: 1000, TotalAllocBytes: 1000, Mallocs: 400}');
 saved = once(saved, 'x.Cost.TotalAllocBytes = 800', 'x.Cost.TotalAllocBytes = 1000\n\t\t\t\t\tx.Cost.Mallocs = 256');
 saved = once(saved, 'r.Trials[1].Cost.TotalAllocBytes = 801', 'r.Trials[1].Cost.TotalAllocBytes = 1001');
 saved = once(saved, 'r.Trials[i].Cost.ElapsedNS = 1250', 'r.Trials[i].Cost.ElapsedNS = 1000');
 saved = once(saved, 'r.Trials[i].Cost.ElapsedNS = 1251', 'r.Trials[i].Cost.ElapsedNS = 1001');
 saved += extraControls;
 for (const text of [probe, saved]) {
  check(!/"(?:legacy|scratch)(?:_prepare|_rank)?"/.test(text) && !text.includes('riido-chi-constructor-scratch-v1') && !text.includes('TestConstructor') && !/"constructor-(?:input|model|output|saved)"/.test(text), 'old_protocol_remaining');
 }
 check(functionText(probe, 'func tokenPlanBaselineScratch(') === once(baselineFn, 'func Prepare(', 'func tokenPlanBaselineScratch('), 'baseline_body_changed');
 return {
  'prepared.go': Buffer.from(src.candidate),
  'token_probe_test.go': Buffer.from(probe),
  'token_saved_test.go': Buffer.from(saved),
  'INPUT.public.v3.json': Buffer.from(src.fixture),
  'go.mod': Buffer.from(moduleText),
  'go.sum': Buffer.from(src.goSum)
 };
}
export function readInputs(repoRoot, baselinePath) {
 const files = {candidate:'pkg/hintprepared/prepared.go',probe:'pkg/hintprepared/constructor_probe_test.go',saved:'pkg/hintprepared/constructor_saved_test.go',fixture:'experiments/short-claim/next-cohort-chi-audit/preview/INPUT.public.v3.json',oracle:'internal/hintlearn/learn.go',scratch:'internal/hintlearn/features_scratch.go',plan:'internal/hintlearn/features_tokenplan.go',goMod:'go.mod',goSum:'go.sum'};
 const input = {baseline: fs.readFileSync(baselinePath)};
 for (const [name, relative] of Object.entries(files)) input[name] = fs.readFileSync(path.join(repoRoot, relative));
 return input;
}
export function outputPath(outputDir, name) {
 check(['prepared.go','token_probe_test.go','token_saved_test.go','INPUT.public.v3.json','go.mod','go.sum'].includes(name), 'output_name');
 const result = path.resolve(outputDir, name);
 check(path.dirname(result) === path.resolve(outputDir), 'output_escape');
 return result;
}
export function derive({repoRoot, baselinePath, outputDir}) {
 check(path.isAbsolute(repoRoot) && path.isAbsolute(baselinePath) && path.isAbsolute(outputDir), 'absolute_paths_required');
 const repo = fs.realpathSync(repoRoot), out = fs.realpathSync(outputDir);
 check(fs.lstatSync(outputDir).isDirectory() && out !== repo && !out.startsWith(repo + path.sep) && fs.readdirSync(out).length === 0, 'owned_empty_output_required');
 const generated = buildSources(readInputs(repo, baselinePath));
 // Relative replacement is reproducible; the private symlink is not published.
 fs.symlinkSync(repo, path.join(out, 'repo'), 'dir');
 const files = [];
 for (const [name, bytes] of Object.entries(generated)) {
  const fd = fs.openSync(outputPath(out, name), 'wx', 0o600);
  try { fs.writeFileSync(fd, bytes); fs.fsyncSync(fd); } finally { fs.closeSync(fd); }
  files.push({path:name,bytes:bytes.length,sha256:digest(bytes)});
 }
 return {schema:'riido-token-plan-derived-sources-v1',report_schema:schema,module:'github.com/teamswyg/laya-tools/internal/constructor_token_plan_probe',source_pins:pins,files,scope:'Source derivation only. No model/Go/training execution; caller owns temporary directory and cleanup.'};
}
if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
 try {
  const args = process.argv.slice(2);
  check(args.length === 6 && args[0] === '--repo' && args[2] === '--baseline' && args[4] === '--out', 'arguments');
  console.log(JSON.stringify(derive({repoRoot:args[1],baselinePath:args[3],outputDir:args[5]})));
 } catch { console.error('token_plan_source_derivation_failed'); process.exitCode = 1; }
}
