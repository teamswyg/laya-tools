# Opaque revision에 묶는 shadow 표시

이미 보유한 본문에 진행 보고·완료 보고·질문 후보를 붙일 위치를 조건부로 확인하는 Go library입니다. 실제 HTTP/native reader·인증·reaction·알림·상태 쓰기는 구현하지 않았습니다. 테스트 reader와 점수는 명시적인 합성 double입니다.

`Anchor`는 opaque scope와 Work/Content 참조, **소유 경계가 제공한 업무 revision**, 본문 SHA-256을 분리합니다. `OwnerContentRevision`이 없으면 빈값 그대로 남깁니다. `TextDigest`는 최대4,096바이트의 유효한 UTF-8 본문을 그대로 해시하며 trimming·소문자화·정규화를 하지 않습니다. 파생 digest는 소유자가 발급한 revision·권한·admission·atomicity 증거가 아닙니다. 시간·행 번호·digest·숫자 일부로 업무 revision을 만들거나 대체하지 않습니다.

`RevisionReader.ReadAnchor(ctx, scope, workRef, contentRef)`는 expected revision/hash나 모델 입력을 받지 않습니다. 현재 SHA는 신뢰할 수 있는 앱 경계가 실제로 보유한 현재 본문에서 계산해야 합니다. 읽을 수 없거나 모르면 digest를 비워 unavailable로 처리합니다. 사용자 JSON의 hash 또는 모델 입력 hash를 현재 내용의 증거로 반사하면 안 됩니다. 이 규칙을 구현하는 실제 native adapter는 아직 없습니다.

`ValidatePlacement(expected, current)`는 scope·Work/Content 참조·opaque 업무 revision·제공된 content revision·SHA를 정확히 비교합니다. 합성 `p12-w9` 같은 revision을 문자열 그대로 허용하며 일부 숫자를 추출하지 않습니다. 필수 metadata가 없거나 잘못되면 unavailable, 변경됐으면 stale입니다. 완전히 같은 경우도 결과는 **`placement_matched`라는 조건부 관측**이고 `actual_verified:false`입니다.

`Adapter{Revisions, Catalog, Predictor}.Propose(ctx, Request)`는 이미 보유한 본문 hash를 확인하고 current anchor를 읽은 뒤 원본 catalog schema를 검증합니다. 같은 scope·Work에 대한 라벨/이모지 연결만 진행·완료 보고·질문으로 제한하고 상태 계획을 끕니다. confidence0.9·margin0.05를 유지하며8개 확률을 재정규화하지 않습니다. 분류 후 anchor를 다시 읽어 변경 또는 미확인 상태면 표시 후보를 폐기합니다. 업무의 opaque revision을 catalog의 기존 숫자 버전에 넣지 않습니다.

모든 결과는 shadow, mutation/state-change false, catalog consistency unqualified입니다. 별도 조회의 일치는 atomic snapshot을 증명하지 않습니다. catalog와 권한은 향후 실제 표시를 적용하는 앱이 다시 확인해야 합니다. native ID·이름·credential과 대응표는 private adapter 안에 유지하고 공개 코드에는 caller opaque 참조만 전달합니다. 보고 결과에 원문·업무 revision·본문 digest를 되풀이하지 않습니다.

반복 호출은 현재 reader를 다시 사용하며 private 공용 cache나 lock이 없습니다. 현재 catalog에 동일 라벨/이모지가 존재하면 추가 제안은 no-op입니다. 분류기 비용과 실제 native metadata 조회·앱 표시 비용은 별도이며, 후자의 실제 검증은0회입니다.
