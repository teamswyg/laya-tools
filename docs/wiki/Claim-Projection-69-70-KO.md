# 아주 작은 힌트 모델: 실제 준비69와 새 분할70

목표는 **에이전트가 먼저 확인할 근거 후보의 순서를 저렴하게 제안**하는 것이다. Laya 본체의 새 학습과 구분하는 encoder-free 주장 힌트 실험이다. 모델이 앞에 놓은 후보는 원문과 요구사항으로 검증하고, 틀려도 나머지 후보와 fallback을 유지한다. 실행 승인이나 정답 증명기가 아니다.

예를 들어 “중첩 디렉터리의 Go 파일을 찾는다”는 요청에 여러 함수 설명이 있으면, 어떤 설명부터 확인할지를 제안한다. 성공 기준은 같은 요청에서 **맞는 후보까지 필요한 확인 횟수**가 줄어드는지다. 이것도 개발용 대리 지표이고 실제 LLM 사용량은 따로 관측해야 한다.

## 완료된 것

- 원래72개 요청의 첫 실제 Go 배열 준비69: Project1회·재시도0·fit0. train90/validation36행 중 손실 가중치0인18/1행도 보존했다. 미확정 후보63개는 nullable 감사 자료에 남기고 calibration은 fitting에서 제외했다.
- 작성자와 별도 검토자의 저장 원본·감독·배열 대조가 통과했다. 검토 도구의 후보 수 가정 오류와 장부 추산 정정도 보존했다. feature 재추출이나 학습 성능 검증은 아니다.
- 새 upstream68 요청4개를 더한76개/226후보/19가족을 전체로 다시 배정했다. 고정 seed1729·첫 Assign1회·재시도0. known-containing 가족은 train11/validation4/calibration3이고 기존 coverage가 통과했다. 과거66 역할 뒤에 새 사례를 붙이지 않았다.

## 숫자를 읽는 법

| 관측 | 실제 값 | 해석 |
|---|---:|---|
| 원래72개 준비의 소유 payload |약1.69MiB|반환 배열·보존 문자열 계산, 임시 할당 제외|
| 같은 전체 프로세스 peak RSS/wall |약53.95MiB /0.466초|읽기·검증·배열·JSON 출력을 포함한 단일 개발 관측|
| 원본 결과 / 공개 요약 |약7.7MB /97KB|요약은 전체 feature를 생략하며 학습 자산을 대체하지 않음|
| 전체76개 역할 준비 peak RSS/wall |약51.19MiB /0.478초|metadata와 역할 준비; 학습·추론이 아님|

CPU1/soft Go heap256MiB/외부300초로 실행했다. soft heap·payload·JSON·RSS는 다른 항목이다. GPU·Go heap·Codex 토큰 절감과 AI 보조 준비·검토 비용은 미측정이다.

## 다음 첫 학습

76개 역할에 맞는 배열을 별도로 고정한 뒤8192차원 FP32 fit1회를 실행한다. 같은 validation에서 네 비학습 기준 중 최선의 하나와 비교하며 가중치0 후보도 순위 평가에 남긴다. calibration은 fit과 epoch 선택에서 제외한다. 새 두 upstream 가족은 실제로 train에 들어갔으므로 이 validation을 그 라이브러리에 대한 일반화 평가라고 부르지 않는다.

준비된 FP32 형식은32,792B, Decode된 float64 계수 배열은65,536B다. 전체 실행 메모리와 반복 추론 비용은 실제 모델을 만든 뒤 측정한다. FP32 개발 효용이 확보되면 INT8/PTQ 자식이나3진 STE 형제를 별도 실험으로 비교한다. 도메인별 고유 보호 평가2400개 이상은 별도 미완료 목표다.

## 기록과 사용법

[첫 실제 준비69](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/projection-execution-69/README.ko.md) · [전체76 역할70](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/combined-role-execution-70/README.ko.md) · [이전 PR92 CI·Wiki 게시 확인](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/publication-proof-92/README.ko.md).

지금 쓸 수 있는 비학습 Go 명령은 `riido-shortclaim`이다. [사용법56](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/USAGE-56.ko.md)의 요청·후보 JSON으로 시작한다. 이 기록이 학습 모델을 기본 등록한 것은 아니다. 계수는 GitHub에 올리지 않고 라이선스·검증·CI를 통과한 한정 버전만 Hugging Face에 별도 게시한다. [English](Claim-Projection-69-70-EN).
