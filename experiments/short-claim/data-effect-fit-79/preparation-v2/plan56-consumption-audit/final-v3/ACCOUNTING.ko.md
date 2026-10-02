이번 기록은 원 PLAN56의 실제 소비와 현재 보존된 파일을 읽어 대조한 결과입니다. 새 학습·모델 Decode·Score·Features·Project·역할 배정·외부 호출·삭제·hardlink는 모두 0입니다. 기존 자료와 결과를 본 비맹검 메타데이터 검토이며 모델 품질이나 일반화 승인이 아닙니다.

실제 과학적 fit은 71과 72의 각 1회, 합계 2회입니다. 서로 다른 모델은 2개(각 32,792 B)이며 원 실행·HF stage·readback의 6개 물리 파일은 합계 196,752 B입니다. 같은 결과를 복사한 13개 파일, 호출 수를 복사한 ledger 5개와 author QA 4개를 추가 학습으로 세지 않았습니다. 두 실행 모두 동일한 projection 76 SHA를 사용했습니다. 확인된 범위에서 별도 INT8/PTQ/STE 파생 모델은 0입니다. 기존 8 fit/16 artifact 한도에서 잔여 6/14는 실행 자격이 아닙니다. 미보고·삭제·원격 history까지 완전하다고 증명하지 않습니다.

이전 semantic-learning/semantic-scale의 16개 서로 다른 hbin export와 40개 물리 복사본은 각각 다른 동결 계획·입력·원천에 연결됩니다. path-cost46의 모델 metadata 10개도 다른 프로토콜입니다. 합성 unit-test 학습, 원격 업로드, tag, readback 또는 collection 항목 수를 PLAN56 학습 수로 합산하지 않았습니다. HF 71/72 공개는 실패 연구 archive 9개 파일씩이고 HF 69는 dataset 20개 파일입니다. 이 검토에서는 HF를 호출하지 않았고 기존 영수증만 확인했습니다.

현재 projection 76은 5개/41,952,310 B, projection 69는 3개/23,099,430 B입니다. 각 파일은 서로 다른 inode, nlink 1인 regular file이고 같은 device에 있지만 owner-writable 상태입니다. 원래 canonical 76은 보존했습니다. 실제 경로·mode·device·inode·nlink는 공개하지 않는 별도 binding에 있습니다. 동일 SHA나 향후 hardlink를 이유로 경로별 bytes를 할인하지 않습니다.

8개 projection, 모델 6개, 기존 combined70 입력 1개만 합쳐도 67,943,460 B입니다. 기존 64 MiB(67,108,864 B)보다 최소 834,596 B 큽니다. 최종 보수적 JSON/JSONL/hbin registry는 2,776 경로, 206,707,066 B이며 각 경로의 size·SHA·regular-file identity를 다시 확인했습니다. 초기 selector는 72 stage/readback 일부를 빠뜨려 2,768 경로/206,581,572 B였고, 그 실패 소스와 partial registry를 보존한 뒤 별도 버전으로 수정했습니다. 현재 registry에는 범위 안의 계획·숫자 검토·출처 영수증·JSONL test-report 지원 파일도 포함했습니다. 코드/Markdown/license 텍스트, 실행 바이너리, 비JSON 원시 로그와 무관한 다른 프로토콜/cache는 제외합니다. 이 제외와 이후 새 파일은 명시적 accounting 경계이지 자동 무료 저장이 아닙니다.

PLAN56 원문은 보존 footprint와 lifetime bytes-written를 구분하지 않았습니다. 반복 checkpoint의 lifetime 쓰기량·원격 실제 할당량은 미측정입니다. 현재 경로별 quota, inode payload, 동일 내용 SHA 합계, 물리 allocated blocks는 서로 다릅니다. 원문을 소급 재해석해 과거 64 MiB PASS로 만들지 않습니다. 현재 SHA가 같은 파일의 내용 합계를 보는 값은 설명용이며 quota 차감에 사용하지 않습니다.

512 MiB 경로별 저장 상한을 공개 versioned amendment로 명시하는 안은 계산상 가능하지만 아직 이 검토가 enact하거나 실행을 승인한 것은 아닙니다. 현재 envelope + 새 projection 64 MiB × 3 + 결과/ledger 8 MiB × 2 + 모델 32,792 B = 424,843,666 B입니다. 512 MiB까지 112,027,246 B가 남습니다. 이 감사의 사후 snapshot 밖 metadata overhead에 별도 16 MiB 예약을 제안하면 잔여는 95,250,030 B입니다. checkpoint 임시 overlap, readback, 지원자료와 이후 새 경로도 쓰기 전에 별도 합산해야 합니다. 결과와 ledger가 각 64 MiB를 실제로 쓰도록 남겨두면 더 넓은 예약이 필요하므로 8 MiB cap의 사전 계획 고정이 필요합니다. lifetime 상한을 만족했다는 주장은 하지 않습니다.

이 안은 seed·loss·평가·fit 8/artifact 16·한 실행 sparse payload 64 MiB·CPU 1·heap soft 256 MiB·RSS 256 MiB·300 s를 그대로 둡니다. 저장 조건 초과를 숨기지 않으며 새 자료/결과 예약이 넘으면 해당 쓰기/실행을 멈춥니다. 24 GiB RAM과 약 474 GiB 여유 디스크는 Parent가 보고한 환경값이며 이 감사가 다시 측정한 자원값은 아닙니다.

주요 기록은 [소비 registry](FIT-ARTIFACT-CONSUMPTION.v1.json), [보존 경로별 공개 registry](RETAINED-SERIALIZED-REGISTRY.v1.json), [예산 비교](STORAGE-ACCOUNTING.v1.json), [기존 근거 핀](EVIDENCE-PINS.v1.json), [장부](REVALIDATION-LEDGER.v1.json), [counter 복사 분리](COUNTER-ALIASES.v1.json), [receipt](RECEIPT.v1.json)입니다. 준비 문법 오류 1회, metadata selector 실패 1회, 수정 후 성공 1회를 보존했습니다. 성공 검사 stdout 한 필드의 stale 산술 상수는 [별도 교정](DISPLAY-CORRECTION.v1.json)에 남겼으며 실제 numeric JSON·assertion은 올바른 값입니다.
