# 상주 자원 실험56d 구현과 실행 경계

[English](IMPLEMENTATION-RESIDENT-56d.en.md) · [보존된 사전 설계](PLAN-RESIDENT-56d.ko.md) · [Go 구현](../../cmd/riido-residentperf/main.go)

이 단계는 작은 힌트 도구를 한 번 켜고 여러 번 요청할 때 드는 비용을 알아보기 위한 것이다. **Laya 모델 추론·학습 실험은 아니다.** 같은 공개 요청을 반복하는 자원 측정과, 서로 다른 요청에서 힌트가 유용한지 확인하는 의미 평가를 분리한다. 이 문서는 결과 수집 전에 작성하는 구현 설명이며, 성공 수치나 CI 완료를 미리 주장하지 않는다.

## 무엇을 실행하는가

기존72개 공개 부모의 요청과 후보 설명만 추출한다. 후보가 원래3개인 부모만 선택하고, 텍스트의 정규화 특징이 같은 투영은 첫 원문을 유지한다. 원래 부모·후보 ID와 선택/제외 이유는 연결 자료로 보존한다. ID·정답·후보 소스는 순위 특징에 넣지 않는다. 서로 다른 특징이2개 이상인지는 준비 단계에서 확인한다.

모든 입력은 정확한 compact JSON bytes를 base64로 저장한다. 보기 좋게 JSON을 다시 저장해도 실제 전송 bytes가 변하지 않게 하기 위해서다. 실행할 때 LF 한 개를 붙이며 LF 포함12KiB 제한을 적용한다. 원문 wire SHA, 입력 digest, 정규화 특징 SHA는 서로 다른 목적의 값이다. 정규화가 문장 부호를 제거하므로 특징 SHA는 원문 문법을 보는 `narrow_rule`의 안전한 캐시 키가 아니다.

실제 child는 `riido-shortclaim --stream`이다. `fixed_order`, `bm25`, `lexical_ordered`, `narrow_rule` 네 가지 Go 기준선은 각 요청의 준비·검증·순위·digest·출력을 다시 계산한다. 기대 응답은 준비 단계에서 같은 공개 구현으로 만든 **프로토콜 비교값**이며 독립적인 정답 라벨이 아니다. 후보는 항상 모두 보존하고 응답 상태도 `unverified_heuristic`이다.

## 결과를 보기 전에 고정하는 것

| 항목 | 실행 정책 |
|---|---|
| 행렬 | 기준선4 × same/distinct2 × 새 child3 = 24행 |
| 행의 요청 | 첫 요청1 + 예열20 + timed1024 = 1045 |
| 풀 시작 위치 | 첫 요청·예열·timed 각 단계에서 위치0부터 시작 |
| 실행 순서 | 반복1·3은 기준선 정방향, 반복2는 역방향; workload 순서는 교대 |
| 제한 | child15초, replay60초, 정리 공통 예산1000ms; 실패 첫 행에서 후속 행 중단 |
| 동시 처리 | child1, 진행 중 요청1, 자동 재시도0 |
| Go 설정 | Go1.27.1, CGO0, trimpath, GOMAXPROCS1, heap soft limit256MiB |
| 신규 공개 자료 | corpus·plan·preparation·recipe·results JSON 합계8MiB 이내 |

총25,080회와 timed24,576회는 **반복 관측 횟수**다. 새로운 학습 요청이나2400개 최종 평가를 확보했다는 뜻이 아니다. 새 child도 OS 캐시가 비워진 cold를 뜻하지 않으며 캐시를 강제로 비우지 않는다. CPU P 수와 heap soft limit은 OS 스레드·전체 메모리의 강제 상한이 아니다.

배포할 driver의 실제 소스22개는 컴파일에 포함된 bytes와 저장소 bytes를 대조한다. 별도 child 소스를 포함한 runtime 파일24개, module 파일2개, 수집 전 검사5개를 합친31개를 계획에 해시로 고정한다. 목록 밖 구현 파일도 거부한다. 빌드 recipe와 실제 두 바이너리의 SHA·buildinfo, 원본 입력 SHA, 플랫폼별 프로토콜 기대 bytes,24행 순서를 공식 실행 전에 봉인한다.

`preparation_source_commit`은 CLI에서40자리 형식만 확인한다. **그 commit의 Git blob31개가 계획의 소스 해시와 같은지는 별도 봉인 작업에서 확인하고 기록한다.** CLI가 Git 이력을 검증했다고 표현하지 않는다. 계획 SHA를 따로 받아 검증하며, 한 번 읽어 해시를 확인한 child bytes를 비공개 위치에 복사해 실행한다. 실제 바이너리와 비공개 경로는 Git에 올리지 않는다.

## CPU·메모리·시간을 읽는 법

Child 종료 뒤 user/system CPU와 lifetime peak RSS를 기록한다. Controller는 자기 CPU의 replay 구간 증가와, 준비를 포함한 전체 lifetime peak RSS를 별도로 기록한다. Child는 다른 프로세스를 만들지 않는다. [Go의 ProcessState](https://pkg.go.dev/os#ProcessState.SysUsage)는 OS별 자원 정보를 제공하며, CPU API는 종료된 프로세스와 그 자식의 시간을 설명한다. 이번 관측 범위를 임의의 자식 프로세스 트리로 일반화하지 않는다.

Linux RSS 원단위는 [getrusage](https://man7.org/linux/man-pages/man2/getrusage.2.html)의 KiB이다. Darwin은 [XNU 자원 구현](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/kern/kern_resource.c)에서 `resident_size_max`를 사용하며 [필드 정의](https://github.com/apple-oss-distributions/xnu/blob/main/osfmk/mach/task_info.h)는 bytes다. 원단위와 변환 bytes를 함께 남긴다. 두 프로세스의 peak를 더해서 동시에 사용한 총 메모리라고 부르지 않는다. RSS는 Go heap·GPU 메모리·현재 warm 메모리도 아니다.

Timed RTT는 쓰기 시작부터 완전한 응답 줄 수신까지다. Controller 검증 시간은 수신부터 검증 완료까지로, 전달 대기와 실제 검증 작업을 별도로 기록한다. 순수 startup·child 처리·pipe 시간은 관측하지 못한다. 각 행의 기대 stream 해시 준비 시간은 child 시작 전이지만 전체 replay CPU·시간 안에 포함한다. 겹치는 시간을 빼거나 더해 숨은 비용을 추정하지 않는다.

완료된 행의 CPU/request 분모는 첫 요청과 예열을 포함한1045다. Timed 지연 분모는1024다. 실패한 행은 실제 시도·전송·수신·검증·실패·미완료·미시도 수와 가능한 CPU/RSS를 보존하며,1045개 성공 평균으로 채우지 않는다. 후속 미시작 행도 삭제하지 않는다. 프로세스 종료 확인, stdout/stderr reader, 취소 watcher가 **한 번의 공통1000ms 정리 예산**을 공유한다. 정리가 늦거나60초 경계가 넘어가면 위반을 남기고 재측정으로 덮지 않는다.

## 준비와 공식 실행의 구분

`--stage prepare`는 새 디렉터리에 `corpus.json`, `preparation.json`, `build-recipe.json`을 만든다. Child 요청·모델 호출·fit·보호된 최종 자료 읽기는0이다. 단위 검사에서 쓰는 작은 가짜 프로세스·잘못된 응답·시간 초과 사례와 packaged build 대조는 수집기 검증이며 공식 성능 행렬이 아니다. 원본 소스 함수의 의미 정답을 새로 감사하지 않는다.

공식 `--stage replay`는 봉인된 plan SHA·corpus·실제 packaged child가 필요하다. 기존 출력은 덮어쓰지 않으며, 첫 child 전에 새 results 파일을 확보한다. 프로토콜 불일치·stderr·시간 초과·종료 실패는 고정 오류 이름으로 기록하고 비공개 인자나 stderr 원문을 게시하지 않는다. 결과 수집 후에만 실제 관측·CI 링크를 별도 결과 문서에 붙인다.

캐시·SIMD·lock·SoA·open-loop 동시성·echo 대조는 이후 독립 계획으로 남긴다. 의미 효용과 도메인별 최소2400개의 서로 다른 보호된 최종 요청 목표도 별도다. 이번 단계의 학습·새 모델·가중치·운영 자동 활성화·LLM 절감 주장은0이다.
