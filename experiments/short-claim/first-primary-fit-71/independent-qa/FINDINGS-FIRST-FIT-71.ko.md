첫 FP32 fit은 실행에 성공했지만, 기존 비교 방식보다 효용이 나빴다. 독립 검토는 이 실패를 그대로 확인했다. 같은 validation A의 15개 알려진 요청(답 있음 10·no_answer 5)에서 가장 좋은 `lexical_ordered`는 27회, 모델은 31회 확인했다. 확인 횟수는 14.8148% 늘었고 Top1은 5/10에서 2/10으로 감소했다. Top3는 둘 다 10/10이다. Oracle 21회 대비 개선 여지 22.2222%는 있었지만, 실제 모델의 기존 효용 기준은 통과하지 못했다. 제외된 unknown 5개는 validation의 원래 분모와 기록에서 유지한다.

저장 파일·수치 검토 1회가 성공했고 실패는 0이었다. 원본 API를 다시 실행하지 않고 stdlib로 FP32 파일 32,792 B의 8,192개 finite 계수·header·SHA를 직접 읽었다. 저장 sparse 배열에 대해 자체 dot 계산 181행(최종 NLL 136행과 validation 점수 45행), 대체 binary cross-entropy 계산 2회, 저장 순위 75행 및 모델 순위 15행을 확인했다. 저장 모델 점수와 차이는 0이었다. 최종 train NLL `0.6514239277140677`, validation NLL `0.6657976915843424`가 일치한다. 50개 저장 epoch의 순서·유한 수치·최초 최소 선택이 일치하며 선택 epoch는 50이다. 각 epoch의 중간 계수나 optimizer 상태를 재구성한 검증은 아니다.

전체 76 요청·226 후보·38 answerable·17 no_answer·21 unknown/63 null을 유지한다. Train 91행 중 18개와 validation 45행 중 1개는 zero loss weight지만 원본 후보·정답·효용 분모에서 제거하지 않는다. 가중치 합은 73/44이며 calibration과 unknown은 fit에 들어가지 않았다. 추가 4개 요청의 source68 supervision은 고정된 유한 관측 범위의 proposed 자료로 구분하며 새로운 일반화 증거로 승격하지 않는다. 원래 정답 A의 후보 전체를 기준으로 비용을 비교하고 완전한 caption B는 별도 진단이다. 원래 seed·mask·그룹·역할·5% 기준은 변경하지 않았다.

Root의 기존 관측은 전체 child RSS 28,295,168 B, controller wall 약 0.805초, 반올림된 user/system CPU 0.07/0.02초다. 새 자원 측정이나 원시 OS 로그 재해석은 하지 않았다. 모델 파일 크기와 decode한 float64 계수 payload 65,536 B는 전체 RAM과 다르다. 학습의 shadow/gradient/loss는 float64이며 FP32 파일을 native FP32 연산·GPU 사용·반복 hint 지연 개선·LLM 절감으로 표현하지 않는다.

검토자는 현재 71 driver·binder·controller·trainer 작성자가 아니지만 이전 70 metadata 작성과 자료·결과 노출 이력이 있는 AI 보조 비블라인드 검토다. 원본 Fit/Features/Baselines/Project/역할/원천/모델 API 재호출, 새 라벨·역할·학습·가중치·공유 변경·공개·protected-final 읽기는 모두 0이다. 자체 수치 검산은 위 횟수로 별도 계수했다. 같은 validation으로 epoch를 고르고 효용을 평가했으므로 holdout 일반화나 재현성 증명은 아니다. `ProductionReady=false`, `PublicationQualified=false`, utility 실패를 보존한다.

정확한 결과 SHA `060eaa602f27f7fd9a67306e2fa2a530f0ef4c5c4548858a7818d9dfdb4bb28a`, 모델 SHA `5ecbeb9b35f67cd91de5c1674d0398e9890bae51029684f45cc955e7407c7049`, mechanics SHA `c48999690d11342dbe7f9688707492bb5b6b71d3dd306ebb9bc83391a9b2f196`를 receipt와 연결한다. 보고서는 계수 본문·private 경로·원시 stderr/profile을 포함하지 않는다.
