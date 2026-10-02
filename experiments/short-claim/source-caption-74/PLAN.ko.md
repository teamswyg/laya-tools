# 공개 소스 동작 문구의 짝 비교 제안

기존 후보 v2의 짧은 upstream 문구와, 공개 소스에서 작성한 동작 요약을 **같은 3개 요청·9개 후보에 짝지어 비교**하는 준비다. 원래 요청/문구/순서/Want/null truth를 덮어쓰지 않았다. 새 문구도 아직 독립 fidelity/coverage 검토 전이며 라벨·역할·가중치·학습 적격성은 모두 미정이다.

새 영어 caption에는 모델이 구분해야 할 실제 API 경계를 담았다. 예를 들어 UnmarshalText는 syntax/bits 때 receiver0, range 때 uint64 최대를 넣고 오류를 반환하지만 MustParse는 Parse 오류를 panic으로 바꾼다. Values의 기본 slice 반복은 순서·중복을 보존하고 empty slice는 생략하며 empty string은 omitempty가 없으면 남긴다. Tokenizer.Next는 오류와 미완성 단일 token을 반환하지만 Split은 완성한 token 목록만 반환한다. 이 사실들은 pinned 공개 source/helper에서 읽었고 native Got 숫자나 답을 넣지 않았다. `CAPTION-ABLATION-DRAFTS.v1.json`은 9개 새 문구, 원래 문구, full source SHA와 26개 정확한 line/byte/span SHA 및 조건별 근거를 담는다. 새 문구 최대203B/31 standalone words이며 한도512B/32words 이내다. 원래 런타임 Normalize/Features를 실행해 검증했다는 뜻은 아니다.

`SATISFACTION-PROPOSAL.v1.json`의 별도9행은 전체 함수/위임 소스가 지원하는 조건, 충돌하는 조건, 아직 알 수 없는 조건을 제안한다. 함수가 receiver가 없거나 panic하는 점, Next가 목록이 아닌 한 단어/typed token을 반환하는 점은 소스만으로 읽을 수 있다. 과거 관찰은 직접 UnmarshalText/Values/Split의 finite probes뿐이며 wrapper·interface·helper·Next를 독립 관찰했다고 주장하지 않는다. 그렇다고 원문에 명시된 API 모양까지 모두 판정 불가로 남기지는 않는다. Source-readable 동작과 관찰된 사실은 분리하며 독립 검토자가 bounded satisfaction을 진전시킬 수 있도록 했다.

원래 문구의 모호성 두 개도 보존했다. “distinct receiver states”를 세 오류의 서로 다른 값으로 읽으면 syntax와 bits가 같은0이므로 충돌한다. Query 요청을 모든 tag/옵션/custom callback에 대한 반복 값 보장으로 읽으면 source가 지지하지 않는다. 두 경우 별도 좁은 요청 버전을 **제안만** 했으며 현재 요청을 바꾸거나 새 부모를 만들지 않았다. Interface의 실제 callback 내용이나 wrapper/adapter의 entrypoint 대체 가능성은 증거가 없으면 unknown이다. Missing evidence를 자동 음성/no_answer로 바꾸지 않는다.

실험 A는 exact upstream v2, B는 새 source-derived 문구다. 향후 독립 검토와 공통 supervision이 확정된 뒤 같은 scorer·후보 순서·whole family·역할·mask·예산·fallback을 유지하고 문구만 바꿔 비교한다. 지금 점수/Top1/checks/효용 수치를 만들지 않았다. 모델 입력은 원래 영어 요청+선택한 caption text뿐이다. Code, source SHA/path/ID, Want/Got, coverage/역할/라벨은 입력이 아니다. KR은 동일 부모의 표시 번역이며 A/B·locale·정상/오류 사례가 3가족을6/18개 독립 학습 예제로 늘리지 않는다. 기존76과 alias/helper/저작 흐름을 먼저 연결하고 새 전체 역할을 배정할 때 가족과 두 문구 버전을 같은 역할에 묶어야 한다. 공유 조직/stdlib만으로 그룹을 합치지는 않는다.

이 문구는 요청에 노출된 AI-assisted source 요약이다. 잘 작동하더라도 기존 scorer에 명시적 의미 정보를 넣는 효과이지, 독립적인 사람 저작 다양성·자동 소스 발견·2,400 fresh-domain 최종 검증 성공의 증거가 아니다. 기존 실패한 학습의 기본 활성화/fallback이나 수치·threshold를 바꾸는 계획도 아니다.

원본 API/init/observer/테스트/Features/Project/Fit/모델/HF·역할/label/가중치 배정·보호 final 읽기·공유 변경은0이다. 로컬 retained 공개 소스3개를 재읽었고 새 다운로드/획득 실패는0이다. Metadata 생성 도구는 두 compile 실패(누락 comma, 미사용 import)를 바이트 그대로 private history에 보존한 후 세 번째 command에서 처음 실행되어 JSON2개를 만들었다. 실패는 source acquisition이나 native/model 실패로 바꾸지 않았다. 정확 ledger/pin은 별도 handoff에 둔다.
