# 66 실제 역할 배정 결과의 읽기 전용 검토

동결된 65 입력과 seed `1729`의 구성·순서 계산이 실제 66 결과와 일치한다. 17개 whole group의 membership SHA/encoded byte 수, 알려진 부모를 포함한 16개 그룹의 seed 순서, 72개 부모의 역할·저장 truth·후보 수를 별도 stdlib 계산으로 확인했다. 현재 결과와 입력 연결에서 중대한 불일치는 발견하지 못했다. 원래 Verify/Assign/MembershipDigest 함수나 실행 파일을 재호출하지 않았다.

이 검토자는 66 CLI 저자가 아니지만 이전 62 public roleplan 모듈의 저자이며 65/66 사전 검토에 노출됐다. 비맹검 검토이고 독립 원천 제작, compiler/binary 신뢰, 의미론적 source 승인이나 fit 준비 완료를 증명하지 않는다. root reservation receipt에 기록된 CI/merge/head를 확인했으며, 이번 검토에서 GitHub API·Git 또는 CI를 별도로 다시 실행하지 않았다.

| 확인한 단위 | train | validation | calibration |
| --- | ---: | ---: | ---: |
| known-containing whole group | 10 | 3 | 3 |
| unknown-only group를 포함한 보존 group | 11 | 3 | 3 |
| 모든 부모 | 44 | 16 | 12 |
| 모든 후보 위치 | 132 | 48 | 36 |
| unknown 부모 | 14 | 4 | 3 |

known-containing16개에 원래 3:1:1 largest-fraction 계산을 적용하면 10/3/3이다. 각 membership을 정렬된 원래 member와 From/Kind/To 관계로 만들고 UTF8 length를 64-bit big-endian으로 붙이는 선언된 encoding을 별도로 계산했다. 17개 digest와 encoded byte 수가 입력 및 결과의 raw32 digest와 일치한다. 순서는 length-prefixed domain·ASCII seed·raw32 membership bytes의 SHA256이다. hex 문자열로 바꾸지 않았고 order hash, membership digest, original group ID의 정해진 tie 순서를 유지했다. Known 순서의 앞10·다음3·마지막3이 결과 역할과 정확히 일치한다. Unknown-only group64는 이 16개 순서·라벨 하한에 넣지 않았고, 완전한 train provenance group으로 보존됐다. 새 seed 검색이나 group split/merge는 하지 않았다.

원래 알려진 truth34·no_answer17·unknown21이 그대로다. exact original parent ID·저장 outcome·후보 수와 whole group reference를 확인했고 각 group의 모든 부모가 같은 역할을 따른다. 72개 parent index는 빠짐없이 한 번씩 나타난다. 후보216개, 구성 member453개와 관계1,922개 및 group 간 member 비중복이 보존된다. Parent와 outcome의 전체 JSON subtree를 원래 SHA로 묶었으므로 기존 다중 acceptable-index·unknown/null·빈 배열을 다시 해석하거나 바꾸지 않는다.

고정 plan의 20개 input/evidence/support/implementation descriptor, 총4,408,642bytes를 exact SHA로 확인했다. 65 snapshot에서 2,934개의 artifact/pointer/value reference를 각 원래 pinned JSON에 대조했다. 값 SHA는 Go `json.Marshal(value)` compact/no-LF 규칙이며, decoded number를 보존했다. 이 확인은 저장된 관계·참조의 정합성이고 누락된 외부 관계가 없다는 증명은 아니다. 현재8개 input과 기존 source/review bindings를 검증했으며 원본 source/API·AST·formatter·SourcePins·Bind·Generate를 실행하지 않았다.

저장 truth의 answerable/no_answer coverage6개는 각각10/3/3을 만족한다. 이것은 저장 truth coverage이며 이후 감독 mask를 적용한 loss-eligible group 수가 아니다. 이번 검토는 mask를 읽어 role과 합치거나 fit gate를 재산출하지 않았다. 기존 unknown은 label-ineligible이고 source/text 불확실성, 영어 유한 Go claim synthetic pipeline의 다양성 한계도 남는다. 개발 validation/calibration 이름이 과거 관측 데이터를 새로운 맹검 또는 protected-final로 바꾸지 않는다. Graph/role은 scorer feature가 아니며 source/diversity/semantic/training readiness는 false다. 새로운 수치 gate를 추가하지 않았다.

결과·root reservation·root execution ledger·worker attempt ledger의 plan/input/binary/seed pin과 상태가 맞다. root의 실제 role process1회·retry0·exit0, membership-set verification1회 및 Assign1회가 기록돼 있다. 내부 카운터의 validation17, order hash16, coverage6, assignment17도 재계산한 수량과 일치한다. 결과의 `metadata_only=false`는 Prepare-only 모드를 벗어나 실제 역할을 배정했다는 의미다. source 추론·모델 학습이 실행됐다는 뜻이 아니다. 특징·새 라벨/loss mask·fit·model·paid·protected-final은0이다.

root의 raw OS 출력과 resource 장부는 단위·값이 일치한다. OS real0.41초, controller wall0.413666042초, OS user0.02초/sys0.00초다. 최대 RSS17,448,960bytes=16.640625MiB이고 peak footprint14,746,128bytes는 별도 측정이다. CPU 시간은 소수 둘째 자리 출력이며 0.00이 CPU 작업0을 의미하지 않는다. 한 번의 cold role metadata process에는 startup, 입력 확인, 원래 membership verification/assignment, 출력이 포함된다. 빌드·CI·이번 감사 helper/test·모델 추론 성능은 포함되지 않는다. Go heap soft limit 및 RSS observation cap256MiB는 hard RSS 제한이 아니며 GPU·모델 품질·실용 hint 성능이나 통계적 처리속도 개선을 측정한 결과가 아니다.

이번 stdlib-only metadata helper1회가 성공했다. bounded read27회/4,488,107bytes, 합성 테스트4개를 담은 uncached race1회와 vet1회가 통과했다. helper/test/vet/필수 pin 실패는0이다. 탐색 중 JSON shape를 잘못 예상한 jq sub-operation2건은 wrapper exit0에도 실패로 남겼다. 출력 예산 잘림4건 및 null guessed field는 실제 key/projection과 pinned helper 검증으로 복구했으며 입력 누락으로 해석하지 않았다. 사전 66 QA의 별도 탐색 실패1건도 이전 장부에 그대로 있다. 이번 검토의 original Verify/Assign/MembershipDigest/binary·source·features·roles API·fit·model·paid·보호 final·공유 수정·Git·게시 실행은 모두0이다. 일반 AI 협업 비용은 측정하지 않았다.
