# Reference 이전 family 구조 확인

`riido-familycohort`는 명시적으로 제공한 metadata bundle 하나의 구조를 확인합니다.
Source 결합 400개, 선언된 family/frame 결합 1,200개(Source당 3개), 제공된 comment
2,400개(family당 `ko`와 `en` 각 1개)가 정확히 있어야 합니다. 내용을 생성하거나
모델·reference 판정을 실행하지 않습니다. 실제 후속 작성에는 전체 Source QA가
별도의 선행 조건으로 계속 필요합니다.

```sh
go build ./cmd/riido-familycohort
./riido-familycohort --root ./supplied-cohort --input bundle.json \
  --sha256 "$BUNDLE_SHA256" --bytes "$BUNDLE_BYTES" --out ./structure-out
```

독립적으로 먼저 고정한 bundle SHA-256과 정확한 인코딩 바이트 수를 제공합니다.
도구는 이 파일을 한 번만 읽고 동기화한 `reviewpacket` start/result 영수증을
남깁니다. 실제 해시·길이, 엄격한 UTF-8·닫힌 JSON 형태와 전체 ID 결합을 확인합니다.
stdout에는 집계 JSON만 나오며 오류에 입력 문장·ID·경로를 넣지 않습니다.
출력 디렉터리는 새로 만들어야 합니다. 파일은 0600, 디렉터리는 0700 모드로
만들고 동기화하며 기존 파일을 덮어쓰지 않습니다. 성공하면 읽은 바이트 그대로를
`BUNDLE.private.json`, 결합·집계를 `STRUCTURE.private.json`, 집계를
`SUMMARY.json`에 보존합니다. 거부 시 집계와 생성된 읽기 영수증은 남기되
검증된 구조를 내보내지 않습니다.

다음 필드는 모두 필수입니다. 알 수 없는 키, 중복·대소문자 별칭, null, 누락,
뒤따르는 JSON과 잘못된 Unicode escape를 거부합니다.

- Bundle: `schema` = `riido-familycohort-bundle-v1`, `frame_schema`(File),
  `frame_schema_version`, `sources`, `families`, `comments`.
- Source: `source_id`, `split`, `dev_style`, `source`(File), `source_review`(File).
  split은 `train`, `dev`, `cal`, `test`입니다. `dev`에서 style은 `short` 또는
  `general`이고 다른 split에서는 빈 문자열입니다. 새 split·label 할당량은 없습니다.
- Family: `family_id`, `source_id`, `frame`(File), `frame_schema`(File).
  schema pin은 bundle에 제공한 schema pin과 정확히 같아야 합니다.
- Comment: `comment_id`, `family_id`, `source_id`, `source`(File), `frame`(File),
  `locale`, `text`. Source ID·파일과 frame 파일은 연결된 family·Source와 같아야
  합니다. split과 DEV style은 이 연결을 통해 상속합니다.
- File: `path`, 소문자 `sha256`, `bytes`. 경로는 root 기준의 정규화된 `/` 상대
  경로이며 절대 경로·상위 이동·`\`·`:`·`./` 같은 별칭을 거부합니다. 바이트 수는
  음수가 아니며 0바이트는 빈 파일 해시와만 결합됩니다. 한 경로에 서로 다른
  해시·바이트 수를 선언할 수 없습니다.

ID는 ASCII 문자·숫자·밑줄·하이픈 1–128바이트입니다. 제공한 frame schema version은
앞뒤 공백이 없는 비어 있지 않은 UTF-8 문자열이며 최대 128바이트입니다.
Source/family/comment ID는 종류별로 중복되면 안 됩니다. 알 수 없는 연결,
누락된 family와 중복 locale 슬롯은 bundle 전체를 거부합니다. 출력의 split/style
집계는 Source에서 상속한 값이며 family·comment를 독립적으로 배분하지 않습니다.

인코딩된 bundle 전체의 **8 MiB** 한도와 comment 하나의 **UTF-8 4,096바이트**
한도는 별개입니다. text는 유효한 UTF-8이며 비어 있거나 NUL을 포함하면 안 됩니다.
도구가 문장을 자르거나 정규화하지 않습니다. 2,400개 comment가 각각 최대 길이면
bundle 한도를 넘을 수 있습니다. 이 초기 전달 방식은 comment 한도를 줄이거나
쉬운 일부만 선택하지 않고 큰 bundle 전체를 거부합니다. 비공개 artifact는 각각
16 MiB, 출력 전체는 64 MiB 이내입니다. 실제 바이트 확인 전에는 수용량을 주장하지
않습니다.

실제로 여는 파일은 bundle 하나뿐입니다. Source·Source 검토·frame·frame schema
File은 형태만 확인하는 불투명한 선언입니다. 검토/schema/frame 파일을 공유할 수
있으며 그 내부의 개별 record·snapshot을 찾아 검증하지 않습니다. 같은 해시·문장은
독립성이나 의미상 중복을 증명하지 않습니다. 제공한 schema pin/version은 결합하지만
내용이나 채택 여부는 검증하지 않습니다. 제안된 8개 QA 축은 이후 검토 차원이며
이 도구가 강제하는 frame 필드가 아닙니다. 축·frame 값·head label을 만들지 않습니다.

성공 상태는 `STRUCTURAL_ONLY_QA_PENDING`입니다. 전체 Source QA, 읽기 근거,
frame schema 의미, 원본 의미, 제공자 신원·독립성, 권리, Source·이중언어 충실성,
자연스러움, blind reference, reference support와 모든 workflow gate는 계속
pending입니다. 의미 적격성, 학습 권한, human Gold 또는 모델 품질을 증명하지 않으며
해당 gate를 대체해서는 안 됩니다.

Go API `familycohort.Validate(Bundle) (Summary, error)`는 파일을 열거나 호출자
slice를 바꾸지 않고 선언을 연결합니다. `Decode([]byte)`는 전달 바이트 한도와
닫힌 형태를, `Run(root, File, out)`는 실제 1회 읽기·무결성 영수증과 비공개 보존을
추가합니다. 자동 연동 등록이나 배포는 없으며 test에는 software fixture만 씁니다.
