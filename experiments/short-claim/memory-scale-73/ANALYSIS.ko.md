# 희소 배열 메모리: 현재 구조와 확장 경계

**특성 index는 이미 uint16이다.** `pairlearn.Dataset`은 행 경계 `Offsets`, `Indices`, `float64 Values`를 분리한 CSR/SoA이며, 라벨·그룹·가중치도 별도 열이다. ranking의 부모·양성 행·음성 행·pair 가중치도 SoA이고 한 실행이 소유하므로 공유 lock이 없다. 원문·감사 메타데이터는 고정 후보 8칸을 가진 부모 구조체 배열에 남는다.

저장된 76개/226후보의 자체 payload는 **1,928,154 B = 1.84 MiB**다. 저장된 열 길이와 64bit 구조체 크기만으로 정확히 재계산했다. train91/validation45행의 105,079 entries에 진단용 CSR이 75,573 entries를 복제한다. index+value는 **1,806,520 B, payload의 93.7%**다. 8192차원을 행마다 조밀하게 저장하지 않으며 실제 fit행 평균은 772.6 stored entries다. 해시 충돌 합산 뒤 0이 된 값도 포함한다.

다음 표는 **76개의 후보 수·unknown/calibration·mask·문장 길이 비율을 그대로 늘린 산술 추정**이다. 실제 요청/역할/특성을 생성한 결과가 아니며 N은 전체 개발 pool의 요청 수다.

| 요청 N | 자체 payload 추정 | 같은 JSON 형식의 파일 추정 |
|---:|---:|---:|
| 128 | 3.10 MiB | 13.48 MiB |
| 256 | 6.19 MiB | 26.95 MiB |
| 512 | 12.38 MiB | 53.91 MiB |
| 2400 | 58.04 MiB | 252.69 MiB |

이 비율에서 자체 payload 64 MiB 경계는 약 2646개지만, **현재 private fitter의 고정 입력파일 64 MiB 계약은 약 607개에서 먼저 걸린다.** 자동 상향하지 않는다. unknown은 nullable 감사 원문만 보유하고 calibration은 fit행이 없다. zero-weight known행은 진단 CSR에서만 빠지므로 비율을 바꾸면 표도 바뀐다.

모든 요청이 known train/validation이고 weight1이라고 가정하면, 현재 평균 entries/행과 문자열 byte 상한을 참고한 2400개×후보2/4/8 시나리오는 **80.18/156.34/308.65 MiB**다. 같은 조건에서 64 MiB에 들어가려면 평균 entries/행이 약 **596/268/105 이하**여야 한다. 현행 32단어 계약은 최대63 terms의 교차+recall로 **행당≤3970 entries**라는 정적 상한을 준다. 이 상한에서는 128개×후보8도 초과한다. 실제 분포 예측이나 새 데이터 생성 결과가 아니다.

64bit payload 식은 `1096 + 576N + retainedTextBytes + 10F + 48R + 32` B다. F/R은 네 CSR을 합친 길이다. rowPlan scratch32 B/fit행, JSON 파싱·allocator·runtime은 별도다. ranking reserve `48(fitRows+2)+32pairs+48epochs`는 2400×후보8 최대pair16/부모에서도 **2.05 MiB**로 작고 별도의 64 MiB cap이다. FP32 모델 파일32,792 B, decoded계수65,536 B, 주요 학습 배열 약4×64 KiB도 구분해야 한다. **Go soft256 MiB는 RSS hard cap이 아니다.** 실제 projection의 JSON8,390,462 B/RSS약47.7 MiB를 새 데이터 RSS로 선형 확대할 수 없다.

무손실 축소로 RowRef의 uint16 부모/uint8 후보 분리 열은 현재 **3289 B**, uint32 offsets는 **1028 B**만 절약한다. **uint16 offsets는 train entries67,104 때문에 불가**하다. 외부 parent/group ID는 임의의 int이므로 좁히려면 역매핑과 원래 cross-role 검사를 보존해야 한다. 진단 CSR을 참조 view로 바꾸면 feature 복제 **755,730 B**를 줄일 여지가 더 크지만 호환 API 작업이 필요하다. FP32 Values는 합·gradient·동률·순서를 바꿀 수 있는 수치/스키마 변경이므로 별도 ablation이며 즉시 적용하지 않는다.

다음 최소 후보는 **FP64 값을 보존하는 compact/sharded 저장 경로**다. 값·인덱스의 exact round-trip과 원문·순서·nullable truth·mask·role·loss를 유지해야 한다. 이는 파일 병목만 줄이며 배열 cap 통과를 보장하지 않는다. 이번 코드/스키마/제한 변경0; 새 데이터 확대가 우선이다. **별도 fresh-domain final≥2400개는 JSONL 한 건씩 평가할 수 있고, 전부 학습 메모리에 올려야 한다는 가정이 아니다.** 이 표는 final 완성·성능·학습 준비 승인이 아니며 보호 final 읽기0이다.

근거: [projection](https://github.com/teamswyg/laya-tools/blob/d506f58ccf9629e2d2b6ca3cd1766ab20789bf64/internal/claimfit/projection.go), [Dataset/trainer](https://github.com/teamswyg/laya-tools/blob/d506f58ccf9629e2d2b6ca3cd1766ab20789bf64/internal/pairlearn/learn.go), [ranking](https://github.com/teamswyg/laya-tools/blob/d506f58ccf9629e2d2b6ca3cd1766ab20789bf64/internal/pairlearn/ranking.go). 세부 source/input SHA와 가정은 `CALCULATIONS.json`에 있다. 이 분석에서 저장 배열 계산 1회/실패0, 읽기 검색 경로 오류1회; 원래 API·학습·추론·Project·Features·benchmark·공유 수정·외부 게시0. AI-assisted 분석 비용은 측정하지 않았다.
