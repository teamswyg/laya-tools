# 입력하면서 보는 riidolaya 힌트

[English](live-claims-demo.en.md) · [README](../README.md)

텍스트를 입력하거나 수정하면 작은 Go 모델이 세 가지 주장을 각각 판정합니다. 예제 버튼을 누르거나 직접 문장을 바꿔서 결과를 비교할 수 있습니다.

- ❓ **응답 요청:** 답변이나 확인을 요청하는 표현인가?
- 🛠️ **진행 보고:** 지금 개발 작업을 하고 있다고 보고하는가?
- ✅ **완료 보고:** 식별할 수 있는 개발 작업을 끝냈다고 보고하는가?

세 항목은 독립적입니다. 한 문장이 진행 상황과 질문을 함께 담을 수 있습니다. 완료 보고는 글쓴이의 주장이고 실제 업무가 끝났다는 검증이 아닙니다.

## 실행하기

저장소에서 Go 1.27.1로 빌드합니다. 웹 파일은 실행 파일에 포함되므로 Node.js나 Python 서버가 필요하지 않습니다.

```sh
go build -o ./bin/riidolaya ./cmd/riidolaya
mkdir -p .cache/live-claims-demo/model
curl --fail --location \
  'https://huggingface.co/JooYoon/riidolaya-development-claim-hints-semantic-contrast-v0.1/resolve/a94997978f4cd4a7dfbf2cbd4c0c22dee143d7ba/claims.rsc' \
  --output .cache/live-claims-demo/model/claims.rsc
./bin/riidolaya demo \
  --model .cache/live-claims-demo/model/claims.rsc \
  --model-sha256 cfd35ee70a23f94a8d470b7dea596244a91ac7f1410475e8a42959693c99864b
```

브라우저에서 **http://127.0.0.1:8877**을 엽니다. 종료하려면 서버를 실행한 터미널에서 `Ctrl+C`를 누릅니다. 포트를 바꾸려면 `--listen 127.0.0.1:8878`을 추가합니다. 로컬 주소만 허용합니다.

모델은 [Apache-2.0 공개 연구 모델의 고정 버전](https://huggingface.co/JooYoon/riidolaya-development-claim-hints-semantic-contrast-v0.1/tree/a94997978f4cd4a7dfbf2cbd4c0c22dee143d7ba)이며 파일은 **73,988바이트**입니다. 위 명령은 전체 파일 SHA-256을 확인한 뒤 로드합니다. Git에는 모델 파일을 넣지 않습니다. Laya 가중치를 변환한 모델이 아니라 자체 Go 특징 분류기이며, 이 시연에 GPU나 외부 추론 API는 필요하지 않습니다.

## 화면 읽기

입력이 잠시 멈추면 약 180ms 뒤 다시 계산합니다. 한글 조합 중에는 기다리고, 오래된 요청의 응답은 화면을 덮어쓰지 못하게 합니다. 빈 입력은 결과를 초기화합니다. 입력 한도는 **UTF-8 4,096바이트**이고 초과하면 잘라서 추론하지 않습니다.

각 카드의 **최종 판정**은 ‘해당·해당하지 않음·판정 보류’입니다. 확률은 모델의 내부 점수이며 정확도 보증이 아닙니다. **점수가 가장 높은 후보**는 최종 판정과 별도로 표시합니다. 신뢰도 0.9와 점수 차이 0.05 기준을 충족하지 못하면 가장 높은 후보가 있어도 보류합니다. 보류를 긍정 제안으로 해석하지 마세요.

화면의 계산 시간은 서버에서 모델을 실행한 시간입니다. 입력 대기·브라우저 통신·화면 갱신 시간과 전체 CPU/GPU·메모리 사용량을 뜻하지 않습니다. 모델과 작업 공간을 한 번 준비하고 공유 작업 공간의 계산을 직렬화합니다.

텍스트는 이 컴퓨터의 로컬 서버에만 전송하며 저장하지 않습니다. 예제는 이 시연을 위해 새로 작성했고 결과는 실제 모델에서 계산합니다. 로컬 API는 `GET /api/status`와 `POST /api/hints` (`{"text":"..."}`)로도 사용할 수 있습니다.

현재 모델은 **운영 품질 미검증**입니다. 이 웹은 입력 변화에 따른 결과와 한계를 관찰하기 위한 시연이며 뤼이도의 라벨·댓글·상태를 실제로 변경하지 않습니다. 보호된 학습·평가 자료는 시연에 사용하지 않습니다.

## 선택형 Go 프로파일링

CPU·메모리 동작을 조사하려면 실행 명령에 `--cpu-profile .cache/demo-cpu.pprof --heap-profile .cache/demo-heap.pprof`를 추가합니다. 저장 경로의 디렉터리는 미리 만들고, 기존 파일이 없는 새 이름을 사용하세요. 기본 실행에서는 프로파일을 만들지 않습니다.

프로파일을 켠 뒤 문장을 시험하고 `Ctrl+C`로 종료하면 CPU 기록과 GC 후 살아 있는 Go heap의 스냅샷을 각각 저장합니다. 서버와 모델은 heap 측정 동안 유지합니다. `go tool pprof -top .cache/demo-cpu.pprof`와 `go tool pprof -top .cache/demo-heap.pprof`로 읽을 수 있습니다. CPU는 샘플링 방식이므로 짧은 시험에서는 표본이 부족할 수 있고, 프로파일링 자체의 비용도 더해집니다.

파일은 최대 각각 16MiB, 권한 0600으로 새로 생성하며 기존 파일을 덮어쓰지 않습니다. 프로파일용 HTTP 주소는 열지 않습니다. **Go heap은 OS RSS나 GPU 메모리가 아닙니다.** heap 프로파일 역시 표본이므로 작은 할당이 표시되지 않을 수 있습니다. 프로파일에는 로컬 코드 경로 등이 포함될 수 있으므로 Git이나 공개 게시에 넣지 말고 분석한 집계만 공유하세요. 저장에 실패하거나 한도를 넘으면 종료 시 오류를 반환하며 해당 파일은 불완전할 수 있습니다.
