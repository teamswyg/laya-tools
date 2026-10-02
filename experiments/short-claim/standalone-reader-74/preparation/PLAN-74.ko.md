기존 76-parent projection JSON과 이미 저장된 compact blob을 각각 새 프로세스 하나에서 읽는 비교 준비입니다. 저장 roundtrip과 saved-bytes control 결과는 그대로 보존합니다. 실제 입력 파일을 읽는 worker/controller 실행은 부모와 별도 검토자의 사전 검토 뒤에만 진행합니다. 지금 단계는 소스 작성, 순수 합성 테스트, compile-only, 소스·바이너리 핀 준비입니다. 원본 API·projection·features·fit·역할·라벨·모델·유료 실행·HF 게시 0입니다.

입력은 JSON 8,390,462 B / SHA f68bff5f48747a66f038c17262568c1bd91251a8c81b3f46f7ae5090c8fe4b99, compact 1,978,694 B / SHA 43736823fa4dfa1206eaa9e939b7ab042e97be5950bc3a73b6196668db33a627입니다. 메타데이터 163,758 B / SHA c9b56cea372707068f48369f8a19f8d1a3f3d6aaefe47bea6517cbc4d77686c5는 schema/membership refs/parent metadata/nullable labels 등을 유지합니다. 네 CSR은 rows 91/45/73/44, NNZ 67104/37975/37916/37657입니다. Offsets/Groups는 signed64, Indices는 이미 uint16, Values/Labels/SampleWeights는 원래 FP64 bit를 유지합니다. 인덱스 축소나 FP32 전환, diagnostic 중복 제거는 하지 않습니다.

새 JSON reader는 기존 ImportJSON을 호출하지 않습니다. 첫 토큰 pass에서 모든 배열 길이·shape·유한 수치와 메타데이터를 확인하고, retained 메타데이터+배열 합계 64MiB 이내임을 확인한 뒤에만 CSR 배열을 만듭니다. 두 번째 pass의 원문 SHA·배열 shape·메타데이터가 같아야 합니다. 모든 JSON 객체에서 escape decoding 뒤 중복 키를 거절하고, 작은 plan/result control JSON은 대소문자 중복도 거절합니다. 기존 ImportJSON의 중복-key overwrite를 호환 동작으로 허용하지 않습니다. Tokenizer 임시 버퍼와 metadata scratch, slice headers/allocator overhead는 retained cap 밖이므로 hard RSS 제한이라고 부르지 않습니다.

compact 쪽은 동결된 기존 Decode를 그대로 쓰고, 이후 양쪽 모두 동일한 모든 배열 검증과 strict metadata 검사를 수행합니다. column별 count/null/SHA와 전체 FP64-bit/uint16/int64 digest, 정확 metadata SHA를 비교합니다. null과 []는 구분되고 -0 bit와 수치 순서를 유지합니다. 이 비교는 SHA 기반 동등성 검사이며 기존 직접 roundtrip equality 실험을 반복하지 않습니다.

실행안은 Go1.27.1 / Darwin arm64 / CGO0 / trimpath / buildvcs=false, CPU1, Go heap soft target 256MiB입니다. controller는 두 입력의 전체 핀을 먼저 읽으므로 warm-file 비교입니다. JSON 다음 compact의 고정 순서로 각 새 child 1회만 실행하고 재시도하지 않습니다. 각 child는 외부 60초 후 process-group kill과 Wait를 사용합니다. pin 읽기·Wait·후처리 때문에 전체 controller가 엄격히 120초 안에 끝난다고 보장하지 않습니다. reserve/plan/intent와 partial counters를 fsync한 후에만 Start하며, 실패·부분 출력·원시 OS log는 보존합니다. 실패값을 성공으로 바꾸지 않습니다.

OS real/user/sys/RSS/footprint는 source/binary 검증부터 읽기·검증·digest·출력까지 전체 측정 프로세스를 포함합니다. Go 마지막 HeapAlloc/TotalAlloc, retained payload, 저장 file bytes는 각각 다른 범위입니다. GPU 실행은 없습니다. 원시 log·입력 본문·private plan의 host 경로는 공개 대상이 아닙니다.

한 번의 비교는 이 호스트와 저장된 이 snapshot에서 두 reader 구현의 전체 읽기 비용을 보여줄 수 있습니다. 고정 순서·warm cache·JSON의 두 pass·메타데이터 검증·소스 핀 읽기까지 포함되므로 형식만의 인과 효과나 cold I/O, 반복 오차, 128→512 규모 확장, 학습/서빙 성능, LLM 비용 절감을 입증하지 않습니다. TrainingReady/PerformanceApproved/Final2400Complete는 모두 false입니다. 작성자는 이전 compact/projection 작업에 노출되었으므로 nonblind이며, 작성자 합성 테스트를 독립 감사라고 부르지 않습니다.
