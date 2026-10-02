# 공개 동작 검증 도구 사용법57

`riido-publicaudit`은 유지보수자가 작은 공개 동작의 정답 근거를 모으는 Go 도구다. 사람이든 에이전트든 결과의 입력·오류·원천 revision을 읽을 수 있다. 모델 선택이나 실제 코드 수정 승인은 하지 않는다. 일반 사용자는 기존 `riidolaya`/`riido-shortclaim`을 계속 사용한다.

## 공개 결과 확인

Go1.27.1에서 저장소 루트의 아래 검사를 실행한다. Linux와 macOS 모두 실제 API를 재생해 보존된 결과·원천·기대값을 대조한다. 역사적 Darwin binary를 다른 OS에서 실행했다는 뜻은 아니며 새 공식 표본을 만들지 않는다.

```sh
go test ./cmd/riido-publicaudit -run TestPublishedFiniteAuditReplay -count=1
```

[입력](probes-57.json)의 raw UTF-8와 hex가 같아야 한다. [원본 결과](results-57.json)는 규범·기술적 API·더 강한 정책을 분리한다. `source_read_hypothesis`는 사전 소스 읽기 예상이며 수락 기준이 아니다. 파생 정답이나 hidden source ID를 모델 문구에 채우는 기능은 없다.

## 공식 수집 명령의 경계

원래 수집은 아래 순서로 수행했다. binary는 Go1.27.1 Darwin/arm64, CGO0에서 [recipe](build-recipe-57.json)의 세 build flag로 만들어 실행 계획의 SHA와 맞췄다. 결과 파일을 덮어쓰지 않으며 새 디렉터리도 기존에 있으면 거절한다. fresh output은 전역 실행 횟수 제한이 아니므로 공식1회 기록은 외부 실험 장부로 관리한다.

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -p=1 -o ./riido-publicaudit ./cmd/riido-publicaudit
./riido-publicaudit plan --repo . --source-commit 9e99914f6d1e35aa9d97413f80fc6e03768e97ce --out NEW_PLAN_FILE
./riido-publicaudit audit --repo . --input-commit 1462c705087f24666049f8bd31d317490c05dbeb --out NEW_RESULT_DIRECTORY
```

`audit`은 저장소의 고정된 `execution-plan-57.json`과 `probes-57.json`을 읽는다. `plan`의 새 출력만으로 이를 자동 교체하지 않는다. 원래 frozen commit들이 로컬 Git에 있어야 하고, main의 squash history만 가져온 checkout은 해당 공개 commit도 별도로 fetch해야 한다. 다른 OS/toolchain/binary의 재수집은 새 계획·입력 freeze를 작성하는 **별도 실험**이며 위 역사적 수집을 덮어쓰지 않는다.

지원 연산은 `strict_parse`, `compare`, `glob_match`, `glob_validate`뿐이다. 원본 코드 문자열의 임의 실행이나 파일시스템 glob은 없다. 잘못된 UTF-8/과한 길이/미등록 파일/변조/duplicate JSON key/extra field/누락된 입력 freeze는 관측 전에 거절한다. 라이브러리의 알려진 오류는 관측하고 panic·미분류 오류는 unknown으로 남긴다.

기대값 불일치는 유효한 연구 결과이므로 그 자체로 exit 실패를 만들지 않는다. 불완전/unknown 관측이나 저장 실패는 nonzero다. 저장 실패 진단은 이미 시작한 관측 수를 표시하며 공식 재시도를 자동으로 수행하지 않는다. 이번 도구는 CPU/RSS/GPU/지연 측정기가 아니다.
