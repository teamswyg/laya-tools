# 71 v4 — 첫 FP32 학습 준비와 중단 기록

작은 주장 모델이 함수 설명 후보를 먼저 확인할 순서를 제안하도록 첫 FP32 학습 1회를 준비했다. 현재는 Go 드라이버 컴파일과 합성 검사까지이며 원본 fit·Features·Project·Baselines·Encode·Decode·NLL 호출은 0이다. v4 템플릿의 `frozen`과 준비 전제는 false이고 전체76 투영 SHA가 비어 있어 실행할 수 없다. 부모가 전체76의 출처·supervision·mask·그룹 역할과 투영을 고정한 뒤 실제 실행한다.

입력 목표는 기존72/216과 출처68의 관련4/10을 합친76/226, known38·no_answer17·unknown21, 전체그룹19·known-containing18이다. 전체76 역할은 새로 배정하며 과거66 역할을 복사하지 않는다. 출처68의 요청4개는 연결된2개 가족이고 human-only 저작이나 광범위 일반화를 증명하지 않는다. 출처와 학습 가능성 판단의 의미는 부모의 동결·검토 책임이며 이 드라이버의 SHA 확인이나 caller flag만으로 증명되지 않는다.

첫 설정은 기존8192차원 `pairlearn.FitWithTrace`, FP32, seed1729, 최대50 epoch, batch128, 학습률0.1, L2 0.0001이다. 기존 binary unit loss와 명시적0/1 가중치를 유지한다. known의 가중치0 행도 보존하고 unknown은 nullable 감사 기록으로만 남긴다. calibration은 fit·epoch 선택·validation 효용에 사용하지 않는다. 효용 A는 모든 known/no_answer 부모의 모든 후보를 보존하며 평가에는 validation 역할만 사용한다. 완전 loss-eligible 부모의 B는 진단이다.

동일 validation A에서 `fixed_order → bm25 → lexical_ordered → narrow_rule` 네 비교 방식의 총 확인 횟수를 비교한다. 가장 적은 방식 하나를 선택하고 동률은 이 고정 배열 순서를 따른다. Top1/Top3는 동률 기준이 아니다. oracle 하한 대비 개선 여지가 기존5%보다 작으면 fit0회로 끝난다. fit을 실행하면 validation NLL이 처음 최소가 된 epoch를 선택하고, 같은 validation에서 확인 횟수5% 개선과 Top1/Top3 보존을 검사한다. epoch 선택과 효용이 같은 validation을 사용하므로 bounded development 결과이며 holdout 일반화나 실제 LLM 비용 절감 관측이 아니다. 원래8fit/16asset 상한 아래 이 드라이버는 primary1회만 실행하고 자동 재시도하지 않는다.

독립 v3 검토에서 확인된 세 문제를 v4에서 좁게 수정했다. 전역 flag 출력 대신 별도 FlagSet과 고정 오류만 사용해 잘못된 인수 원문을 stderr에 내보내지 않는다. 계획 파일은 regular-file·1MiB 크기를 먼저 확인하고 제한된 읽기만 하므로 전체 파일을 먼저 할당하지 않는다. 실행 결과와 장부는 최초 O_EXCL 생성 뒤 매 checkpoint마다 갱신·fsync하고 각 validation 부모의 비교 callback과 Fit·Encode·Decode·NLL 호출 직전에 예약 카운터를 저장한다. 반환 카운터는 실제 callback 반환 후 증가하며 이후 checkpoint에 남는다.

`Attempts`는 호출 직전의 **예약된 dispatch 의도**다. checkpoint 실패라면 호출0회여도 예약 카운터가1일 수 있다. 외부 강제 종료 시 마지막으로 완전히 저장된 snapshot만 확인 가능하며 그 이후 in-flight 호출의 반환·trainer 내부 epoch는 관측하지 못한다. 파일 갱신 중 종료되면 JSON이 부분 파일일 수 있으므로 외부 제어기는 raw SHA/bytes와 OS 결과를 보존하고 파싱 실패를 실패 장부로 기록해야 한다. 이를 내부 trainer 완전 추적이나 강제 종료 후 정확한 모든 호출 횟수 보장으로 설명하지 않는다. 현재 준비는 실제 trainer 강제 종료 실험을 하지 않았다.

저장 모델은 헤더24B와 float32 계수32768B, 합계32792B다. decode 계수 배열은 float64 65536B이고 shadow·gradient·loss 누적과 reference 점수도 float64다. 기존 Result.Weights에는 이미 `json:"-"`가 있다. 명시적 FitSummary DTO는 결과 계약을 좁힌 것이며 기존 유출 결함 수정이 아니다. 모델 계수와 바이너리는 private asset으로만 보존하고 Git·HF 게시를 자동 실행하지 않는다.

CPU1·동시 실제fit1개·Go soft heap256MiB·전체300초·payload/new assets64MiB를 유지한다. 힙 목표와 직렬화 후 크기 확인은 OS RSS 하드 상한이 아니다. 외부 제어기가 프로세스 전체 CPU/RSS·실행 시간을 측정하고 시간 초과를 종료해야 한다. 현재 모델을 추가로 GPU에 올리지 않는다.

사람과 에이전트가 같은 CLI에 계획 SHA·source root·출력 경로를 전달한다. 저장소 source/plan56은 source root 기준, 나머지 입력은 계획 디렉터리 기준 상대 경로다. 공개 포팅 전 private module replace와 로컬 경로를 제거한다.

```sh
GOMAXPROCS=1 GOMEMLIMIT=256MiB ./riido71fit \
  --plan fit-plan.json --plan-sha256 "$FROZEN_PLAN_SHA" \
  --source-root . --output-dir .cache/primary-fit-attempt-1
```

v1–v3 source/binary/template/receipt는 수정하지 않고 보존했다. 수정 중 중복 선언 문법 오류로 gofmt1과 race4가 실패했으며 원본 호출0이다. 이를 고친 뒤 race5의10개 test·8개 subcase, vet4·compile4가 통과했다. 합성 callback은 성공한 suite1/2/3에서 각9회, suite5에서10회, 총37회이며 실제 trainer fit은0이다. 실패한 suite4는 컴파일 단계에서 중단되어 callback0회다. 별도 모델/API/benchmark0은 AI-assisted 준비와 협업 비용이0이라는 뜻이 아니며 그 비용은 측정하지 않았다. 새76 역할·투영·준비성 검토의 실제 결과와 외부 실행 제어기를 고정하는 것이 다음 단계다.
