# 새 개발 자료의 실제 동작부터 확인하기

두 모델이 확인 비용을 줄이지 못했으므로, 다음 학습 전에 서로 다른 공개 코드의 동작 자료를 확보한다. 첫 단계는 datasize의 단위·오류 후 상태, query.Values의 태그·다중값, shlex.Split의 인용·불완전 입력이다. 사람이나 에이전트가 후보 코드를 고를 때 놓치기 쉬운 구체적인 조건이다.

**3개 동작 목표를 24개 고정 입력으로 검사한다.** 24개를 독립 요청 24개나 새로운 학습 부모 24개로 세지 않는다. 기존 76개 개발 자료와 별도로 준비하며, 도메인별 2,400개 최종 요청 요건을 충족했다고 주장하지 않는다.

[사전 Want](wants.v1.json)는 실행 전에 고정한 정확한 정수·오류·반환 상태다. [작성자 인계](FIRST-THREE-OBSERVER-HANDOFF.v1.json)는 24개 입력과 Want, 원천 revision·SHA, 빌드·준비 횟수를 포함한다. 원천 10파일·39,013B의 라이선스와 모듈은 그대로 보존했고 [import 빌드](FIRST-THREE-SOURCE-BUILD-HANDOFF.v1.json) 1회가 통과했다. 관찰기 작성자의 순수 stub 테스트·vet·컴파일도 각각 1회 통과했다. 원본 API와 관찰기 실행은 작성자 준비에서 0회다.

다른 검토자는 관찰기 Want를 읽기 전에 [독립 원문 메모](ORACLE-NOTES.ko.md)와 [receipt](SOURCE-REVIEW-RECEIPT.json)를 고정했다. 기대값은 공개 원문의 유한 동작을 해석한 결과이며, 실행 결과를 보고 새 정답으로 바꾸지 않는다. 전체 API 의미, 후보 설명, 학습 정답과 `training_ready` 승인은 별도 문제다.

[24개 Want의 독립 검토](independent-want/FINDINGS.ko.md)는 실행 전 의미 blocker0이었다. [receipt](independent-want/RECEIPT.json)와 [수치·핀 확인](independent-want/MECHANICS.json)은 큰 uint64와 21개 closure, 관찰기 바이너리의 파일 SHA·빌드 정보를 원본 실행 없이 확인한다. 이는 고정 유한 기대값과 관측 변환에 한정되며 넓은 의미·학습 적격성 승인으로 확대하지 않는다.

[실행 경계의 독립 검토](OBSERVER-MECHANICS-RECEIPT.v1.json)에서 차단 사유가 없었고, root가 고정 원문·Want·바이너리·controller를 읽어 확인한 뒤 **실제 child를 한 번 시작했다**. [결과](results.json)는 직접 API 예약24·반환24·일치24·불일치0·panic0이다. [Root 장부](ROOT-ACTUAL-LEDGER.v1.json)는 child1·retry0·timeout0·종료 코드0과 최종/부분 결과가 같은 SHA임을 보존한다. 원래 Want와 작성자 준비 기록은 변경하지 않았다.

외부 60초 제한·CPU 1개·Go soft heap 256MiB 설정에서 전체 child peak RSS는18,743,296B(약17.9MiB), footprint16,122,360B, OS real/user/sys는0.85/0.01/0.04초였다. 원본 초기화·pin 확인·wrapper·checkpoint 저장을 포함하고 parent controller는 제외한다. 이것은 모델 추론 속도나 GPU 메모리 측정이 아니다. 각 직접 API 호출 전에 예약 카운터를 저장했다. 프로세스 시작 시 원본 Go package 초기화가 발생했으므로 작성자 준비의 초기화0회와 구분하며, 초기화 종류와 내부 호출 수를 동적으로 계측한 것은 아니다.

[저장 결과 대조](saved-result-qa/RESULT-QA.ko.md) 1회도 통과했다. 큰 uint64를 float64로 바꾸지 않고 24개 Probe·Want·Got·Matches와 final/partial·plan·OS 장부를 대조했다. 오류8개는 예상된 오류이며 API 호출24개가 모두 정상 성공했다는 뜻은 아니다. 검토자는 controller 작성자로서 저장 파일을 확인했으며 원본 API를 다시 실행하지 않았다.

유한 관찰의 일치가 새 학습 부모·후보 정답·caption 승인은 아니다. [3개 부모×3개 후보 설계안](candidate-proposal/PROPOSAL.ko.md)은 실행 결과 요약 노출 전에 고정한 v2다. 원문 선언 순서로 후보를 고정하고 KR/EN을 같은 부모로 묶었다. 짧은 caption에 필요한 오류·부분 반환 조건이 없어 모든 coverage는pending, 정답·역할·가중치는null이다. 실제 새 부모·정답·모델·학습·최종 자료 읽기는 모두0이고 `training_ready=false`다. 다음에는 caption 조건과 후보 만족 근거를 별도로 검토한다. 기존v1과 실패 준비 helper의 SHA는 인계서에 보존했고, 호스트 경로가 있는 helper는 공개하지 않는다.

관찰기 원문은 [spec](observer-spec.go.txt)·[계획 타입](observer-plan.go.txt)·[순수 테스트](observer-spec-test.go.txt)·[실행 wrapper](observer-main.go.txt) 텍스트로 보존한다. 이는 실행 소스의 읽기용 archive다. 로컬 절대 경로가 있는 module·준비 helper·실행 계획, 바이너리, 원시 로그는 공개하지 않는다. 공개 Want와 원천 핀은 유한 사례 검토에 사용할 수 있지만 이 폴더만으로 실행 설치가 완료되지는 않는다.

[다음 원천 목록](../development-expansion-73/README.ko.md) · [메모리 검토](../memory-scale-73/ANALYSIS.ko.md) · [English](README.en.md)
