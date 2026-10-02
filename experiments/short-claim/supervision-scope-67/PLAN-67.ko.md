# 원래 텍스트와 정답을 보존한 감독 범위 점검67

목표는 기존 유한 영어 요청의 **전체 명시 조건**을 후보 설명과 연결하는 작은 힌트 학습이다. 설명을 일부만 이해했다는 새 목표에 기존 함수 정답을 그대로 빌려 쓰지 않는다. 데이터나 의미 판정·정답을 다시 작성하지 않고, 이미 고정한 60/61 검토를 기계적으로 연결하여 후보별 제안 loss mask를 점검한다.

첫 실제 mask 점검 전에 아래 규칙을 고정한다. 원래 72부모·216후보·전체 허용 후보 집합과 17그룹을 보존한다. 원래 unknown은 nullable audit만 남고 label/fit row를 만들지 않는다. known/no_answer 후보의 원래 0/1 label은 그대로 보존한다. 후보 mask를 false로 해도 라벨이나 그룹 구성원을 삭제하거나 바꾸지 않는다.

부모 request_contract, 해당 prototype의 contract_observation, 후보 source closure와 source_fidelity가 모두 consistent_with_scoped_evidence여야 한다. 원래 label1은 candidate coverage가 consistent_with_scoped_evidence일 때만, label0은 contradicts_scoped_evidence일 때만 loss mask=true다. omission/uncertain/pending이나 저장 정답과 맞지 않는 coverage는 mask=false이고 정확한 원래 상태와 사유를 보존한다. 검토 판정을 새로운 positive/negative label로 승격하지 않는다. 보조 observation_fields/explicit_negative_boundaries는 그대로 감사에 연결하며 원래 coverage 판정을 덮어쓰지 않는다.

이 mask는 감독을 제안하는 metadata이며 학습 준비 승인이 아니다. 같은 후보의 label과 전체 평가 후보·순서·분모를 유지한다. 원천 다양성은 아직 uncleared다. 실제 역할/seed/order/특징 추출·투영·fit·모델·원래 API·보호 최종 읽기는 이번 점검에 없다. 공개71 등의 새 숫자 기준이나 유리한 seed 검색을 추가하지 않는다. 기존 최소15그룹·역할9/3/3·필요 효용5%와 별도의 도메인별 최종2400요청 목표를 유지한다.

점검은 원래 56/56b probes와 저장 결과, 60/61 review 원문 SHA를 먼저 확인하고 원래 index/ID/후보 수/738개 원래 entry의 완전한 대응을 확인한다. 파일별 최대2MiB와 전체16MiB의 읽기 상한을 적용한다. 메모리 한도는 파일 읽기의 payload bound이며 peak RSS 측정이 아니다. 원래 AST/formatter/Bind/Generate/SourcePins/API는 호출하지 않는다.

결과를 보고 이 규칙을 완화하거나 mask를 바꿔 fit 자격을 맞추지 않는다. 각 전체 그룹의 positive-weight positive/negative 후보와 원래 unknown 분모를 보고한다. 역할에 따른 fit 자격과 기존 동일 scope의 5% headroom 재계산은 별도 고정 실행에서 한다. 지금 보고서가 valid여도 authoring_diversity_cleared=false, training_ready=false다.

AI 보조 기획·구현·검토를 사용했다. 일반 Codex 협업 사용량과 비용은 측정하지 않았고 별도 모델 호출0을 전체 AI 비용0으로 표현하지 않는다. 실제 점검 횟수·실패는 별도 장부에 남긴다.
