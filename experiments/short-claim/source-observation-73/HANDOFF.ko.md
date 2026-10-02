# 첫 3원천 관찰기의 바깥 controller

private 준비가 끝났으며 실제 controller·관찰기·원 API 실행은 모두 0회입니다. 원래 원천 패키지를 import하지 않는 표준 라이브러리 Go 코드와 합성 테스트만 빌드·검증했습니다. 고정 파일 21개와 Want, 바이너리는 읽기 및 메타데이터 검사로 연결했습니다.

Root가 `first-three-controller`를 1회 실행하며, 플래그는 `--draft`, `--draft-sha256`, `--handoff`, `--handoff-sha256`, `--control-dir`입니다. 입력 SHA는 `CONTROLLER-HANDOFF.v1.json`의 exact 값이며 바깥 결과 디렉터리는 아직 존재하지 않아야 합니다. 현재 이 문서는 실행 기록이 아닙니다.

controller는 원래 draft의 `Frozen` 토큰 1개만 false에서 true로 바꿉니다. 새 디렉터리·O_EXCL 예약·호출 장부를 fsync한 뒤에만 `/usr/bin/time -l`로 child를 시작합니다. child에는 CPU 1, Go soft heap 256 MiB와 최소 환경만 전달합니다. 60초가 지나면 프로세스 그룹을 kill하고 Wait를 정확히 1회 join합니다. Kill/Wait 뒤의 총 지연을 숨기지 않으며 RSS 하드 제한으로 설명하지 않습니다.

최종 파일과 중간 파일의 SHA, nullable 호출 카운터, OS CPU/RSS를 장부에 보존합니다. 최종 JSON이 손상되어도 유효한 중간 수치를 지우지 않습니다. 모든 24회가 끝났더라도 원래 Want와 다른 관측은 그대로 남고, 자동 재시도·라벨 변경은 없습니다. 24개 probe는 서로 다른 부모 요청 24개가 아니라 행동 목표 3개의 유한 입력입니다.

controller 테스트 10개와 메타데이터 테스트 1개, 변경된 자원 로그 해석 테스트 1회가 통과했습니다. 실제 원천/초기화/API 호출, 학습, 모델 실행, 공유 저장소 수정·외부 게시는 모두 0입니다. 장부의 테스트 시간은 제품 성능 측정이 아닙니다.

controller 작성자의 합성 검사는 독립적인 관찰기 mechanics 검토 및 별도 Want 검토와 구분합니다. 컴파일러 메타데이터와 source/binary 핀은 출처 증거이며 컴파일러를 형식적으로 증명하지 않습니다. 캡션·부모 라벨·일반화·학습 자격은 여전히 false입니다. 원시 로그와 private 절대 경로가 있는 원래 draft는 공개 자산이 아닙니다.
