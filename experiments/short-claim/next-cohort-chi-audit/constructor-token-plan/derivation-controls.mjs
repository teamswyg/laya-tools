// SPDX-License-Identifier: Apache-2.0
// Source-only controls. Reads public Go/fixture text, never model or Go APIs.
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import assert from 'node:assert/strict';
import {buildSources,readInputs,functionText,outputPath,pins,schema,moduleText} from './derive-source.mjs';
assert.equal(process.argv.length,2,'no arguments: public source paths come from this file');
const here=path.dirname(fileURLToPath(import.meta.url));
const inputs=readInputs(path.resolve(here,'../../../..'),path.join(here,'reference/prepared-baseline.go.txt'));
const generated=buildSources(inputs),again=buildSources(inputs);
assert.deepEqual(Object.keys(generated),['prepared.go','token_probe_test.go','token_saved_test.go','INPUT.public.v3.json','go.mod','go.sum']);
for(const name of Object.keys(generated)) assert(generated[name].equals(again[name]));
assert(generated['prepared.go'].equals(inputs.candidate));
assert(generated['INPUT.public.v3.json'].equals(inputs.fixture));
assert.equal(generated['go.mod'].toString(),moduleText);
assert(generated['go.sum'].equals(inputs.goSum));
const baseline=functionText(inputs.baseline.toString(),'func Prepare(');
const probe=generated['token_probe_test.go'].toString(),saved=generated['token_saved_test.go'].toString();
assert.equal(functionText(probe,'func tokenPlanBaselineScratch(').replace('func tokenPlanBaselineScratch(','func Prepare('),baseline);
assert(probe.includes('Schema: "'+schema+'"'));assert(saved.includes('r.Schema != "'+schema+'"'));
assert(probe.includes('BaselineScratchAnchor')&&probe.includes('TokenPlanAnchor'));
assert(probe.includes('method != "baseline_scratch" && method != "token_plan"'));
const interval=functionText(probe,'func tokenPlanInterval(');
assert(interval.indexOf('prepareAPI, rankAPI :=')<interval.indexOf('x.Cost = measure('));
assert(interval.includes('tokenPlanCall(prepareAPI, prepare)')&&interval.includes('tokenPlanCall(rankAPI, func()'));
assert(!interval.includes('tokenPlanCall(method+'));
for(const text of [probe,saved]) {
 assert(!/"(?:legacy|scratch)(?:_prepare|_rank)?"/.test(text));
 assert(!/"constructor-(?:input|model|output|saved)"/.test(text));
}
assert(saved.includes('b.Cost.TotalAllocBytes > a.Cost.TotalAllocBytes'));
assert(saved.includes('b.Cost.Mallocs > a.Cost.Mallocs'));
assert(saved.includes('a.Cost.Mallocs-b.Cost.Mallocs < 144'));
assert(saved.includes('medians[group] > 1.0'));
assert(!saved.includes('> .8')&&!saved.includes('> 1.25'));
assert(saved.includes('TestTokenPlanSavedMallocAndPairControls'));
assert(saved.includes('"INPUT.public.v3.json"')&&!saved.includes('../../experiments/'));
// A valid-looking edited parent must fail before any source is derived/written.
for(const name of Object.keys(pins)) {
 const changed=Buffer.from(inputs[name]);changed[changed.length-1]^=1;
 assert.throws(()=>buildSources({...inputs,[name]:changed}),new RegExp('^Error: source_pin_'+name+'$'));
 assert.throws(()=>buildSources({...inputs,[name]:inputs[name].subarray(1)}),new RegExp('^Error: source_pin_'+name+'$'));
}
assert.throws(()=>{pins.baseline[0]=0;},TypeError);
assert.throws(()=>outputPath('/owned/temp','../prepared.go'),/output_name/);
assert.throws(()=>outputPath('/owned/temp','repo'),/output_name/);
assert.equal(outputPath('/owned/temp','prepared.go'),path.resolve('/owned/temp/prepared.go'));
console.log('Pinned derivation/source binding and protocol mapping controls passed; no output files, Go or model calls.');
