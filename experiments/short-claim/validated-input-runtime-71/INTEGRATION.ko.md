# 전체 회귀 검사와 원본 보존

입력을 한 번 검증하는 최적화의 여섯 Go 파일은 작성자 커밋 `5aab32e8a0695815e2d10e8a875e13d459f30e61` 이후 동일하다. 추가 수정은 과거 소스 핀을 검증하는 회귀 테스트와 정확한 원본 텍스트 archive에 한정한다. Runtime 검증기, 과거 계획·결과·기준값은 변경하지 않았다.

부모의 전체 race 검사 세 번은 실패했다. 초기 출력 일부가 잘려 첫 두 실행의 실패 목록 전체를 확정하지 않는다. 확인된 실패는 과거 소스 해시와 변경된 현재 소스를 같은 것으로 비교하던 검사들이었다. behavior/property/typed 회귀 세 개, 이후 resident 측정 기록, scope 준비 및 stored utility 회귀를 차례로 수정했다. 실패 실행을 성공으로 바꾸어 기록하거나 원래 측정을 다시 실행하지 않았다.

과거 source/support는 `9d204c2c700505658c108297d5fa769a835a60c3`의 정확한 Git blob 일곱 개로 SHA와 바이트 수를 확인한다. 현재 compiled source와 source closure는 별도로 엄격히 검증한다. 현재 수치 재현은 원래 점수 허용 오차·후보 순서·정답·비용·fallback·그룹·상태 비교를 유지한다. 실제 공개 CLI는 현재 소스와 맞지 않는 과거 실행 계획을 계속 거절한다. 역사 archive를 runtime fallback이나 feature로 사용하지 않는다.

독립 읽기 검토는 여섯 최적화 파일의 동일성, 일곱 역사 기록의 원본 동일성, 추가 회귀 세 개와 새 archive 두 개를 확인했고 blocker가 없었다. 중복 학습·benchmark를 실행하지 않았다. 작성자 대상 race/vet 검사들은 통과했다. 부모의 네 번째 전체 `go test -race ./... -json`도 종료 코드0으로 통과했고, 모든 JSON 실패 이벤트는0개다. 전체 vet·owned-source gofmt·diff 검사도 통과했다. 최종 GitHub CI는 별도 merge gate다.

정렬 할당 81→0이라는 한 번의 관측은 보존한다. 프로파일 설정이 달랐으므로 속도·RSS 개선률을 확정하지 않으며, 종료 Go heap 증가도 유지한다. 이번 회귀 수정으로 새 benchmark, 공식 학습, 새 정답, 역할 재배정, 유료 호출은 추가하지 않았다.

후속 정렬 학습 코드와 PR94 문서를 함께 통합한 뒤 전체 race 검사 다섯 번째도 통과했다(80패키지·test/subtest 1,873개·실패0). 전체 vet도 통과했다. README 문맥을 잘못 지정한 문서 patch 두 번은 적용 전에 거절됐고 수정한 patch만 적용했다. Runtime·원본 결과·학습 횟수에는 영향이 없다.

[사용과 측정](README.ko.md) · [초기 대상 회귀 기록](FINDINGS.v1.ko.md) · [English](INTEGRATION.en.md)
