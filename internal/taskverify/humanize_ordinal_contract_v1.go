package taskverify

// The signed decimal oracle uses arbitrary-precision arithmetic and an explicit
// residue table, independently of the authored correct candidate's string tails.
const humanizeOrdinalContractTests = `package humanize_test

import (
 "math"
 "math/big"
 "strconv"
 "testing"
 "github.com/dustin/go-humanize"
)

var ordinalSuffixByResidue = [100]string{
 1:"st",2:"nd",3:"rd",
 21:"st",22:"nd",23:"rd",31:"st",32:"nd",33:"rd",
 41:"st",42:"nd",43:"rd",51:"st",52:"nd",53:"rd",
 61:"st",62:"nd",63:"rd",71:"st",72:"nd",73:"rd",
 81:"st",82:"nd",83:"rd",91:"st",92:"nd",93:"rd",
}
func independentOrdinal64(x int64) string {
 var signed,magnitude,residue big.Int
 signed.SetInt64(x); magnitude.Abs(&signed)
 residue.Mod(&magnitude,big.NewInt(100))
 suffix:=ordinalSuffixByResidue[residue.Int64()]; if suffix=="" { suffix="th" }
 return signed.String()+suffix
}
func checkOrdinal64(t *testing.T,x int64) {
 t.Helper()
 if got,want:=humanize.Ordinal64(x),independentOrdinal64(x); got!=want { t.Fatalf("Ordinal64(%d): got %q, want %q",x,got,want) }
}
func TestTaskVerifyOrdinal64SignatureAndExamples(t *testing.T) {
 var format func(int64) string=humanize.Ordinal64
 for _,tc:=range []struct{x int64; want string}{
  {0,"0th"},{1,"1st"},{2,"2nd"},{3,"3rd"},{4,"4th"},{10,"10th"},{11,"11th"},{12,"12th"},{13,"13th"},
  {21,"21st"},{22,"22nd"},{23,"23rd"},{101,"101st"},{111,"111th"},{112,"112th"},{113,"113th"},
  {-1,"-1st"},{-2,"-2nd"},{-3,"-3rd"},{-4,"-4th"},{-10,"-10th"},{-11,"-11th"},{-12,"-12th"},{-13,"-13th"},
  {-21,"-21st"},{-22,"-22nd"},{-23,"-23rd"},{-101,"-101st"},{-111,"-111th"},{-112,"-112th"},{-113,"-113th"},
  {math.MinInt64,"-9223372036854775808th"},{math.MaxInt64,"9223372036854775807th"},
 } { if got:=format(tc.x); got!=tc.want { t.Fatalf("explicit Ordinal64(%d): got %q, want %q",tc.x,got,tc.want) } }
}
func TestTaskVerifyOrdinal64EverySuffixResidue(t *testing.T) {
 // Every magnitude residue is exercised at small, wide, float-inexact and
 // near-MaxInt64 magnitudes, with both signs and every teen exception.
 for _,block:=range []int64{0,1,2,10,1000,21474836,42949673,90071992547409,92233720368547757} {
  for residue:=int64(0);residue<100;residue++ {
   x:=block*100+residue
   checkOrdinal64(t,x); checkOrdinal64(t,-x)
  }
 }
}
func TestTaskVerifyOrdinal64FullRangeDigits(t *testing.T) {
 for offset:=int64(0);offset<=256;offset++ {
  checkOrdinal64(t,math.MinInt64+offset); checkOrdinal64(t,math.MaxInt64-offset)
 }
 for _,boundary:=range []int64{1<<31,1<<32,1<<53,1<<54,1<<62} {
  for offset:=int64(-128);offset<=128;offset++ {
   x:=boundary+offset; checkOrdinal64(t,x); checkOrdinal64(t,-x)
  }
 }
}
func TestTaskVerifyOrdinal64DeterministicRange(t *testing.T) {
 // Fixed public arithmetic covers high and low signed bits reproducibly.
 // These are verifier controls for one task, not separate golden requests.
 state:=uint64(0x5bcd1234d00d2026)
 var samples [4096]int64
 for i:=range samples {
  state=state*6364136223846793005+1442695040888963407
  samples[i]=int64(state); checkOrdinal64(t,samples[i])
 }
 // Repeated calls in a different order cannot reuse another input's result.
 for i:=len(samples)-1;i>=0;i-- { checkOrdinal64(t,samples[i]) }
}
func TestTaskVerifyOriginalIntOrdinal(t *testing.T) {
 var legacy func(int) string=humanize.Ordinal
 for x:=0;x<=1200;x++ {
  if got,want:=legacy(x),independentOrdinal64(int64(x)); got!=want { t.Fatal("existing positive int API changed") }
  if x!=0 { if got,want:=legacy(-x),strconv.Itoa(-x)+"th"; got!=want { t.Fatal("existing negative int API changed") } }
 }
 for _,x:=range []int{math.MinInt32,math.MaxInt32} {
  want:=strconv.Itoa(x)+"th"; if x>=0 { want=independentOrdinal64(int64(x)) }
  if legacy(x)!=want { t.Fatal("existing int boundary behavior changed") }
 }
 if strconv.IntSize==64 {
  lo,hi:=int64(math.MinInt64),int64(math.MaxInt64)
  if legacy(int(lo))!="-9223372036854775808th" || legacy(int(hi))!="9223372036854775807th" { t.Fatal("existing wide int boundary behavior changed") }
 }
}
`
