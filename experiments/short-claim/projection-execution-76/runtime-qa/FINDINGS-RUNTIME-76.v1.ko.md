# 전체76 실제 투영 결과의 읽기 전용 확인

최초 Project1회 성공 결과8390462B와 SHA를 그대로 읽고 입력·배열을 대조했다. 나는 worker76 작성자와 다른 검토자이지만 claimfit64 포팅·69/71 드라이버 작성자이며 앞선68/70 검토에 노출됐다. 자체69 DTO audit 패턴을 재사용했으므로 이 저작·노출 한계를 숨기지 않는다. 이번 검사에는 저장된 JSON을 읽는 stdlib Go만 사용했고 원본 Project·Features·Normalize·Validate·role·fit을 재실행하지 않았다.

30개 source/input/binary pin8662721B와 원래 입력을 다시 확인했다. frozen 계획은 draft의 frozen 한 값만 바뀌었다. 전체76 요청·226개 후보의 raw text·ID·후보 순서·acceptable 순서·그룹 역할·mask·nullable 정답과 결과 parent audit가 일치한다. 원래72개·153 known/no_answer 후보의 정답은 그대로이고 source68의4개 관련 요청·10개 finite proposed 후보를 구분한다. unknown21개/63후보는 nullable로 보존한다. 알려진 가중치0 행은 삭제하지 않고 원래 label과 weight0을 함께 남긴다. no_answer 정답은 전부0이다.

| 실제 배열 | Train | Validation | Calibration |
|---|---:|---:|---:|
| 부모/후보 |44/130|20/60|12/36|
| unknown 부모/nullable 후보 |13/39|5/15|3/9|
| fit row |91|45|0|
| weight0 row |18|1|0|
| weight1 diagnostic row |73|44|0|
| diagnostic positive/negative |22/51|13/31|0/0|

Calibration의 known 후보27개는 runtime에 보존하고 fit row는 만들지 않았으며 unknown 후보9개도 audit에 남는다. 같은 그룹은 같은 역할이고 train/validation row 참조는 원래 parent·candidate 순서를 따른다. sparse offset·길이·8192 범위 index·중복·finite value·label/weight/group를 확인했다. Diagnostic 데이터는 원본 weight1 행과 같은 feature slice·label·weight·group를 보존하며 Data.Excluded0과 unknown 분모39/15를 분리한다. AUC 함수는 실행하지 않았다.

원본 장부의 direct Validate76/76·Project1/1·FeatureScans272와 childexit0/retry0이 결과·외부 장부와 일치한다. 내부 ValidatePrepared/Normalize 호출은 개별 계측되지 않았으므로 수를 새로 추정하지 않는다. 실제 normalized text는 최대32words·239B이며 정규화를 다시 수행하지 않고 저장된 값의 UTF-8·크기·단어 범위만 확인했다. pin으로 고정된 claimfit source는 원래 request와 candidate.Text만 Features에 전달한다. role/truth/source/review metadata가 인자로 들어가지 않는 코드 경계를 확인했지만 feature 의미값을 독립적으로 다시 계산한 것은 아니다.

동일 DTO 필드·현재64bit ABI와 배열 길이로 payload를 재계산하면 기록과 같은1928154B다. 이는 보유 배열·문자열의 계산값이며 allocator·임시 JSON·전체 RSS를 포함하지 않는다. 원본 JSON8390462B와도 다른 값이다. 외부 측정은 child 전체 wall0.475047292초·peakRSS50036736B·peak footprint41255464B다. 표시 CPU user0.08/sys0.01초는 반올림된 전체 프로세스 지표이며 Project만의 비용이나 hard RSS 상한으로 확대하지 않는다.

공개할 수 있는 compact mechanics에는76개 부모의 raw text SHA·label/mask/role과253개 fit/diagnostic row의 feature 수·SHA, 네 dataset column SHA를 남겼다. 원본 feature 값 배열이나 모델 계수를 복제하지 않는다. canonical SHA 규칙은 이 DTO를 Go json.Marshal한 값이며 RFC8785 주장이나 원본 byte SHA 대체가 아니다. 원본8MB 파일의 SHA와 bytes는 별도로 유지하고 원본 파일을 변경하지 않았다. 원본·compact 모두 private absolute path는0개다. sample weights는 학습행의 loss weights이고 학습된 모델 계수가 아니다.

자체 runtime array checker1회와 vet1이 통과했고 실패0이다. 추가 읽기에서 child 파일2개를 바깥 control 폴더에서 잘못 찾은 오류는 기록했으며 실제 checker는 올바른 child 경로에서 그 파일을 이미 확인했다. 별도 모델·유료 API·fit·새 role/label/mask/weight·공유수정·게시0이다. AI-assisted 검토와 협업 비용은 측정하지 않았다.

이 결과는 실제 데이터 투영의 정합성을 확인한 것이다. 새2가족은 모두train이며 source68 supervision은 좁은 finite proposed 상태다. 모델 성능·held-out 일반화·전체 외부 graph 독립성·학습 준비성·출처 다양성을 승인한 결과가 아니다. 기존 실제 결과의 training_ready/source_diversity_cleared=false를 유지하고 새 fit gate를 만들지 않았다. 부모가 별도의71 동결 계획으로 최초 FP32 학습을 결정하고 실행할 때 이 exact ref를 사용할 수 있다.
