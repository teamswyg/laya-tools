// SPDX-License-Identifier: Apache-2.0
// Called on Go list metadata by maintainers, before any original API trial.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
const sha=b=>crypto.createHash('sha256').update(b).digest('hex');
export function decodeStream(text){
 const objects=[];let depth=0,start=-1,inString=false,escape=false;
 for(let i=0;i<text.length;i++){
  const c=text[i];
  if(inString){if(escape)escape=false;else if(c==='\\')escape=true;else if(c==='"')inString=false;continue;}
  if(c==='"'){inString=true;continue;}
  if(c==='{'){if(depth===0)start=i;depth++;}
  if(c==='}'){if(--depth===0){objects.push(JSON.parse(text.slice(start,i+1)));start=-1;}}
 }
 if(depth!==0||inString||objects.length===0)throw Error('metadata_stream');
 return objects;
}
export function closure(text){
 const rows=decodeStream(text),all=[],nonstandard=[];
 const fields=['GoFiles','CgoFiles','CFiles','CXXFiles','MFiles','HFiles','FFiles','SFiles','SysoFiles','EmbedFiles'];
 for(const p of rows){
  if(p.Error||p.Incomplete)throw Error('incomplete_compiler_package');
  const files=[];
  for(const field of fields)for(const name of p[field]||[]){
   const b=fs.readFileSync(path.join(p.Dir,name));
   files.push({field,path:name,bytes:b.length,sha256:sha(b)});
  }
  files.sort((a,b)=>a.field.localeCompare(b.field)||a.path.localeCompare(b.path));
  const record={package:p.ImportPath,standard:!!p.Standard,files};all.push(record);
  if(!p.Standard)nonstandard.push({...record,ignored_go_files:p.IgnoredGoFiles||[]});
 }
 all.sort((a,b)=>a.package.localeCompare(b.package));nonstandard.sort((a,b)=>a.package.localeCompare(b.package));
 return {schema:'riido-btree-compiler-selected-source-v1',selected_packages:all.length,selected_files:all.reduce((n,p)=>n+p.files.length,0),selected_bytes:all.reduce((n,p)=>n+p.files.reduce((m,f)=>m+f.bytes,0),0),aggregate_sha256:sha(Buffer.from(JSON.stringify(all)+'\n')),aggregate_recipe:'Package, standard flag, field/file/bytes/SHA records; sorted package, then field/file with JS localeCompare; UTF8 JSON+LF. Absolute paths excluded.',nonstandard,scope:'Actual non-race Go list source fields before build; not generated compiler intermediates, compiler internals, SDK or complete C toolchain provenance.'};
}
