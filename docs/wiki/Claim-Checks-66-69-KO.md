# 작은 주장 힌트를 학습 데이터로 연결하기66–69

[English](https://github.com/teamswyg/laya-tools/wiki/Claim-Checks-66-69-EN) · [이전 준비64–67](https://github.com/teamswyg/laya-tools/wiki/Claim-Preparation-64-67-KO)

목표는 짧은 요청과 후보 설명을 읽고, 사람이거나 에이전트가 **먼저 확인할 후보를 제안하는 아주 작은 모델**이다. 예를 들어 중첩 경로까지 찾고 싶은 작업에 여러 패턴을 보여 줄 때, 맞을 가능성이 큰 패턴을 앞에 제안한다. 최종 판단은 실제 코드 검사로 확인한다. 확인할 후보가 없거나 모델 입력에 맞지 않는 작업은 기존 흐름으로 넘긴다.

이번 단계에서는 새 모델을 내기 전에 원래 문장·정답·연결 그룹·학습 제외 표시를 Go 배열에 정확히 연결할 준비를 진행했다. 실제 역할 배정과 공개 함수 관측은 각각 한 번 실행했다. 새 학습이나 실제 Codex 비용 절감 결과는 아직 없다.

| 완료한 일 | 관측 결과 | 다음 학습에서 쓰는 뜻 |
|---|---|---|
| [역할 배정66](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/role-execution-66/README.ko.md) | 라벨 그룹 train10·validation3·calibration3 | 같은 코드 가족을 여러 역할에 나누지 않는다. |
| [upstream 설명68](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/source-caption-68/README.ko.md) | 영어 요청4개·후보10개·코드 가족2개. 고정 호출12개와 String3개를 관측해15개 비교가 모두 일치 | 미리 존재하던 MIT 설명을 한정된 의미 목표에 연결한다. |
| [학습 제외 연결69](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/mask-role-69/README.ko.md) | 적격 양성35·음성95, known 제외23·unknown63 | 제외는 학습 가중치0으로 표현하고 원래 후보와 정답을 보존한다. |
| [입력 한도69](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/input-boundary-69/README.ko.md) | 원래 요청72개와 설명216개 모두 기존512bytes·32단어 한도 안에 있음 | 문장을 줄이거나 한도를 바꾸지 않고 변환한다. |

## 어떤 문장이 학습에 들어가나

원래 함수 검사의 정답과 그 함수를 설명하는 문장의 품질은 구분한다. 함수의 동작이 맞더라도 설명이 부족할 수 있다. 그런 설명은 후보 목록에 남기면서 손실 계산에서만 제외한다. 요청21개의 unknown 상태도 보존하며 음성 정답으로 바꾸지 않는다.

학습에 반영되는 후보가 있는 그룹 수는 train9·validation3·calibration3이다. 기존 준비 하한을 그대로 만족한다. Go 투영 도구는 known train90행·validation36행을 모두 보존하고, 이 가운데 제외된18·1행에는 가중치0을 넣도록 준비했다. Calibration의 known27·unknown9후보는 감사 자료로 남기고 fit에 넣지 않는다. 실제 원래72개 투영은 별도의 고정 계획으로 진행한다.

## 공개 설명을 어떻게 검증했나

Semver 두 요청은 정확히 `v1.2.3`에서 leading `v`를 허용할지만 묻는다. NewVersion 후보의 인용은 함수 전용 주석이 아닌 package-level bullet이므로, 같은 문서의 parsing 맥락과 함수 경로를 함께 기록했다. Glob 두 요청은 명시한 두 파일 이름에서 중첩 경로를 포함할지만 묻는다. 이 유한 범위의 후보 정답 제안은 양성5·음성5다. 보조 관측인 missing-patch acceptance와 String은 모델 목표에 추가하지 않았다.

원본 함수의 결과가 예상과 맞는 것, 설명이 명시한 한정 범위에서 이를 전달하는 것, 새 학습 데이터가 여러 작성 원천을 충분히 포함하는 것은 서로 다른 검토다. 기존72개는 같은 합성 영어 작성 흐름이고, upstream의 네 요청도 AI-assisted 초안이다. 문서 번역·검사 항목 수·후보 수를 독립 작업 수로 세지 않는다. 원문과 전체 MIT 고지를 보존하고, 모델 본체는 GitHub에 넣지 않는다.

## 메모리 수치를 읽는 방법

역할 배정66의 전체 자식 프로세스 peak RSS는16.64MiB, 공개 함수 관측68은6.89MiB였다. 이는 자료 준비와 함수 관측 프로세스의 값이다. Go heap, GPU 사용량, 작은 모델의 추론 비용, 동시 처리 성능은 별도 측정해야 한다. 0.00초로 반올림된 CPU 시간도 CPU 작업이 없었다는 뜻은 아니다. 원시 로그·개인 경로·로컬 바이너리는 공개 기록에서 제외했다.

## 다음 실행

[고정 역할의 저장 순위 집계69](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/stored-role-utility-69/STORED-ROLE-UTILITY.ko.md)는 원래 함수 정답51요청을 유지했다. 검증 역할의 단일 최선 기준은 lexical22회, 정답을 아는 순서는17회였다. 기존5% 필요조건을 만족하는 개선 여지이며 학습 성과는 아니다. 전체는 단일 BM25기준91회/oracle73회를 유지한다. 역할마다 다른 최선 기준을 합친90회로 전체 기준을 바꾸지 않는다. 설명 mask가 모두 true인 요청만 보는 B진단에는 별도 gate를 적용하지 않았다.

원래72개를 먼저 고정 역할·손실 mask대로 배열에 변환하고, 같은 역할과 원래 함수 정답에서 고정 기준의 검사 비용을 확인한다. 그 뒤 새 upstream4개를 별도 버전으로 결합할 때 전체 코드 가족 관계와 역할을 다시 고정한다. 결과를 보고 seed·역할·후보·mask를 바꾸지 않는다. FP32 부모가 실제 효용을 보인 뒤 INT8/PTQ 자식과 삼진 STE 형제를 비교한다. 도메인별2400개 보호 최종 요청 목표는 별도로 유지한다.

[진행 이슈19](https://github.com/teamswyg/laya-tools/issues/19)에서 CI·실행·게시의 완료 상태를 확인할 수 있다. 이전 준비 문서의 당시 미실행 상태는 역사 기록으로 남는다.
