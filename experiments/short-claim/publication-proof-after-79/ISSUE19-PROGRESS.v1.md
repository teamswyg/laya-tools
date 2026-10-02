### PDCA 진행: 79개 데이터 추가 실험은 실행 완료, 성능 개선은 미확인

[PR #105](https://github.com/teamswyg/laya-tools/pull/105)에 이번 실험의 실행 결과, 메모리 관측, 입력 보존 검사와 검토 기록을 제출했습니다. 원본 76개 요청에 학습 전용 요청 3개를 추가하고, 기존 FP32 실험과 같은 조건으로 **Project 1회·Fit 1회** 실행했습니다. 누적 학습 소비는 기존 2회와 이번 1회를 합한 **3/8회**, 논리 모델은 **3/16개**입니다. 같은 79개 자료의 추가 seed·epoch·손실 탐색은 진행하지 않습니다.

| 비교 | 모의 후보 검증 횟수, 적을수록 좋음 | 첫 후보 정답, 10개 중 |
|---|---:|---:|
| 가장 좋은 기존 lexical 비교군 | 27 | 5 |
| 첫 FP32 실험 71 | 31 | 2 |
| pair+BCE 실험 72 | 33 | 1 |
| 데이터 추가 실험 79 | 31 | 2 |

이번 결과는 **유틸리티 기준 실패**입니다. 학습 프로그램이 정상 종료한 것과 쓸 만한 모델이라는 판정은 구분합니다. 요청 20개의 기존 개발 검증 자료에서 정답 있는 10개와 정답 없는 5개를 평가하고, unknown 5개는 제외했습니다. 검증 횟수는 저장된 정답에 따른 모의 확인이며 실제 Codex 호출·토큰·요금 절감 측정이 아닙니다. 보호된 최종 평가 자료도 아닙니다.

모델 파일은 32,792 B입니다. 외부 관측 기준 Project 전체 프로세스는 약 2.09초·최대 RSS 33.45 MiB, Fit 전체 프로세스는 약 1.24초·35.08 MiB였습니다. Go CPU 학습이며 GPU/MPS나 원본 Laya transformer 학습은 이번 실행에서 0회입니다. 작은 모델이 기존 단순 비교군보다 유용하다는 증거가 없어 비활성 실패 아카이브로만 준비합니다.

- [x] 기존 76개 prefix·validation·calibration 보존을 확인하고 79개 Project 1회 실행
- [x] 고정 BCE 학습 1회 실행, 결과와 실패 판정 보존
- [x] 다음 경로 3개를 비교한 한국어·영어 독립 분석 준비
- [ ] PR #105의 Ubuntu·macOS·secrets·quality CI 모두 통과 후 병합
- [ ] Hugging Face 비활성 실패 아카이브 게시 후 immutable revision·파일 지문 확인
- [ ] Wiki 한국어·영어 사용 설명과 실제 결과 동기화
- [ ] 새로운 의미 계약 60개 확보, 30개 확보 시 중간 검토
- [ ] 새 연결 원천 그룹을 점수 확인 전에 분리하고 다음 제한 실험 고정

현재 CI의 Ubuntu 작업은 `golang.org/x/text@v0.25.0` 다운로드 중 Go proxy의 HTTP/2 `INTERNAL_ERROR`로 실패했습니다. [CI 실행](https://github.com/teamswyg/laya-tools/actions/runs/37032449547)의 macOS 작업이 진행 중이며, 완료 후 실패한 작업을 다시 실행합니다. 로컬 전체 race 테스트·vet·비밀정보 검사는 통과했습니다. CI 통과나 HF 게시가 완료됐다고 아직 기록하지 않습니다.

다음 자료 확보에서는 입력 예시·번역·후보 3개를 요청 수로 세지 않습니다. 파싱, 오류 뒤 상태/소유권, 설정/변환, 정확한 산술, 파일/I/O, 취소/수명 등 여섯 구간에서 서로 다른 동작 계약을 추립니다. 기존 79개와 원천 조사 120개 초안의 중복을 확인하고, 기존 요청의 새로운 검사 계약과 완전히 새로운 의미 요청을 따로 셉니다. **60개는 목표이고 현재 이 새 배치의 적격 계약은 0개**입니다. 각 주장 도메인의 보호 final 최소 2,400개도 여전히 별도 목표입니다.

현재 특징에는 이미 단어·인접 두 단어의 교차가 있습니다. 다음 표현 가설은 조건과 행동·부정의 적용 범위를 보존하는 제한된 특징입니다. 이것이 실패 원인이라는 결론은 아니며, 새로운 개발 원천에서 한 가지 변경과 기존 방식의 비교를 사전 고정합니다. 양자화는 유용성 검증과 별도로 진행합니다. 작은 디스크 파일이 곧 작은 실행 메모리는 아닙니다.

**English:** One data-addition Project and one fixed-recipe BCE Fit completed; utility failed (31 simulated checks and Top-1 2/10, versus the best control's 27 and 5/10). Cumulative fits are 3/8 and logical models 3/16. These are exposed development results, not protected-final evidence or measured LLM savings. Both old 91 and new 100 training rows fit within batch 128, so batch count is unchanged. The next priority is 60 distinct qualified semantic contracts, with a 30-contract acquisition checkpoint, connected-source-group splits frozen before scoring, and a separate condition/action feature hypothesis. The new batch currently has zero qualified contracts. At least 2,400 protected final requests per claimed domain remain outstanding. PR CI and inactive HF publication are pending; the Ubuntu CI failure was a dependency-download network error. No default model activation, paid evaluation or GPU/Laya-transformer training was performed.
