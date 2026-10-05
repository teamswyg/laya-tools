// Original portable CI proof authoring; no child/API/model execution from this helper.
import fs from 'node:fs';
import path from 'node:path';
import {createHash} from 'node:crypto';
const [phase,mode,proof,priv]=process.argv.slice(2),module='riido.example/jwt151-cost-preview';
const need=(x,m)=>{if(!x)throw new Error(m)};
const hash=b=>createHash('sha256').update(b).digest('hex');
const read=p=>{const s=fs.lstatSync(p);need(s.isFile()&&!s.isSymbolicLink()&&s.size<=32*1024*1024,'bounded regular file');const b=fs.readFileSync(p);need(b.length===s.size,'stable bytes');return b};
const pin=p=>{const b=read(p);return {path:p.replaceAll('\\','/'),bytes:b.length,sha256:hash(b)}};
const write=(name,v)=>fs.writeFileSync(path.join(proof,name),JSON.stringify(v)+'\n',{flag:'wx',mode:0o644});
const hostPath=/\/(?:private|Users|home|opt|Library|Applications|tmp|var|usr|etc|Volumes|System|root|work|runner|workspace)\/|(?:^|[\s"'(])[A-Za-z]:[\\/]/;
const publicationSafe=b=>need(!hostPath.test(b.toString('utf8')),'absolute host path in public proof');
if(phase==='publication'){
  need(mode==='all'&&proof==='ci-artifacts/proof'&&priv==='ci-private/publication','fixed publication scope');
  const records=[];
  function visit(dir,relative=''){
    const st=fs.lstatSync(dir);need(st.isDirectory()&&!st.isSymbolicLink(),'proof directory');
    for(const leaf of fs.readdirSync(dir).sort()){
      need(/^[A-Za-z0-9_.-]+$/.test(leaf),'safe proof filename');
      const p=path.join(dir,leaf),rel=relative?relative+'/'+leaf:leaf,s=fs.lstatSync(p);
      if(s.isDirectory()){need(relative===''&&['normal','race','producer'].includes(leaf),'fixed proof directory depth');visit(p,rel)}
      else{need(s.isFile()&&!s.isSymbolicLink(),'proof leaf');const b=read(p);publicationSafe(b);records.push({path:rel,bytes:b.length,sha256:hash(b)})}
    }
  }
  visit(proof);need(records.length>0,'nonempty first proof');
  const gate={schema:'riido-retrieval153-publication-safety-v1',all_retained_public_leaves_scanned:true,records,scope:'Full originals retained before this scan. A failure blocks every artifact upload; no redaction, dropped stream or success reinterpretation. Private helper and locator captures are not published.'};
  publicationSafe(Buffer.from(JSON.stringify(gate)));write('PUBLICATION-SAFE.json',gate);
}else{
need(['environment','sources','finish'].includes(phase)&&['normal','race','producer'].includes(mode)&&proof===`ci-artifacts/proof/${mode}`&&priv===`ci-private/${mode}`,'fixed CI scope');
const env=JSON.parse(read(path.join(proof,'environment.stdout'))),locations=JSON.parse(read(path.join(priv,'sdk-locations.stdout')));
// go env reports the effective configuration-file path: GOENV=off becomes "".
// Check both the disabled OS setting and its reported effective value.
need(env.GOVERSION==='go1.27.1'&&env.GOTOOLCHAIN==='local'&&env.GOWORK==='off'&&process.env.GOENV==='off'&&env.GOENV===''&&env.GODEBUG==='goindex=0'&&env.GOFLAGS===''&&env.GOEXPERIMENT===''&&env.CGO_ENABLED===(mode==='race'?'1':'0'),'exact Go/environment');
if(mode==='producer')need(env.GOOS==='darwin'&&env.GOARCH==='arm64'&&process.platform==='darwin'&&process.arch==='arm64','native producer');
const safeRelative=p=>typeof p==='string'&&p.length>0&&!path.isAbsolute(p)&&!p.split('/').includes('..')&&!/[\x00-\x1f|]/.test(p);
const sums=read('SHA256SUMS').toString('utf8').trimEnd().split('\n').map(line=>{const m=/^([a-f0-9]{64})  (.+)$/.exec(line);need(m&&safeRelative(m[2]),'source checksum row');return {sha256:m[1],path:m[2]}});
need(sums.length===45&&new Set(sums.map(x=>x.path)).size===45,'exact source45');
const source45=sums.map(e=>{const b=read(e.path);need(hash(b)===e.sha256,'source45 changed');return {...e,bytes:b.length,mode:fs.lstatSync(e.path).mode&0o7777}});
if(phase==='environment')write('ENVIRONMENT.json',{schema:'riido-retrieval153-ci-environment-v1',mode,env,version:pin(path.join(proof,'version.stdout')),source45_checksum:pin('SHA256SUMS'),original_worker_starts:0,original_fixture_calls:0});
if(phase==='sources'){
  const packages=[],selected=[],fields=['GoFiles','CgoFiles','CFiles','CXXFiles','MFiles','HFiles','FFiles','SFiles','SwigFiles','SwigCXXFiles','SysoFiles','EmbedFiles'];
  for(const line of read(path.join(proof,'selected-source-rows.stdout')).toString('utf8').split('\n').filter(Boolean)){
    const cells=line.split('|');need(cells[0]!=='E','selected package error');
    if(cells[0]==='P'){
      need(cells.length===4&&['true','false'].includes(cells[2]),'package row');
      need(cells[2]==='true'||cells[3]===module||cells[3]==='github.com/golang-jwt/jwt/v5'||cells[1]===module+'.test','unexpected external module');
      packages.push({import_path:cells[1],standard:cells[2]==='true',module_path:cells[3]||null});
    }else{
      need(cells[0]==='S'&&cells.length===4&&fields.includes(cells[2])&&safeRelative(cells[1])&&safeRelative(cells[3]),'selected SDK row');
      const rel=path.posix.join('src',cells[1],cells[3]),b=read(path.join(locations.GOROOT,rel));
      selected.push({package:cells[1],field:cells[2],path:rel,bytes:b.length,sha256:hash(b)});
    }
  }
  need(packages.some(x=>x.import_path===module)&&selected.length>0,'selected closure empty');
  let generated=null;
  if(mode!=='producer'){
    const locator=read(path.join(priv,'testmain-locator.stdout')).toString('utf8').trim().split('\n');
    need(locator.length===1&&path.basename(locator[0]).endsWith('-d')&&path.isAbsolute(locator[0])&&path.relative(locations.GOCACHE,locator[0]).split(path.sep).every(x=>x!=='..'),'generated cache origin');
    const body=read(locator[0]);need(body.length<=16384&&body.toString('utf8').startsWith("\n// Code generated by 'go test'. DO NOT EDIT."),'bounded generated testmain');
    fs.writeFileSync(path.join(proof,'generated-testmain.go'),body,{flag:'wx',mode:0o644});generated=pin(path.join(proof,'generated-testmain.go'));
  }
  const notices=[];
  for(const leaf of ['LICENSE','PATENTS']){const b=read(path.join(locations.GOROOT,leaf));fs.writeFileSync(path.join(proof,'Go-'+leaf),b,{flag:'wx',mode:0o644});notices.push(pin(path.join(proof,'Go-'+leaf)))}
  for(const [src,out]of [['LICENSE','OWN-LICENSE'],['NOTICE','OWN-NOTICE'],['jwt-source/LICENSE','JWT-LICENSE']]){const b=read(src);fs.writeFileSync(path.join(proof,out),b,{flag:'wx',mode:0o644});notices.push(pin(path.join(proof,out)))}
  const tools=[['go',path.join(locations.GOROOT,'bin/go')],...['compile','asm','link','vet'].map(n=>[n,path.join(locations.GOTOOLDIR,n)])].map(([name,p])=>{const b=read(p);return{name,bytes:b.length,sha256:hash(b)}});
  write('SOURCE-QUALIFICATION.json',{schema:'riido-retrieval153-portable-ci-source-qualification-v1',mode,env,source45,tools,packages,selected,generated_testmain:generated,selection_scope:mode==='producer'?'ordinary go list -deps without -test; CGO0 producer':'test superset and full generated testmain; CGO0 normal or CGO1 race',notices,scope:'Portable exact selected-source/tool hashes and full generated source; host locator paths omitted. Not full raw Go-list JSON/host-path recovery, target execution or original fixture evidence.'});
}
if(phase==='finish'){
  const first=[];
  for(const f of fs.readdirSync(proof).filter(x=>x.endsWith('.status')).sort()){
    const label=f.slice(0,-7),status=read(path.join(proof,f)).toString('utf8');need(status==='0\n','retained CI stage failed');
    first.push({stage:label,status:0,argv:pin(path.join(proof,label+'.argv')),stdout:pin(path.join(proof,label+'.stdout')),stderr:pin(path.join(proof,label+'.stderr'))});
  }
  const shapeLog=mode==='race'?'owned-race.stdout':mode==='normal'?'owned-tests.stdout':null;let shape=null;
  if(shapeLog){const matches=read(path.join(proof,shapeLog)).toString('utf8').split('\n').filter(x=>x.includes('riido-retrieval-actual-go-shape-proof-v1'));need(matches.length===1,'exact one shape report');shape=JSON.parse(matches[0].slice(matches[0].indexOf('{')));for(const k of ['Rows','Pairs','Wire','ScoreScalar','Padding','Extension','Upper','Cap','Headroom','JWT','Rank','Model'])need(Number.isSafeInteger(shape[k])&&shape[k]>=0,'shape integer');need(shape.Wire>0&&shape.ScoreScalar>0&&shape.Upper>0&&shape.schema==='riido-retrieval-actual-go-shape-proof-v1'&&shape.Rows===128&&shape.Pairs===160&&shape.ScoreScalar<=32&&shape.Padding===128*8*(32-shape.ScoreScalar)&&shape.Extension===4096&&shape.Upper===shape.Wire+shape.Padding+shape.Extension&&shape.Cap===196608&&shape.Headroom===shape.Cap-shape.Upper&&shape.JWT===0&&shape.Rank===0&&shape.Model===0&&/^[a-f0-9]{64}$/.test(shape.synthetic_wire_sha256),'shape equation/unknown');write('SHAPE.json',shape)}
  let binary=null;
  if(mode==='producer'){
    binary=pin('ci-artifacts/bin/retrieval-preview');need(binary.bytes>0&&binary.bytes<=4194304,'prospective binary transfer bound');
    const info=JSON.parse(read(path.join(proof,'buildinfo.stdout')));const list=Array.isArray(info)?info:[info];need(list.length===1&&list[0].GoVersion==='go1.27.1','buildinfo Go');
    const settings=list[0].Settings??[];const setting=k=>settings.find(x=>x.Key===k)?.Value;
    need(setting('GOOS')==='darwin'&&setting('GOARCH')==='arm64'&&setting('CGO_ENABLED')==='0'&&setting('-trimpath')==='true'&&setting('-ldflags')==='-s -w'&&!settings.some(x=>x.Key.startsWith('vcs')),'producer build settings');
    need(fs.existsSync('ci-artifacts/proof/normal/SHAPE.json'),'normal actual shape before producer');
  }
  const run={repository:process.env.GITHUB_REPOSITORY??'',commit:process.env.GITHUB_SHA??'',run_id:process.env.GITHUB_RUN_ID??'',run_attempt:process.env.GITHUB_RUN_ATTEMPT??''};
  need(/^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+$/.test(run.repository)&&/^[a-f0-9]{40}$/.test(run.commit)&&/^\d+$/.test(run.run_id)&&/^\d+$/.test(run.run_attempt),'CI public identity');
  const source=pin(path.join(proof,'SOURCE-QUALIFICATION.json'));
  const finalProof={schema:'riido-retrieval153-ci-producer-and-synthetic-proof-v1',mode,run,env,first,source,binary,shape,normal_shape_ref:mode==='producer'?pin('ci-artifacts/proof/normal/SHAPE.json'):null,original_worker_starts:0,original_fixture_API_calls:0,key_generations:0,model_calls:0,Fit_calls:0,synthetic_rank_calls:'not_counted',upstream_package_startup:'occurs in test processes; not measured',scope:'Owned synthetic tests/source compilation only. A downloaded binary remains unexecuted until a separately frozen local one-shot. Race CGO1 proof is distinct from CGO0 producer authority; no speed/generalization claim.'};
  for(const leaf of fs.readdirSync(proof)){const p=path.join(proof,leaf);const s=fs.lstatSync(p);need(s.isFile()&&!s.isSymbolicLink(),'proof leaf');const b=read(p);publicationSafe(b);}
  publicationSafe(Buffer.from(JSON.stringify(finalProof)));write('PROOF.json',finalProof);
}
}
