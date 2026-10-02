# 원본72 첫 실제 투영 — 작성자 배열 감사

첫 실제 투영은 1회·재시도0·종료0으로 완료됐다. 이는 기존 Go feature와 학습 입력을 만든 실행이며 모델 학습이나 Laya 모델 추론이 아니다. 작성자가 별도의 표준 Go 전용 checker로 원본 저장 JSON과 실제 결과 배열을 읽기 전용으로 대조했다. 이 검토자는 69/71 드라이버 작성자이므로 독립 심사자가 아니며, 부모와 별도 독립 검토자의 확인은 분리한다.

요청72개·후보216개, known34·no_answer17·unknown21, 전체 그룹17개가 유지됐다. 원문 요청·후보 문자열과 순서, acceptable 배열, 전체 그룹 역할, nullable 라벨, mask와 감사 기록이 원래 핀에 일치한다. normalized 문자열은 저장된 길이·단어 수와 대조했으며 정규화 함수를 다시 호출하지 않았다.

학습 입력은 train90행과 validation36행이다. 원래 라벨을 유지한 가중치0 행18개·1개도 물리적으로 남았다. 미확정 후보는 역할별42·12·9개이며 fit 행이 없다. calibration의 알려진 후보27개도 fit에 들어가지 않았다. 진단 뷰는 양수 가중치72행·35행만 포함하고 `Excluded=0`이며 미확정 수는 별도 분모에 정확히 남는다. feature index의 범위·행 내 중복·유한 float와 sparse 열의 shape를 검사했으며 진단 열이 원래 열의 올바른 부분집합인지 확인했다. feature 의미나 값을 재추출한 검증은 아니다.

원본 실행 기록은 direct Validate72회, Project1회, 반환된 feature scan252회다. Project 내부 정규화 횟수는 별도 계측하지 않았다. 이번 감사는 원본 Validate/Project/Features/Assign/Fit/AUC/BM25/모델 호출 모두0회다. 감사 checker 실행1회·실패0이며 원본 실행은 반복하지 않았다.

소유한 준비 payload는 **1,774,364B**로 Go ABI와 저장 배열·문자열에 대한 기존 계산식을 다시 대조해 일치했다. JSON 원본은 들여쓰기와 반복 진단 열 때문에 **7,699,810B**다. 외부 제어기가 측정한 전체 자식 프로세스 peak RSS는 **56,573,952B**, peak memory footprint는 **49,857,016B**, controller wall은 **0.465949875초**다. 표시된 real0.46/user0.07/sys0.01초는 반올림된 OS 측정값이며 feature 함수만의 비용이 아니다. 준비 payload, JSON 파일 크기, OS RSS는 서로 다른 측정이다.

공개용 compact derivative는 **97,255B**다. 원본 SHA `348af320a8bdcbee433f3102cc72aba06404ddc568a9ae9ddf6ff8f8d2980a4f`와 크기를 보존하고, 요청/후보 텍스트 해시·원래 감독 정보·행 참조·feature 수·행별 canonical `{Indices,Values}` SHA를 남긴다. canonical 규칙은 Go1.27.1 `encoding/json.Marshal`과 마지막 LF이며 RFC8785를 주장하지 않는다. 큰 feature 배열을 생략한 요약이므로 자체만으로 모델 학습을 재현할 수 없으며 원본을 대체하지 않는다. 모델 계수는 없다.

원본 result와 compact/감사 문서에는 로컬 절대 경로가 없다. 고정 실행 계획의 입력·소스·출력 경로는 승인된 private namespace의 재현 참조이며 해당 원본 계획은 공개하지 않는다. 모델·유료 API·benchmark 프로세스0은 이 AI 보조 검토의 에이전트 사용이나 비용0을 뜻하지 않는다. 협업 비용은 측정하지 않았다. 이 결과는 새 학습 준비·출처 다양성·효용·일반화 승인이 아니다.
