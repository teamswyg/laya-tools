# 68 실제 관측과 고정 요청의 범위 검토

첫 native 실행 결과를 읽어 도출한 허용 후보 집합은 `[0]`, `[1]`, `[1]`, `[0, 2]`로, 실행 전 제안과 일치한다. 네 요청의 10개 후보 위치를 모두 확인했다. 마지막 요청은 두 후보를 동시에 허용하며, 하나를 임의로 선택하지 않았다. 현재의 좁고 명시적인 목표 범위에서 중대한 관측·도출·원문 문맥 불일치는 발견하지 못했다.

이 검토자는 68의 요청 초안, v4 계획·계약, native worker, upstream 문구를 작성하지 않았다. 이전 62·67 구현과 이번 사전 검토에 노출된 비맹검 검토다. 독립 원천 제작, 사람만의 저작 여부, compiler/binary 신뢰를 증명하는 검토로 부르지 않는다.

| 고정 요청 | 후보별 실제 target Got | 실제 관측에서 도출한 허용 후보 |
| --- | --- | --- |
| `v1.2.3`의 leading-v 거부 | Strict `[false]`, NewVersion `[true]` | `[0]` |
| 같은 입력의 leading-v 허용 | Strict `[false]`, NewVersion `[true]` | `[1]` |
| `src/main.go`와 `src/lib/main.go` 둘 다 매칭 | `src/*.go` `[true,false]`, `src/**/*.go` `[true,true]`, `src/**.go` `[true,false]` | `[1]` |
| 직접 경로 매칭·중첩 경로 거부 | 같은 세 실제 벡터 | `[0,2]` |

요청의 목표는 계약의 후보별 Want에서 역으로 만들지 않고 고정된 요청에서 각각 `[false]`, `[true]`, `[true,true]`, `[true,false]`로 해석했다. 실제 Got을 그 목표에 대조해 후보 라벨과 전체 허용 집합을 도출한 다음, 별도로 실행 전 Want·제안과 비교했다. 후보의 원문 설명이 정확해도 요청을 충족하지 않으면 이 범위의 유효한 음성이다. 원문 미지원 또는 관측 실패를 새 음성 사실로 바꾸지 않는다.

Strict는 `v1.2.3`에서 반환했고 error가 있으며 version은 없어 수락 Got이 false였다. NewVersion은 error 없이 version을 반환해 true였다. 각 glob 호출은 error·panic 없이 반환했다. Got 정의 여부, 비교·일치 플래그와 두 층의 호출 카운터를 메타데이터에서 재계산했다. Strict3/New3/Match6의 entrypoint12, New.String3, Strict.String0, flag 읽기2가 일치한다. 두 flag는 true/true로 실행 전후 동일하다. 15개 Want 비교는 15일치·0불일치·0미비교로 재계산된다.

이 15개는 target과 보조 관측의 혼합이며 모델 정확도가 아니다. 고유 target API 관측은 leading-v2개와 glob6개인 8개다. 네 요청의 10개 후보 위치가 이를 재사용한다. Semver의 다른 입력 `1.2.3`·`1.2`에 대한 수락4개와 New.String 정규화3개인 보조7개는 leading-v 라벨에 넣지 않았다. 정상화·missing-patch·error 종류·no-panic·모든 입력에 대한 문장 약속으로 확장하지 않았다.

원문은 사전 검토에서 동결한 계약과 계획 그대로다. 실제 결과의 계약 SHA, 계획 SHA와 사전 검토 receipt를 연결하고, doc.go·match.go의 exact file bytes/SHA 및 10개 위치에 재사용되는 완전 원문5개의 byte/line/SHA를 다시 확인했다. Strict의 valid-v2 완전 문장과 digit-only source 연결은 이 고정 입력의 거부를 설명한다. optional-v 문구는 패키지 능력 문구이며 함수별 주석이 아니다. 기존 parsing 문맥·coercing regex와 실제 NewVersion(v1.2.3)에 한정된 문맥 연결을 유지한다. 이 문구로 모든 API의 v 허용을 주장하지 않는다. Glob의 원문 `.txt` 예제도 바꾸지 않았고, 고정 `src/`·`.go`·whole-name/slash 조건 및 같은 패턴 구조에서만 해석한다. 이 설명에는 사전 검토의 source 맥락 확인을 사용했으며 버전 파서나 regex를 실행하지 않았다.

별도 `PROPOSED-OBSERVED-SUPERVISION-68.v1.json`은 고정 유한 목표의 후보별 nullable 라벨과 제안 loss mask만 기록한다. 현재 10개는 모두 알려진 관측으로 양성5·음성5이며, 각 제안 mask는 이 좁은 문맥·목표에만 true다. 실제 학습 weight는 null이고 기존 dataset 라벨·mask에는 반영하지 않았다. 실패/미반환/panic 및 glob error는 unknown/null로 남기고 제안 mask를 false로 만드는 의미를 합성 테스트로 확인했다. 반환된 version error는 기존 관측 계약에 따른 알려진 거부다. 후보 전체와 다중 허용을 유지하며, 부분 관측을 완전한 허용 집합이라고 주장하지 않는다.

두 source family와 다섯 upstream 문구를 재사용한 네 요청이지 네 독립 source group이 아니다. Strict 함수 이름은 어휘 shortcut의 여지를 유지한다. 원천·ID·source binding·Want/Got·라벨·검토 값은 옆자료이며 feature나 전체 caption 승인 scalar가 아니다. 사람만의 저작·다양성 확보·실용 hint 품질·일반화·2400 final·비용 절감·학습 준비 완료를 승인하지 않는다. 기존72부모/216후보/unknown21/whole17그룹 및 기존 감독 mask·역할은 변경하지 않았다. 새로운 수치 gate도 추가하지 않았다.

측정 원문과 root 장부는 단위·값이 일치한다. OS real은 0.36초, controller wall은 0.366244042초다. 최대 RSS는 7,225,344 bytes = 6.890625 MiB, peak footprint는 4,735,480 bytes = 약 4.516106 MiB로 서로 다른 측정값이다. OS user/sys 0.00초는 소수 둘째 자리 출력이며 CPU를 전혀 안 썼다는 의미가 아니다. GOMAXPROCS1과 Go heap soft limit268,435,456bytes는 설정이다. RSS observation cap도 hard RSS 제한이 아니며 worker가 resource를 직접 측정한 것으로 보지 않는다. 한 번의 cold process 수치에 init·guard·adapter·observer·출력이 포함된다. API만의 지연시간이나 Laya/LLM/GPU 성능·반복 실행의 안정성은 이 결과에서 알 수 없다.

이번 stdlib-only 메타데이터 helper1회가 성공했다. bounded read11회/87,074bytes, 합성 테스트3개를 담은 uncached race1회와 vet1회가 통과했다. runtime 검토 중 실패 명령0·필수 pin 실패0이다. 초기 출력 예산 잘림1건과 틀린 예상 JSON 키가 반환한 null은 실제 aux 키·전체 필요한 projection을 다시 읽어 복구했으며 누락으로 해석하지 않았다. 이전 사전 검토의 탐색 읽기 실패2건은 기존 receipt에 보존했다. 이 검토의 native/upstream API/init/AST/formatter/Bind/Generate/Features/roles/fit/model/paid/protected-final/공유 수정·Git·게시 추가 실행은 모두0이다. 일반 AI 협업 비용은 측정하지 않았다.
