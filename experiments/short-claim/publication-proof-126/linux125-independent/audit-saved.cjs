// SPDX-License-Identifier: Apache-2.0
// Alternative saved-JSON bookkeeping only; no Reader, scorer, model or original APIs.
'use strict';
const fs=require('node:fs'),crypto=require('node:crypto'),assert=require('node:assert/strict');
assert.equal(process.argv.length,6);
const paths=process.argv.slice(2,5),out=process.argv[5];
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const pins=[
[242297,'9a7b15eec028f5014791bec9a9aed7de0f60f1bc6eee404073d53e2d5a47670e'],
[242305,'4d0467a15d974605497142db3b00ac4f9c40d9fa377a478de240f537a72d9f97'],
[109300,'dbd9dc093ec4176275da272af06898e08d67ef14e09df980f4b7345ae5c07fd8']];
const [mac,linux,diff]=paths.map((p,i)=>{assert(fs.lstatSync(p).isFile());const b=fs.readFileSync(p);assert.equal(b.length,pins[i][0]);assert.equal(hash(b),pins[i][1]);return JSON.parse(b);});
const kinds=['fixed_order','bm25','lexical_ordered','narrow_rule'],pairs=[[0,1],[0,2],[1,2]];
const eq=(x,y)=>{try{assert.deepStrictEqual(x,y);return true;}catch{return false;}};
const epsilon=x=>1e-12*Math.max(1,Math.abs(x));
function perms(n){const result=[];function extend(p){if(p.length===n){result.push([...p,...Array(3-n).fill(0)]);return;}for(let j=0;j<n;j++)if(!p.includes(j))extend([...p,j]);}extend([]);return result;}
const emptyAggregate=kind=>({kind,Cases:0,CandidateVisitsWithoutOrdering:0,SimulatedVisitsToFirstPositive:0,Top1KnownPositive:0,Top3KnownPositive:0,FallbackCases:0,simulated_visit_reduction:0});
function derive(report){
assert.equal(report.Schema,'riido-development33-nonlearned-bias-audit-v1');assert.equal(report.State,'complete');
assert.equal(report.SimulatedOnly,true);assert.equal(report.ActualVerificationWorkObserved,false);assert.equal(report.ExcludedUnknownCandidatesNotRead,true);assert.equal(report.ModelInferenceCalls,0);assert.equal(report.FitCalls,0);assert.match(report.DataSHA256,/^[0-9a-f]{64}$/);
assert.equal(report.Parents.length,33);assert.equal(report.Cases.length,219);
const Primary=kinds.map(emptyAggregate),Permuted=kinds.map(emptyAggregate),positive=[0,0,0],parentIDs=[];
let at=0,selected=0,permutations=0,rankings=0,activeScores=0,activePairs=0;
for(let pi=0;pi<33;pi++){
 const p=report.Parents[pi],n=p.count;assert(n===2||n===3);
 assert(typeof p.stable_id==='string'&&p.stable_id.length>0&&!parentIDs.includes(p.stable_id));parentIDs.push(p.stable_id);
 assert.equal(p.candidate_ids.length,3);assert.equal(p.frozen_known_labels.length,3);assert(p.style.RequestBytes>0&&p.style.RequestNormalizedWords>0);
 for(let j=0;j<3;j++)if(j<n){assert(typeof p.candidate_ids[j]==='string'&&p.candidate_ids[j].length>0&&!p.candidate_ids.slice(0,j).includes(p.candidate_ids[j]));assert(typeof p.frozen_known_labels[j]==='boolean');assert(p.style.CandidateBytes[j]>0&&p.style.CandidateNormalizedWords[j]>0);positive[j]+=Number(p.frozen_known_labels[j]);}else{assert.equal(p.candidate_ids[j],'');assert.equal(p.frozen_known_labels[j],false);assert.equal(p.style.CandidateBytes[j],0);assert.equal(p.style.CandidateNormalizedWords[j],0);}
 assert.equal(p.frozen_known_labels.filter(Boolean).length,1);selected+=n;
 const all=perms(n);permutations+=all.length;
 for(const [primary,mapping] of [[true,all[0]],...all.map(m=>[false,m])]){
  const c=report.Cases[at++];assert.equal(c.parent_index,pi);assert.equal(c.primary,primary);assert.equal(c.count,n);assert.deepStrictEqual(c.display_to_selected,mapping);assert.equal(c.ranks.length,4);
  for(let k=0;k<4;k++){
   const r=c.ranks[k],s=r.scores_by_display_index;assert.equal(r.kind,kinds[k]);assert(typeof r.fallback_reason==='string'&&r.fallback_reason.length<=128);if(k!==3)assert.equal(r.fallback_reason,'');
   assert.equal(s.length,3);assert.equal(r.DisplayOrder.length,3);assert.equal(r.SelectedOrder.length,3);assert.equal(r.tied_pairs_01_02_12.length,3);
   for(let j=0;j<3;j++){assert(Number.isFinite(s[j]));if(j>=n){assert.equal(s[j],0);assert.equal(r.DisplayOrder[j],0);assert.equal(r.SelectedOrder[j],0);}}
   if(k===0)assert.deepStrictEqual(s,[0,0,0]);
   const predicted=Array.from({length:n},(_,j)=>j).sort((i,j)=>s[i]>s[j]?-1:s[i]<s[j]?1:i-j);
   assert.deepStrictEqual(r.DisplayOrder.slice(0,n),predicted);
   const selectedOrder=predicted.map(j=>mapping[j]);assert.deepStrictEqual(r.SelectedOrder.slice(0,n),selectedOrder);
   assert.deepStrictEqual(r.tied_pairs_01_02_12,pairs.map(([i,j])=>j<n&&s[i]===s[j]));
   const visit=selectedOrder.findIndex(j=>p.frozen_known_labels[j])+1;assert(visit>0);assert.equal(r.simulated_first_positive_visit,visit);
   const a=(primary?Primary:Permuted)[k];a.Cases++;a.CandidateVisitsWithoutOrdering+=n;a.SimulatedVisitsToFirstPositive+=visit;a.Top1KnownPositive+=Number(visit===1);a.Top3KnownPositive+=Number(visit<=3);a.FallbackCases+=Number(r.fallback_reason!=='');rankings++;activeScores+=n;activePairs+=n===3?3:1;
  }
 }
}
assert.equal(at,219);assert.equal(selected,96);assert.equal(permutations,186);assert.equal(rankings,876);
for(const a of [...Primary,...Permuted])a.simulated_visit_reduction=(a.CandidateVisitsWithoutOrdering-a.SimulatedVisitsToFirstPositive)/a.CandidateVisitsWithoutOrdering;
const counts={ReaderAttempts:33,ReaderReturned:33,ReaderSucceeded:33,BaselinesAttempts:219,BaselinesReturned:219,PrimaryCases:33,PermutationCases:186,RankingRecords:876};
assert.deepStrictEqual(report.Counts,counts);assert.deepStrictEqual(report.Primary,Primary);assert.deepStrictEqual(report.Permuted,Permuted);assert.deepStrictEqual(report.PrimaryPositivePositions,positive);assert.equal(report.PositionIndexOneHits,positive[1]);
return {counts,Primary,Permuted,primary_positive_positions:positive,position_index_one_hits:positive[1],activeScores,activePairs};
}
const f=derive(mac),a=derive(linux);
const metadata=r=>{const {Cases,Primary,Permuted,...fixed}=r;return fixed;};assert.deepStrictEqual(metadata(mac),metadata(linux));
const changes=[],scoreOnly=[],behavior=[],fieldCounts=[0,0,0,0,0,0];
let maxAbsolute=0,maxFraction=0,robustPairs=0,overlapPairs=0,scoreChecks=0;
for(let ci=0;ci<219;ci++){
 const fc=mac.Cases[ci],ac=linux.Cases[ci],{ranks:fr,...fm}=fc,{ranks:ar,...am}=ac;assert.deepStrictEqual(am,fm);
 for(let k=0;k<4;k++){
  const x=fr[k],y=ar[k],n=ac.count;assert.equal(x.kind,y.kind);assert.equal(x.fallback_reason,y.fallback_reason);
  const overlaps=pairs.map(([i,j])=>j<n&&Math.abs(x.scores_by_display_index[i]-x.scores_by_display_index[j])<=epsilon(x.scores_by_display_index[i])+epsilon(x.scores_by_display_index[j]));
  for(let j=0;j<3;j++){const delta=Math.abs(x.scores_by_display_index[j]-y.scores_by_display_index[j]),e=epsilon(x.scores_by_display_index[j]);assert(delta<=e);maxAbsolute=Math.max(maxAbsolute,delta);maxFraction=Math.max(maxFraction,delta/e);scoreChecks++;}
  pairs.forEach(([i,j],p)=>{if(j>=n)return;if(overlaps[p]){overlapPairs++;return;}robustPairs++;assert.equal(Math.sign(x.DisplayOrder.indexOf(i)-x.DisplayOrder.indexOf(j)),Math.sign(y.DisplayOrder.indexOf(i)-y.DisplayOrder.indexOf(j)));});
  const flags=[!eq(x.scores_by_display_index,y.scores_by_display_index),!eq(x.tied_pairs_01_02_12,y.tied_pairs_01_02_12),!eq(x.DisplayOrder,y.DisplayOrder),!eq(x.SelectedOrder,y.SelectedOrder),x.simulated_first_positive_visit!==y.simulated_first_positive_visit,x.kind!==y.kind||x.fallback_reason!==y.fallback_reason];flags.forEach((v,i)=>fieldCounts[i]+=Number(v));
  if(!eq(x,y)){
   changes.push({case:ci,rank:k,FrozenScores:x.scores_by_display_index,ActualScores:y.scores_by_display_index,FrozenDisplay:x.DisplayOrder,ActualDisplay:y.DisplayOrder,FrozenSelected:x.SelectedOrder,ActualSelected:y.SelectedOrder,FrozenTies:x.tied_pairs_01_02_12,ActualTies:y.tied_pairs_01_02_12,FrozenVisit:x.simulated_first_positive_visit,ActualVisit:y.simulated_first_positive_visit,overlap_pairs_01_02_12:overlaps});
   if(flags[0]&&!flags.slice(1).some(Boolean))scoreOnly.push([ci,k]);else behavior.push([ci,k,ac.parent_index,ac.primary,...flags.map(Number)]);
  }
 }
}
const tri=(frozen,actual,recomputed)=>frozen.map((x,k)=>({Frozen:x,Actual:actual[k],RecalculatedActual:recomputed[k]}));
const expectedDiff={schema:'riido-bias-portable-frozen-difference-v1',state:'complete_portable_reproduction',validation_error:'',score_rule:'abs(actual-frozen) <= 1e-12*max(1,abs(frozen)); robust pair gap > bound_i+bound_j; actual ties use ==',frozen_report_sha256:pins[0][1],actual_report_sha256:pins[1][1],FrozenCases:219,ActualCases:219,rank_changes:changes,Primary:tri(mac.Primary,linux.Primary,a.Primary),Permuted:tri(mac.Permuted,linux.Permuted,a.Permuted),actual_recalculation:{counts:a.counts,Primary:a.Primary,Permuted:a.Permuted,primary_positive_positions:a.primary_positive_positions,position_index_one_hits:a.position_index_one_hits},actual_recalculation_complete:true};
assert.deepStrictEqual(diff,expectedDiff);
const subtract=(frozen,actual)=>actual.map((x,k)=>Object.fromEntries(Object.entries(x).map(([name,value])=>[name,typeof value==='number'?value-frozen[k][name]:value])));
const report={schema:'riido-CI125-Linux-independent-saved-audit-v1',state:'complete_saved_only_bookkeeping_audit',input_pins:pins.map(([bytes,sha256],i)=>({id:['frozen-Mac-report','saved-Linux-report','saved-Linux-difference'][i],bytes,sha256})),checks:{parents:33,selected:96,cases:219,rankings:876,reported_counts_consistent:true,active_score_slots:a.activeScores,all_score_slots_compared:scoreChecks,active_pairs:a.activePairs,robust_pairs_checked:robustPairs,overlap_pairs_recorded:overlapPairs,strict_order_exact_ties_mapping_visits_aggregates:true,Mac_score_bounds:true,entire_sidecar_semantically_exact:true,max_absolute_score_delta:maxAbsolute,max_delta_over_existing_bound:maxFraction},changes:{total:changes.length,field_names:['score','ties','display_order','selected_order','first_positive_visit','rank_metadata'],field_counts:fieldCounts,score_only_count:scoreOnly.length,score_only_case_rank_positions:scoreOnly,behavior_count:behavior.length,behavior_tuple_fields:['case','rank','parent','primary','score','ties','display_order','selected_order','visit','rank_metadata'],behavior_tuples:behavior},Primary:{Actual:linux.Primary,Difference:subtract(mac.Primary,linux.Primary)},Permuted:{Actual:linux.Permuted,Difference:subtract(mac.Permuted,linux.Permuted)},relationships:{nonblind:true,reviewer_prior_bias_source_preparation_and_cross_platform_contract_author:true,current_portable_checker_nonauthor:true,independent_scope:'Alternative Node saved-JSON bookkeeping, not independent supervision or sample count.'},limits:['Saved records validate reported counts, not independent observation of Reader/Baselines callbacks. Root artifact receipt supplies GitHub provenance.','No scoring arithmetic/Reader/original API/model/Fit/network is executed.','186 permutations repeat33 exposed development parents;876 rankings are not independentN.','Top3 is trivial in this two-or-three candidate one-positive pool; simulated visits are not actual work savings or accuracy.','No labels/roles/qualification/protected evaluation/current GJSON outcome interpretation.'],execution:{saved_json_audit:1,Go:0,Reader:0,Baselines:0,original_API:0,native:0,model:0,Fit:0,network:0}};
const raw=Buffer.from(JSON.stringify(report)+'\n');assert(raw.length<=8500);
const fd=fs.openSync(out,'wx',0o600);try{fs.writeFileSync(fd,raw);fs.fsyncSync(fd);}finally{fs.closeSync(fd);}const dfd=fs.openSync(require('node:path').dirname(out),'r');try{fs.fsyncSync(dfd);}finally{fs.closeSync(dfd);}assert.equal(hash(fs.readFileSync(out)),hash(raw));
console.log(JSON.stringify({bytes:raw.length,sha256:hash(raw),changes:report.changes.field_counts,score_only:scoreOnly.length,behavior:behavior.length,primary_differences:report.Primary.Difference,permuted_differences:report.Permuted.Difference}));
