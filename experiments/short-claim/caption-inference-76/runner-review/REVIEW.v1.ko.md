# 고정 실패 모델 설명문 진단76: runner 실행 전 정적 검토

현재 봉인된 worker와 wrapper를 실행할 때 문제가 될 제어 흐름이나 입력 연결 오류를 찾지 못했습니다. 이는 한 번의 개발용 진단 실행에 관한 정적 검토이며, 모델 품질·학습 준비·독립 평가의 승인이 아닙니다.

요청과 설명문만 Features에 전달됩니다. 별도로 봉인된 정답은 순위가 나온 뒤 checks·Top1·Top3를 계산하는 곳에서 쓰입니다. 코드, 정답, 후보 ID, 역할, 마스크, 출처 정보는 모델 특징 인자로 전달되지 않습니다. 양쪽 A/B 설명문에서 전체 9개 후보를 유지합니다. 세 부모의 허용 후보 [0]/[1]/[2]는 앞선 별도 정답 검토와 연결되어 있고, 원래 null 자료를 수정하지 않습니다.

두 고정 모델을 각각 32,792바이트와 SHA로 확인한 뒤 Decode합니다. 유한한 FP64 계수 8,192개만 허용하며 출력에 계수를 담지 않습니다. 예상 직접 호출은 Prepare6, Baselines6, Decode2, Features36, Score36입니다. 이것은 향후 실행 예산이며 이번 검토에서 해당 API를 실행한 횟수는 모두 0입니다. Score의 동일 점수는 원래 후보 순서, 최소 checks인 통제의 동점은 fixed_order→bm25→lexical_ordered→narrow_rule 순서를 유지합니다. 좁은 규칙의 기존 BM25 fallback을 기록합니다. 부모마다 후보가 세 개라 Top3는 유용한 구별 지표가 아닙니다.

worker는 계획·정답·런타임 소스 다섯 파일을 확인한 뒤 새 하위 attempt 디렉터리를 예약합니다. 각 callback 호출 전 fsync한 partial에 예약 카운터를 남기고, 반환 뒤 유효 접두부와 유한한 점수를 보존합니다. 예약은 실제 API 진입의 증명이 아닙니다. 최종 results.json은 O_EXCL이고 stdout은 동일한 최종 구조를 직렬화합니다. 저장 자체가 실패하면 마지막 내구성 있는 접두부만 남을 수 있고, 오류 반환과 worker exit1은 wrapper가 거절합니다.

wrapper는 새 부모 디렉터리와 사전 예약을 만든 뒤 worker를 한 번 시작하며, 출력 1MiB·stderr64KiB를 제한합니다. 최소 환경에서 CPU1·256MiB Go soft heap을 설정하고 60초에 process group을 종료한 뒤 Wait 한 번을 합류합니다. 시작·대기·exit·timeout·출력 초과 오류는 실패입니다. wrapper PASS는 정상 종료만 뜻하며, 실제 결과 상태·내용·카운터의 판정은 후속 실제 결과 검토가 맡습니다.

Go soft heap은 OS RSS 상한이 아니며 Go 메모리·worker wall은 최종 저장 이전의 snapshot입니다. mutable 파일에 대한 TOCTOU 방어나 예약 디렉터리의 crash durability, 전체 반환의 60초 보장은 주장하지 않습니다. root 소유 봉인 파일이 유지된다는 범위입니다.

별도 stdlib metadata helper 1회가 19개 파일의 bytes/SHA와 실제 worker build metadata(Go1.27.1, CGO0, Darwin arm64, trimpath)를 확인했습니다. 원래 테스트·worker·모델·상위 API 실행 0, 원본과 공유 파일 수정 0입니다. 검토자는 이 runner·wrapper·정답의 저자가 아니지만, 원래 finite 관찰기와 이전 projection/binder를 작성했으며 사전 결과를 본 nonblind 검토자입니다. byte/build metadata 확인은 compiler의 독립적·hermetic 신뢰 증명이 아닙니다.
