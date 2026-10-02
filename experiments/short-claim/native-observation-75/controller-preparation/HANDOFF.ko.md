# Wrap/Marshal 75: 한 번 실행할 외부 제어기 준비

새 제어기는 기존 first-three 제어기를 수정하지 않고 별도로 만들었습니다. Go 표준 라이브러리만 사용하며 Go 1.27.1·CGO 0·trimpath·buildvcs=false로 한 번 빌드했습니다. **새 제어기와 원본 관찰기 바이너리는 아직 실행하지 않았습니다.** 이 준비에서는 원 API·init·역할·라벨·모델·학습·원격 게시가 0회입니다.

실행 권한은 부모 작업이 유지합니다. 제어기는 `--draft`, `--draft-sha256`, `--handoff`, `--handoff-sha256`, `--control-dir`를 받습니다. 실제 경로는 비공개 인자이고 진단에는 출력하지 않습니다. 기존 원본 draft의 `Frozen: false` 토큰 하나만 `true`로 바꿔 새 파일과 SHA를 만듭니다. 원래 draft와 handoff는 덮어쓰지 않습니다.

실행 전에는 root가 제공한 draft/handoff SHA, 순서가 고정된 14개 source/license/module/document 핀, Want seal, 24개 Want와 handoff Specification의 동일성, binary SHA와 Go/OS/arch/CGO/trimpath/module replace 정보를 읽어서 확인합니다. 임의 추가 Go 원문은 거절합니다. 원본 함수나 관찰기 실행으로 이 검사를 대신하지 않습니다. 이 핀들은 출처 결속이며 trusted compiler의 독립적 증명은 아닙니다.

실행 순서는 다음과 같습니다.

1. 새 외부 폴더를 독점 생성하고 frozen plan·invocation·before-start ledger를 독점 저장한 뒤 파일/폴더를 동기화합니다.
2. 자식 폴더의 부재를 확인하고 독점 생성합니다. 자식이 요구하는 outside-reservation을 독점 저장하고 동기화합니다. `results.json`은 제어기가 만들지 않으며 worker가 독점 예약합니다.
3. 최소 환경인 CPU 1·256 MiB soft Go heap·기본 PATH/locale만 전달해 `/usr/bin/time -l`로 자식을 한 번 시작합니다. 인증 정보나 사용자 환경은 전달하지 않습니다. 제어기도 CPU 1·같은 soft heap 설정을 적용합니다.
4. 60초가 되면 process group을 SIGKILL하고 단 한 번의 Wait를 끝까지 합류합니다. 자동 재시도는 없습니다. Kill/Wait 이후 반환까지 정확히 60초 이내라고 보장하지 않습니다.
5. final과 partial 원문 SHA를 모두 보존합니다. 각 저장 record의 Probe/Want·순서·Got와 Matches·panic/return/dispatch 산술을 고정 Want에 대조합니다. 차이가 있는 완료와 24개 Want의 완전 일치를 별도 상태로 기록합니다.

빈 결과나 실행 전 거절에서는 API 수를 만들어내지 않습니다. 시작 실패, 시작 후 비정상 종료, timeout, 유효한 부분 prefix, 잘못된 final은 구분합니다. corrupt final 뒤에도 검증 가능한 partial 수를 남기되 final 오류를 숨기지 않습니다. 예약 수는 호출 의도이며 마지막 in-flight 반환의 증명이 아닙니다. 완전한 24개 관찰이 끝나도 차이나 panic은 그대로 남습니다.

Darwin time의 real/user/sys, 최대 RSS와 footprint는 nullable입니다. RSS 단위는 바이트이며 자식 init·preflight·wrapper/API·checkpoint I/O를 포함하고 부모 제어기는 제외합니다. soft Go heap은 총 RSS의 하드 제한이 아닙니다. 원시 stdout/stderr·plan과 경로는 비공개이며 안전한 ledger에는 고정 진단 종류·카운터·SHA·관측값만 있습니다.

검사는 합성/metadata-only race 1회(14개 테스트)와 vet 2회, 최종 compile-only 1회가 통과했습니다. race 이후 main의 CPU/heap 설정 두 줄만 추가했고 최종 vet/build로 확인했습니다. stub lifecycle은 실제 process를 시작하지 않습니다. source/Want/binary를 읽는 검사도 init/API를 실행하지 않습니다. 작성자는 제어기 작성자이므로 독립적인 원천 의미·실제 실행·Want 정확성 판정이라고 부르지 않습니다. 두 행동 목표의 24개 finite probe는 24개의 독립 부모가 아니며 학습·caption·parent-label·broad-truth 자격은 false로 유지합니다.
