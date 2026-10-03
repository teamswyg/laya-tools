# 공개 개발 자료를 안전하게 읽는 Go reader

`pkg/shortclaimdata`는 공개 개발 요청 한 행에서 모델용 문장, 정답·가중치, 출처 기록을 나누어 꺼내는 작은 Go API입니다. 후속 학습 코드를 단순하게 만들고, 정답이나 출처 ID를 문장 특징에 우연히 섞는 일을 줄이려고 만들었습니다.

- `Example.Input()`: 요청·후보 문장을 검증하고 정규화한 기존 `shortclaim.ValidatedInput`입니다. 후보 ID는 기록용으로 남습니다.
- `Example.Supervision()`: 후보 순서와 같은 정답·가중치 배열입니다. Boolean 라벨과 정수 리터럴 `1` 가중치가 필수입니다.
- `Example.Metadata()`: 요청 ID, 연결 그룹, 출처 가족·리비전, 관측 범위입니다. 인코더 문장에 넣으면 안 됩니다.

입력 provenance는 `finite-development-v1`로 고정합니다. 원 출처 정보는 metadata에 보존하므로 원 입력 파일과 바이트가 같다는 뜻은 아닙니다.

## 사용

`io`, 프로젝트의 `pkg/shortclaim`, `pkg/shortclaimdata`를 import한 코드에서 다음처럼 한 행을 읽습니다. 이 예제는 추론이나 학습을 실행하지 않습니다.

```go
func readRow(r io.Reader) (shortclaim.ValidatedInput, shortclaimdata.Supervision, error) {
    example, err := shortclaimdata.LoadDevelopmentRow(r)
    if err != nil {
        return shortclaim.ValidatedInput{}, shortclaimdata.Supervision{}, err
    }
    return example.Input(), example.Supervision(), nil
}
```

JSON 객체 하나가 최대 16 KiB, 후보는 1–8개입니다. JSONL은 크기를 제한한 한 행씩 전달합니다. 전체 파일을 한 번에 읽는 API가 아닙니다. `development_train` 역할만 받고, 필수 필드 누락·중복·대소문자 별칭·잘못된 Unicode를 거절합니다. 오류 메시지는 입력 본문이나 경로를 담지 않습니다.

내부에는 불변 문자열과 소유한 고정 배열을 유지합니다. 정답과 가중치는 별도 배열로 두고, 접근자는 값 복사를 반환합니다. 반환 배열을 바꿔도 원 자료는 바뀌지 않습니다. reader에는 공유 가변 map·slice나 lock이 없습니다.

## 현재 확인한 범위

Go 1.27.1 로컬 race 검사에서 테스트·하위테스트 합계 46건과 vet 검사를 통과했습니다. 실패·건너뜀은 0건입니다. 별도 작성자의 소스 검토에서도 차단 문제는 없었습니다. [검증 기록](../../experiments/short-claim/publication-proof-111/READER-LOCAL-QA.v1.json)은 로컬 검사와 해당 변경의 CI를 구분합니다.

출처와 라벨을 바꿔도 정규화 문장이 같다는 테스트는 입력 분리를 확인합니다. 모델 품질·속도·메모리·GPU 개선은 측정하지 않았습니다.

현재 로컬 [개발 자료](../../experiments/short-claim/next60-development-twentythree/README.ko.md)는 요청 23개·라벨 67개(+23/−44)입니다. 실제 reader 호출·반환·값 일치가 각각 23회인 [대응 검사](../../experiments/short-claim/next60-development-twentythree/INPUT-VALIDATION.v1.json)를 통과했습니다. 마지막 확인된 [HF 게시 버전은 21개 요청](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-21-finite-v1)이며, 23개 게시 완료 여부는 별도 게시 증거로 확인합니다. 이전 자료를 보존하고, 113개 유한 입력과 335개 원 관측을 요청 수나 가중치로 늘려 세지 않습니다. 라벨에 선택한 관측은 331개이며 미상을 거짓으로 바꾸지 않습니다. 2,400개 보호 평가 자료가 아니며, reader의 구조 검사가 권리·역할·정답 의미나 학습 실행 권한을 부여하지 않습니다.
