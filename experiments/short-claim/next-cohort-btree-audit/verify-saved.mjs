// SPDX-License-Identifier: Apache-2.0
// Independent saved finite comparison. No original API, model, feature or Fit.
import fs from 'node:fs';
import path from 'node:path';
import {gunzipSync} from 'node:zlib';
import crypto from 'node:crypto';
import {fileURLToPath} from 'node:url';
const sha=b=>crypto.createHash('sha256').update(b).digest('hex');
const assert=(ok,c)=>{if(!ok)throw Error(c);};
const same=(a,b)=>JSON.stringify(a)===JSON.stringify(b);
export function compareRanking(saved,fresh){
 const tolerance=a=>1e-12+1e-12*Math.abs(a);
 const meta=r=>Object.fromEntries(Object.entries(r).filter(([k])=>k!=='order'&&k!=='scores'));
 assert(same(meta(saved),meta(fresh)),'fresh_ranking_metadata');
 for(const r of [saved,fresh]){
  assert(r.order.length===8&&r.scores.length===8&&same([...r.order].sort((a,b)=>a-b),[0,1,2,3,4,5,6,7])&&r.scores.every(Number.isFinite),'fresh_ranking_shape');
  for(let j=1;j<8;j++){const a=r.order[j-1],b=r.order[j];assert(r.scores[a]>r.scores[b]||(r.scores[a]===r.scores[b]&&a<b),'fresh_own_score_order');}
 }
 let maxScoreDifference=0,wellSeparatedPairs=0;
 const positions=Array(8);fresh.order.forEach((v,i)=>positions[v]=i);
 for(let i=0;i<8;i++){
  const difference=Math.abs(saved.scores[i]-fresh.scores[i]);
  assert(difference<=tolerance(saved.scores[i]),'fresh_score_tolerance');
  maxScoreDifference=Math.max(maxScoreDifference,difference);
  for(let j=i+1;j<8;j++){
   const gap=saved.scores[i]-saved.scores[j];
   if(Math.abs(gap)>tolerance(saved.scores[i])+tolerance(saved.scores[j])){
    assert((gap>0)===(positions[i]<positions[j]),'fresh_well_separated_order');wellSeparatedPairs++;
   }
  }
 }
 return{max_score_difference:maxScoreDifference,well_separated_pairs:wellSeparatedPairs,order_changed:!same(saved.order,fresh.order)};
}
export function verify(packet,freshPath){
 const inputBytes=fs.readFileSync(path.join(packet,'INPUTS.public.v1.json'));
 assert(sha(inputBytes)==='1612003555b5a0293092d6667fa1afc1cd7620bc54ef048ea3fc7834cd776ecb','input_pin');
 const input=JSON.parse(inputBytes);
 const raw=gunzipSync(fs.readFileSync(path.join(packet,'OBSERVATIONS.actual.public.v1.json.gz')),{maxOutputLength:131073});
 const saved=JSON.parse(raw);
 const check=r=>{
  assert(r.input_sha256===saved.input_sha256&&r.selected_generic_source_sha256===saved.selected_generic_source_sha256&&r.schema===saved.schema&&r.scope===saved.scope,'source_and_input_binding');
  assert(r.development_parents===12&&r.connected_families===1&&r.fixtures===34&&r.original_candidate_calls===272&&r.candidate_returns===272&&r.model_calls===0&&r.Fit_calls===0&&!r.training_ready&&!r.qualification_performed,'ledger');
  assert(r.observations.length===272&&r.rows.length===12,'shape');
  let ordinal=0;const checks=[0,0,0,0],work=[0,0,0,0],top1=[0,0,0,0],top3=[0,0,0,0];
  let oracleChecks=0,oracleNS=0,positive=0,negative=0,answerable=0,baselineTogetherNS=0;
  for(let p=0;p<12;p++){
   const row=r.rows[p],q=input.requests[p],states=Array(8).fill('T'),ns=Array(8).fill(0);
   assert(row.nonlearned_rankings.length===4&&row.counterfactual_checks.length===4&&row.counterfactual_verification_work_ns.length===4,'row_shape');
   assert(row.id===q.id,'parent_identity');
   for(let f=0;f<q.fixtures.length;f++)for(let c=0;c<8;c++){
    const v=r.observations[ordinal++],fixture=q.fixtures[f],g=v.Got;
    assert(v.parent===p&&v.fixture===f&&v.candidate===c&&Number.isSafeInteger(v.elapsed_ns)&&v.elapsed_ns>0,'ordinal_or_time');
    assert(g.returned===true&&g.panicked===false&&g.count>=0&&g.count<=7&&g.callbacks===g.count&&g.values.length===8,'complete_Got');
    assert(g.values.slice(g.count).every(x=>x===0),'unused_Got_slots');
    // A second, literal direct-API interpretation checks every candidate Got,
    // rather than accepting only the observer's generated labels.
    const functions=[k=>k>=fixture.lower&&k<fixture.upper,k=>k<fixture.upper,k=>k>=fixture.lower,()=>true,k=>k<=fixture.upper&&k>fixture.lower,k=>k<=fixture.upper,k=>k>fixture.lower,()=>true];
    let expected=fixture.keys.filter(functions[c]);if(c>=4)expected.reverse();
    const stopped=fixture.stop_after!==0&&expected.length>=fixture.stop_after;
    if(fixture.stop_after!==0)expected=expected.slice(0,fixture.stop_after);
    assert(same(g.values.slice(0,g.count),expected)&&g.callbacks===expected.length&&g.last_callback_returned_false===stopped,'independent_direct_API_Got');
    const state=same(expected,fixture.Want)&&g.callbacks===fixture.Want_callbacks&&stopped===fixture.Want_stopped?'T':'F';
    if(state==='F')states[c]='F';ns[c]+=v.elapsed_ns;
   }
   assert(same(row.finite_states,states)&&same(row.all_fixture_verification_ns,ns),'finite_state_or_work');
   const n=states.filter(s=>s==='T').length;positive+=n;negative+=8-n;
   const klass=n===0?'known_none':n===8?'known_all':n===1?'known_one':'known_many';
   assert(row.finite_class===klass,'class');if(n>0)answerable++;
   const oc=n===0?8:1,on=n===0?ns.reduce((a,b)=>a+b,0):Math.min(...ns.filter((_,i)=>states[i]==='T'));
   assert(row.oracle_checks===oc&&row.optimistic_oracle_verification_ns===on,'oracle');oracleChecks+=oc;oracleNS+=on;
   assert(Number.isSafeInteger(row.prepare_and_four_controls_ns)&&row.prepare_and_four_controls_ns>0,'preparation_time');baselineTogetherNS+=row.prepare_and_four_controls_ns;
   for(let m=0;m<4;m++){
    const ranking=row.nonlearned_rankings[m],kinds=['fixed_order','bm25','lexical_ordered','narrow_rule'];
    assert(ranking.kind===kinds[m]&&ranking.count===8&&ranking.order.length===8&&ranking.scores.length===8&&same([...ranking.order].sort((a,b)=>a-b),[0,1,2,3,4,5,6,7]),'ranking_permutation');
    assert(ranking.scores.every(Number.isFinite),'score_finite');
    for(let j=1;j<8;j++){const a=ranking.order[j-1],b=ranking.order[j];assert(ranking.scores[a]>ranking.scores[b]||(ranking.scores[a]===ranking.scores[b]&&a<b),'stable_saved_score_order');}
    let count=0,cost=0;for(const i of ranking.order){count++;cost+=ns[i];if(states[i]==='T')break;}
    assert(row.counterfactual_checks[m]===count&&row.counterfactual_verification_work_ns[m]===cost,'method_cost');checks[m]+=count;work[m]+=cost;
    if(n>0&&states[ranking.order[0]]==='T')top1[m]++;
    if(n>0&&ranking.order.slice(0,3).some(i=>states[i]==='T'))top3[m]++;
   }
  }
  return{parents:12,source_families:1,candidates:96,positive,negative,unknown:0,answerable,checks,work_ns:work,top1,top3,top_denominator:answerable,oracle_checks:oracleChecks,oracle_ns:oracleNS,prepare_and_four_controls_ns:baselineTogetherNS,scope:'Saved finite interpretation and measured single-sample counterfactual work; no actual speed or LLM savings.'};
 };
 const summary=check(saved);
 if(freshPath){
  const fresh=JSON.parse(fs.readFileSync(freshPath)),freshSummary=check(fresh),differences=[];
  let maxScoreDifference=0,wellSeparatedPairs=0;
  for(let p=0;p<12;p++){
   assert(same(fresh.rows[p].finite_states,saved.rows[p].finite_states),'fresh_semantics');
   for(let m=0;m<4;m++){
    const d=compareRanking(saved.rows[p].nonlearned_rankings[m],fresh.rows[p].nonlearned_rankings[m]);
    maxScoreDifference=Math.max(maxScoreDifference,d.max_score_difference);wellSeparatedPairs+=d.well_separated_pairs;
    if(d.order_changed)differences.push({parent:p,kind:fresh.rows[p].nonlearned_rankings[m].kind,saved_order:saved.rows[p].nonlearned_rankings[m].order,fresh_order:fresh.rows[p].nonlearned_rankings[m].order,saved_checks:saved.rows[p].counterfactual_checks[m],fresh_checks:fresh.rows[p].counterfactual_checks[m]});
   }
  }
  for(let i=0;i<272;i++)assert(same(fresh.observations[i].Got,saved.observations[i].Got),'fresh_Got');
  summary.fresh_replay={summary:freshSummary,absolute_tolerance:1e-12,relative_tolerance:1e-12,max_score_difference:maxScoreDifference,well_separated_pairs:wellSeparatedPairs,near_tie_order_differences:differences,saved_metrics_replaced:false};
 }
 return summary;
}
if(process.argv[1]&&path.resolve(process.argv[1])===fileURLToPath(import.meta.url))try{console.log(JSON.stringify(verify(process.argv[2],process.argv[3])));}catch(e){console.error('btree_saved_verification_failed: '+e.message);process.exitCode=1;}
