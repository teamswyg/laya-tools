# 공개 Go 투영 port 검토64

별도 검토자 root는 공개 `internal/claimfit` source를 원형과 비교했습니다. 알고리즘 차이는 없으며 SPDX·역사 source SHA 상수·caller binding 설명만 추가했습니다. 원형/공개 source SHA와 8개 역사 보관본은 [port 장부](PORT-LEDGER-64.v1.json) 및 archive guard로 확인합니다. 검토자는 원형/port 저자가 아니지만 계획·이전 검토에 노출되어 있으므로 blind 검토가 아닙니다.

전체 acceptable 집합, weight0 행의 물리적 보존, unknown null, calibration 제외, 원문만 받는 feature 경로와 전체 그룹의 cross-role 거절을 다시 읽었습니다. 공개 port의 원형7·이전 독립5·archive/enum2 합성 테스트를 Go1.27.1로 다시 실행해 상위14개·subtest 포함29개 pass, fail0과 종료0을 확인했습니다. Vet·whitespace 검사도 통과했습니다. 이는 이전 독립 테스트의 재실행이며 새로 만든 독립 테스트29개라는 뜻이 아닙니다.

출력 sparse 배열과 진단 view가 분리되어 있고, preflight와 두 feature pass가 구현돼 있습니다. 64MiB는 반환 payload 상한이며 caller 입력·allocator·임시 feature/정렬 배열을 합친 peak RSS 보장이 아닙니다. Private injectable feature의 동일 길이/서로 다른 유한 값 비교는 여전히 하지 않지만 exported Project는 기존 순수 feature 함수만 받습니다. 동시 입력 수정·batch 밖 graph 누락·source/정답/역할/감독 적격성의 증명은 caller 책임입니다.

현재 port를 막는 새 코드 결함은 발견하지 않았습니다. 실제 corpus의 fit 준비를 승인한 결과는 아닙니다. Corpus 투영/특징 추출·역할/seed·fit/AUC scoring·원본 행동 API·새 label·모델·가중치·보호 최종 읽기는 이번 검토에서0입니다. 준비 중 source 읽기 한 경로를 잘못 지정해 실패1회 후 실제 파일 이름을 확인했습니다. Source 비교의 종료1은 의도한 comment/provenance 차이이며 test 실패가 아닙니다. Raw private test 출력은 공개하지 않고 hash만 장부에 남깁니다. 일반 AI 보조 협업 비용은 측정하지 않았습니다.
