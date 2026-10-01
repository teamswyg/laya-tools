package taskverify

// The event-envelope request is checked through the public API. Raw item values
// stay opaque; expected counts and totals are independent authored arithmetic.
const eventKeyBoundsContractTests = `package taskoutcome_test

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "fmt"
 "reflect"
 "strings"
 "testing"
 "github.com/teamswyg/laya-tools/pkg/taskoutcome"
)

const contractThread = "{\"type\":\"thread.started\"}\n"
const contractTurn = "{\"type\":\"turn.started\"}\n"
const contractCompleted = "{\"type\":\"turn.completed\",\"usage\":{\"input_tokens\":12,\"cached_input_tokens\":3,\"output_tokens\":5,\"reasoning_output_tokens\":1}}\n"
var envelopeKinds = []string{"thread.started","turn.started","item.started","item.updated","item.completed","turn.completed","turn.failed","error","future.control"}

func contractEnvelope(kind string,extras []string) string {
 return contractEnvelopeAt(kind,extras,0)
}
func contractEnvelopeAt(kind string,extras []string,typePosition int) string {
 fields:=make([]string,len(extras)+1)
 copy(fields,extras[:typePosition]); fields[typePosition]=fmt.Sprintf("\"type\":%q",kind); copy(fields[typePosition+1:],extras[typePosition:])
 return "{"+strings.Join(fields,",")+"}"
}
func contractExtras(n int) []string {
 out:=make([]string,n); for i:=range out { out[i]=fmt.Sprintf("\"future_%02d\":null",i) }; return out
}
func contractContext(kind,event string) string {
 switch kind {
 case "thread.started": return event+"\n"+contractTurn+contractCompleted
 case "turn.started": return contractThread+event+"\n"+contractCompleted
 case "turn.completed","turn.failed": return contractThread+contractTurn+event+"\n"
 case "error": return contractThread+contractTurn+contractCompleted+event+"\n"
 default: return contractThread+contractTurn+event+"\n"+contractCompleted
 }
}
func contractReject(t *testing.T,trace string,want taskoutcome.ErrorCode) {
 t.Helper(); s,err:=taskoutcome.Summarize(strings.NewReader(trace),taskoutcome.Metadata{})
 if err!=want || !reflect.DeepEqual(s,taskoutcome.Summary{}) { t.Fatal("malformed envelope did not return fixed error and empty summary") }
}
func contractAccept(t *testing.T,trace string) taskoutcome.Summary {
 t.Helper(); s,err:=taskoutcome.Summarize(strings.NewReader(trace),taskoutcome.Metadata{})
 if err!=nil { t.Fatal("valid boundary envelope rejected") }
 h:=sha256.Sum256([]byte(trace))
 if s.TraceSHA256!=hex.EncodeToString(h[:]) || s.TraceBytes!=int64(len(trace)) || s.ObservedModel!="unknown" || s.ProcessExit!="unknown" || s.TaskAcceptance!="unknown" { t.Fatal("trace binding or unknown identity semantics changed") }
 return s
}
func contractRetainsMarker(value reflect.Value,marker string) bool {
 if !value.IsValid() { return false }
 switch value.Kind() {
 case reflect.String: return strings.Contains(value.String(),marker)
 case reflect.Interface,reflect.Pointer:
  return !value.IsNil() && contractRetainsMarker(value.Elem(),marker)
 case reflect.Struct:
  for i:=0;i<value.NumField();i++ { if contractRetainsMarker(value.Field(i),marker) { return true } }
 case reflect.Array,reflect.Slice:
  if value.Type().Elem().Kind()==reflect.Uint8 {
   raw:=make([]byte,value.Len()); for i:=range raw { raw[i]=byte(value.Index(i).Uint()) }
   if strings.Contains(string(raw),marker) { return true }
  }
  for i:=0;i<value.Len();i++ { if contractRetainsMarker(value.Index(i),marker) { return true } }
 case reflect.Map:
  entries:=value.MapRange(); for entries.Next() {
   if contractRetainsMarker(entries.Key(),marker) || contractRetainsMarker(entries.Value(),marker) { return true }
  }
 }
 return false
}
func TestTaskVerifyEnvelopeDuplicateKeys(t *testing.T) {
 for _,kind:=range envelopeKinds {
  // Exercise every unknown-key position in a complete 64-key envelope,
  // with separated duplicate partners; type itself is the first key.
  for position:=0;position<63;position++ {
   extras:=contractExtras(63); partner:=(position+31)%63; extras[partner]=extras[position]
   for _,typePosition:=range []int{0,63} { contractReject(t,contractContext(kind,contractEnvelopeAt(kind,extras,typePosition)),taskoutcome.ErrDuplicate) }
  }
  // Vary type through all64 slots with a fixed separated duplicate pair.
  // This catches guards that begin only once the discriminator is decoded.
  for typePosition:=0;typePosition<64;typePosition++ {
   extras:=contractExtras(63); extras[31]=extras[0]
   contractReject(t,contractContext(kind,contractEnvelopeAt(kind,extras,typePosition)),taskoutcome.ErrDuplicate)
  }
  for _,pair:=range []string{
   "\"future\":1,\"\\u0066uture\":2",
   "\"é\":1,\"\\u00e9\":2",
   "\"😀\":1,\"\\ud83d\\ude00\":2",
   "\"\":1,\"\":2",
  } {
   contractReject(t,contractContext(kind,contractEnvelope(kind,[]string{pair})),taskoutcome.ErrDuplicate)
  }
  // JSON names are case-sensitive, and an empty unknown key is valid.
  contractAccept(t,contractContext(kind,contractEnvelope(kind,[]string{"\"future\":1","\"Future\":2","\"\":null"})))
 }
 // Escaped original control keys keep their existing duplicate semantics.
 for _,event:=range []string{
  "{\"type\":\"thread.started\",\"\\u0074ype\":\"thread.started\"}",
  "{\"type\":\"thread.started\",\"thread_id\":\"a\",\"\\u0074hread_id\":\"b\"}",
 } { contractReject(t,event,taskoutcome.ErrDuplicate) }
}
func TestTaskVerifyEnvelopeKeyCountBoundaries(t *testing.T) {
 for _,kind:=range envelopeKinds {
  // Count every original control field as well as unknown extensions. All
  // 16 combinations are valid envelopes for these phases; item stays opaque
  // during an active turn and null usage keeps its existing unknown meaning.
  controls:=[]string{"\"thread_id\":\"AUTHORED_THREAD\"","\"usage\":null","\"item\":null","\"error\":null"}
  for mask:=0;mask<16;mask++ {
   var known []string; for i,field:=range controls { if mask&(1<<i)!=0 { known=append(known,field) } }
   for _,count:=range []int{1,2,3,63,64,65,66,128} {
    if count<1+len(known) { continue }
    extras:=append(append([]string(nil),known...),contractExtras(count-1-len(known))...)
    positions:=[]int{0}
    if count==64 || count==65 { positions=make([]int,count); for i:=range positions { positions[i]=i } }
    for _,typePosition:=range positions {
     trace:=contractContext(kind,contractEnvelopeAt(kind,extras,typePosition))
     if count<=64 { contractAccept(t,trace) } else { contractReject(t,trace,taskoutcome.ErrJSON) }
    }
   }
  }
 }
}
func TestTaskVerifyDecodedKeyByteBoundaries(t *testing.T) {
 for _,kind:=range envelopeKinds {
  for _,key:=range []string{"",strings.Repeat("a",127),strings.Repeat("a",128),strings.Repeat("a",129),strings.Repeat("é",64),strings.Repeat("é",65),strings.Repeat("😀",32),strings.Repeat("😀",33)} {
   quoted,err:=json.Marshal(key); if err!=nil { t.Fatal("authored key") }
   for _,typePosition:=range []int{0,1} {
    trace:=contractContext(kind,contractEnvelopeAt(kind,[]string{string(quoted)+":null"},typePosition))
    if len(key)<=128 { contractAccept(t,trace) } else { contractReject(t,trace,taskoutcome.ErrJSON) }
   }
  }
  // Wire size exceeds128; decoded UTF-8 bytes remain exactly128.
  for _,encoded:=range []string{strings.Repeat("\\u0061",128),strings.Repeat("\\u00e9",64)} {
   for _,typePosition:=range []int{0,1} { contractAccept(t,contractContext(kind,contractEnvelopeAt(kind,[]string{"\""+encoded+"\":null"},typePosition))) }
  }
 }
}
func TestTaskVerifyPerEventResetAndUsage(t *testing.T) {
 trace:=contractEnvelope("thread.started",contractExtras(63))+"\n"+
  contractEnvelope("turn.started",contractExtras(63))+"\n"+
  "{\"type\":\"turn.completed\",\"future\":null,\"usage\":{\"input_tokens\":12,\"cached_input_tokens\":3,\"output_tokens\":5,\"reasoning_output_tokens\":1}}\n"+
  contractEnvelope("turn.started",contractExtras(63))+"\n"+
  "{\"type\":\"turn.completed\",\"future\":null,\"usage\":{\"input_tokens\":9,\"cached_input_tokens\":6,\"output_tokens\":7,\"reasoning_output_tokens\":2}}"
 s:=contractAccept(t,trace)
 if s.Events!=5 || s.ThreadsStarted!=1 || s.TurnsStarted!=2 || s.TurnsCompleted!=2 || !s.LifecycleComplete || !s.UsageComplete || s.TraceStatus!="completed_turns" { t.Fatal("envelope tracking escaped its event or changed lifecycle") }
 for _,check:=range []struct{field taskoutcome.FieldUsage; total int64}{{s.Usage.Input,21},{s.Usage.Cached,9},{s.Usage.Output,12},{s.Usage.Reasoning,3}} {
  if !check.field.Complete || check.field.Total==nil || *check.field.Total!=check.total || check.field.ObservedTotal==nil || *check.field.ObservedTotal!=check.total || check.field.ObservedTurns!=2 || check.field.MissingTurns!=0 { t.Fatal("usage arithmetic or subset meaning changed") }
 }
 partial:=contractAccept(t,contractThread+contractTurn+contractCompleted+contractTurn+"{\"type\":\"turn.completed\",\"future\":null,\"usage\":null}\n")
 if partial.UsageComplete || partial.Usage.Input.Total!=nil || partial.Usage.Input.ObservedTotal==nil || *partial.Usage.Input.ObservedTotal!=12 || partial.Usage.Input.ObservedTurns!=1 || partial.Usage.Input.MissingTurns!=1 { t.Fatal("partial observations erased or whole totals invented") }
}
func TestTaskVerifyOpaqueItemsAndRedaction(t *testing.T) {
 const marker="AUTHORED_OPAQUE_MARKER_NOT_REAL_DATA"
 // Unknown envelope names and JSON values must not enter the public Summary.
 // Exercise both discriminator orders and each lifecycle-valid event kind;
 // object/array markers also catch retention that skips scalar strings.
 unknowns:=[]string{
  "\"AUTHORED_UNKNOWN_KEY_NAME\":\"AUTHORED_UNKNOWN_STRING_VALUE\"",
  "\"AUTHORED_UNKNOWN_OBJECT_KEY\":{\"nested\":\"AUTHORED_UNKNOWN_OBJECT_VALUE\"}",
  "\"AUTHORED_UNKNOWN_ARRAY_KEY\":[\"AUTHORED_UNKNOWN_ARRAY_VALUE\",false,null,20261001531234567]",
  "\"AUTHORED_UNKNOWN_NUMERIC_KEY\":20261001531234567",
  "\"AUTHORED_UNKNOWN_BOOLEAN_KEY\":true",
  "\"AUTHORED_UNKNOWN_NULL_KEY\":null",
 }
 for _,kind:=range envelopeKinds {
  for _,typePosition:=range []int{0,len(unknowns)} {
   summary:=contractAccept(t,contractContext(kind,contractEnvelopeAt(kind,unknowns,typePosition)))
   encoded,err:=json.Marshal(summary)
   if err!=nil { t.Fatal("summary serialization failed") }
   for _,raw:=range []string{"AUTHORED_UNKNOWN_KEY_NAME","AUTHORED_UNKNOWN_STRING_VALUE","AUTHORED_UNKNOWN_OBJECT_KEY","AUTHORED_UNKNOWN_OBJECT_VALUE","AUTHORED_UNKNOWN_ARRAY_KEY","AUTHORED_UNKNOWN_ARRAY_VALUE","AUTHORED_UNKNOWN_NUMERIC_KEY","20261001531234567","AUTHORED_UNKNOWN_BOOLEAN_KEY","AUTHORED_UNKNOWN_NULL_KEY"} {
    if strings.Contains(string(encoded),raw) || contractRetainsMarker(reflect.ValueOf(summary),raw) { t.Fatal("unknown envelope name or JSON value retained in output or Summary memory") }
   }
  }
 }
 longKey,err:=json.Marshal(strings.Repeat("z",129)); if err!=nil { t.Fatal("authored nested key") }
 fields:=append(contractExtras(130),string(longKey)+":null","\"opaque\":\""+marker+"\"","\"opaque\":null")
 item:="{"+strings.Join(fields,",")+"}"
 trace:=contractThread+contractTurn+"{\"type\":\"item.completed\",\"item\":"+item+"}\n"+contractCompleted
 s:=contractAccept(t,trace)
 if !s.UsageComplete || s.DiscardedItemEvents!=1 || s.StartupErrorItems!=0 || s.Events!=4 { t.Fatal("opaque active-turn item acquired envelope rules") }
 serialized,err:=json.Marshal(s)
 if err!=nil || strings.Contains(string(serialized),marker) || strings.Contains(string(serialized),"future_129") || strings.Contains(string(serialized),strings.Repeat("z",129)) { t.Fatal("unknown names/contents retained") }
 // The narrow startup diagnostic keeps its existing nested-control rules;
 // unknown nested keys are not subject to the new envelope key count.
 startupFields:=append(contractExtras(130),string(longKey)+":null","\"opaque\":\""+marker+"\"","\"opaque\":null")
 startup:="{\"id\":\"AUTHORED_ITEM_ID\",\"type\":\"error\",\"message\":\""+marker+"\","+strings.Join(startupFields,",")+"}"
 startupTrace:=contractThread+"{\"type\":\"item.completed\",\"item\":"+startup+"}\n"+contractTurn+contractCompleted
 s=contractAccept(t,startupTrace)
 if s.StartupErrorItems!=1 || s.UsageComplete || s.TraceStatus!="observed_failure" || s.Usage.Input.ObservedTotal==nil || *s.Usage.Input.ObservedTotal!=12 || s.Usage.Input.Total!=nil { t.Fatal("startup diagnostic meaning changed") }
 serialized,err=json.Marshal(s)
 if err!=nil || strings.Contains(string(serialized),marker) || strings.Contains(string(serialized),"AUTHORED_ITEM_ID") { t.Fatal("startup content retained") }
}
func TestTaskVerifyExistingNestedControls(t *testing.T) {
 for _,tc:=range []struct{event string; code taskoutcome.ErrorCode}{
  {"{\"type\":\"turn.completed\",\"usage\":{\"future\":1,\"future\":2}}",taskoutcome.ErrDuplicate},
  {"{\"type\":\"turn.completed\",\"usage\":{\"input_tokens\":1,\"cached_input_tokens\":2}}",taskoutcome.ErrUsage},
  {"{\"type\":\"turn.completed\",\"usage\":{\"input_tokens\":-1}}",taskoutcome.ErrUsage},
  {"{\"type\":\"turn.completed\",\"usage\":{\"output_tokens\":1,\"reasoning_output_tokens\":2}}",taskoutcome.ErrUsage},
 } { contractReject(t,contractThread+contractTurn+tc.event,tc.code) }
 contractReject(t,contractThread+"{\"type\":\"item.completed\",\"item\":{\"type\":\"error\",\"id\":\"a\",\"id\":\"b\",\"message\":\"authored\"}}",taskoutcome.ErrDuplicate)
 contractReject(t,contractThread+contractTurn+"{\"type\":\"turn.started\"}",taskoutcome.ErrLifecycle)
 // Existing usage-object limits keep ErrUsage, distinct from the new event
 // envelope's ErrJSON. These apply even when usage is not counted for a turn.
 controls:=[]string{"\"input_tokens\":null","\"cached_input_tokens\":null","\"output_tokens\":null","\"reasoning_output_tokens\":null"}
 for _,kind:=range envelopeKinds {
  for mask:=0;mask<16;mask++ {
   var known []string; for i,field:=range controls { if mask&(1<<i)!=0 { known=append(known,field) } }
   for _,count:=range []int{64,65} {
    fields:=append(append([]string(nil),known...),contractExtras(count-len(known))...)
    event:=contractEnvelope(kind,[]string{"\"usage\":{"+strings.Join(fields,",")+"}"})
    trace:=contractContext(kind,event)
    if count==64 { contractAccept(t,trace) } else { contractReject(t,trace,taskoutcome.ErrUsage) }
   }
  }
  for _,key:=range []string{strings.Repeat("a",128),strings.Repeat("a",129),strings.Repeat("é",64),strings.Repeat("é",65)} {
   quoted,err:=json.Marshal(key); if err!=nil { t.Fatal("authored usage key") }
   trace:=contractContext(kind,contractEnvelope(kind,[]string{"\"usage\":{"+string(quoted)+":null}"}))
   if len(key)<=128 { contractAccept(t,trace) } else { contractReject(t,trace,taskoutcome.ErrUsage) }
  }
  for _,tc:=range []struct{encoded string; valid bool}{{strings.Repeat("\\u0061",128),true},{strings.Repeat("\\u0061",129),false},{strings.Repeat("\\u00e9",64),true},{strings.Repeat("\\u00e9",65),false}} {
   trace:=contractContext(kind,contractEnvelope(kind,[]string{"\"usage\":{\""+tc.encoded+"\":null}"}))
   if tc.valid { contractAccept(t,trace) } else { contractReject(t,trace,taskoutcome.ErrUsage) }
  }
 }
}
`
