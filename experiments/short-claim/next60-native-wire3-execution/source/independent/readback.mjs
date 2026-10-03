// SPDX-License-Identifier: Apache-2.0
// Saved-only independent audit, not yet executed; Root executes.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
const UNKNOWN = Symbol('unobserved');
const sha = b => crypto.createHash('sha256').update(b).digest('hex');
const same = (a, b) => JSON.stringify(a) === JSON.stringify(b);
function need(ok, id) { if (!ok) throw new Error(id); }
const inputs = [
['evidence', 'next60-native-wire3-cache-reset-actual-run-Root-v2/evidence.jsonl', 14303, '08cde54e854e824156acbf3b7438c640567317cb8e1ea64d49252a08b5e065e1', 65536],
['journal', 'next60-native-wire3-cache-reset-actual-run-Root-v2/journal.jsonl', 404858, '4de5a9f30f180ce872cce2d6667835291a148a6c2373bddd4f802e2214d115fe', 1310720],
['fixtures', 'next60-union-duplicate-literal-preparation52-v1/FIXTURES-WANTS.proposed.v1.json', 11940, 'a12a05374ba8b67390c3b779ec2d9485ce14fc00a6ca9ff7aa1534c0d7efb0db', 11940],
['captions', 'next60-union-duplicate-literal-preparation52-v1/REQUESTS-CAPTIONS.proposed.v1.json', 3155, 'a459a046db273e684d6558daea4b1675d24235094a86fe9b3cf778ebe0c22f4d', 3155],
['freeze', 'UNION-DUPLICATE-LITERAL-FREEZE.actual.public.v1.json', 16071, 'dac90a4be5acaa26e289c90451701462b74d7b7c437af5144f4ce84a970f58b8', 16071],
['comparison', 'NATIVE-CACHE-RESET-WANTED-COMPARISON.actual.public.v1.json', 38595, 'ad9d33c619767a725ccadfc25bf97ee77cc62a75c99bbd6eeae77eab82f17345', 65536],
];
const root = process.argv[2];
need(process.argv.length === 3 && path.isAbsolute(root), 'one_absolute_root');
const data = Object.create(null);
for (const [id, relative, bytes, hash, cap] of inputs) {
const p = path.join(root, relative), st = fs.lstatSync(p);
need(st.isFile() && !st.isSymbolicLink() && st.nlink === 1 && fs.realpathSync(p) === p && st.size === bytes && bytes <= cap, 'input_metadata_' + id);
const b = fs.readFileSync(p);
need(b.length === bytes && sha(b) === hash, 'input_pin_' + id);
data[id] = b;
}
function jsonLines(buffer, count, limit) {
const out = []; let start = 0;
while (start < buffer.length) {
const end = buffer.indexOf(10, start);
need(end >= start && end - start + 1 <= limit, 'LF_bound');
const body = buffer.subarray(start, end);
need(!body.includes(13) && body.length > 0, 'LF_payload');
const value = JSON.parse(body.toString('utf8'));
out.push({offset: start, length: end - start + 1, body, value});
start = end + 1;
}
need(out.length === count, 'line_count'); return out;
}
const carrier = JSON.parse(data.fixtures), freeze = JSON.parse(data.freeze), captions = JSON.parse(data.captions);
const saved = JSON.parse(data.comparison);
const el = jsonLines(data.evidence, 41, 1536), jl = jsonLines(data.journal, 528, 2048);
const pins = {f: inputs[4][3], i: inputs[2][3], c: inputs[3][3]};
need(same(el[0].value, {v: 1, p: pins}) && same(el[40].value, {v: 1, r: 39, c: true}), 'file_envelopes');
need(carrier.records.length === 13 && freeze.dispatches.length === 39 && captions.requests.length === 2, 'frozen_dimensions');
need(same(freeze.family_roles.whole_existing_family_groups, [79, 83]) && freeze.family_roles.new_numeric_group_or_role_assignments === 0, 'whole_groups');
for (let q = 0; q < 2; q++) need(freeze.requests[q].family_context.existing_group === [79, 83][q] && freeze.requests[q].family_context.existing_role === 'development_train', 'fixed_role');
const after = Array(39), method = Array.from({length: 39}, () => []), cbN = Array(39).fill(0);
const stack = [], counters = Array.from({length: 4}, () => [0, 0, 0, 0]);
const stageUnit = {1: 0, 2: 0, 3: 1, 4: 1, 5: 2, 6: 2, 7: 3, 8: 3};
function atom(b) { return b.a && b.k ? b.n ? null : b.v : UNKNOWN; }
function bool(mask) {
need([0, 1, 3, 7, 11].includes(mask), 'bool_mask');
return mask < 3 ? UNKNOWN : mask === 7 ? null : mask === 11;
}
function hasKnown(mask) { return (mask & 3) === 3; }
function maskBool(b) { return +b.a + 2 * +b.k + 4 * +b.n + 8 * +b.v; }
for (let i = 0; i < jl.length; i++) {
const z = jl[i].value;
need(z.v === 3 && z.q === i + 1, 'frame_identity');
if (i === 0 || i === jl.length - 1) {
need(z.k === (i === 0 ? 0 : 9) && z.d === 255 && z.s === 0 && z.p === 0 && same(z.b, pins), 'boundary_frame');
need(stack.length === 0 && same(z.n, counters), 'boundary_counter'); continue;
}
const unit = stageUnit[z.k];
need(unit !== undefined && z.d < 39, 'stage_domain');
const frozen = freeze.dispatches[z.d], ids = [frozen.request_ordinal, frozen.fixture_index, frozen.display, frozen.display];
need(same(z.x, ids), 'frame_schedule');
if (z.k % 2 === 1) {
need(z.s === z.q && z.p === (stack.at(-1)?.q ?? 0), 'before_parent');
if (unit === 0) need(stack.length === 0, 'row_parent');
if (unit === 1) need(stack.at(-1)?.unit === 0, 'candidate_parent');
const cand = unit === 1 ? z.q : stack.find(x => x.unit === 1)?.q;
if (unit >= 2) need(cand !== undefined, 'active_candidate');
const span = {q: z.q, unit, parent: z.p, d: z.d, x: z.x, m: z.m, c: z.c, candidate: cand};
if (unit === 2) { need(z.x[0] === 0 ? z.m === 1 : z.m >= 2 && z.m <= 5, 'method_domain'); method[z.d].push(span); }
if (unit === 3) { need(stack.at(-1)?.unit === 2, 'callback_parent'); cbN[z.d]++; }
stack.push(span); counters[unit][0]++;
need(same(z.n, counters), 'before_counter'); counters[unit][1]++;
} else {
const b = stack.pop();
need(b && b.unit === unit && b.q === z.s && b.parent === z.p && b.d === z.d && same(b.x, z.x) && b.m === z.m && b.c === z.c && z.f !== null, 'after_pair');
counters[unit][2]++;
if (atom(z.f.p) === true) counters[unit][3]++;
need(same(z.n, counters), 'after_counter');
if (unit === 2) { b.after = z.q; b.source = z.f.a; }
if (unit === 1) { need(!after[z.d] && z.r !== null, 'unique_row_ref'); after[z.d] = z; }
}
}
const final = jl.at(-1).value;
need(same(counters, [[39,39,39,0],[39,39,39,0],[96,96,96,0],[89,89,89,0]]) && final.z.c === true && final.z.r === 0 && final.z.s === null && final.z.n === null, 'complete_observed_counts');
need(same(final.e, {n: data.evidence.length, h: sha(data.evidence), c: 39}), 'final_file_ref');
const objectPaths = [null, '', '/outer', '/0', '/1'];
const decodedKeys = [null, '', 'a', 'b', 'k', 'outer', 'z'];
const entryNames = [null, 'a', 'b', 'c'];
const errorIDs = {io_EOF: 2, owned_layer_cause: 3, owned_duplicate_key: 4, owned_invalid_JSON: 5};
function errID(s) { need(s === null || Object.hasOwn(errorIDs, s), 'Wanted_identity'); return s === null ? 1 : errorIDs[s]; }
function small(s, dictionary) {
if (!hasKnown(s.m)) return UNKNOWN;
if (s.m === 7) return null;
if (dictionary) return s.v > 0 && s.v < dictionary.length ? dictionary[s.v] : Symbol('invalid_value');
return s.v;
}
function tri(actual, wanted) { return actual === UNKNOWN ? 'U' : same(actual, wanted) ? 'T' : 'F'; }
function outcome(predicates) { return predicates.some(x => x.verdict === 'F') ? 'F' : predicates.some(x => x.verdict === 'U') ? 'U' : 'T'; }
function judge(row, phase, wanted) {
const p = [], add = (name, actual, expected) => p.push({name, verdict: tri(actual, expected)});
const err = (name, got, present, identity) => {
add(name + '.present', hasKnown(got.m) ? !!(got.m & 4) : UNKNOWN, present);
add(name + '.identity', hasKnown(got.m) && (got.m & 8) ? got.v : UNKNOWN, errID(identity));
};
add('returned_normally', atom(phase.n), wanted.returned_normally);
add('panic_present', atom(phase.p), wanted.panic_present);
if (bool(row.p) === true) add('observed_source_panic', true, wanted.panic_present);
need(row.f === false, 'observer_fault');
if (row.x[0] === 0) {
const u = row.u, doneK = hasKnown(u.c) || hasKnown(u.s);
add('read_count', doneK ? u.r.length : UNKNOWN, wanted.reads.length);
wanted.reads.forEach((w, i) => {
const g = u.r[i], key = 'reads.' + i;
if (!g) { add(key, UNKNOWN, true); return; }
add(key + '.call', g.k ? g.c : UNKNOWN, w.call_index);
add(key + '.normal', bool(g.o[0]), w.returned_normally);
add(key + '.panic', bool(g.o[1]), false);
add(key + '.nil', hasKnown(g.d.m) ? g.d.m === 7 : UNKNOWN, w.entries_nil);
const values = g.d.v.map(v => ({name: entryNames[v[0]], marker: v[1]}));
add(key + '.ordered_entries', hasKnown(g.d.m) ? values : UNKNOWN, w.entries ?? []);
err(key + '.error', g.e, w.error_present, w.error_identity);
});
add('logical_entry_descriptors_unchanged', bool(u.d), wanted.logical_entry_descriptors_unchanged);
add('complete_call_list', bool(u.c), wanted.complete_call_list);
add('stopped_after_source_error', bool(u.s), wanted.stopped_after_source_error);
need(typeof u.u === 'string', 'suffix_base64');
const suffix = Buffer.from(u.u, 'base64'); need(suffix.toString('base64') === u.u, 'canonical_base64');
add('unexecuted_suffix', doneK ? Array.from(suffix) : UNKNOWN, wanted.unexecuted_suffix);
} else {
const j = row.j, present = hasKnown(j.e.m) ? !!(j.e.m & 4) : UNKNOWN;
add('syntax_valid', bool(j.b[0]), wanted.syntax_valid);
add('published_view', bool(j.b[1]), wanted.published_view);
add('duplicate_error_channel_available', present === UNKNOWN ? UNKNOWN : true, wanted.duplicate_error_channel_available);
err('error', j.e, wanted.error_present, wanted.own_error_identity);
add('duplicate_present', bool(j.b[2]), wanted.duplicate_present);
add('duplicate_object_pointer', small(j.o[0], objectPaths), wanted.duplicate_object_pointer);
add('decoded_key', small(j.o[1], decodedKeys), wanted.decoded_key);
add('key_occurrence', small(j.o[2]), wanted.key_occurrence);
const prefix = j.p.map(v => v.v[3] === 1 ? {object: objectPaths[v.v[0]], key: decodedKeys[v.v[1]]} : {array: objectPaths[v.v[0]], index: v.v[4]});
const prefixEqual = same(prefix, wanted.ordered_visit_prefix);
add('ordered_visit_prefix', j.k ? prefix : UNKNOWN, wanted.ordered_visit_prefix);
const walk = bool(j.b[3]), stopped = bool(j.b[5]);
add('walk_complete', walk, wanted.walk_complete);
add('input_bytes_unchanged', bool(j.b[4]), wanted.input_bytes_unchanged);
const available = present !== UNKNOWN && typeof walk === 'boolean' && j.k && (!wanted.error_present || typeof stopped === 'boolean');
const implication = wanted.error_present ? present === true && walk === false && stopped === true && prefixEqual : present === false && walk === true && prefixEqual;
add('suffix_after_rejection_not_visited', available ? implication : UNKNOWN, wanted.suffix_after_rejection_not_visited);
}
return {d: row.d, x: row.x, verdict: outcome(p), predicates: p};
}
const checked = [], groups = [79, 83], sourceP = {unknown: 0, observed_positive: 0}, unknowns = [];
const candidates = Array.from({length: 2}, (_, r) => Array.from({length: 3}, (_, s) => ({Request: r, Display: s, T: 0, F: 0, U: 0, Verdict: ''})));
for (let n = 0; n < 39; n++) {
const line = el[n + 1], row = line.value, a = after[n], frozen = freeze.dispatches[n];
need(a && row.d === n && same(row.x, a.x) && a.r.o === line.offset && a.r.n === line.length && a.r.h === sha(line.body), 'row_identity_' + n);
const record = carrier.records[(row.x[0] === 0 ? 0 : 6) + row.x[1]];
need(record.fixture_id === frozen.fixture_id && record.request_ordinal === row.x[0] && record.fixture_index === row.x[1], 'Wanted_row_identity');
need(row.a.length === method[n].length && row.c === cbN[n], 'row_thunk_coverage');
const used = new Set(); let observedPanic = false;
row.a.forEach((api, slot) => {
const m = method[n][slot], s = m.source;
need(m.after === api.q && m.candidate === a.s && m.m === api.m && m.after < a.q && !used.has(api.q), 'method_after_binding'); used.add(api.q);
need(s && s.a === true, 'source_channel_available');
const flags = + (atom(s.e) === true) + 2 * +(atom(s.d) === true) + 4 * +s.d.k + 8 * +(atom(s.r) === true) + 16 * +s.r.k + 32;
need(api.b === flags && api.p === maskBool(s.p), 'source_mask_binding');
need(atom(s.e) === true && atom(s.d) === true && atom(s.r) === true && atom(s.p) === false && atom(s.c) === true, 'observed_source_return_or_copy_unknown');
observedPanic ||= atom(s.p) === true;
});
need(row.p !== 3 && (row.p === 11) === observedPanic, 'pre_after_source_panic');
if (bool(row.p) === UNKNOWN) sourceP.unknown++; else if (bool(row.p) === true) sourceP.observed_positive++;
if (row.u) row.u.r.forEach((g, i) => need(g.n === record.input.calls[i], 'frozen_call_count'));
const result = judge(row, a.f, record.Want), expected = saved.rows[n];
need(expected && result.d === expected.d && same(result.x, expected.x) && result.verdict === expected.verdict, 'saved_row_verdict_' + n);
const actual = [...result.predicates].sort((a,b) => a.name < b.name ? -1 : a.name > b.name ? 1 : 0);
const prior = [...expected.predicates].sort((a,b) => a.name < b.name ? -1 : a.name > b.name ? 1 : 0);
need(new Set(actual.map(x => x.name)).size === actual.length && new Set(prior.map(x => x.name)).size === prior.length && same(actual, prior), 'saved_predicates_' + n);
result.predicates.filter(p => p.verdict === 'U').forEach(p => unknowns.push({d: n, whole_group: groups[row.x[0]], name: p.name, row_verdict: result.verdict}));
checked.push(result); candidates[row.x[0]][row.x[2]][result.verdict]++;
}
candidates.flat().forEach(c => { c.Verdict = c.F ? 'F' : c.U ? 'U' : 'T'; });
need(same(candidates, saved.candidates) && saved.EvidenceSHA === sha(data.evidence) && saved.JournalSHA === sha(data.journal), 'saved_aggregate_or_hash');
need(saved.rows.length === 39 && saved.parent_requests === 2 && saved.fixture_inputs === 13 && saved.protected_rows === 0 && saved.new20to60_rows === 0 && saved.Fit === 0 && saved.original_replay === 0, 'saved_scope');
const result = {
schema: 'riido-independent-native-saved-Want-readback-v1',
source_sha256: sha(fs.readFileSync(new URL(import.meta.url))),
saved_pins: inputs.map(([id,,bytes,sha256]) => ({id, bytes, sha256})),
exact_rows: checked.length, exact_predicates: checked.reduce((n,r) => n+r.predicates.length,0),
row_verdicts: Object.fromEntries(['T','F','U'].map(v => [v,checked.filter(r => r.verdict === v).length])),
candidates, unknown_predicates: unknowns.length,
known_failure_rows_with_unknowns: checked.filter(r => r.verdict === 'F' && r.predicates.some(p => p.verdict === 'U')).map(r => r.d),
source_panic_pre_after: sourceP, observed_direct_source_returns: method.reduce((n,x) => n+x.length,0),
callback_attempts: cbN.reduce((n,x) => n+x,0), frames: jl.length,
predicate_verdicts_sha256: sha(Buffer.from(JSON.stringify(checked))),
unknown_predicates_sha256: sha(Buffer.from(JSON.stringify(unknowns))),
whole_groups: groups, whole_roles: ['development_train','development_train'],
parents: 2, inputs: 13, qualifications_added: 0, labels_added: 0, protected_rows: 0, new_cohort_rows: 0,
original_calls: 0, models: 0, Fit: 0, full_process_or_Sync_reexecuted: false,
qualified_for_next20to60_or_protected: false,
limitations: ['Nonblind saved-value implementation.', 'Init/nested/GPU/cost/utility unverified.', 'Sync/ACK/process/source provenance uses separate Root receipts.', 'Qualification/dedup/fixed roles/materializer readback still required.'],
};
const encoded = JSON.stringify(result) + '\n'; need(Buffer.byteLength(encoded) <= 16384, 'public_output_bound');
process.stdout.write(encoded);
