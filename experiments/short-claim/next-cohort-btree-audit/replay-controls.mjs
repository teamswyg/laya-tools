// SPDX-License-Identifier: Apache-2.0
// Synthetic platform controls only; never invokes original APIs or models.
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {gunzipSync} from 'node:zlib';
import {compareRanking,verify} from './verify-saved.mjs';
const packet=path.dirname(fileURLToPath(import.meta.url));
const ranking=scores=>({kind:'synthetic',count:8,fallback:'unchanged',scores,order:[0,1,2,3,4,5,6,7].sort((a,b)=>scores[b]-scores[a]||a-b)});
const baseline=ranking([1,1+Number.EPSILON,0,-1,-2,-3,-4,-5]);
const near=ranking([1,1,0,-1,-2,-3,-4,-5]);
assert.equal(compareRanking(baseline,near).order_changed,true);
assert.equal(compareRanking(baseline,baseline).order_changed,false);
assert.throws(()=>compareRanking(baseline,ranking([1,1.000001,0,-1,-2,-3,-4,-5])));
assert.throws(()=>compareRanking(baseline,{...near,order:baseline.order}));
assert.throws(()=>compareRanking(baseline,{...near,fallback:'changed'}));
assert.throws(()=>compareRanking(baseline,ranking([NaN,1,0,-1,-2,-3,-4,-5])));
assert.throws(()=>compareRanking(baseline,ranking([Infinity,1,0,-1,-2,-3,-4,-5])));
assert.throws(()=>compareRanking(baseline,ranking([0,1,-1,-2,-3,-4,-5,-6])));
const zero=ranking([0,0,0,0,0,0,0,0]);
assert.equal(compareRanking(zero,ranking([5e-13,0,0,0,0,0,0,0])).max_score_difference,5e-13);
assert.throws(()=>compareRanking(zero,ranking([2e-12,0,0,0,0,0,0,0])));
const large=ranking([1e6,0,0,0,0,0,0,0]),epsilon=1e-12+1e-12*1e6;
assert.ok(compareRanking(large,ranking([1e6+epsilon/2,0,0,0,0,0,0,0])).max_score_difference>0);
assert.throws(()=>compareRanking(large,ranking([1e6+epsilon*2,0,0,0,0,0,0,0])));
const saved=JSON.parse(gunzipSync(fs.readFileSync(path.join(packet,'OBSERVATIONS.actual.public.v1.json.gz'))));
const scratch=fs.mkdtempSync(path.join(os.tmpdir(),'riido-btree-synthetic-'));
try{
 const checkMutation=(mutate,reject)=>{
  const copy=structuredClone(saved);mutate(copy);
  const f=path.join(scratch,'synthetic.json');fs.writeFileSync(f,JSON.stringify(copy));
  if(reject)assert.throws(()=>verify(packet,f));else assert.equal(verify(packet,f).fresh_replay.saved_metrics_replaced,false);
 };
 checkMutation(()=>{},false);
 checkMutation(r=>r.rows[0].counterfactual_checks[0]++,true);
 checkMutation(r=>r.rows[0].finite_states[0]='F',true);
 checkMutation(r=>r.observations[0].Got.values[7]=1,true);
 checkMutation(r=>r.Fit_calls=1,true);
 checkMutation(r=>r.input_sha256='changed',true);
 checkMutation(r=>r.rows[0].nonlearned_rankings[0].order[1]=0,true);
}finally{fs.rmSync(scratch,{recursive:true});}
console.log(JSON.stringify({schema:'riido-btree149-synthetic-platform-controls-v1',controls:19,passed:19,absolute_and_relative_branches_checked:true,near_tie_actual_order_preserved:true,score_drift_stale_order_cost_truth_padding_metadata_input_permutation_and_Fit_rejected:true,original_calls:0,model_calls:0,Fit_calls:0}));
