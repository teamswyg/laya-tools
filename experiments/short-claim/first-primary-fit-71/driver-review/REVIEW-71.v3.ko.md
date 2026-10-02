# 71 v3 독립 사전 검토 — 수정 필요

검토자는 이 학습 실행기를 작성하지 않았다. 고정된 private v3 소스를 읽고 별도 복사본에서 작은 합성 입력만 검사했다. 원본 76개 투영·후보·특징·순위·학습·역할 알고리즘은 호출하지 않았다. 이것은 실행기의 기계적 사전 검토이며 문장 의미, 데이터 다양성, 모델 품질이나 학습 적격성을 새로 승인하지 않는다.

대상 계획 SHA는 `41c648c5019a2b7604326d3973186a8cf381ad301d4d51a1a6cc42f1b14c06b9`, 작성자 receipt SHA는 `1c867ffd7f84f3008f9b09d8dadf2950fde855571d0d4fa98767967e65fb6a18`다. 전체 작성자 폴더를 `frozen-v3/`에 그대로 복사해 보존했다. 45개 receipt asset, 원본 저장소 source 15개, 준비 장부 1개에서 byte 크기와 SHA가 모두 일치하고 source/driver/binary의 계획 연결 23개도 일치했다. `PIN-RECEIPT-71.v3.json`은 `880c5e90c49a3e1e695f2777328d6ab841b3a964aaf76d2404a4e0e3fecb2ed6`이다.

수정해야 하는 범위는 세 가지다.

1. `main.go:157–161`의 전역 `flag.Parse()`는 잘못된 옵션 이름을 원문으로 stderr에 출력하고 프로세스를 종료한다. 별도 합성 subprocess가 존재하지 않는 옵션 이름의 노출과 exit 2를 실제 재현했다. 계획·원본 API에 도달하지 않는다. `ContinueOnError`인 전용 FlagSet, 버리는 출력, 고정 오류로 처리하면 된다.
2. `main.go:165–167`은 `os.ReadFile`이 전체 파일을 읽은 뒤 1 MiB 상한을 확인한다. 따라서 거절되는 큰 파일도 먼저 메모리에 올라간다. 코드 읽기와 작성자 확인의 근거이며 큰 파일·RSS 측정은 하지 않았다. 제한된 reader로 상한+1 byte만 읽고 거절해야 한다.
3. `main.go:213–218, 298–307`의 results 파일은 처음에 비어 있고 before-fit checkpoint는 planned fit 1/config/headroom만 담는다. `fit.go:103`의 실제 fit 시도 카운터와 완료된 control 호출 수는 메모리에만 남다가 `main.go:257–258`에서 종료 시 기록된다. 외부 시간/RSS 제어기가 프로세스를 종료하면 부분 호출 장부를 복원할 수 없다. 정상 반환·오류·panic 때는 카운터가 보존되지만 강제 종료에는 부족하다. 호출 전의 durable checkpoint와 완료 prefix를 기록하고, 종료 중인 시도는 미완료로 남겨야 한다. 작성자도 기존 외부 controller가 복구한다고 가정하지 않았음을 확인했다. 실제 kill·fit 재현은 하지 않았다.

작성자는 위 세 항목을 인정했고 부모는 원문 v1/v2/v3를 보존한 v4 수정을 승인했다. 이 보고서는 v3가 바로 원본 학습에 투입될 수 있다는 READY가 아니다.

핵심 학습·효용 범위는 일치한다. FP32 설정은 차원 8192/seed 1729/50 epoch/batch 128/학습률 0.1/L2 0.0001로 고정되고 호출은 최대 1회다. 별도 상위 예산의 8 fits/16 assets는 자동 sweep 권한이 아니다. `pairlearn`의 계수는 채점 전에 float32로 반올림하지만 shadow 가중치·gradient·loss·reference 채점은 float64다. 저장 모델 32,792 B와 Decode 계수 65,536 B는 전체 프로세스 메모리와 다르다.

원래 정답 관점 A는 validation의 모든 known/no_answer 후보를 유지한다. 마스크 false 후보도 순위와 확인 횟수에 포함되며 no_answer는 후보 전부를 센다. unknown은 감사 분모에 남고 정답 효용에서 제외된다. B는 모든 후보의 loss 자격이 있는 부모만의 별도 진단이다. 보정 역할은 학습·epoch 선택·이번 효용에서 제외된다. 가장 강한 비학습 비교자는 총 확인 횟수가 최소인 하나이고 동률은 fixed_order→bm25→lexical_ordered→narrow_rule 순서다. Top1/Top3는 동률 선택에 쓰이지 않는다. 선택 epoch는 validation NLL이 처음 최소가 되는 epoch다. 같은 validation으로 epoch와 효용을 평가하므로 최종 일반화 증거가 아니다.

전제의 의미는 부모 freeze QA의 신뢰 경계다. `main.go:104,119–128,198–201`은 전제 pin의 byte/hash와 caller flag를 확인하지만 full76 역할·mask·출처 의미를 직접 재평가하지 않는다. 합성 검사에서도 임의 한 개의 고정 전제 pin과 true caller flag가 plan 검사를 통과했다. 이는 문서에 이미 명시된 설계이며 새로운 과학적 gate를 추가할 이유는 아니다. 실제 실행 전에 부모가 full76 역할·원래 정답·순서·mask·projection의 동일성을 연결해야 한다. `fit.go:64`와 `checkProjection`은 그 투영 안의 자기 정합성을 확인한다. 고정한 SHA만으로 입력 작성의 의미 정확성이나 trusted compiler를 증명하지 않는다.

합성 race 1회는 작성자의 pure toy 검사 8개와 독립 검사 3개, 총 top-level 11개와 subcase 8개가 모두 통과했다. 작성자 fake fit callback 9회, 독립 flag-only subprocess 1회, fake control callback 합계 9회다. 실제 trainer/Encode/Decode/NLL/Features/Prepare/Project/Baselines/역할/원천 API/모델 호출은 모두 0이다. vet 1회와 pin verifier 1회도 통과했다. 새 측정·새 라벨·새 역할·보호 final 열람·공유 파일 수정·공개·커밋은 0이다. 준비와 읽기에 AI가 참여했으며 일반 협업 비용은 측정하지 않았다.
