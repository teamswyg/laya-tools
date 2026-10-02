# 전체76 투영 실행 전 독립 확인

현재 frozen=false인 v1 계획과 실행 코드의 준비 범위를 확인했다. 검토자는 worker76 작성자와 다른 사람이며69/71 코드 작성·68 source/caption·70 저장 결과 검토에 노출됐으므로 nonblind다. 실제 worker나 원본 Validate·Normalize·Project·Features·role·fit·model은 실행하지 않았다. 준비 코드를 읽고 저장된 JSON을 stdlib Go로 연결한 검사와 독립 복사본의 합성 검사만 수행했다.

30개 실행 pin8662721B는 입력8개·공개source15개·driver6개·binary1개다. 원본14개 non-test Go 파일이7개 package에 포함되고 그 디렉터리에 unpinned non-test Go 파일이 없는 현재 범위를 확인했다. binary의 Go1.27.1·CGO0·darwin/arm64·trimpath·exe 정보와 byte SHA가 일치한다. `-buildvcs=false`는 빌드 recipe의 값이며 buildinfo에 그 flag가 직접 저장됐다고 주장하지 않는다. 이는 현재 source와 build metadata 확인이며 hermetic compiler 증명은 아니다.

독립 stdlib metadata join1회에서 원래72 요청·216 후보의 text·정답·acceptable 순서·mask와 source68의 좁은 finite proposed4/10을 대조했다. 새 supervision은 original label로 바꾸지 않고 proposed 상태를 유지한다. 입력 loss weights는 모두null이고 실제 role70의 parent index→row 배열 연결을 사용해 원래 row 순서를 바꾸지 않는다. 같은 그룹의 역할이 나뉘지 않고 전체76/226·known38/no_answer17/unknown21·그룹19가 유지된다.

| 실행 후 기대되는 배열 | Train | Validation | Calibration |
|---|---:|---:|---:|
| 부모/후보 |44/130|20/60|12/36|
| unknown 부모/nullable 후보 |13/39|5/15|3/9|
| fit row |91|45|0|
| 가중치0 row |18|1|0|
| 가중치1 positive/negative |22/51|13/31|0/0|

이 값은 저장된 역할·정답·mask에서 계산한 **기대값**이다. 실제 Project 관측값이 아니다. direct Validate76회와 Project1회를 호출할 계획이고 성공 시 feature scans는2×(91+45)=272다. unknown은 nullable audit에만 남고 calibration은 fit row를 만들지 않는다. known의 가중치0 행은 삭제하지 않으며 no_answer의 모든 기존0 정답도 유지한다. diagnostic AUC view는 가중치1 행만 포함하고 Data.Excluded0, unknown 분모39/15를 별도로 확인한다. AUC 함수나 모델 효용은 실행하지 않는다.

원본 함수에 넘기는 feature 입력은 원래 request·caption Text뿐이다. source ID·role·truth·mask·review 정보는 feature가 아니다. Project의 sparse payload64MiB preflight와 파일 출력64MiB를 구분한다. CPU1·soft heap256MiB는 설정이며 OS RSS 하드 상한이 아니다. 결과 pin·prepared raw text·candidate 순서·labels·weights·groups·sparse shape·finite index/value·diagnostic 필터를 함수 반환 뒤 확인한다.

외부 controller는30pins와 review receipt SHA를 먼저 확인하고 frozen flag 한 값만 바꾼 계획과 실행 예약을 child Start 전에 저장한다. process group300초 종료와 전체 child CPU/RSS 측정을 맡는다. v1에서 JSON `null`/`{}`가 빈 summary로 통과할 수 있다는 좁은 문제를 알렸고 root가 v2로 수정했다. v2는 schema·state·counter presence와 성공 시 projection/payload/272scans를 확인한다. empty/partial/missing/invalid 결과는 raw SHA/bytes·OS 지표·unknown 카운터를 보존하고 실패로 종료하는 경로를 정적으로 확인했다. controller도 실제 실행하지 않았다.

worker는 O_EXCL 예약·결과 파일과 before-validate/before-project marker를 보존한다. 강제 종료 시 마지막 marker 뒤 개별 Validate나 Project의 진입·반환을 정확히 재구성할 수 있다는 보장은 없다. Project 중 종료됐다면 실제 FeatureScans는unknown으로 기록해야 하며272를 관측값으로 채워 넣지 않는다. 이 검토는 실제 강제 종료/RSS 실험을 수행하지 않았다.

독립 복사본의 합성7개 test·10개 subcase는 race1/vet1이 통과했다. injected Validate callbacks78회·Project callbacks2회(정상1·panic1)이며 원본 API0이다. 자체 metadata checker는3회였다. 첫 시도는 변수 재선언으로 컴파일 실패했고 두 번째는 buildinfo에 저장되지 않는 buildvcs flag를 필수로 오인했다. 이를 고친 세 번째가 통과했고 실제 metadata join은1회만 수행했다. patch context 실패1회도 장부에 남겼다. 실제 worker·입력·모듈·정답·역할·mask는 수정하지 않았다.

남은 concrete code blocker는 없다는 준비 범위 결론이다. 학습 준비성·출처 다양성·모델 성능 승인이 아니다. 새2가족은 모두train이므로 그 가족의 validation 일반화를 확인한 것도 아니다. AI-assisted 검토 비용은 측정하지 않았고 별도 model/API/paid benchmark0이 이 대화의 비용0을 뜻하지 않는다. 원본 투영 실행과 이후 실제 결과 검증은 부모가 고정된 계획에서 별도로 수행한다.
