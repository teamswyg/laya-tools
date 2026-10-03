# 다음 60개를 위한 원문 계약 확장 제안 v2

기존 13개 가설에 원문으로 뒷받침되는 변경 가설 9개를 더해 **22개**를 준비했습니다. 목표 30개에는 **8개가 부족**합니다. 새 관측·실행 가능한 독립 요청·자격 부여는 모두 0입니다. 가설 수는 실제 데이터나 골드 부모 수가 아닙니다. 이전 18개 중 기존 기능으로 확인되어 제외한 5개와 모든 원 봉인은 그대로 유지했습니다.

[계약 JSON](TASKS.proposed.v2.json)은 기존 13개 객체를 보존하고 새 9개의 짧은 요청·유한 입력 프로필·관찰 채널·후보 캡션 제안을 담습니다. Want/Got/라벨/역할/가중치/숫자 그룹은 미정입니다. 새 후보 코드는 아직 없으므로 캡션은 코드 충실성 승인이 아닙니다. [제외 기록](DISCOVERY-REJECTIONS.v2.json)과 [관찰 전 균형안](PRE-OBSERVATION-BALANCE.proposed.v2.json)을 함께 읽어야 합니다.

| 추가 변경 가설 | 원문 차이와 유한 관찰 |
|---|---|
| 비트 추출 목적지 교체 | 기존 OR·길이 확장과 달리 기존 비트 제거·정확한 packed 길이; 별도 목적지의 전후 상태 |
| 정확한 decimal 나눗셈 ties-to-even | 기존 DivRound의 ties-away와 중간 반올림을 구분; 정확한 몫·나머지 기반 |
| 호출별 decimal JSON 정책 | 전역 quoting/zero/notation 변경 없이 호출별 옵션·전역 전후 상태 |
| 엄격 JSON-lines prefix | 한 줄 한 값·실패 줄·이전 callback 보존·이후 호출 중단 |
| 잘못된 UTF8 거절 append | 기존 U+FFFD 치환과 달리 오류 전 목적지 불변; 유효 UTF8 escaping 유지 |
| parser별 modifier registry | 기존 전역 resolver 대신 소유 등록 스냅샷·두 parser 격리 |
| uint256 → binary64 반올림 | 기존 다중 word fraction 버림과 ties-even·carry를 bit 단위로 구분 |
| full512 곱의 ceiling MulDiv | 기존 floor와 비영 나머지 increment·overflow carry를 구분 |
| xxhash short-state 정합성 | magic/크기뿐 아니라 seed lane 관계·total·unused tail 검사 |

GJSON의 Escape·UTF8 치환·owned byte 결과, Chi의 Clone·문서상 GetHead, uint256의 SSZ·과학 표기 overflow 검사는 이미 있습니다. 이들을 새 요청으로 세지 않았습니다. RLP short-write 수정은 원문 근거가 있지만 이전 Cobra writer 계약과의 의미 중복을 확인해야 하므로 22개 밖에서 보류했습니다. Chi panic 후 pool 반환은 관찰 가능한 신뢰 경계가 없으므로 일반 RAM 개선이나 유한 실패를 주장하지 않았습니다.

확보된 여섯 저장소를 새 미노출 독립 가족이라고 단정하지 않습니다. 기존 7개 계보·의존성 edge, 누락된 bitset generated 파일, GJSON match/pretty 원문·notice, Go BSD/Intx/evmone/radix 정확한 기원 범위가 여전히 보류입니다. 같은 원천·fork·helper·alias는 whole connected family로 처리하며 공통 Go 저작권이나 stdlib 사용만으로 자동 합치지 않습니다. 이전 3 train/2 validation/1 calibration 분배는 조건부 제안이며 실제 역할이나 숫자 그룹이 아닙니다.

권장 첫 작은 준비는 uint256 Float64 반올림, full-product ceiling, 비트 목적지 교체입니다. 원문 차이가 분명하지만 각각 정확한 독립 oracle·literal Want·full source/notices/compiler closure를 먼저 봉인해야 합니다. xxhash 상태 조립, decimal 전역 옵션 관찰, GJSON resolver fork는 별도 ABI·입력 정책이 더 필요합니다. 컴파일/관찰 실행은 이 제안이 허용하지 않습니다.

후보 위치는 여섯 permutation을 미리 배정하는 제안이며 어떤 슬롯도 reference/positive로 지정하지 않습니다. 문체와 길이는 정답과 독립적으로 맞추고 코드와 캡션의 범위를 먼저 확인합니다. 모델에는 요청·캡션만 들어가며 IDs/slot/Want/Got/역할/자격 메타데이터는 배제합니다. unknown-only는 negative가 아닙니다. A/B 문장·fixture·근접 변형은 새 부모가 아닙니다. 현재 30개 자료를 재정렬하거나 고치지 않습니다.

작성자는 기존 계약·어댑터 작성 이력이 있는 비맹검 source reader입니다. 114는 uint256/xxhash 미사용 API의 읽기 의견만 제공했고, 이는 실행된 독립 oracle 검증이 아닙니다. 새 HTTP/Go/native/model/학습/보호자료/HF/Git 호출은 0입니다. 원문 body는 이 폴더에 복사하지 않았고 공개하지 않습니다. [입력 핀](INPUT-REFS.v2.json)과 [장부](ATTEMPT-LEDGER.v2.json)에 역사 기록을 연결했습니다.
