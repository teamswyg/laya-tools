package taskverify

// This source is independently authored acceptance, never candidate-provided.
// Its own source digest is bound into the behavioral report's task specification.
const contractTests = `package catalog_test

import (
 "encoding/json"
 "reflect"
 "testing"
 "github.com/teamswyg/laya-tools/pkg/catalog"
)

func verifyConfig() catalog.Config {
 return catalog.Config{Models: []catalog.Model{
  {ID:"small", Quality:.9, Context:128, Price:&catalog.Price{Input:1,Output:1}},
  {ID:"boundary", Quality:.9, Context:256, Price:&catalog.Price{Input:2,Output:2}},
  {ID:"large", Quality:.9, Context:1024, Price:&catalog.Price{Input:3,Output:3}},
 }, QualityFloors:[]float64{.5}, MinConfidence:.9}
}
func verifyRequest() catalog.Request {
 return catalog.Request{Tier:0,Confidence:.99,InputTokens:17,OutputTokens:23}
}
func hasReason(xs []string, want string) bool {
 for _, x := range xs { if x == want { return true } }; return false
}

func TestTaskVerifyMinimumContext(t *testing.T) {
 c := verifyConfig()
 maxInt := int(^uint(0)>>1)
 for _, minimum := range []int{0,1,127,128,129,255,256,257,1023,1024,1025,maxInt} {
  r := verifyRequest(); r.MinContext = minimum
  got, err := catalog.Select(c,r)
  if err != nil { t.Fatalf("minimum %d rejected",minimum) }
  if len(got.Candidates) != len(c.Models) { t.Fatal("candidate coverage lost") }
  expected := ""
  for _,m := range c.Models {
   eligible := m.Context >= minimum
   var found *catalog.Candidate
   for j := range got.Candidates { if got.Candidates[j].ID == m.ID { found = &got.Candidates[j] } }
   if found == nil || found.Eligible != eligible || hasReason(found.Reasons,"min_context_required") != !eligible {
    t.Fatalf("minimum %d: incorrect eligibility/reason for %s",minimum,m.ID)
   }
   if eligible && expected == "" { expected = m.ID }
  }
  if got.Model != expected { t.Fatalf("minimum %d selected wrong model",minimum) }
 }
 for _, negative := range []int{-1,-19,-maxInt-1} {
  r := verifyRequest(); r.MinContext = negative
  if _,err := catalog.Select(c,r); err == nil { t.Fatal("negative minimum accepted") }
 }
 // New zero-value behavior must agree with the same request decoded without the new key.
 r := verifyRequest()
 data,err := json.Marshal(r)
 if err != nil { t.Fatal("marshal") }
 var fields map[string]json.RawMessage
 if json.Unmarshal(data,&fields) != nil || fields["min_context"] != nil { t.Fatal("zero field is not optional") }
 var old catalog.Request
 if json.Unmarshal([]byte("{\"tier\":0,\"confidence\":0.99,\"input_tokens\":17,\"output_tokens\":23}"),&old) != nil { t.Fatal("old JSON") }
 a,err := catalog.Select(c,r); if err != nil { t.Fatal("zero") }
 b,err := catalog.Select(c,old); if err != nil || !reflect.DeepEqual(a,b) { t.Fatal("zero behavior changed") }
 var fromJSON catalog.Request
 if json.Unmarshal([]byte("{\"tier\":0,\"confidence\":0.99,\"input_tokens\":17,\"output_tokens\":23,\"min_context\":256}"),&fromJSON) != nil || fromJSON.MinContext != 256 { t.Fatal("minimum JSON not read") }
 x,err := catalog.Select(c,fromJSON); if err != nil || x.Model != "boundary" { t.Fatal("JSON minimum ignored") }
 fromJSON.MinContext = 128; fromJSON.InputTokens = 128; fromJSON.OutputTokens = 1
 x,err = catalog.Select(c,fromJSON); if err != nil || x.Model != "boundary" { t.Fatal("token capacity lost") }
 for _,v := range x.Candidates { if v.ID=="small" && (v.Eligible || !hasReason(v.Reasons,"context_limit") || hasReason(v.Reasons,"min_context_required")) { t.Fatal("distinct context constraints lost") } }
 // An exact maximum-int context is admissible without summing untrusted requirements.
 c.Models = []catalog.Model{{ID:"maximum",Quality:.9,Context:maxInt,Price:&catalog.Price{}}}
 r.MinContext = maxInt
 x,err = catalog.Select(c,r); if err != nil || x.Model != "maximum" { t.Fatal("maximum boundary rejected") }
}
`
