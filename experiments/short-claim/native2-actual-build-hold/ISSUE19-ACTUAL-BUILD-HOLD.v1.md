PR [#107](https://github.com/teamswyg/laya-tools/pull/107)의 필수 CI 네 개가 첫 시도에서 모두 통과해 자동 병합됐습니다. 두 실행 파일도 실제 Go1.27.1/Apple Silicon/CGO0로 빌드했습니다. 관측기는 4,537,554B, 바깥 제어기는 5,118,610B이며, 모델 파일이나 추론 메모리 수치가 아닌 실행 도구의 디스크 크기입니다.

실제 바이너리 검토에서 시작 전에 문제를 발견했습니다. Go는 `-trimpath`일 때 `-ldflags`를 buildinfo에 넣지 않습니다. 기존 제어기는 두 조건을 동시에 요구해 실제 바이너리를 받을 수 없는 상태였습니다. 더미 테스트 15개 PASS가 이 실제 빌드 호환성을 검증하지 않았다는 점도 기록합니다. Go1.27.1의 `src/cmd/go/internal/load/pkg.go:2469–2479`와 실제 빌드 메타데이터를 대조했습니다.

- [x] 실제 원본/제어기 바이너리 빌드와 크기·SHA 확인
- [x] 원본 시작 전 호환성 문제 발견; 기존 계획·실행 파일·검사 이력 보존
- [ ] 새 제어기의 바이너리 고정값 검사와 독립 검토·CI
- [ ] 새로운 소스·권리·자원 계획으로 원본 단회 관측
- [ ] 예상값 대조 및 별도 적격성 검토

원본·제어기 Start, 초기화·원본 관측은 모두 0회입니다. once marker도 생성하지 않았고, 실패한 관측을 재시도한 상황이 아닙니다. 다음 버전은 바이너리를 실행하지 않고 Mach-O의 세 Go 문자열 고정값을 읽어 계획의 SHA와 비교합니다. 기존 봉인은 바꾸지 않습니다. 현재 실제 corpus Fit은 누적 3/8회, 논리 모델 3/16개이며 추가 학습·활성화는 없습니다.

English: PR #107 passed all four required CI jobs on attempt 1 and merged automatically. Root built the real Go1.27.1 arm64 worker (4,537,554 B) and controller (5,118,610 B); these are executable disk sizes, not model or RAM measurements. Before any Start, actual metadata revealed that Go deliberately omits `-ldflags` from buildinfo with `-trimpath`, making the v4 gate incompatible. Fifteen passing fake tests did not cover real binary compatibility. All old proofs, plans and binaries are retained. A fresh controller will statically inspect the three actual Mach-O Go string values, then receive independent review and exact-source CI. Original/controller starts, original observations and once markers remain zero. No corpus fit, activation or scientific observation retry was added; cumulative fits/models remain 3/3.
