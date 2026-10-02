# 69 기존 입력 한도 결과의 읽기 전용 검토

원래 probes56/56b의 요청72개·caption216개 모두 기존 decoded-input 계약에 들어간다. 별도 stdlib 계산으로 총288개 텍스트의 raw bytes, 정규화 bytes, 정규화 단어 수를 실제69 결과의 모든 행과 대조했고 차이가 없다. Supported72/unsupported0, 한도 초과 caption0, 최대 요청32단어·caption32단어가 일치한다. 정적 길이 초과 의심은 실제 결과와 이번 검토에서 확인되지 않았다. 문장 축약·삭제·한도 변경·라벨 변경은 필요하지 않다.

| 기존 텍스트 길이 확인 | 요청 | caption |
| --- | ---: | ---: |
| 텍스트 수 | 72 | 216 |
| 최대 raw bytes | 236 | 252 |
| 최대 normalized bytes | 225 | 239 |
| 최대 normalized words | 32 | 32 |
| 정확히32단어인 텍스트 수 | 8 | 24 |

규칙은 raw512bytes 이하, 정규화512bytes 이하, 정규화1~32단어다. 32는 허용되며33부터 거부한다. 원래 pinned splitter는 Unicode letter/digit을 소문자로 바꾸고 다른 문자를 token 경계로 처리한다. lower/digit 뒤 uppercase 및 uppercase acronym 뒤 lower가 이어질 때 분리하며 token 순서·반복을 유지한다. keyword 제거·단어 집합 deduplication·문장 truncation은 정규화에 없다. 이 문장을 input.go와 lexicalhint/features.go의 실제 source SHA에 연결해 읽고, 같은 조건을 별도 streaming token 조립 방식으로 계산했다. 이 규칙만 사용했으며 lexical feature vector·ranking을 계산하지 않았다.

모든 parent/candidate ID·순서·후보 수와 원래 입력 SHA를 확인했다. 기존 decoded Validate의 candidate1~8, candidate ID1~64bytes의 flat ASCII 규칙, 중복 ID 거부, UTF8 조건과 고정 schema/provenance도 현재 원본에서 충족된다. 전체 parent 결과의 supported와 빈 fixed error가 이 조건들에 맞다. 이 결과는 decoded Validate 계약 검토이며 Load의 별도 JSON wire12KiB 제한이나 모든 임의 입력·forged Prepared에 대한 새 검증 실험은 아니다.

실행 전 plan1,677bytes의 SHA, actual result88,420bytes의 SHA, invocation421bytes의 SHA, source3개와 input2개의 정확한 byte/SHA가 맞다. root의 한 번의 audit process는 Validate72회와 명시적인 metric NormalizeText288회를 기록한다. Validate 내부 정규화 호출은 개별 계측하지 않았고 더하지 않는다. Project/Features·original behavior/role API·fit·model·paid·protected-final은0이다. 이번 검토의 original Validate/NormalizeText/Project/Features 및 실제 role/source/fit 재실행은0이다.

새 `MECHANICS-BOUNDARY-69.v1.json`의 raw text SHA는 원래 문자열의 digest다. normalized text SHA는 이번 별도 규칙 계산의 digest이며 원래 함수가 출력·기록한 digest로 가장하지 않는다. root69 결과에는 정규화 문자열/digest가 없고 길이 정보만 있다. 정규화 규칙과 모든 길이·상태를 비교한 근거와 별도 재구성값을 구분한다. 원문 텍스트·정규화 텍스트·특징 벡터는 새 공개 후보 산출물에 넣지 않았다. 예전66/68 receipt와 원본 결과도 수정하지 않았다.

계약 지원은 runtime 입력 범위의 통과를 뜻한다. Source/caption 의미 충실성·supervision eligibility·loss mask·whole group별 유효 라벨 수·다양성·학습 또는 hint 성능을 승인하지 않는다. 기존72부모/216후보·unknown/라벨·mask·group/role을 변경하지 않았고 새 cap·scope·수치 gate를 추가하지 않았다. 현재 원본은 수용되지만 이후 다른 문장이 들어가면 동일한 기존 검사로 다시 판별해야 한다.

이 검토자는69 root helper의 저자가 아니며 이전62/67 구현과66/68 검토에 노출된 비맹검 검토자다. Source origin·사람만의 저작·compiler/binary 신뢰나 final 블라인드 검증을 증명하지 않는다. 한 번의 stdlib-only metadata helper가 성공했고8개 pinned file/205,177bytes를 읽었다. 합성 테스트3개를 담은 uncached race1회와 vet1회도 통과했다. 테스트는 camel/acronym/digit/Unicode·token 반복·정확한32/512경계·Unicode 정규화 byte 증가·빈 텍스트/invalid UTF8·flat ID를 확인한다. helper/test/vet/필수 pin 실패는0이다.

준비 중 source 위치를 잘못 예상한 탐색 명령1건은 실패로 보존했다. 실제 pinned 경로 pkg/shortclaim/input.go와 internal/lexicalhint/features.go를 읽어 복구했고 원본 source가 없다고 판단하지 않았다. 공유 수정·Git·게시·실제 학습·추가 original API 실행은0이며 일반 AI 협업 비용은 측정하지 않았다. 이번 검토에서 resource benchmark를 실행하거나 root69의 성능 수치를 만든 일은 없다.
