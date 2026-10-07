# Public training reference audit / 공개 학습 참조 점검

This maintainer command audits the single immutable, original public training
corpus used by the semantic-contrast study. It runs no model and reads no
evaluation, calibration, private task or profile. A corpus audit is evidence about
training references, not model accuracy or readiness to change application state.

이 유지보수 명령은 의미 대비 실험에 사용한 자체 제작 공개 학습 자료 한 버전만
점검합니다. 모델 실행이나 학습은 없으며, 평가·보정 자료, 실제 업무 댓글이나
프로파일을 입력으로 사용하지 않습니다. 학습 참조의 분포를 확인하는 도구이며,
모델 성능이나 실제 앱 상태를 바꿔도 된다는 근거를 제공하지 않습니다.

From the repository root, with Go 1.27.1, on macOS or Linux:

```sh
go run ./cmd/riido-statehint-claims-training-audit \
  --training .cache/statehint/v4-foundation/three-claims-study/three-claim-semantic-contrast-v1/TRAIN1680.REVIEWED.variant-v1.jsonl \
  --training-sha256 c5b5a0adbc721d73213d4565c6a7c430094f31b8e7502fb5249e7ed40a3da837
```

The example assumes that the pinned public corpus already exists locally; the
command does not download anything. Other file contents or hashes cannot be
enabled by flags. The input must be a bounded regular file, with no path aliases,
symlink components or hardlinks. The SHA supplied on the command line is checked
before file access; actual bytes are checked before row decoding. JSON objects
must have every expected field exactly once, with no extra fields. Paired
families must contain one Korean and one English row in the same group; each
group must contain three complete pairs.

위 예시는 해당 공개 학습 파일이 이미 로컬에 있을 때 사용합니다. 자동 다운로드
기능은 없습니다. 다른 자료나 해시를 옵션으로 허용할 수 없습니다. 상대 경로의
일반 파일만 허용하고, 경로 별칭·심볼릭 링크·하드 링크는 거부합니다. 명령에
입력한 해시는 파일을 읽기 전에 확인하고, 실제 바이트의 해시는 행을 해석하기
전에 확인합니다. JSON 중복 키, 추가·누락 필드, 잘못된 증거 범위나 자료 구조도
거부합니다. 한 가족에는 같은 그룹의 한국어·영어 행이 하나씩 있어야 하고,
그룹에는 세 가족이 있어야 합니다.

Standard output is one JSON object of aggregate counts only:

- Locale, head/state row counts and distinct group supports.
- UTF-8 text byte-length buckets: 1–64, 65–128, 129–256, 257–4096, also counted
  separately by locale. Bytes describe input size; the same meaning can require
  different byte lengths in Korean and English.
- Unicode rune (code point) buckets: 1–16, 17–32, 33–64, 65–4096, with global,
  locale and head/state counts. A rune count is not a grapheme count, word count,
  model token count or measure of semantic complexity.
- Empty/nonempty evidence counts and evidence span/text byte fractions by
  head/state. Fractions include empty false evidence as zero; means are row
  weighted. These are annotation coverage descriptions, not evidence quality.
- Number of bilingual families with differing targets on each head. Different
  targets are counted, not corrected or forced to agree.
- Exact duplicate text clusters after Unicode lowercase and whitespace collapse,
  redundant rows, and clusters/pairs with any differing target. No Unicode
  compatibility normalization is performed.

출력은 집계 JSON 한 개입니다. 언어·주장·상태별 행 수와 그룹 수, 언어별 본문
바이트 길이와 유니코드 코드 포인트(rune) 수, 비어 있는 증거와 증거 비율,
언어 간 라벨 차이, 정규화 후 동일 문장의
라벨 충돌 수를 보여 줍니다. 빈 false 증거는 비율 0으로 평균에 포함합니다.
바이트 수는 입력 크기이며, 같은 의미라도 한국어와 영어의 바이트 수는 다를
수 있습니다. 코드 포인트 수는 글자 묶음·단어·모델 토큰 수나 의미 복잡도를
뜻하지 않습니다.
증거 비율은 범위의 크기이며 증거의 품질 점수가 아닙니다. 언어 간 라벨을
강제로 맞추거나 수정하지 않습니다. 중복 검사에 쓰인 문장·해시·행 ID·이유는
출력하지 않습니다.

The corpus pin is public and appears in the report for reproducibility. Group
counts do not establish semantic independence. AI references still require
independent semantic and human/product review. This command has no Fit, Predict,
model load, selection, network request or application-write path.

재현을 위한 공개 자료 해시는 출력에 포함합니다. 그룹 수가 의미적 독립성을
보장하지는 않습니다. AI 참조는 독립적인 의미 검토와 사람·제품 기준 검증이
필요합니다. 이 도구에는 학습·추론·모델 로드·선택·네트워크·앱 변경 경로가
없습니다.
