# 처음 시작하기

[English](https://github.com/teamswyg/laya-tools/wiki/Getting-Started-EN) · [홈](https://github.com/teamswyg/laya-tools/wiki)

처음에는 키워드 검색부터 사용하세요. Laya 다운로드, Codex 연결, 인증 정보 입력 없이도 결과를 확인할 수 있습니다.

## 1. 실행 파일 설치

[최신 릴리스](https://github.com/teamswyg/laya-tools/releases/latest)에서 Apple Silicon macOS는 `darwin-arm64`, x86-64 Linux는 `linux-amd64`를 선택합니다. Intel Mac과 Linux ARM용 바이너리는 현재 제공하지 않습니다. 압축파일과 SHA256SUMS를 내려받고, 압축파일의 SHA-256을 해당 항목과 비교한 뒤 압축을 풉니다.

예를 들어 배포된 v0.3.0 Apple Silicon 파일은 다음처럼 확인합니다.

```sh
shasum -a 256 riidolaya-v0.3.0-darwin-arm64.tar.gz
# 출력된 해시가 SHA256SUMS의 해당 줄과 같은지 확인한 뒤 진행합니다.
tar -xzf riidolaya-v0.3.0-darwin-arm64.tar.gz
./riidolaya version
```

Linux는 `sha256sum`과 linux-amd64 파일을 사용합니다. 재배포할 때 LICENSE, NOTICE, licenses/도 함께 보존합니다. `./riidolaya`로 직접 실행할 수 있으며, 아래 예시는 실행 파일을 PATH에 등록했다고 가정합니다.

Go 1.27.1과 C 컴파일러(macOS Command Line Tools)가 있다면 소스 빌드도 가능합니다.

```sh
git clone https://github.com/teamswyg/laya-tools.git
cd laya-tools
go build -trimpath -o bin/riidolaya ./cmd/riidolaya
./bin/riidolaya version
```

소스 빌드 직후에는 아래의 `riidolaya` 대신 `./bin/riidolaya`를 사용합니다. 사용자 실행에는 Python이 필요 없습니다.

## 2. 모델 없이 첫 결과 확인

Git 저장소 루트에서 실행합니다.

```sh
riidolaya search --root . --lexical --json 'routing confidence'
```

laya-tools에서는 구현 코드를 검색합니다. 다른 프로젝트에서는 해당 프로젝트의 식별자로 바꿉니다. 경로·줄 범위·코드 일부가 반환되고, 일치하는 단어가 없으면 결과가 비어 있을 수 있습니다. 이 명령은 소스를 외부 공급자에게 전송하지 않습니다. 저장소 하위 폴더나 일반 폴더를 검색하려면 `rg`(ripgrep)가 필요합니다.

옵션은 질문 앞에 둡니다. 읽어도 되는 폴더를 root로 지정하세요.

## 3. 저장소 preview와 모델 계획 시험

아래 예제 파일은 laya-tools를 clone한 폴더에 있습니다. 실행 파일만 내려받은 경우에는 포함되지 않습니다.

```sh
riidolaya repo-preview --catalog examples/repositories/catalog.json --json 'refund invoices'
riidolaya plan --config examples/planner/config.json --request examples/planner/request.json --json
```

첫 명령은 가상의 billing 저장소를 키워드 후보로 반환합니다. 두 번째는 가상 모델·가격으로 계획을 계산합니다. 둘 다 실제 작업이나 유료 모델 호출을 하지 않습니다. 상태를 해석하기 전에 [기능별 사용법](https://github.com/teamswyg/laya-tools/wiki/Workflows-KO)을 확인하세요.

## 4. 필요할 때만 로컬 추론 추가

```sh
riidolaya setup
riidolaya doctor
riidolaya bench --iterations 10
```

현재 모델 압축은 약 446MiB이며 런타임도 별도로 내려받습니다. 모델 파일은 약 572MiB, 모델을 포함한 프로세스 메모리는 측정상 약 1.4GiB입니다. `doctor`는 경로와 파일 존재 여부를 보여주며 정확도 검사가 아닙니다. `bench`가 실제 모델을 로딩하고 호출합니다.

캐시는 macOS `~/Library/Caches/laya-tools`, Linux `~/.cache/laya-tools`이며 `LAYA_CACHE`로 바꿀 수 있습니다. 현재 setup은 고지를 보완한 models-v2를 설치하고 체크섬을 검사합니다. 기존 `LAYA_*` 변수도 유지됩니다.

다음: [기능 선택하기](https://github.com/teamswyg/laya-tools/wiki/Workflows-KO). 단순 사용을 위해 MCP 등록, daemon 실행, Codex 연결을 먼저 할 필요는 없습니다.

## 개발 실험의 근거 읽기

[외부 Go 작업의 공정 비교 55](https://github.com/teamswyg/laya-tools/blob/main/docs/fair-upstream-comparison.ko.md)는 실제 입력과 독립 검사의 조건을 설명합니다. v2 ID는 기존 humanize·UUID 논리 요청의 버전입니다. [실제 결과55](https://github.com/teamswyg/laya-tools/wiki/Task-Results-55-KO)는 두 요청을 두 프로필로 각각 한 번 실행한 네 기록입니다. 서수는 둘 다 통과했고 UUID는 Sol6 요청만 통과했습니다. 버전·반복·검사를 고유 요청으로 늘려 세지 않습니다.

두 계획은 한 durable ledger의 순서가 정해진 예약 네 개를 공유했습니다. 예약·시작·종료 영수증 모두 네 개, 불명확한 시작은 0이며 실행·인증을 정리했습니다. 실패를 환불하거나 재시도하지 않았습니다. 이 제한은 그 ledger에만 적용하며 호스트 전체나 공급자 내부 호출의 한도가 아닙니다. 입력·원본 언어·계획·예산과 사전 라우터 관측을 결과 전에 고정했습니다. Laya는 두 전체 입력 모두 잘려 보류했으며 초저자원·절감 목표를 달성했다는 결과가 아닙니다.

24개 초기 probe와 개발 기록은 최종 성능 평가를 대신하지 않습니다. [골든셋 규모 안내](https://github.com/teamswyg/laya-tools/blob/main/docs/golden-set-scale.ko.md)의 목표는 도메인마다 서로 다른 보호된 최종 요청 최소 2,400개입니다. 비교 준비, 실제 모델 실행, 학습 라벨과 최종 적격 수를 따로 읽으세요.
