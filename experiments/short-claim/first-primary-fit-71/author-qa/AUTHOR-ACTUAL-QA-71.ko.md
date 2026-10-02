# 실제 71 결과의 제작자 측 기록 확인

부모의 최초 실행은 **fit 1회 완료, 효용 검사 실패**다. 이 확인은 binder/감시기/76 worker 제작자가 저장된 바이트·결과 필드·예약·OS 기록을 읽은 좁은 검사이며 독립 모델 평가가 아니다. 원본 fitter·controller·특징·기준 도구·NLL·모델 decoder를 재실행하지 않았다.

원 결과는 71,849바이트, SHA256 `060eaa602f27f7fd9a67306e2fa2a530f0ef4c5c4548858a7818d9dfdb4bb28a`다. 동결 계획 `51b3b99904d1f757a29281e2ecc440e95839cb21f9349e65b0e89bc3d0e03d1d`는 초안 `87ed7c3a…`의 두 false를 true로 바꾼 바이트와 정확히 같다. 예약과 결과가 같은 계획·driver·기존 QA pins를 가리키고, 입력 40개/17,100,056바이트의 SHA 연결이 유지됐다.

저장된 카운터는 Fit attempt/returned/successful 각각 1, Encode/Decode 각각 1, NLL 재확인 2, 기준 도구 평가 부모 15, 캐시 점수 행 45다. 이번 fit worker에서 새 Features/Prepare/Project/roles/labels/paid/final 호출은 0이다. 이는 명시적으로 기록한 외부 호출 수이며 trainer 내부 Quantize/NLL/score 호출 수 전체를 뜻하지 않는다. Trace에는 50 epoch, train 91행/zero 18/weight sum 73, validation 45행/zero 1/weight sum 44가 저장됐다. 선택 epoch 50, reported train NLL 0.6514239277140677, validation NLL 0.6657976915843424이며 이 손실이나 점수를 다시 계산하지 않았다.

기록된 headroom 검사는 통과했지만 check reduction과 Top1 검사는 실패했고 Top3 검사는 통과했다. 따라서 완료 상태를 효용 성공으로 바꾸지 않았다. `ProductionReady`와 `PublicationQualified`는 false다. 이 확인에서는 후보 순위를 다시 만들거나 새로운 label·실험 조건을 추가하지 않았다.

private FP32 모델은 32,792바이트, SHA256 `5ecbeb9b35f67cd91de5c1674d0398e9890bae51029684f45cc955e7407c7049`이며 파일을 byte hash로만 확인했다. 계수·모델 원문을 이 기록에 넣지 않았다. FP32 직렬화와 FP64 decoder/shadow/gradient 메모리의 차이는 준비 문서에 유지돼 있다.

저장된 raw time 기록과 root ledger의 OS 단위를 대조했다. 전체 자식 RSS는 **28,295,168바이트**, peak footprint는 25,821,712바이트다. 표시 시간은 real 0.8초, user 0.07초, sys 0.02초이며 감시기 경과 시간은 0.805194625초다. 이 수치는 기준 도구·fit·trace·모델 입출력·checkpoint를 포함한 전체 자식 파이프라인이며 순수 학습 또는 실제 서비스 추론 속도 측정이 아니다. RSS는 기존 256 MiB 관측 예산 이내였고 Go heap soft cap을 hard cap으로 해석하지 않았다.

제작자 metadata checker는 총 2회 실행했다. 첫 실행은 driver/repository에 같은 `go.mod` ID가 존재하는데 ID만으로 경로 루트를 정한 checker 오류로 실패했다. 원 source와 private failure log를 보존하고 ID와 Path를 함께 비교하는 별도 v2에서 1회 성공했다. 실패는 실제 fit 실패나 model 재실행이 아니다. 두 실행 모두 원본 API 호출 0이며 성공한 1회가 이전 실패 기록을 지우지 않는다. 확인 결과 JSON은 `AUTHOR-ACTUAL-QA-71.v1.json`, SHA256 `69cd3bd5b0a06544850611a70c205bbb364e3a79ea3410b7deac22ca14d3141b`다.
