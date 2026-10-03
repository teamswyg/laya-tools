// SPDX-License-Identifier: Apache-2.0
// Read-only saved archive audit. No disk restores/deletes or scientific execution.
'use strict';
const fs=require('node:fs'),z=require('node:zlib'),crypto=require('node:crypto'),path=require('node:path'),a=require('node:assert/strict');
const [R,out]=process.argv.slice(2);a.equal(process.argv.length,4);a.equal(fs.realpathSync(R),R);
const D=path.join(R,'hf33-completed-directory-archives-root-v1'),Q=path.join(R,'resource-HF33-directory-archive-Root-QA-v1');
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
function read(p,cap=262144){const s=fs.lstatSync(p);a(s.isFile()&&!s.isSymbolicLink());a.equal(fs.realpathSync(p),p);a(s.size<=cap);const b=fs.readFileSync(p),t=fs.lstatSync(p);a.equal(b.length,s.size);a.equal(t.ino,s.ino);a.equal(t.mtimeMs,s.mtimeMs);a.equal(t.size,s.size);return b;}
function pin(p,cap){const b=read(p,cap);return {path:p,bytes:b.length,sha256:hash(b)};}
function check(p,cap=262144){a(Number.isSafeInteger(p.bytes)&&p.bytes>=0&&/^[a-f0-9]{64}$/.test(p.sha256));const b=read(p.path,cap);a.equal(b.length,p.bytes);a.equal(hash(b),p.sha256);return b;}
const json=p=>JSON.parse(read(p));
function absent(p){try{fs.lstatSync(p);a.fail('absent path exists');}catch(e){a.equal(e.code,'ENOENT');}}
const selectionPath=path.join(Q,'ROOT-SELECTION.private.v1.json'),rawSelection=read(selectionPath);
a.equal(rawSelection.length,170774);a.equal(hash(rawSelection),'5a82ecde42a0e6dd1c14350b582ed429085149204eda51abb2557f5582e9c130');
const s=JSON.parse(rawSelection),reservation=json(path.join(D,'RESERVATION.json')),final=json(path.join(D,'FINAL.json'));
a.equal(s.targets.length,2);a.equal(s.canonical_science_excluded,true);a.equal(s.models_tokens_raw_observations_excluded,true);
const config=JSON.parse(check(reservation.config)),ad=JSON.parse(check(reservation.admission));
a.deepEqual(reservation.selection,pin(selectionPath));a.deepEqual(config.selection,reservation.selection);a.deepEqual(config.admission,reservation.admission);a.deepEqual(ad.selection,reservation.selection);a.deepEqual(config.prior_account,ad.prior_account);check(config.prior_account);check(ad.census);ad.helper_sources.forEach(p=>check(p));
a.equal(config.root,R);a.equal(config.output_base,path.basename(D));a.equal(ad.cap_bytes,536870912);a.deepEqual(ad.archive_caps,[327680,327680]);a.equal(ad.restore_payload_cap_bytes,1048576);a.equal(ad.tar_inflate_cap_bytes,2097152);a.equal(ad.new_control_cap_bytes,131072);a.equal(ad.writers_quiescent,true);a.equal(ad.fixed_public_copies_only,true);
for(const k of ['cap_bytes','archive_caps','restore_payload_cap_bytes','new_control_cap_bytes'])a.deepEqual(reservation[k],ad[k]);
const table=Array.from({length:256},(_,i)=>{let c=i;for(let j=0;j<8;j++)c=c&1?0xedb88320^(c>>>1):c>>>1;return c>>>0;});
function crc(b){let c=0xffffffff;for(const x of b)c=table[(c^x)&255]^(c>>>8);return(c^0xffffffff)>>>0;}
a.equal(crc(Buffer.from('123456789')),0xcbf43926);
function oct(b){const t=b.toString('ascii').replace(/\0/g,'').trim();a(/^[0-7]*$/.test(t));const v=t?parseInt(t,8):0;a(Number.isSafeInteger(v)&&v>=0);return v;}
function text(b){const i=b.indexOf(0),raw=i<0?b:b.subarray(0,i);a([...raw].every(x=>x>=32&&x<127));return raw.toString('ascii');}
function rel(p){a(p&&p.length<=256&&!p.includes('\\')&&!p.includes('\0')&&!p.startsWith('/'));a.equal(path.posix.normalize(p),p);a(p!=='.'&&p!=='..'&&!p.startsWith('../'));}
function pax(b){let at=0,v={};while(at<b.length){const sp=b.indexOf(32,at),ntext=b.subarray(at,sp).toString('ascii');a(sp>at&&/^[0-9]+$/.test(ntext));const end=at+Number(ntext);a(end<=b.length&&end>sp+2&&b[end-1]===10);const t=b.subarray(sp+1,end-1).toString('utf8'),eq=t.indexOf('=');a(eq>0);const k=t.slice(0,eq);a.equal(k,'path');a(!Object.hasOwn(v,k));v[k]=t.slice(eq+1);rel(v[k].endsWith('/')?v[k].slice(0,-1):v[k]);at=end;}a.equal(at,b.length);return v;}
function tar(b,t){
 a(b.length<=2097152&&b.length%512===0);let at=0,ix=0,p=null,ph=0,bytes=0,zero=0,ended=false;
 const dirs=[...t.directories].sort((x,y)=>x.relative_path.split('/').length-y.relative_path.split('/').length||(x.relative_path<y.relative_path?-1:x.relative_path>y.relative_path?1:0));
 const expected=[...dirs.map(x=>({kind:'5',...x})),...t.files.map(x=>({kind:'0',...x}))];
 while(at<b.length){
  const h=b.subarray(at,at+512);a.equal(h.length,512);
  if(h.every(x=>x===0)){a(at+1024<=b.length&&b.subarray(at).every(x=>x===0));a.equal(p,null);at=b.length;ended=true;break;}
  let sum=0;for(let j=0;j<512;j++)sum+=j>=148&&j<156?32:h[j];a.equal(sum,oct(h.subarray(148,156)));a.equal(h.subarray(257,263).toString('ascii'),'ustar\0');a.equal(h.subarray(263,265).toString('ascii'),'00');
  for(const [l,r] of [[108,116],[116,124],[136,148],[329,337],[337,345]])a.equal(oct(h.subarray(l,r)),0);
  for(const [l,r] of [[157,257],[265,297],[297,329]])a.equal(text(h.subarray(l,r)),'');
  const name=text(h.subarray(0,100)),prefix=text(h.subarray(345,500)),kind=String.fromCharCode(h[156]||48),n=oct(h.subarray(124,136)),mode=oct(h.subarray(100,108)),start=at+512,end=start+n,next=start+Math.ceil(n/512)*512;
  a(next<=b.length&&b.subarray(end,next).every(x=>x===0));at=next;
  if(kind==='x'){a.equal(p,null);a(n<=4096);p=pax(b.subarray(start,end));ph++;continue;}
  a(kind==='0'||kind==='5');const full=p?.path??(prefix?prefix+'/'+name:name);p=null;a(kind!=='5'||full.endsWith('/'));const file=kind==='5'?full.slice(0,-1):full;rel(file);
  a(ix<expected.length);const x=expected[ix++];a.equal(file,x.relative_path);a.equal(kind,x.kind);a.equal(mode,parseInt(x.mode_octal,8));
  if(kind==='5')a.equal(n,0);else{a.equal(n,x.bytes);a.equal(hash(b.subarray(start,end)),x.sha256);bytes+=n;zero+=Number(n===0);}
 }
 a(ended);a.equal(ix,expected.length);a.equal(bytes,t.files.reduce((v,x)=>v+x.bytes,0));
 return {tar_bytes:b.length,files:t.files.length,directories:t.directories.length,entries:ix,pax_headers:ph,payload_bytes:bytes,zero_byte_files:zero};
}
const records=[];
for(let i=0;i<2;i++){
 const t=s.targets[i],cp=path.join(D,'COMMIT.'+i+'.json'),ap=path.join(D,'CLEANUP-ACK.'+i+'.json'),c=json(cp),ack=json(ap);
 a.equal(c.target_id,t.id);a.deepEqual(c.selection,reservation.selection);a.deepEqual(c.admission,reservation.admission);a.deepEqual(c.prior_account,config.prior_account);a.equal(c.inventory_sha256,hash(Buffer.from(JSON.stringify(t))));
 a.equal(c.archive.path,path.join(D,i+'.tar.gz'));a(c.archive.bytes<=327680);a.equal(c.restored_root,path.join(R,config.restore_bases[i]));a.equal(c.payload_bytes,t.files.reduce((v,x)=>v+x.bytes,0));a.equal(c.regular_files,t.files.length);
 for(const k of ['full_eof_crc_verified','restored_files_bytes_sha_and_modes_match','file_and_directory_sync_returned','before_any_original_unlink'])a.equal(c[k],true);
 const zipped=check(c.archive,327680);a.equal(zipped.subarray(0,10).toString('hex'),'1f8b08000000000000ff');
 const first=z.inflateRawSync(zipped.subarray(10),{info:true,maxOutputLength:2097152}),tail=10+first.engine.bytesWritten;
 a.equal(tail+8,zipped.length);a.equal(zipped.readUInt32LE(tail),crc(first.buffer));a.equal(zipped.readUInt32LE(tail+4),first.buffer.length>>>0);
 const decoded=z.gunzipSync(zipped,{maxOutputLength:2097152});a(decoded.equals(first.buffer));const inventory=tar(decoded,t);
 const original=path.join(R,t.id),mp=path.join(original,'.RECOVERABLE-ARCHIVE.v1.json'),m=json(mp);a.deepEqual(m.commit,pin(cp));a.deepEqual(m.selection,reservation.selection);a.deepEqual(m.archive,c.archive);a.deepEqual(ack.commit,pin(cp));a.deepEqual(ack.marker,pin(mp));a.deepEqual(ack.archive,c.archive);
 for(const k of ['original_regulars_absent','original_root_marker_only','restored_sibling_absent','parent_sync_returned'])a.equal(ack[k],true);
 const st=fs.lstatSync(original,{bigint:true});a(st.isDirectory());a.equal(st.dev.toString(),t.root_metadata.device);a.equal(st.ino.toString(),t.root_metadata.inode);a.equal('0'+Number(st.mode&0o7777n).toString(8),t.root_metadata.mode_octal);a.deepEqual(fs.readdirSync(original),['.RECOVERABLE-ARCHIVE.v1.json']);absent(c.restored_root);
 records.push({target:t.id,...inventory,archive_bytes:zipped.length,archive_sha256:c.archive.sha256,single_gzip_stream:true,deflate_consumed:first.engine.bytesWritten,CRC_ISIZE_match:true,tar_bytes_SHA_modes_match:true,marker_commit_ACK_pins_match:true,current_marker_only_same_root_inode:true,restore_sibling_absent:true});
}
const jp=path.join(D,'UNLINK-JOURNAL.jsonl'),jb=read(jp,131072);a(jb.length&&jb.at(-1)===10);const lines=jb.toString('utf8').split('\n');lines.pop();a.equal(lines.length,672);
let seq=0,prev=pin(path.join(D,'RESERVATION.json')).sha256,removed=0,bytes=0;
function event(k,t,i,c){const e={s:++seq,k,t,i,p:prev};if(c)e.c=c;const raw=Buffer.from(lines[seq-1]+'\n');a(raw.equals(Buffer.from(JSON.stringify(e)+'\n')));prev=hash(raw);}
for(let t=0;t<2;t++){event(0,t,-1,pin(path.join(D,'COMMIT.'+t+'.json')).sha256);s.targets[t].files.forEach((f,i)=>{event(1,t,i);event(2,t,i);removed++;bytes+=f.bytes;});event(3,t,-1,pin(path.join(D,'CLEANUP-ACK.'+t+'.json')).sha256);}
a.equal(seq,672);a.equal(final.complete,true);a.equal(final.completed_targets,2);a.equal(final.removed_ack_files,removed);a.equal(final.removed_ack_logical_bytes,bytes);a.equal(final.unacknowledged_remove_possible,false);a.equal(final.failure,'');a.equal(final.automatic_retries,0);a.deepEqual(final.archive_bytes,records.map(x=>x.archive_bytes));a.equal(final.reclaimed_logical_bytes,bytes-records.reduce((v,x)=>v+x.archive_bytes,0)-131072);
const old=path.join(R,'COMPLETED-BINARY-ARCHIVES.actual.v3'),ores=json(path.join(old,'RESERVATION.json')),sel=JSON.parse(check(ores.selection_pin)),binary=[];
a.equal(sel.files.length,5);a.equal(sel.archive_max_bytes,3145728);
for(let i=0;i<5;i++){
 const cp=path.join(old,'COMMIT.'+(i+1)+'.json'),c=json(cp),d=json(path.join(old,'CLEANUP.'+(i+1)+'.json'));
 a.equal(c.index,i);a.deepEqual(c.selection_pin,ores.selection_pin);a.deepEqual(c.original_pin,sel.files[i]);a.deepEqual(c.prior_account_pin,ores.account_pin);check(c.prior_account_pin);a.equal(c.archive_pin.path,path.join(old,'ARCHIVE.'+(i+1)+'.gz'));check(c.archive_pin,3145728);a.deepEqual(d.commit_pin,pin(cp));a.equal(d.index,i);
 for(const k of ['original_removed_and_parent_synced','restore_removed_and_parent_synced'])a.equal(d[k],true);
 for(const k of ['full_gzip_eof_crc_verified','actual_restore_bytes_sha_mode_verified','archive_and_restore_file_directory_synced'])a.equal(c[k],true);
 for(const k of ['bytes','sha256','mode_octal'])a.equal(c.restored_pin[k],c.original_pin[k]);absent(c.original_pin.path);absent(c.restored_pin.path);
 const st=fs.lstatSync(c.archive_pin.path);a.equal(st.dev,c.archive_pin.device);a.equal(st.ino,c.archive_pin.inode);a.equal('0'+(st.mode&0o7777).toString(8),c.archive_pin.mode_octal);
 binary.push({id:c.original_pin.id,bytes:c.archive_pin.bytes,sha256:c.archive_pin.sha256,current_pin_selection_commit_cleanup_match:true});
}
(async()=>{
 const {census}=await import(path.join(R,'resource-v4-helper-preparation52-v1/census.private.v4.mjs')),account=path.join(R,'resource-recoverable-projection-leaves-v1/ACCOUNTING.private.v4.8.json'),c=census(account);a(c.current_bytes+24576<=c.cap_bytes);a.equal(c.archive_records,20);
 const report={schema:'riido-HF33-completed-archive-independent-saved-audit-v1',state:'verified_saved_archives_and_current_cleanup',directories:records,journal:{bytes:jb.length,sha256:hash(jb),events:672,commits:2,intents:334,unlink_ACKs:334,cleanup_ACKs:2,raw_LF_chain_order_match:true,final_chain_sha256:prev},FINAL:final,binary5_current_pins:binary,counts:{legacy_census_mapped:20,binary5:5,directory2:2,actual_archive_records:27,regular_payloads:334,tar_directories:91,tar_logical_entries:425},resource:{current_before_report:c.current_bytes,cap:c.cap_bytes,plus24576_within_cap:true,non_atomic:true,account_bytes:c.account_basis.bytes,account_sha256:c.account_basis.sha256,disk_logical_only:true},limits:['This is nonblind saved archive bookkeeping; prior archive-source review context is shared.','The two current gzip flags0 headers plus first-deflate consumed length,CRC,ISIZE and exact trailer EOF prove one stream with no concatenated/trailing member for these exact archives.','Prior actual physical restore/mode and Sync-before-unlink facts are inherited Root evidence, not independently replayed. Current tar SHA/modes, markers and missing restore siblings are directly checked.','Prior5 binaries rehashed/metadata checked, not decompressed here. Legacy20 current archive pins verified by unchanged census helper/account;27 records does not change legacy mapping20 or quota.','No disk restore/delete/recompression/new archive, Reader/original/native/model/Fit/Go/network; no source/label/protected interpretation or model RAM claim.'],execution:{saved_archive_audit:1,restores:0,deletes:0,Go:0,Reader:0,original_API:0,native:0,model:0,Fit:0,network:0}};
 const b=Buffer.from(JSON.stringify(report)+'\n');a(b.length<=5500);const fd=fs.openSync(out,'wx',0o600);try{fs.writeFileSync(fd,b);fs.fsyncSync(fd);}finally{fs.closeSync(fd);}const d=fs.openSync(path.dirname(out),'r');try{fs.fsyncSync(d);}finally{fs.closeSync(d);}a.equal(hash(fs.readFileSync(out)),hash(b));console.log(JSON.stringify({bytes:b.length,sha256:hash(b),files:334,tar_entries:425,archives:27,census:c.current_bytes}));
})().catch(e=>{console.error(e);process.exitCode=1;});
