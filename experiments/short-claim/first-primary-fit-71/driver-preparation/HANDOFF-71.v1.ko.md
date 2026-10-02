# 71 — 작은 FP32 주장 모델의 첫 학습 준비

목표는 공개 함수 설명 후보를 먼저 살펴볼 순서를 제안하는 작은 모델 한 개를 실제로 학습하는 것이다. 문장을 생성하거나 코드의 정답을 증명하지 않는다. 지금 산출물은 컴파일된 Go 드라이버와 합성 검사이며, 원본 데이터의 학습·추론·투영은 아직 0회다. 실행 계획의 `frozen`과 준비 전제는 모두 false이고, 76개 투영 결과·전제 기록의 정확한 SHA가 아직 비어 있으므로 현재 템플릿은 실행을 거절한다.

기존 `hintlearn.Dimension`은 정확히 8192다. 첫 실험은 `pairlearn.FitWithTrace`를 사용하여 FP32, seed1729, 최대 50 epoch, batch128, 학습률0.1, L2 0.0001을 고정한다. 기존 unit binary loss와 명시적 0/1 가중치를 그대로 쓰며 새 class balance·listwise loss·가중치 조절은 넣지 않는다. FP32라는 이름은 채점 계수를 float32로 반올림한다는 뜻이다. shadow 가중치·gradient·loss 누적과 reference 채점은 float64다. 저장 모델은 헤더24B와 계수32768B, 합계 **32792B**이며, Decode된 계수 배열은 float64 **65536B**다. 모델 저장 크기와 전체 프로세스 메모리를 혼동하지 않는다.

다음 입력은 원래72개 요청·216개 후보와 출처68의 관련 요청4개·후보10개를 합친 76/226이다. 기존 known34·no_answer17·unknown21에 좁은 유한 계약의 proposed known4개가 더해지는 방향이며, 새 supervision과 실제 가중치는 부모가 별도로 고정해야 한다. 전체 그룹은19개, known을 포함하는 그룹은18개다. 이전66 역할은 원래72개에 대한 역사 기록이므로 복사하지 않고, 고정 seed1729로 전체76 역할을 별도로 배정한다. 관련 요청4개가 독립 그룹4개라는 뜻은 아니며, 새 출처2개도 광범위 일반화나 human-only 저작을 증명하지 않는다.

드라이버는 부모가 고정한 투영을 읽고 배열·참조·nullable 정답·전체 그룹 역할·후보 순서·가중치와 분모를 확인한다. 원래 known/no_answer 55개 전체가 효용 관점 A의 대상이며 실제 평가에는 그중 validation 역할만 사용한다. 모든 알려진 후보를 유지하므로 학습 가중치0 후보도 채점 순서와 평가에서 사라지지 않는다. unknown21은 nullable 감사 기록으로 보존하고 fit과 정답 효용에서 제외한다. calibration은 fit과 epoch 선택, 이번 validation 효용에서 제외한다. 완전히 loss-eligible인 부모만의 B는 별도 진단이며 A를 대체하지 않는다. train/validation 행 수는 새76 역할에 따라 결정되므로 원래72개 때의 90/36을 복사하지 않는다.

학습 전에 같은 validation A에서 `fixed_order`, `bm25`, `lexical_ordered`, `narrow_rule` 네 방식을 비교한다. 여기서 비용은 실제 verifier 실행 시간이 아니라 첫 acceptable 후보까지 필요한 확인 횟수의 대리 지표다. no_answer는 모든 후보 수를 확인 횟수로 센다. 가장 적은 총 확인 횟수의 방식 하나를 고정 비교 대상으로 삼고 동률은 위 배열 순서를 따른다. Top1/Top3는 동률 선택 기준이 아니다. Oracle 하한 대비 가능한 개선 여지가 기존 기준5%보다 작으면 학습0회로 종료하고 그 사실을 기록한다.

개선 여지가 있으면 primary fit을 **1회**만 호출한다. 50개 epoch의 유한 NLL trace를 남기고 validation NLL이 처음 최소가 된 epoch를 고른다. 원래 8회 fit·16개 모델 asset 상한을 유지하지만 이 드라이버는 최초 primary 1회만 수행하며 자동 재시도나 sweep을 하지 않는다. fit 오류·panic·불완전 trace·잘못된 epoch는 실패 장부로 남긴다. 성공 계수는 private FP32 asset으로 1회 Encode/Decode하고 train/validation NLL 2회를 재확인한다. 이번 준비의 합성 검사는 이 실제 trainer와 Encode/Decode를 호출하지 않는다.

학습 뒤 기존 sparse feature를 그대로 재사용하여 validation의 모든 알려진 후보를 채점한다. source ID·정답·역할·검토 결과는 feature 입력이 아니다. 새 `Features`·`Prepare`·`Project` 호출 없이 고정된 배열을 읽는다. 가장 좋은 고정 방식보다 확인 횟수가5% 이상 감소하고 같은 A의 Top1/Top3가 떨어지지 않았을 때에만 bounded development utility를 통과했다고 기록한다. epoch 선택과 효용 평가가 같은 validation을 사용하므로 최종 holdout 검증이나 통계적 일반화 보장은 아니다. 실제 LLM 토큰 비용 절감을 관측한 것도 아니다.

기존 `pairlearn.Result.Weights`에는 이미 `json:"-"`가 있다. 추가한 `FitSummary` DTO는 기존 정보 유출 결함을 고친 것이 아니라 공개 결과 계약을 config·선택 epoch·NLL로 명시적으로 좁힌 것이다. 공개할 수 있는 결과에는 trace50·확인 횟수·Top1/Top3·모델 SHA/bytes/ref만 남기며 모델 계수와 바이너리는 Git에 올리지 않는다. 효용이 없는 성공 fit도 결과를 기록하고 모델은 private에 보존한다. HF 게시·CI·라이선스·재현성과 자원 검증은 부모의 별도 기존 배포 절차이며 이 드라이버가 자동으로 승인하거나 게시하지 않는다.

메모리와 저장장치 부담을 줄이기 위해 CPU1, 동시 실제 fit1개, Go 힙 소프트 목표256 MiB, 전체 실행300초, payload/new assets64 MiB 범위로 시작한다. 외부 제어기가 전체 프로세스 CPU/RSS를 측정하고 시간 초과를 종료해야 한다. Go 소프트 힙 목표와 출력 직렬화 후 크기 검사는 RSS의 하드 상한이 아니다. 전제 기록에는 전체76의 출처/검토/mask 경계가 확인된 bounded development 범위를 고정하며, 학습 가능성 판단을 드라이버의 파일 SHA 검사만으로 증명했다고 말하지 않는다.

최종 사용 형태는 사람과 에이전트가 같은 Go CLI에 입력 계획·SHA·출력 경로를 전달하는 것이다. 저장소 소스와 원래 plan56은 `--source-root` 기준 상대 경로, 드라이버와 나머지 고정 입력은 계획 디렉터리 기준 상대 경로로 읽는다. 아래는 **private 준비 실행 방식**이며 현재 false 템플릿으로 실제 학습은 실행되지 않는다. 공개 포팅 시 로컬 replace와 private 경로를 제거해야 한다.

```sh
GOMAXPROCS=1 GOMEMLIMIT=256MiB ./riido71fit \
  --plan fit-plan.json --plan-sha256 "$FROZEN_PLAN_SHA" \
  --source-root . --output-dir .cache/primary-fit-attempt-1
```

직접 필요한 다음 단계는 전체76 supervision/출처 경계와 전체 그룹 역할을 부모가 고정하고, 같은 계약으로76개를 한 번 투영한 뒤, 결과 SHA와 기존 자원/효용 정책을 이 계획에 채우는 것이다. 그 다음 독립 정적 검토와 외부 실행 제어기 아래 첫 fit1회를 실행한다. 새로운 숫자 gate나 사용자 승인을 요구하지 않는다. 유용한 FP32 기준이 확보되면 기존 `pairlearn`의 ternary_ste를 후속 형제로 검토할 수 있다. `ternarytrain.Train`의 1024차원 Laya feature/다중 클래스 head는 이 8192차원 binary feature 계약과 달라 그대로 대체하지 않는다.

준비에는 AI 에이전트가 참여했고 협업 비용은 측정하지 않았다. 별도 모델/유료 API/benchmark 프로세스0회는 이 대화의 AI 사용이나 비용0을 뜻하지 않는다. 합성 검사3회·vet3회·컴파일3회는 모두 성공했다. 원본 trainer 호출과 합성 trainer fit은 모두0이며, 합성 반환값 callback은 각 검사 실행에서9회, 총27회 호출했다. 최초 작업 디렉터리 생성 실패1회와 잘못 추측한 파일 glob 읽기2회는 실행 전 준비 실패로 보존한다. v1/v2 소스·실행 파일·템플릿은 역사 자료로 보존한다.
