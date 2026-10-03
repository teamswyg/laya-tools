# Pure saved-JSON comparison. No original code, model, replay or labels.
# Inputs are supplied as --slurpfile fixture and --slurpfile worker.
# uint64/int64 decimal amounts remain strings: no floating conversion.
def predicate($field; $got; $want; $known):
  {field:$field, got:$got,
   state:(if $known|not then "unknown"
          elif $got==$want then "satisfied" else "known_mismatch" end)};
def compare($o; $w):
  [predicate("error_present"; $o.error_present; $w.error_present;
     (($o|has("error_present")) and $o.error_present!=null)),
   (if $w.required_error_kind!=null then
      predicate("required_error_kind"; $o.error_kind; $w.required_error_kind;
        ($o.error_present==false or
         ($o.error_present==true and $o.error_kind!=null and
          $o.error_kind!="unclassified_error"))) else empty end),
   (if $w.required_error_type!=null then
      predicate("required_error_type"; $o.error_type; $w.required_error_type;
        ($o.error_present==false or
         ($o.error_present==true and $o.error_type!=null))) else empty end),
   (("int64_decimal","uint64_decimal","name_after") as $key |
      if $w[$key]!=null then
        predicate($key; $o[$key]; $w[$key];
          (($o|has($key)) and $o[$key]!=null)) else empty end),
   (if $w.conflicting_keys!=null then
      predicate("conflicting_keys"; $o.conflicting_keys; $w.conflicting_keys;
        ($o.returned_normally and ($o|has("conflicting_keys"))))
      else empty end),
   (("live_after","snapshot_after","legacy_after") as $key |
      if $w[$key]!=null then
        predicate($key; $o[$key]; $w[$key];
          (($o|has($key)) and $o[$key]!=null)) else empty end)] as $checks |
  {state:(if ($o.returned_normally!=true or $o.panicked or
              $o.error_method_panicked or $o.unknown_reason!="") then "unknown"
          elif any($checks[]; .state=="known_mismatch") then "known_mismatch"
          elif any($checks[]; .state=="unknown") then "unknown"
          else "satisfied" end),
   checks:$checks,
   mismatched_fields:[$checks[]|select(.state=="known_mismatch")|.field],
   unknown_fields:[$checks[]|select(.state=="unknown")|.field],
   diagnostic:{returned_normally:$o.returned_normally,
     error_present:$o.error_present,error_type:$o.error_type,
     error_kind:$o.error_kind,error_text_observed:$o.error_text_observed,
     panicked:$o.panicked,error_method_panicked:$o.error_method_panicked,
     unknown_reason:$o.unknown_reason,
     original_api_calls_measured:$o.original_api_calls_measured,
     startup_init_calls_measured:$o.startup_init_calls_measured}};
$fixture[0] as $f | $worker[0] as $v |
if ($f.requests|length)!=4 or ($v.rows|length)!=57 or
   ([ $f.requests[].fixtures[] ]|length)!=19 or
   $v.header.fixture_sha256!="1ee48c079f85f9e4c3051b53c7f51a3e2d9c0eadc8dd7599f34afbcf46ef42ba" or
   any($v.rows[]; .attempted!=true or .truth!=null or .label!=null or
       .role!=null or .weight!=null or .state!="observed_not_compared") or
   any($f.requests[].fixtures[]; .Got!=null or .truth!=null or
       .label!=null or .role!=null or .weight!=null)
then error("frozen_source_shape_or_null_supervision") else . end |
[ $f.requests[] as $r | $r.fixtures[] as $ff |
  {request_index:$r.request_index,original_ordinal:$r.original_ordinal,
   parent_id:$r.parent_id,fixture_ordinal:$ff.fixture_ordinal,
   case_id:$ff.fixture.case_id,input:$ff.fixture,
   wanted_contract:$ff.wanted_contract,
   candidates:[range(0;3) as $ci |
     [$v.rows[]|select(.dispatch.RequestIndex==$r.request_index and
       .dispatch.OriginalOrdinal==$r.original_ordinal and
       .dispatch.CandidateIndex==$ci and
       .dispatch.FixtureIndex==$ff.fixture.fixture_index and
       .dispatch.FixtureOrdinal==$ff.fixture_ordinal)] as $rows |
     if ($rows|length)!=1 then error("dispatch_not_unique") else
       {candidate_index:$ci,candidate_id:$r.candidates[$ci].candidate_id} +
       compare($rows[0].observation;$ff.wanted_contract) end]} ] as $cases |
[ $f.requests[] as $r | range(0;3) as $ci |
  [$cases[]|select(.request_index==$r.request_index)|
    .candidates[]|select(.candidate_index==$ci)] as $cs |
  {request_index:$r.request_index,original_ordinal:$r.original_ordinal,
   parent_id:$r.parent_id,candidate_index:$ci,
   candidate_id:$r.candidates[$ci].candidate_id,
   finite_inputs:$r.fixture_count,
   satisfied:[$cs[]|select(.state=="satisfied")]|length,
   known_mismatch:[$cs[]|select(.state=="known_mismatch")]|length,
   unknown:[$cs[]|select(.state=="unknown")]|length,
   finite_satisfaction:(if all($cs[];.state=="satisfied") then "all_satisfied"
     elif any($cs[];.state=="known_mismatch") then "has_known_counterexample"
     else "unknown" end),label:null,role:null,weight:null} ] as $summaries |
{schema:"riidolaya-four-selector-saved-full-Want-comparison-v1",
 state:"independent_nonblind_saved_result_comparison_not_qualification",
 source_pins:{fixture:{bytes:47256,sha256:"1ee48c079f85f9e4c3051b53c7f51a3e2d9c0eadc8dd7599f34afbcf46ef42ba"},
   worker_result:{bytes:53036,sha256:"fabad6186f4f0213ce8449a2c0cceee454267ef6951a94c9f7a61b6b642c89ed"},
   outside_result:{bytes:3179,sha256:"fb7e59bc648406ec82e2bb5d84a3ca2fdfa16448562ef0696667ca73b50df046"}},
 scope:{requests:4,finite_inputs:19,candidate_positions:12,
   saved_candidate_observations:57,compared_predicates:"Every nonnull optional Want plus mandatory error_present; exact strings and MapImage including nil/entries shape",
   optional_Want_null:"No predicate, not an unknown observation or an invented equality requirement",
   native_or_init_call_counts:"Uninstrumented null, excluded from satisfaction predicates",
   error_text:"Diagnostic observation; no error-text equality predicate was requested",
   labels_created:0,roles_assigned:0,new_parent_requests:0,replays:0,model_calls:0},
 summary:{satisfied:[$cases[].candidates[]|select(.state=="satisfied")]|length,
   known_mismatch:[$cases[].candidates[]|select(.state=="known_mismatch")]|length,
   unknown:[$cases[].candidates[]|select(.state=="unknown")]|length,
   all_satisfied_candidates:[$summaries[]|select(.finite_satisfaction=="all_satisfied")]|length,
   candidates_with_known_counterexample:[$summaries[]|select(.finite_satisfaction=="has_known_counterexample")]|length},
 candidate_summary:$summaries,cases:$cases}
