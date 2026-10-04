// SPDX-License-Identifier: Apache-2.0
// Source preparation only. No compiler, upstream API, model or network calls.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {gunzipSync} from 'node:zlib';
import {fileURLToPath} from 'node:url';
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const assert=(ok,code)=>{if(!ok)throw Error(code);};
export function derive(packet,repo,out){
 assert(path.isAbsolute(packet)&&path.isAbsolute(repo)&&path.isAbsolute(out),'absolute_arguments');
 packet=fs.realpathSync(packet);repo=fs.realpathSync(repo);out=fs.realpathSync(out);
 assert(!out.startsWith(repo+path.sep)&&!out.startsWith(packet+path.sep)&&fs.readdirSync(out).length===0,'owned_empty_output');
 const pins=JSON.parse(fs.readFileSync(path.join(packet,'SOURCE-PINS.public.v1.json')));
 for(const p of pins.packet){const b=fs.readFileSync(path.join(packet,p.path));assert(b.length===p.bytes&&hash(b)===p.sha256,'packet_pin');}
 for(const p of pins.repo){const b=fs.readFileSync(path.join(repo,p.path));assert(b.length===p.bytes&&hash(b)===p.sha256,'repo_pin');}
 const zipped=fs.readFileSync(path.join(packet,'upstream-source.json.gz'));
 const raw=gunzipSync(zipped,{maxOutputLength:80344});assert(raw.length===80343,'source_pack_bound');
 const source=JSON.parse(raw);
 assert(source.schema==='riido-btree-source-pack-v1'&&source.revision==='aeba20f7a1e1315badec4eca4fdc9f754f5f880a'&&source.files.length===5,'source_identity');
 const names=['README.md','LICENSE','go.mod','btree.go','btree_generic.go'];
 fs.mkdirSync(path.join(out,'upstream'),{mode:0o700});
 for(let i=0;i<names.length;i++){
  const f=source.files[i],b=Buffer.from(f.text);
  assert(f.path===names[i]&&b.length===f.bytes&&hash(b)===f.sha256,'source_file_pin');
  const blob=crypto.createHash('sha1').update(Buffer.concat([Buffer.from('blob '+b.length+'\0'),b])).digest('hex');
  assert(blob===f.git_blob,'git_blob');
  fs.writeFileSync(path.join(out,'upstream',f.path),b,{flag:'wx',mode:0o600});
 }
 const input=fs.readFileSync(path.join(packet,'INPUTS.public.v1.json'));
 const main=fs.readFileSync(path.join(packet,'main.go.txt'),'utf8');
 assert(main.split('@INPUT_SHA@').length===2,'input_hash_anchor');
 fs.writeFileSync(path.join(out,'main.go'),main.replace('@INPUT_SHA@',hash(input)),{flag:'wx',mode:0o600});
 fs.copyFileSync(path.join(packet,'main_test.go.txt'),path.join(out,'main_test.go'),fs.constants.COPYFILE_EXCL);
 fs.writeFileSync(path.join(out,'INPUTS.public.v1.json'),input,{flag:'wx',mode:0o600});
 fs.symlinkSync(repo,path.join(out,'repo'),'dir');
 fs.writeFileSync(path.join(out,'go.mod'),'module github.com/teamswyg/laya-tools/internal/btree_utility_probe\n\ngo 1.27.1\n\nrequire (\n github.com/google/btree v0.0.0\n github.com/teamswyg/laya-tools v0.0.0\n golang.org/x/text v0.25.0 // indirect\n)\n\nreplace github.com/google/btree => ./upstream\nreplace github.com/teamswyg/laya-tools => ./repo\n',{flag:'wx',mode:0o600});
 fs.copyFileSync(path.join(repo,'go.sum'),path.join(out,'go.sum'),fs.constants.COPYFILE_EXCL);
 return {schema:'riido-btree-derived-source-v1',input_sha256:hash(input),original_calls:0,model_calls:0,files:['main.go','main_test.go','go.mod','go.sum','INPUTS.public.v1.json',...names.map(n=>'upstream/'+n)].map(p=>{const b=fs.readFileSync(path.join(out,p));return{path:p,bytes:b.length,sha256:hash(b)};})};
}
if(process.argv[1]&&path.resolve(process.argv[1])===fileURLToPath(import.meta.url)){
 try{assert(process.argv.length===5,'arguments');console.log(JSON.stringify(derive(...process.argv.slice(2))));}
 catch{console.error('btree_source_preparation_failed');process.exitCode=1;}
}
