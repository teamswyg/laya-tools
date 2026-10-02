# 요청75 v4의 범위·보존 검토

검토 범위에 남은 material blocker는0이다. FINAL-HANDOFF.v2(5462ea…)와 v4(e1290b…)는 v3(d3bc8988…)의 부모·후보·순서·요청·A/B 설명·code satisfaction·마스크·source/license metadata를 보존하고 pair 기여 설명만 정정했다. 작성자 v1/v2/v3와 각 invocation 핀이 그대로이며 v3 peer receipt(c9c5df…)와 연결된다. 부모3개·후보9개 모두 실제 truth/acceptable set/role/group/weight/pair 자격은 null이다. 새 membership·역할·정답 배정은 없다.

원본3파일 전체를 읽고 고정 SHA 및26개 근거 span의 byte/line/SHA,9개 candidate declaration과9개 A 원문 span을 확인했다. caption74 catalog는 그대로다. 요청은173/237/134B와27/32/21 identifier-split words이며 A/B18개 설명도 기존512B/32word 내에 있다. 이 단어 계산은 stdlib byte 연산이고 원래 Normalize/Validate/Features 호출이 아니다. 새 설명 B는 AI-assisted 요약이며 upstream의 독립 인간 저작이 아니다. 같은3부모의 새 요청 버전·A/B·번역을 추가 독립 요청으로 세지 않는다.

고정 원문에서 좁은 직접 선택 API의 충족 제안은 UnmarshalText·Values·Split의3긍정, 다른6후보의 직접 형태/정책 불충족이다. unseen adapter나 Encoder 인터페이스의 미관측 구현을 가정하지 않는다. datasize는 nonnil receiver에서 ASCII digit prefix/binary unit branch와 syntax/bits→0, range→uint64 max의 반환 상태를 지원한다. 빈 입력 거부나 모든 정수 표현·nil receiver 안전성을 새로 주장하지 않는다. Parse의 local receiver와 MustParse의 panic은 직접 요청된 receiver/returned-error 조건을 충족하지 않는다.

query는 exported/nonignored/default primitive field·slice, custom encoder 제외라는 요청 맥락이다. 원문은 omitempty/IsZero에 따른 생략을 중첩 재귀보다 먼저 검사한다. 따라서 **included named nonnil non-time struct**에 한정한 bracket scope 주장은 지원된다. anonymous flattening, nil pointer, time formatting, custom callbacks나 delimiter/numbered 등의 모든 옵션을 포괄하지 않는다. B의 짧은 중첩 설명을 요청 조건 없이 보편 주장으로 승인하지 않는다. shlex는 기본 string splitter가 빈 따옴표 토큰을 누적하고 닫히지 않은 quote/escape 오류 전의 완료 목록을 반환하는 경로를 지원한다. Next 두 종류의 단일 반환값을 누적 목록과 같다고 처리하지 않는다. 외부 reader의 모든 오류·panic 부재를 입증한 것은 아니다.

소스가 요청을 충족하는지와 모델에 보이는 설명이 충분한지는 별개다. source-SAT 제안3긍정·6부정은 전체9후보 축이고, 설명 A의 적격 제안은0긍정·4부정, B는3긍정·6부정이다. 정보 생략/unknown을 새 negative truth로 바꾸지 않는다. 공통 교집합은 긍정0으로 nonzero pair 제안도0이다(A0/B6/common0). 이는 실제 pair 생성이나 fit 수치가 아니다. 이 작은 slice만으로 학습 가능한 paired 비교를 선언하지 않는다. 평가를 한다면 어려운 A 후보를 분모에서 삭제해서 성능을 높이지 않고 고정 source-code acceptable 제안 전체를 보존하는 범위다.

현재72의 정정은 정확하다. source531ce807…의 pair weight는 양·음 endpoint SampleWeights 곱이며0이면 gradient와 NLL에서 skip한다. 고정 learn e25ca8…와 saved driver fit fca4a111…/main40efd681…의8개 span도 같고, 실제 frozen plan17afbf…가 일치했다. driver는 동일 projection datasets와0/1 weights를 유지하므로 audit에 남긴0쌍이 학습에 기여했다고 해석할 수 없다. 독립 endpoint policy는 미래 설계 개념이며 현72의 동일0가중치 처리를 위해 새 구현이 필수인 것은 아니다. 역사72의16/2 audit 수치는 기존 root/peer 결과 기록의 참조이고 여기서 재계산하지 않았다. 기존 코드·결과·모델은 변경되지 않았다.

다음 고정71/72모델의 A/B diagnostic을 설계하는 데 이 scoped source-supported rubric을 사용할 수 있다. 같은3요청·9후보의 A/B는18개 입력 위치이며 두 모델 결과도 독립 부모 수를 늘리지 않는다. 기존 모델을 고정하고 fit 없이 설명 정보에 따른 출력 차이를 보는 탐색은, 새 학습 가능한 A/B cohort 주장과 다르다. 아직 실행·점수·성능 결과는0이다. 새 요청 scope/version, corpus membership와 역할이 pending이고 prior Want/결과/실패 연구에 노출됐으므로 진짜 독립 heldout·보편 code truth·일반화·LLM 절감 평가라고 부를 수 없다. 이 검토는 새 역할 배정, source-diversity clearance, global training-ready나 최종2,400 평가 승인을 만들지 않는다.

검토자는 request75/caption74/v4 작성자가 아니지만 이전 native observer·source/projection 준비에 참여해 nonblind이다. 첫 stdlib metadata checker는 존재하지 않는 request field 이름을 참조한 제 오류로 실패했고 초기 source를 history에 보존했다. 정확한 필드로 고친 둘째 시도는 통과했다. 이는 artifact 오류나 원본 실행 실패가 아니다. 원 API/init/observer/AST/Features/Decode/Score/Project/Fit/roles 실행, 저자 파일·공유 수정, 외부 게시는0이다.
