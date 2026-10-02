# 선택형 packed 가중치 읽기

`hintweights`는 개발자가 명시적으로 선택하는 작은 Go 라이브러리입니다. RIIDOH01 형식의 8,192개 FP32·INT8·ternary 가중치를 소유 바이트에서 직접 읽습니다. 전체 계수를 `[]float64`로 펼치지 않으며 production 코드는 Go 표준 라이브러리만 사용합니다. 기존 라우터·CLI·기본 `Decode`/`Score`·모델 정책은 그대로입니다.

```go
import "github.com/teamswyg/laya-tools/pkg/hintweights"

view, err := hintweights.New(rawModelBytes)
if err != nil {
    return err
}
score, err := view.Score([]hintweights.Feature{
    {Index: 0, Value: 1},
    {Index: 64, Value: 0.5},
})
```

`rawModelBytes`는 호출자가 준비한 RIIDOH01 바이트입니다. 파일 읽기·다운로드·모델 선택은 이 패키지가 수행하지 않습니다. `New`가 읽는 동안 입력을 바꾸면 안 됩니다. 반환 후에는 원 입력을 바꾸어도 View가 가진 복사본에 영향을 주지 않습니다. View를 값으로 복사하면 내부의 변경되지 않는 저장소를 공유하며 동시 읽기가 가능합니다. 내부 바이트를 노출하는 accessor는 없습니다.

| API | 의미 |
|---|---|
| `New(raw)` | 형식·유한 계수·count·padding을 검사한 뒤 소유 복사본 생성 |
| `Coefficient(index)` | 한 계수를 FP64로 읽음 |
| `Score(features)` | 입력 순서·중복을 유지하는 FP64 곱셈과 누적 |
| `OwnedPayloadBytes()` | 소유 byte payload와 ternary prefix 표의 데이터 크기 |

nil/zero View는 `ErrView`, 범위 밖 index는 `ErrIndex`, 잘못된 모델 바이트는 `ErrFormat`입니다. Score는 0 계수의 곱셈도 생략하지 않아 Inf·NaN 동작을 보존합니다. 각 형식의 루프를 분리했지만 scale을 마지막에 곱하거나 feature를 재정렬하지 않습니다.

소유 payload는 FP32 32,792 bytes, INT8 8,216 bytes입니다. ternary는 `24 + 1,024 + ceil(nonzero/8) + 258` bytes이며 nonzero 13개면 1,308 bytes, 8,192개면 2,330 bytes입니다. 구조체·slice header·allocator padding·호출자 입력·feature 배열·Go heap·OS RSS는 이 수치에 포함되지 않습니다. 원 입력과 소유 복사본이 동시에 남을 수 있습니다. 표가 작다고 CPU나 전체 프로세스 메모리가 반드시 줄어드는 것은 아닙니다.

이 라이브러리는 Laya의 문장 encoder, 학습기, 새 모델이 아닙니다. ternary 저장은 native 1.58-bit 연산·SIMD 실행을 뜻하지 않습니다. 모델 정확도·학습 자격·production 승격·Codex 요금 절감을 증명하지 않습니다. 실제 속도는 입력과 사용 경로에 따라 별도로 측정해야 합니다. 공개 합성 검사에서 형식 호환·계수/점수 비트 일치·소유권·동시 읽기·signed zero·Inf/NaN·sign-byte 경계를 확인하며 실제 가중치나 HF 자산을 테스트에 넣지 않습니다.
