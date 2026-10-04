# 작은 힌트 표현을 직접 비교하기

`riido-hintpreview`는 같은 문장이 어떻게 숫자로 바뀌는지 보여주는 Go 실험 도구입니다. 기존 단어 해시, 비교 기호를 남기는 해시, 범위 관계 배열을 한 번에 출력합니다. 모델 다운로드나 Python 없이 실행할 수 있습니다. 생성·학습·라우터 변경은 하지 않습니다.

```sh
go build -trimpath -o bin/riido-hintpreview ./cmd/riido-hintpreview
./bin/riido-hintpreview <<'JSON'
{"request":"Return keys in increasing order, keeping x >= lower and x < upper.","candidates":["AscendRange calls the iterator for every value in the tree within the range [greaterOrEqual, lessThan), until iterator returns false.","DescendRange calls the iterator for every value in the tree within the range [lessOrEqual, greaterThan), until iterator returns false."]}
JSON
```

사람은 `agreement_order`의 후보 번호를 확인하고, 에이전트는 JSON으로 다음 확인 순서를 받습니다. 번호는 입력 배열의 0부터 시작하는 위치입니다. 후보를 제거하거나 실행하지 않습니다. 점수는 인식한 관계가 같으면 +1, 다르면 −1이며 확률이나 정답 판정이 아닙니다. 동점은 현재 표시 순서를 유지합니다.

`legacy`와 `symbol`의 SHA는 특징 배열의 지문입니다. 각 index를 uint16 LE, 값을 float64 LE로 이어서 계산합니다. 기호 없는 유효 입력은 기존 값과 순서까지 같지만, `<`와 `<=`를 남기면 지문이 달라질 수 있습니다. 같은 차원8192라도 의미가 달라 기존 가중치를 재사용하면 안 됩니다.

`relation_columns`는 방향·하한·상한·멈춤 각각의 일치와 충돌, 총8열입니다. 최대8후보의 float32 특징 자체는256B입니다. 이는 프로그램·문장·스택·JSON·전체 RSS를 포함한 메모리 크기가 아닙니다. `grammar_recognition`의 Coverage는 방향1·하한2·상한4·멈춤8을 합친 인식 비트입니다. 인식 실패가 오답을 뜻하지 않습니다.

지원 문법은 `Return keys/all keys ... in increasing/decreasing order`와 `keeping/with x >= lower`, `x < upper` 같은 명시적 조건입니다. 피연산자 순서도 뒤집을 수 있습니다. 후보는 예제의 전체 traversal 설명 문법을 사용합니다. 감소 순서에서는 interval의 왼쪽이 상한, 오른쪽이 하한입니다. 알려지지 않은 endpoint 단어는 부분 인식으로 남지만, 빈 endpoint·중첩·추가 쉼표는 지원하지 않습니다. 부정·인용·조건문·한국어 등 일반 문장을 이해하는 도구가 아닙니다.

조기 종료의 횟수는 현재 표현에 없고 `stop_conflict`는 현재 지원 문법에서 항상0입니다. 빈 트리 요청처럼 인식되지 않는 요청은 모두0점으로 표시 순서를 유지합니다. 라벨·역할·학습 mask에는 영향을 주지 않습니다.

입력 JSON은16KiB, 각 문자열512 UTF-8 바이트·32단어·64토큰, 후보1~8개로 제한합니다. 관계 문법의 단어 수는 ASCII word run을 camel case로 나누기 전 기준이며, 기호 해시는 Unicode 단어 기준입니다. JSON의 두 필드는 필수이고 null·중복 키·추가 필드·잘못된 Unicode를 거절합니다. 오류는 stderr와 종료코드1로 알리고 원문을 출력하지 않습니다. 코드 탐색·Codex 라우팅 기본 동작과 독립된 opt-in 도구입니다.

[실제 개발 진단과 자원 측정](../../experiments/short-claim/text-representation-preview/README.ko.md)을 확인하세요.
