package taskverify

// Independently authored external-package assertions for the new exact text
// policy. Fixed byte vectors and encoding/hex provide expectations without
// treating upstream Parse or candidate-authored assertions as the oracle.
const uuidCanonicalContractTests = `package uuid_test

import (
 "encoding/hex"
 "errors"
 "strings"
 "sync"
 "sync/atomic"
 "testing"
 "github.com/google/uuid"
)

const canonicalSample = "01234567-89ab-cdef-0123-456789abcdef"
func canonicalReject(t *testing.T, s string) {
 t.Helper(); got,err:=uuid.ParseCanonical(s)
 if err==nil || got!=(uuid.UUID{}) { t.Fatal("noncanonical input did not return error and zero UUID") }
}
func TestTaskVerifyCanonicalFixedVectors(t *testing.T) {
 vectors:=[]struct{text string; want uuid.UUID}{
  {"00000000-0000-0000-0000-000000000000",uuid.UUID{}},
  {"ffffffff-ffff-ffff-ffff-ffffffffffff",uuid.UUID{255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255}},
  {canonicalSample,uuid.UUID{1,35,69,103,137,171,205,239,1,35,69,103,137,171,205,239}},
  {"ff001122-3344-5566-7788-99aabbccddee",uuid.UUID{255,0,17,34,51,68,85,102,119,136,153,170,187,204,221,238}},
 }
 for _,v:=range vectors { got,err:=uuid.ParseCanonical(v.text); if err!=nil || got!=v.want { t.Fatal("fixed canonical vector or unrestricted bit pattern changed") } }
}
func TestTaskVerifyCanonicalEveryBytePosition(t *testing.T) {
 for position:=0;position<36;position++ {
  separator:=position==8 || position==13 || position==18 || position==23
  for value:=0;value<256;value++ {
   changed:=[]byte(canonicalSample); changed[position]=byte(value); input:=string(changed)
   valid:=value==int('-'); if !separator { valid=(value>=int('0')&&value<=int('9')) || (value>=int('a')&&value<=int('f')) }
   if !valid { canonicalReject(t,input); continue }
   raw,err:=hex.DecodeString(strings.ReplaceAll(input,"-","")); if err!=nil || len(raw)!=16 { t.Fatal("independent authored hex expectation failed") }
   var want uuid.UUID; copy(want[:],raw)
   got,err:=uuid.ParseCanonical(input); if err!=nil || got!=want { t.Fatal("accepted byte-position matrix lost the exact decoded value") }
  }
 }
}
func TestTaskVerifyCanonicalAlternativeLengths(t *testing.T) {
 raw:=strings.ReplaceAll(canonicalSample,"-","")
 alternatives:=[]string{
  strings.ToUpper(canonicalSample),"urn:uuid:"+canonicalSample,"URN:UUID:"+canonicalSample,
  raw,strings.ToUpper(raw),"{"+canonicalSample+"}","["+canonicalSample+"]",
  " "+canonicalSample,canonicalSample+" ",canonicalSample+"\n","\t"+canonicalSample,
  canonicalSample+"\x00",canonicalSample[:35],canonicalSample+"0",
  strings.ReplaceAll(canonicalSample,"-",""),"é"+canonicalSample[1:],"😀"+canonicalSample[1:],
  canonicalSample[:18]+"--"+canonicalSample[19:],strings.Repeat("0",4096),
 }
 for _,input:=range alternatives { canonicalReject(t,input) }
 for length:=0;length<=100;length++ { canonicalReject(t,strings.Repeat("0",length)) }
}
func TestTaskVerifyCanonicalFailureIsZero(t *testing.T) {
 const allNonzero="ffffffff-ffff-ffff-ffff-ffffffffffff"
 for position:=0;position<len(allNonzero);position++ {
  if allNonzero[position]=='-' { continue }
  for _,bad:=range []byte{'g','G',0,255} {
   changed:=[]byte(allNonzero); changed[position]=bad; canonicalReject(t,string(changed))
  }
 }
}
func TestTaskVerifyLegacyParserCompatibility(t *testing.T) {
 want:=uuid.UUID{1,35,69,103,137,171,205,239,1,35,69,103,137,171,205,239}
 raw:=strings.ReplaceAll(canonicalSample,"-","")
 for _,input:=range []string{canonicalSample,strings.ToUpper(canonicalSample),"urn:uuid:"+canonicalSample,"URN:UUID:"+strings.ToUpper(canonicalSample),raw,strings.ToUpper(raw),"{"+canonicalSample+"}","["+canonicalSample+"]"} {
  for _,bytesInput:=range []bool{false,true} {
   var got uuid.UUID; var err error
   if bytesInput { got,err=uuid.ParseBytes([]byte(input)) } else { got,err=uuid.Parse(input) }
   if err!=nil || got!=want { t.Fatal("legacy accepted encoding was narrowed or decoded differently") }
  }
 }
 for _,tc:=range []struct{input string; category error}{
  {"",uuid.ErrInvalidLength}, {canonicalSample+"0",uuid.ErrInvalidLength},
  {"bad:uuid:"+canonicalSample,uuid.ErrInvalidURNPrefix},
  {"g"+canonicalSample[1:],uuid.ErrInvalidUUIDFormat},
  {strings.Replace(canonicalSample,"-","_",1),uuid.ErrInvalidUUIDFormat},
 } {
  _,err:=uuid.Parse(tc.input); _,bytesErr:=uuid.ParseBytes([]byte(tc.input))
  if !errors.Is(err,tc.category)||!errors.Is(bytesErr,tc.category) { t.Fatal("legacy parse error category changed") }
 }
}

type canonicalEntropyProbe struct { calls int64; size int64 }
var canonicalEntropyError=errors.New("authored forbidden entropy read")
func (p *canonicalEntropyProbe) Read(b []byte) (int,error) { atomic.AddInt64(&p.calls,1); atomic.StoreInt64(&p.size,int64(len(b))); return 0,canonicalEntropyError }
type canonicalPoolProbe struct { calls int }
func (p *canonicalPoolProbe) Read(b []byte) (int,error) { p.calls++; for i:=range b { b[i]=byte(i) }; return len(b),nil }
func TestTaskVerifyCanonicalDeterminismAndEntropy(t *testing.T) {
 uuid.DisableRandPool(); defer uuid.DisableRandPool()
 probe:=new(canonicalEntropyProbe); uuid.SetRand(probe); defer uuid.SetRand(nil)
 want:=uuid.UUID{1,35,69,103,137,171,205,239,1,35,69,103,137,171,205,239}
 var group sync.WaitGroup
 for worker:=0;worker<16;worker++ { group.Add(1); go func(){ defer group.Done(); for repeat:=0;repeat<64;repeat++ {
  got,err:=uuid.ParseCanonical(canonicalSample); if err!=nil||got!=want { t.Error("canonical concurrent/repeated output changed"); return }
  got,err=uuid.ParseCanonical(canonicalSample[:35]+"g"); if err==nil||got!=(uuid.UUID{}) { t.Error("canonical concurrent failure changed"); return }
 } }() }
 group.Wait(); if atomic.LoadInt64(&probe.calls)!=0 { t.Fatal("canonical parser read random entropy") }
 _,err:=uuid.NewRandom(); if !errors.Is(err,canonicalEntropyError)||atomic.LoadInt64(&probe.calls)!=1||atomic.LoadInt64(&probe.size)!=16 { t.Fatal("canonical parser changed configured randomness or pool enablement") }
 pool:=new(canonicalPoolProbe); uuid.DisableRandPool(); uuid.EnableRandPool(); uuid.SetRand(pool)
 if _,err:=uuid.NewRandom(); err!=nil||pool.calls!=1 { t.Fatal("authored pool setup failed") }
 for _,input:=range []string{canonicalSample,canonicalSample[:35]+"g"} { _,_=uuid.ParseCanonical(input) }
 next,err:=uuid.NewRandom(); expected:=uuid.UUID{16,17,18,19,20,21,70,23,152,25,26,27,28,29,30,31}
 if err!=nil||pool.calls!=1||next!=expected { t.Fatal("canonical parser changed pool contents, cursor or enablement") }
}
`
