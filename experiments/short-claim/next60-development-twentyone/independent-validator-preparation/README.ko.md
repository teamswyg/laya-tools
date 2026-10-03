# 21개 개발 자료의 독립 reader 검증 준비

현재는 검증기 소스와 합성 검사만 준비한 단계입니다. 실제 21개 자료와 공개 Go reader는 이 준비에서 실행하지 않았습니다. Root가 자료를 생성하고 해시를 고정한 다음 소스를 검토해 별도 실행합니다.

기존 16개 파일의 모든 바이트와 마지막 LF가 새 파일의 접두부여야 합니다. 새 5행은 봉인된 Wanted의 요청·후보 설명, fixture의 Wanted·ID, Root 채택 기록, 저장 비교 보고서를 각각 대조합니다. 실제 reader 반환 객체의 요청·후보 ID·원문·라벨·가중치·메타데이터·개수·미사용 슬롯도 독립 파싱 값과 대조합니다. Valid() 통과만으로 완료하지 않습니다. Prepared()는 이미 읽은 값의 getter이며 Features·Score·Project·Fit 호출이 없습니다.

새 후보 수는 3/2/3/3/3입니다. 오류 writer의 기존 후보는 관측 불가여서 제외됩니다. 남은 reference는 학습용 위치 0이지만 원래 위치는 1이고 metadata_id=reference-design을 유지합니다. known false와 unknown이 함께 있는 음성은 알려진 반례로 채택될 수 있으나 unknown 위치를 지우지 않습니다. 양성은 모든 요구가 known true여야 합니다. 채택은 Root가 이미 봉인한 결정이며 검증기가 새 라벨을 만들지 않습니다.

예상 성공 출력은 21개 요청·61개 라벨·양성 21·음성 40·입력 103입니다. 행의 finite_scope는 **선택된 후보의 관측 301개**를 셉니다. 제외 후보까지 포함한 원 실행 증거는 **305개**입니다. 새 5개만 보면 65개와 69개입니다. 두 수치를 섞지 않습니다.

후보 위치 1만 선택하는 단순 대조군은 18/21, 약 85.7%입니다. 오류 writer의 양성이 위치 0으로 이동하므로 19/21이 아닙니다. 모든 후보를 음성으로 고르는 라벨 단위 대조군은 40/61입니다. 두 값 모두 모델 정확도가 아닙니다. baseline/reference/wrong-seed 이름과 문장의 문체도 정답 단서가 될 수 있습니다. 이 준비는 후보 순서나 문구를 바꾸지 않습니다. 향후 순서 치환·ID 차단·문체 통제는 별도 실험으로 검증해야 합니다.

각 입력은 최대 1MiB, 한 행은 16KiB, 결과는 64KiB입니다. 입력은 regular 파일·크기·SHA-256으로 고정하고, 출력은 새 절대 경로에만 O_EXCL로 생성한 뒤 Write·Sync·Close를 확인합니다. 경로·오류 본문은 결과에 담지 않습니다. 파일 Sync 호출이 전원 장애 내구성이나 저장 장치의 실제 영속성을 증명한다는 주장은 없습니다.

CLI는 정확히 다음 10개 옵션을 한 번씩 `--이름 값` 형태로 받습니다. 등호 형태·중복·위치 인자는 거부합니다.

```
--data DATA21 --data-sha256 ROOT_FROZEN_DATA21_SHA
--previous-data DATA16 --fixtures FROZEN_FIXTURES
--wants FROZEN_WANTS --adoption ROOT_QUALIFICATION_V2
--adoption-sha256 3c1468970a99bc494f3674c6fcbaf476f485e3d491a850ab7aebb09abadd3bfd
--finite-comparison SAVED_FINITE_REPORT --saved-results SAVED_WORKER_RESULTS
--output FRESH_ABSOLUTE_REPORT_PATH
```

module/validate.go와 source/reader-main.go.txt를 프로젝트 모듈 내부의 별도 검토용 디렉터리에 놓고 사용합니다. 준비용 module의 검사는 표준 라이브러리와 가짜 reader만 사용하며 bridge를 빌드하지 않습니다. synthetic 자료는 실제 후보·원본·Go reader·모델 증거가 아닙니다.

첫 formatter 경로 오류와 테스트의 미사용 변수로 인한 build-fail 로그를 보존했습니다. 수정 후 Go 1.27.1의 race 검사 12개 상위+10개 하위 검사가 통과했고 vet도 통과했습니다. 이후 예상 성공 JSON 바이트를 직접 대조하는 검사를 추가해 최종 race 13개 상위+10개 하위 검사와 vet가 다시 통과했습니다. 두 성공 기록과 첫 실패 기록을 모두 보존합니다. race는 Go 검사 환경상 CGO=1, vet는 CGO=0이며 원본 패키지를 import하지 않습니다. 공개 reader bridge의 빌드·실행, actual LoadDevelopmentRow, 모델·학습·feature·projection·score, Git·HF·네트워크는 모두 0입니다.

검토자는 Wanted·후보·observer·Root 채택·자료 생성기의 작성자가 아니며, 이전 소스와 실제 저장 비교를 알고 있는 눈가림 없는 검토입니다. 검증기 소스는 이 검토자가 작성했습니다. 일반 정확성·새 모델 품질·토큰 절감·라이선스 보증·새 CI 통과를 주장하지 않습니다.

[계획](PLAN.v1.json) · [예상 성공 JSON](EXPECTED-OUTPUT.v1.json) · [위치·문체 단서 검토](LEAKAGE-REVIEW.v1.json)
