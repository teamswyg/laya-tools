# 71 v4 독립 영향 재검토 — 코드 준비 범위 READY

v3의 세 구현 문제를 고친 v4를 별도 `frozen-v4/`에 보존하고 변경분과 작은 합성 검사를 확인했다. 이 검토자는 학습 실행기를 작성하지 않았다. v3의 실패 지적·원문·합성 기록은 그대로 남는다. 원본 데이터의 FitWithTrace/Features/Baselines/Project/역할/원천 API를 호출하지 않았다.

고정 template은 6,730 B, SHA `708da907944294953fd7bfb3358187afbf338007fd6e32dd1c740947b177081c`, 작성자 v2 receipt는 16,449 B, SHA `e1475cca9cff3c894652ebf12d7d597707e46cab801feee167612bcd0efc648d`다. 실행 파일은 4,369,938 B, SHA `a9f1634ec71e9661b51982adda5123808b015141048d31b435bc9006b24cee29`다. 작성자 receipt의 65개 asset, 저장소 source 15개, 장부 1개의 byte/hash가 일치하고 계획의 23개 source/driver/binary 연결도 일치한다. 이전 버전의 asset들도 보존됐다. 독립 byte receipt SHA는 `f4e7b715826df6c4b3f23e67ae41afc7baea57652addf4f4b2db736463b4ae40`이다.

전용 `ContinueOnError` FlagSet은 parser 출력을 버리고 고정 문자열 `fit71_arguments`만 남긴다. 잘못된 옵션·help·누락 값·잘못된 SHA·positional 인자를 반복해도 원문이나 private 값이 나오지 않았다. 읽기 오류도 경로를 노출하지 않았다. `readPlan`은 regular 파일 크기를 먼저 검사하고 최대 1 MiB+1 byte의 reader만 사용한다. 합성 파일 0 B/1 B/정확히 1 MiB/1 MiB+1 B와 디렉터리에서 포함 경계와 거절을 확인했다. 큰 파일의 메모리나 RSS를 측정한 것은 아니다.

results와 ledger는 먼저 exclusive하게 만들고 초기 상태부터 fsync한다. control과 fit의 호출 전 시도 예약을 두 파일에 기록하며 반환한 prefix도 후속 checkpoint에 담는다. 독립 fake callback은 호출 전에 두 파일에서 fit intent 1/returned 0을 읽었고, 고정 오류 반환 후 returned 1과 raw error 비노출을 확인했다. checkpoint 실패는 fit/control callback에 들어가지 않았다. control이 한 부모의 결과를 반환한 뒤 checkpoint가 실패하면 두 번째 부모를 실행하지 않고 완료 prefix를 유지했다.

여기서 attempt는 **durable dispatch intent**다. checkpoint 실패로 attempt 1이 있어도 callback 진입은 0일 수 있다. 외부 강제 종료 때 in-flight 호출의 반환 여부·내부 epoch는 미관측이다. 두 파일은 순차적으로 제자리 갱신하므로 종료 중 부분 JSON이나 서로 다른 snapshot이 남을 수 있다. 따라서 외부 controller는 각 원문 byte/SHA와 OS 결과·decode 실패를 보존하고, 마지막 완전 snapshot이 증명하는 범위만 보고해야 한다. 실제 trainer를 죽이는 실험이나 새 자원 측정은 하지 않았다. 이 한계를 v4 작성자 한영 문서도 명시하므로 무제한 부분 호출 복원 보장을 주장하지 않는다.

설정·비교자·기준·정답·mask·calibration 정책은 그대로다. FP32 8192/seed 1729/50 epoch/batch 128/LR 0.1/L2 0.0001, primary fit 최대 1회, 원래 8 fits/16 assets/300초/soft heap 256 MiB/64 MiB payload 및 asset 범위를 유지한다. 가장 적은 확인 횟수의 고정 비교자 하나와 목록 순서의 tie 정책, A의 모든 known/no_answer 및 masked 후보, 별도 B 진단, calibration 제외, 첫 최소 validation NLL epoch 선택은 변하지 않았다. FP32 저장 32,792 B/Decode 계수 65,536 B와 float64 shadow/gradient/loss/reference 설명도 그대로다. 학습 계수는 공개 DTO에 없다.

준비 전제의 의미 연결은 여전히 부모의 frozen QA 신뢰 경계다. 실제 full76 역할·원래 정답·순서·mask·projection과 제한된 개발 목적의 준비 근거를 실행 전에 고정해야 한다. 파일 SHA 검사는 의미 정확성이나 trusted compiler의 독립 증명이 아니다. 지금 template의 false 전제와 비어 있는 projection pin은 실제 실행을 거절한다. 이 READY는 **수정된 코드의 준비 범위**에 한정하며 실제 학습 적격성, 모델 품질, 일반화, 배포나 production 승인이 아니다.

영향 race 1회는 작성자 수정 검사 2개와 독립 검사 4개, 총 top-level 6개를 통과했다. fake fit callback 2회와 fake control callback 3회만 사용했다. vet 1회와 byte verifier 1회도 통과했다. 원래 cold matrix를 불필요하게 재실행하지 않았다. v3 검사까지 합친 독립 작업은 race 2회/17 top-level 검사, vet 2회, pin verifier 2회이고 실패는 0회다. 실제 trainer·Features·Prepare·Project·Baselines·NLL·Encode·Decode·역할·원천 API·모델·새 라벨·새 역할·보호 final·공유 파일 수정·커밋·공개는 모두 0이다. AI가 읽기와 준비에 참여했으며 협업 비용은 미측정이다.
