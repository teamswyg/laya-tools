# 가변 요청을 받는 Go 학습기 준비

기존 Fit79 소스를 실제로 찾았습니다. 역사 계획의 실행기11개 파일과 현재 저장소 핵심 의존16개가 모두 당시 핀과 일치합니다. 알고리즘을 새로 만들기보다, 79부모·235후보를 강제하는 입력 검사를 가변 입력의 동결 계약으로 바꾸는 범위를 권합니다. 이번은 설계만이며 코드 이관·학습·새 가중치 생성은0입니다.

seed1729·8192차원·FP32 scoring·50epochs·batch128·LR0.1·L2 0.0001·pair0 및 validation NLL 최초 엄격 최솟값 선택을 유지합니다. 현재 fp32는 점수용 계수를 float32로 반올림하지만 그림자 계수·기울기·손실은 float64입니다. 이를 모두 float32로 바꾸면 다른 실험입니다.

33개와 이후 약60개의 개발 요청은 기존 Go reader로 한 줄씩 읽고 Root가 고른 known 후보만 연결합니다. unknown-only 후보는 label/weight=null 근거 장부에 남겨 제외합니다. 알려진 반례와 unknown을 함께 가진 후보는 Root 채택이 있다면 음성 라벨을 유지할 수 있습니다. all-false를 자동으로 no-answer로 만들지 않습니다. 79+33=112는 산술이며 독립 요청 수가 아닙니다.

기존 가족 역할·검증/보정 배열·loss mask·0가중치 행은 보존합니다. 그룹82/83을 부모수33 크기 배열에 인덱싱하지 말고, 정렬 group-role slice와 후보[8]/역할[3] 집계를 사용합니다. 특징에는 요청/후보 텍스트만 보내고 정답·ID·역할·관측 수는 메타데이터로 남깁니다.

실행 전 실제 자료 생성/Go reader 확인, 기존79와 의미 중복·가족/역할 충돌 검토, 약60의 별도 역할/coverage 준비, 합류 명세·최신 자원 장부가 필요합니다. 수량만으로 자동 Fit하지 않습니다. 노출된 validation은 개발 회귀/선택이며 새로운 held-out이 아닙니다. protected2400을 읽거나 모든 개발 Fit의 선행요건으로 추가하지 않습니다.

새 버전은 기존 oneFitCheckpoint·완전한50epoch trace와4baselines·5% 필요효용·Top1/Top3 보존을 재사용합니다. 고정79 숫자와76..78 train-origin 구간을 Root 합류 범위로 바꿉니다. 새로운 seed/features/loss/ternary는 분리합니다. Fit3/8·model3/16과 실패 모델 비활성 상태는 그대로입니다.

CSR/연속 배열과 단일 worker를 유지합니다. 학습 시간·OS peak RSS·Go heap·CSR bytes는 이후 추론의 정규화/특징/점수/순위 비용과 따로 측정해야 합니다. model32,792B는 이론 저장 크기이지 RAM 측정값이 아닙니다. 프로파일용 corpus Fit 반복도 Fit에 기록하고, fake callback과 실제 synthetic test-fitting을 구분합니다.

[MIGRATION](MIGRATION.v1.json)·[측정 계약](PROFILE.v1.json)·[소스 핀](SOURCE-PINS.v1.json)을 참조하세요. agent114의 이전 after30 계획은 별도 핀으로 보존했습니다. 작성자는 관련 comparer/materializer 저자지만 Fit79 실행기 비저자이며 비맹검 소스/메타데이터 준비입니다.
