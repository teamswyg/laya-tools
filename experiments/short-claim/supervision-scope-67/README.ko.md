# 원래 정답을 보존하는 감독 범위 점검67

`riido-supervision`은 이미 저장한 요청·후보·정답과 의미 검토를 연결하여, **어떤 후보를 원래 목표의 학습 loss에 사용할 수 있는지 제안하는 Go 명령어**다. 모델을 학습하거나 추론하지 않는다. 검토가 부족한 후보를 새로운 오답으로 만들지 않고, 원래 정답과 전체 후보 집합을 유지한다.

목표는 유한 영어 요청에 적힌 **전체 명시 조건**과 후보 설명을 연결하는 작은 힌트 모델이다. 예를 들어 정답이 여러 개인 요청은 허용 후보 전체를 보존하고, `no_answer`는 원래의 모든 음성 후보를 유지한다. `unknown`은 라벨을 만들지 않는다. 설명이 원래 요청의 일부 조건만 다룬다면 기존 함수 정답을 다른 목표의 정답으로 빌려 쓰지 않는다.

## 첫 실제 점검 결과

부모 실행 장부의 첫 metadata 점검은 1회 성공, 실패0, 재시도0이었다. 아래 수치는 변경하지 않은 [원본 결과](results-67.json)와 [완료 장부](ROOT-ACTUAL-EXECUTION-LEDGER-67.v1.json)의 집계다.

|항목|원래 집합 또는 제안 결과|
|---|---:|
|부모 요청 / 후보 위치 / 전체 그룹|72 / 216 / 17|
|known 또는 no_answer 부모 / unknown 부모|51 / 21|
|원래 entry / 보조 감사 entry|738 / 432|
|핀 검증 파일 / 검증 bytes / 검증 review records|21 / 8,586,507 / 1,170|
|감독 적격 양성 후보 / 음성 후보|35 / 95|
|라벨은 보존하되 mask=false인 known 후보|23|
|unknown 후보, nullable 라벨|63|
|적격 후보가 하나 이상인 그룹|15|
|적격 양성 후보가 하나 이상인 그룹|12|

`35 + 95 + 23 + 63 = 216`으로 모든 후보 위치를 보존했다. 그룹52의 known 후보9개는 모두 mask=false이며 삭제하지 않았다. 원래 known을 포함한 그룹은16개다. “적격 후보가 있는 그룹15개”와 같은 수치가 아니다. 양성 포함 그룹12개는 서술 통계이며 새 학습 하한이 아니다.

이것은 **감독 범위의 metadata 제안**이다. 실제 특징 투영·역할 배정·seed 선택·fit·모델/유료 호출·원래 source behavior API·보호 최종 읽기는 모두0이다. `training_ready=false`, `authoring_diversity_cleared=false`를 유지한다. 역할별 적격 범위와 기존 같은 scope의 효용/headroom 계산은 아직 수행하지 않았다. 이 결과로 mask를 완화하거나 유리한 seed를 찾지 않는다.

## 사용법

Go1.27.1을 사용하는 저장소 루트에서 실행한다. 원본 JSON21개가 각각 고정한 bytes/SHA와 일치해야 한다. 이 페이지 아래의 archive 파일만으로 입력을 대신할 수 없다.

```sh
go run ./cmd/riido-supervision \
  --input-root . \
  --output supervision-67-new.json \
  --plan-sha256 c9a07943f8047b69e32d93354e812ca5b158b1dc7476264bd5b8505a86242408
```

바이너리로 사용하려면 다음과 같이 빌드할 수 있다. Python, 네트워크 호출, 모델 다운로드가 실행 요건에 없다.

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o riido-supervision ./cmd/riido-supervision
./riido-supervision --help
```

`--output`은 새 파일이어야 한다. 기존 파일은 덮어쓰지 않으며, 실패 뒤 자동 재시도하지 않는다. 입력은 read-only root 안의 regular file만 읽는다. 파일당2MiB, 합계16MiB 상한을 적용한다. 이 상한은 읽은 payload 상한이고 프로세스 RSS나 Go heap 상한은 아니다.

사람은 고정 안내와 실패 코드를 읽을 수 있고, 에이전트는 종료 코드와 JSON envelope를 읽을 수 있다. 성공은 종료0과 `passed_metadata_proposed_supervision_only`다. 검증 실패는 종료1과 `failed` envelope이며, 빈 report와 실패 지점까지의 검증 장부를 남긴다. 잘못된 플래그나 필수 인수는 종료2다. 출력 파일을 예약하거나 저장하지 못하면 envelope가 완성되지 않을 수 있다. 오류는 공급한 경로나 플래그 값을 그대로 출력하지 않는다.

## 제안 mask의 규칙

부모 요청의 contract, prototype의 contract observation, 후보의 source closure와 caption fidelity가 모두 `consistent_with_scoped_evidence`여야 한다. 원래 label1은 후보 coverage도 consistent일 때, label0은 coverage가 `contradicts_scoped_evidence`일 때만 mask=true다. omission/uncertain/pending이나 저장 정답과 맞지 않는 coverage는 mask=false다.

원래 acceptable 집합·후보 순서·평가 분모·전체 그룹은 그대로다. known/no_answer 후보의0/1 라벨을 다시 작성하지 않고, unknown 라벨은 null이다. 보조 observation/boundary 기록은 감사 자료이며 원래 coverage나 라벨을 덮어쓰지 않는다. source closure만 확인한 기록을 caption fidelity 승인으로 세지 않는다. 공개 source 이름과 digest/pointer는 추적용 옆자료이며 runtime 특징이나 전체 승인 scalar가 아니다.

단계67은 fit용 행을 생성하지 않는다. 나중 투영할 때 mask=false인 known 후보를 weight0으로 유지하고, unknown은 audit에만 유지해야 한다는 계약을 보고한다. 같은 그룹의 일부 후보만 제거하여 역할이나 학습 하한을 채우지 않는다. 기존56계획의 역할9/3/3과 같은 scope의 효용5%를 바꾸거나, 매 실험마다2400개나 새로운 다양성 수치를 요구하지 않는다.

## 속도와 한계

원래 private 바이너리의 단일 cold metadata 프로세스 측정은 real0.71초, user CPU0.13초, system CPU0.01초, peak RSS27,508,736bytes였다. 별도의 peak memory footprint는24,887,824bytes다. 두 메모리 수치는 같은 지표가 아니다. 원본 진단 로그는 비공개로 두고 SHA만 완료 장부에 기록했다.

이 수치는 모델 추론·모델 메모리·GPU·Go heap·warm 라우터 속도나 일반적인 성능 분포를 설명하지 않는다. 공개 포트는 빌드와 합성 테스트만 수행했으며 원본 입력 실행0이다. 실제1회 점검은 원래 private 바이너리로 수행했다.

일반 join/count에는 배열·slice와 정렬을 사용하고 lock을 추가하지 않았다. 개별 검토 record의 기존 canonical JSON SHA를 재현할 때만 기존 hash recipe의 map 표현을 사용한다. 이번 단계에 SIMD나 추론 속도 개선을 입증한 벤치마크는 없다.

검토를 작성하고 loader를 만든 AI 보조 작업자는 원래 의미 검토와 정답을 알고 있었다. 합성 테스트는 제작자 측 검증이고 독립 원천이나 blind 평가가 아니다. loader는 이미 저장한 판정의 binding과 mask 규칙을 확인하며, 그 의미 판단의 정확성·원천 다양성·학습 준비를 새로 증명하지 않는다. 일반 AI 협업 비용은 측정하지 않았다.

## 파일과 재현 범위

- [원본 고정 계획](PLAN-67.ko.md): 내용과 bytes를 그대로 보존한다.
- [원본 결과](results-67.json): 원본 envelope의 exact copy다.
- [원본 실행 예약](ROOT-INVOCATION-RECEIPT-1.json)과 [완료 장부](ROOT-ACTUAL-EXECUTION-LEDGER-67.v1.json): 예약과 완료 상태를 각각 보존한다.
- [준비 장부](PREPARATION-LEDGER-67.json): 실제 실행 전의 역사적 상태0을 그대로 보존한다. 현재 실행 횟수는 완료 장부를 읽어야 한다.
- `prototype-*.txt`: 원래 helper·합성 테스트·module 파일의 exact text archive다. 바이너리는 게시하지 않는다.
- [공개용 호출 설명](INVOCATION-PUBLIC-67.json): 비공개 경로를 placeholder로 바꾼 별도 새 artifact다. 원래 호출 계획과 SHA가 같다고 주장하지 않는다.
- [포트 장부](PORT-LEDGER-67.json)와 [artifact provenance](PROVENANCE-67.json): 원본/공개 포트/변환 자료의 bytes와 SHA를 구분한다.

이 명령어는 maintainer용 metadata 도구다. Codex 등록은 별도 opt-in이며, 이번 포트는 현재 router나 모델을 바꾸지 않는다. CI 검사와 외부 게시 과정은 저장소의 기존 절차를 따른다.
