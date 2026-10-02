# 작은 가중치 표를 Go에서 선택해 쓰기

`pkg/hintweights`는 문장을 생성하거나 모델을 학습하는 기능이 아닙니다. 이미 준비한 RIIDOH01 형식의 8,192개 계수로 점수를 계산할 때, 전체 계수를 FP64 배열로 풀지 않고 압축 바이트에서 읽는 선택형 Go API입니다. Laya 문장 encoder 전체를 이 크기로 실행한다는 뜻은 아닙니다.

FP64 계수 배열의 데이터는 65,536바이트입니다. 새 reader의 소유 payload는 FP32 32,792바이트, INT8 8,216바이트, ternary는 0이 아닌 계수 13개에서 1,308바이트·8,192개에서 2,330바이트입니다. 구조체·할당 반올림·호출자 입력·특징 배열·Go heap·전체 프로세스 RSS는 이 표에 포함하지 않습니다. 생성 중에는 원 입력과 복사본이 함께 남을 수 있습니다.

속도에는 대가가 있습니다. 공개 합성 특징 266개를 쓰는 private prototype의 점수 계산은 처음 구현에서 기존 배열보다 약4~5배 느렸습니다. 형식마다 루프를 분리한 후속 구현에서는 중앙값이 FP32 213.1ns, INT8 189.6ns, ternary sparse 248.9ns·dense 418.8ns로, 기존 배열보다 약1.21~2.68배 느렸습니다. 측정한 점수 계산은 모두 0 bytes/op·0 allocs/op였습니다. 이 비교는 한 Mac의 warm 합성 반복이며 두 버전은 별도 프로세스에서 측정했습니다. 공개 패키지 포팅은 별도 벤치마크하지 않았습니다.

따라서 메모리가 중요한 개발자는 선택해 시험하고, 점수 계산 속도가 중요하면 기존 경로를 유지할 수 있습니다. 기본 라우터는 새 reader를 자동 선택하지 않습니다. `New`는 입력을 검증한 뒤 소유 복사본을 만들며, 생성 중 입력은 안정적으로 유지해야 합니다. 반환된 View는 읽기만 공유하므로 별도 lock 없이 동시 읽기가 가능합니다. 입력 순서·중복·FP64 곱셈/누적을 유지하고, Inf/NaN 동작 때문에 0 계수의 곱셈도 생략하지 않습니다.

[사용 예제와 API](https://github.com/teamswyg/laya-tools/blob/main/pkg/hintweights/README.ko.md) · [원본 측정과 실패 기록](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/packed-weights-78). 현재 bitmap/sign 저장은 native 1.58-bit 연산·bulk SIMD·전체 SoA 구현을 뜻하지 않습니다. 실사용 RAM·모델 정확도·LLM 사용량 절감은 별도 검증이 필요합니다.

개발 자료도 넓히고 있습니다. [관찰77](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/native-observation-77)은 logfmt와 percent escape의 24개 유한 입력을 한 번 관측해 예상과 일치했습니다. 이는 독립 작업 24개나 새 학습 정답 수가 아닙니다. [원천 제안79](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/source-growth-79)은 UUID·humanize 등 6계열을 검토하되, 원문 취득·라이선스·그룹 관계·관찰·학습 자격을 따로 확인하도록 합니다. 보호된 2,400개 최종 평가와 새 모델의 효용 증명은 아직 완료하지 않았습니다.

[English](https://github.com/teamswyg/laya-tools/wiki/Packed-Weights-78-EN)
