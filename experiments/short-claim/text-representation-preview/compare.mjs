// SPDX-License-Identifier: Apache-2.0
// Saved, exposed development diagnostic. No original source behavior replay,
// model, coefficients, Fit, protected evaluator, or external inference API.
import fs from 'node:fs';
import path from 'node:path';
import {createHash} from 'node:crypto';
import {gzipSync,gunzipSync} from 'node:zlib';
import {spawnSync} from 'node:child_process';

const sha=b=>createHash('sha256').update(b).digest('hex');
const assert=(x,s)=>{if(!x)throw Error(s);};
const enc=v=>Buffer.from(JSON.stringify(v)+'\n');
const repo=path.resolve(process.argv[2]||'.');
const binary=path.resolve(process.argv[3]||'');
const packet=path.join(repo,'experiments/short-claim/text-representation-preview');
const frozen=JSON.parse(fs.readFileSync(path.join(packet,'FREEZE.public.v1.json')));
for(const p of frozen.pins){const b=fs.readFileSync(path.join(repo,p.path));assert(b.length===p.bytes&&sha(b)===p.sha256,'frozen byte pin: '+p.path);}
const input=JSON.parse(fs.readFileSync(path.join(repo,frozen.input.path)));
const observations=JSON.parse(gunzipSync(fs.readFileSync(path.join(repo,frozen.observations.path))));
assert(input.requests.length===12&&input.candidates.length===8&&observations.rows.length===12,'fixed exposed dimensions');
const capture=[];
for(const r of input.requests){
 const result=spawnSync(binary,[],{input:enc({request:r.text,candidates:input.candidates.map(c=>c.text)}),timeout:5000,maxBuffer:16384,encoding:'buffer'});
 capture.push({request_id:r.id,status:result.status,signal:result.signal,error_code:result.error?.code||null,stdout_base64:result.stdout?.toString('base64')||'',stderr_base64:result.stderr?.toString('base64')||''});
}
// Optional owned first-capture destination; always store ALL bounded responses
// and statuses before parsing or accepting them. The caller reserves its cap.
if(process.env.RIIDOLAYA_REPRESENTATION_CAPTURE){
 const p=path.resolve(process.env.RIIDOLAYA_REPRESENTATION_CAPTURE);
 const packed=gzipSync(enc({schema:'riido-representation-first-capture-v1',binary_sha256:sha(fs.readFileSync(binary)),capture}),{level:9});
 assert(packed.length<=32768,'reserved first-capture limit');
 fs.writeFileSync(p,packed,{flag:'wx',mode:0o600});
 const f=fs.openSync(p,'r');fs.fsyncSync(f);fs.closeSync(f);
 const d=fs.openSync(path.dirname(p),'r');fs.fsyncSync(d);fs.closeSync(d);
}
const rows=[];
const totals={fixed_checks:0,agreement_checks:0,fixed_saved_verifier_ns:0,agreement_saved_verifier_ns:0,answerable_with_negatives:0,answerable_with_negatives_top1:0,answerable_with_negatives_top3:0,all_positive:0,no_answer:0};
for(let i=0;i<capture.length;i++){
 const c=capture[i],r=input.requests[i],old=observations.rows[i];
 assert(c.status===0&&!c.signal&&!c.error_code&&!c.stderr_base64,'bounded preview failed');
 const v=JSON.parse(Buffer.from(c.stdout_base64,'base64'));
 assert(v.schema==='riido-hint-representation-preview-v1'&&v.preview_only&&v.model_calls===0&&v.Fit_calls===0&&v.candidates.length===8,'preview shape or model-call drift');
 assert(old.id===r.id&&old.finite_states.length===8&&old.finite_states.every(x=>x==='T'||x==='F'),'saved finite scope drift');
 assert(old.all_fixture_verification_ns.length===8&&old.all_fixture_verification_ns.every(Number.isSafeInteger),'saved cost shape');
 const order=v.agreement_order;
 assert(order.length===8&&[...order].sort((a,b)=>a-b).every((x,j)=>x===j),'permutation');
 const fixed=[0,1,2,3,4,5,6,7];
 const checks=ord=>{const n=ord.findIndex(x=>old.finite_states[x]==='T');return n<0?8:n+1;};
 const work=ord=>ord.slice(0,checks(ord)).reduce((n,x)=>n+old.all_fixture_verification_ns[x],0);
 const positive=old.finite_states.filter(x=>x==='T').length;
 const stratum=positive===0?'no_answer':positive===8?'all_positive':'answerable_with_negatives';
 const n=checks(order),f=checks(fixed),ns=work(order),fs=work(fixed);
 totals.fixed_checks+=f;totals.agreement_checks+=n;
 totals.fixed_saved_verifier_ns+=fs;totals.agreement_saved_verifier_ns+=ns;
 totals[stratum]++;
 if(stratum==='answerable_with_negatives'){
  totals.answerable_with_negatives_top1+=Number(n===1);
  totals.answerable_with_negatives_top3+=Number(n<=3);
 }
 rows.push({id:r.id,stratum,finite_states:old.finite_states,preview:v,fixed_checks:f,agreement_checks:n,fixed_saved_verifier_ns:fs,agreement_saved_verifier_ns:ns});
}
const result={schema:'riido-exposed-btree-representation-diagnostic-v1',input_sha256:sha(fs.readFileSync(path.join(repo,frozen.input.path))),observations_gzip_sha256:sha(fs.readFileSync(path.join(repo,frozen.observations.path))),relation_schema:'riido-interval-relations8-v1',parents:12,connected_families:1,whole_group:84,role:'development_train',totals,rows,model_calls:0,Fit_calls:0,original_API_calls:0,new_weights:0,changed_labels_roles_masks:0,independent_evaluation:false,actual_speedup_proved:false,LLM_savings_proved:false};
if(process.argv[4]){
 const saved=JSON.parse(gunzipSync(fs.readFileSync(process.argv[4])));
 assert(JSON.stringify(saved)===JSON.stringify(result),'exact saved diagnostic mismatch');
 process.stdout.write(JSON.stringify({saved_exact:true,parents:12,totals,model_calls:0,Fit_calls:0})+'\n');
}else process.stdout.write(JSON.stringify(result)+'\n');
