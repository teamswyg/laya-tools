# 37개 개발 요청과 108개 후보 라벨

작은 주장·힌트 모델이 검증 순서를 제안하도록 준비한 개발 자료입니다. 기존35행 48,618바이트를 그대로 보존하고, [실제 원본 관측](../next60-native-wire3-execution/README.ko.md)과 별도 비교기로 검증한 Union·JSON 요청2행을 추가했습니다. 새로 학습하거나 모델을 활성화하지 않았습니다.

| 항목 | 실제 값 |
| --- | ---: |
| 개발 요청 / 후보 라벨 | 37 / 108 |
| 양성 / 음성 | 37 / 71 |
| 고정 입력 | 184 |
| 전체 저장 후보 관측 / 선택 후보 관측 | 548 / 539 |
| 양성 후보 위치0·1·2 | 6 · 28 · 3 |
| 추가 요청 / 라벨 | 2 / 6 (+2·−4) |
| 프로젝트 Reader 시도·반환·전체 값 일치 | 각각37 |
| 추가 정규화 입력 대조 | 37 |
| 새 Fit / 독립 원천 / 보호 평가 추가 | 0 / 0 / 0 |

새13입력·39관측을 고정 Wanted의819조건으로 비교했고, 독립 구현의 판정이 모두 일치했습니다. Union의 캐시 커서 후보와 JSON의 객체별 해독 키 정책 후보가 각각 모든 고정 조건을 통과했습니다. 나머지 네 후보에는 구체적인 반례가 있습니다. JSON Map 후보의 미확인68조건은 그대로 남겼습니다. 이전35의 post23·선택 범위46개와 연결한 scoped114개이며 전체 과거 이력의 미확인 총계가 아닙니다. SourcePanic pre-AFTER 미확인39개도 원래 증거에 보존되지만 새 Wanted predicate 수로 더하지 않습니다.

[ROOT-QUALIFICATION](ROOT-QUALIFICATION.actual.public.v1.json)은 라벨을 채택한 근거이고, [ROOT-QUALIFIED-ROWS](ROOT-QUALIFIED-ROWS.v1.json)는 단위 가중치의 두 행입니다. [MATERIALIZATION](MATERIALIZATION.actual.public.v1.json)과 [닫은 후 읽기 검사](ROOT-POST-CLOSE-READBACK.actual.public.v1.json)는 실제 Go 생성과 전체 Reader 값 일치를 기록합니다. Reader가 의미적 정답을 새로 정한 것은 아닙니다. 특징에는 요청·후보 설명 텍스트만 허용하며 라벨·그룹·원천·증거는 입력에서 분리합니다.

37은 독립 평가 표본 수가 아닙니다. 두 요청은 기존 group79·83의 `development_train`을 유지합니다. Union은 기존 catalog19 ID를 재사용했습니다. JSON은 최상위 JSONL envelope의 중복 키를 다루는 기존 이웃 과제와 연결되어 있으며, 이번 범위는 재귀적인 객체별 검사·collapse 전 거부·순서 보존입니다. 새 독립 그룹·20→60 코호트·보호2400 평가에 더한 행은0입니다. 후보 위치 편향도 여전히 크므로 이 데이터만으로 학습 성공이나 비용 절감을 주장하지 않습니다.

사용자는 [data/train.jsonl](data/train.jsonl)을 Reader로 읽거나 아래 명령으로 같은 파일과 영수증을 재현할 수 있습니다.

```sh
bash scripts/verify-next60-thirtyseven.sh
```

Go1.27.1이 필요합니다. 임시 폴더에서 이전35와 고정 채택행으로37을 생성하고, 바이트와 Reader 입력·정답·메타데이터를 모두 대조합니다. 원본 작업이나 모델을 재실행하지 않습니다. 새 테스트를 통과시킨다고 새 Fit·보호 평가가 자동 허가되지 않습니다.

HF 공개판은 아직 [35행 고정 버전](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-35-finite-v1)입니다. 37행은 새 CI를 통과하고 별도 버전·전체 공개 파일 읽기 검증을 마친 뒤 게시할 대상으로 준비했습니다. 원본 구현·실행 파일·원시 저널·개인 경로·모델 본체를 GitHub에 추가하지 않습니다. 공개 소유 요청·짧은 설명·primitive 라벨의 범위이며, 원천 구현의 라이선스나 독립성을 확대하지 않습니다.

[이슈19](https://github.com/teamswyg/laya-tools/issues/19) · [English](README.en.md)
