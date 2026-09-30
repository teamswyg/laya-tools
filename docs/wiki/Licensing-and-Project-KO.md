# 프로젝트 목적과 라이선스

[English](https://github.com/teamswyg/laya-tools/wiki/Licensing-and-Project-EN) · [홈](https://github.com/teamswyg/laya-tools/wiki)

## 왜 사용하는 도구인가요?

팀이 모델 선택을 품질·예산 정책으로 관리할 수 있는지 실험합니다. 작은 모델이라고 자동으로 경제적인 것은 아닙니다. 실패·재시도·지연도 비용에 포함됩니다. 공급자나 로컬 분류기가 항상 맞는 선택을 한다고 가정하기보다 판단 근거를 측정하려는 도구입니다.

저장소·모듈 이름은 `laya-tools`, 명령어는 `riidolaya`입니다. Laya는 선택을 분류하고, 명시적으로 실행할 때 Codex가 코딩합니다. 저장소 선택은 preview이며 riido-daemon 연결, 실제 절감, 진행 중인 대화의 모델 변경은 구현·입증된 기능이 아닙니다.

## 사용하거나 재배포해도 되나요?

자체 코드는 Apache-2.0이며 확인한 외부 구성 요소는 Apache-2.0·MIT·BSD-3-Clause 조건을 따릅니다. 고지 보존 등 조건 아래 수정·상업적 사용·재배포가 가능합니다. Go로 옮겼다고 원저작권이 사라지지는 않습니다.

재배포할 때 LICENSE, NOTICE, licenses 폴더를 보존합니다. 모델 묶음의 라이선스·출처·변환 고지·provenance도 함께 유지합니다. Laya·laya.tools·원저작자가 공식 보증하는 도구처럼 표현하지 않습니다.

초기 모델·CLI 배포에는 일부 고지 누락이 있었습니다. models-v2와 현재 바이너리에 보완했고 이전 릴리스에도 부속 문서를 제공했습니다. [전체 검토 문서](https://github.com/teamswyg/laya-tools/blob/main/docs/license-audit.ko.md)에 구체적인 수정 내용을 기록했습니다.

공개 라이선스가 모든 학습 데이터나 특허의 권리를 독립적으로 입증하는 것은 아닙니다. 검토는 기술적 배포 점검이며 무조건적인 법률 보증이 아닙니다. 별도 계약상 보증이나 비공개 데이터 미세조정에는 추가 검토가 필요합니다.

## 감사와 출처

[Laya와 SDK](https://github.com/NandhaKishorM/laya), 독립 커뮤니티 디렉터리 [laya.tools](https://laya.tools/), [system-one-router](https://github.com/mmornati/system-one-router), [pi-pignon](https://github.com/siiick/pi-pignon)에 감사드립니다. 후보 선택 정책과 캐시 전환 로직을 고정 출처·고지와 함께 이식했습니다. 다른 프로젝트의 속도·정확도 주장을 우리 결과로 사용하지 않습니다.

[이식 상세](https://github.com/teamswyg/laya-tools/blob/main/docs/ecosystem.ko.md) · [NOTICE](https://github.com/teamswyg/laya-tools/blob/main/NOTICE)

## 변경은 어떻게 검토하나요?

PR은 포맷·테스트·경쟁 상태·정적 검사·실제 추론·비밀정보·라이선스 목록 CI를 통과해야 합니다. 신뢰된 동일 저장소 PR은 사람 리뷰 승인 없이 병합할 수 있습니다. 이는 개발 병합 정책이며 Codex 실행 권한을 우회하는 설정이 아닙니다.

Wiki 원본은 [docs/wiki](https://github.com/teamswyg/laya-tools/tree/main/docs/wiki)에 있습니다. 두 언어를 함께 수정하고 저장소 CI를 통과한 뒤 Wiki에 게시하는 방식입니다. Wiki를 직접 수정하면 메인 브랜치 보호 규칙이 자동 적용되지는 않습니다. 문서를 위해 새 인증 정보나 자동 게시 서비스를 설치하지 않았습니다.
