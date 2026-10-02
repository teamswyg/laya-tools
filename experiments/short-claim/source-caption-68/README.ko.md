# 실제 upstream 설명을 이용한 작은 행동 힌트 계약68

[English](README.en.md) · [동결 계약](contract-68.v4.json) · [원본 복사 목록](archive-manifest-68.json)

이번 실험은 우리가 직접 만든 설명만 사용하던 범위에, 미리 존재하던 공개 라이브러리 설명을 실제 후보로 붙여 보는 작은 단계다. 모델을 학습한 결과가 아니라, **원문 설명·고정 호출·기대 행동을 연결한 뒤 원본 Go API를 한 번 실행한 기록**이다. 같은 코드 가족을 사용하는 영어 요청4개를2개 전체 source 가족으로 묶었다. 요청4개를 독립 그룹4개로 세지 않는다. 기존72요청·216후보·unknown21·17그룹·정답·mask는 변경하지 않았다.

## 무엇을 물었나

Semver의 두 요청은 정확히 `v1.2.3`에서 leading `v`를 거부할지 허용할지만 묻는다. 후보는 `StrictNewVersion`와 `NewVersion`이다. Strict 후보는 완전한 valid-v2 문장, New 후보는 원래 package-level bullet 전체를 사용한다. 이 bullet은 함수 전용 주석이 아니다. 같은 문서의 parsing 설명과 pinned NewVersion의 coercing 경로를 함께 근거로 연결했다. 모든 Semver 함수·모든 입력·오류·panic·String 값을 약속한다고 확대하지 않는다.

Glob의 두 요청은 명시된 whole-name slash matching, 공통 `src/`·`.go`, 고정 이름 `src/main.go`와 `src/lib/main.go`에서 중첩 경로를 허용하거나 거부하는 작업이다. 후보 패턴은 `src/*.go`, `src/**/*.go`, `src/**.go`다. star·standalone globstar·mid-component doublestar의 완전 원문을 보존했다. 원문의 txt 예시를 go로 고치지 않고, 같은 패턴 구조라는 연결만 기록했다. 파일시스템 탐색·Windows·임의 패턴을 검증한 것은 아니다.

원래 영어 요청과 인용은 [contract](contract-68.v4.json)에 정확한 순서·raw byte·SHA와 함께 있다. `target_Want_not_Got`는 후보별 예상 행동이다. 서로 반대인 Semver 두 요청에서도 이 배열은 `[false],[true]`로 같으며, 요청 목표와 제안 acceptable `[0]`/`[1]`은 별개다. Missing-patch acceptance와 NewVersion.String은 부가 관측이며 모델 목표에 추가하지 않았다.

## 이전 한계도 보존한다

| 기록 | 뜻 |
|---|---|
| [v1](archive/plans/PLAN-68.v1.ko.md) | 실행 전 잘못된 DetailedNewVersionErrors 기대값이 발견된 원래 계획. 원본 보존. |
| [v2](archive/plans/PLAN-68.v2.ko.md) | 두 flag 선언을 true로 바로잡은 계획. 원본 실행은 아직 없었음. |
| [v3](archive/plans/PLAN-68.v3.ko.md) |12entrypoint와 String observer3개를 분리. Generic attempts 문장이 전체 성공·정규화를 약속하지 못하는 한계는 [검토](archive/reviews/source-v3/INDEPENDENT-REVIEW-68.v1.json)에 그대로 남음. |
| [v4](archive/plans/PLAN-68.v4.ko.md) | Prefix 목표로 좁힌 별도 계약. [prefix 검토](archive/reviews/prefix/PREFIX-ONLY-CRITICAL-REVIEW-68.v1.json), [최종 정적 검토](archive/reviews/final-v4/FINAL-V4-REVIEW-68.v1.json), [실행 코드 검토](archive/reviews/worker/INDEPENDENT-WORKER-REVIEW-68.v1.json)를 보존. |

후속 v4 관측이 성공했다고 v3의 부족했던 caption을 사후에 통과로 바꾸지 않았다. 요청은 AI-assisted 초안이며 root가 선택했다. Upstream 문구는 그보다 먼저 존재하던 다른 wording origin이지만, individual/human-only 작성 인증·일반화·authoring-diversity clearance를 뜻하지 않는다. 검토자의 과거 작성 참여와 metadata 노출도 각 receipt에 기록했다.

## 실제로 실행한 범위

[결과](archive/actual/results-68.json)와 [실행 장부](archive/actual/ROOT-ACTUAL-EXECUTION-LEDGER-68.v1.json)는 첫 native 실행1회·재시도0·exit0을 보고한다. 직접 API adapter는 Strict3/New3/Match6=12회, NewVersion.String observer는 별도3회다. Boolean12개와 String3개의 사전 Want 비교15개가 일치했다. Flags는 실행 전후 모두 true였고 변하지 않았다. Init와 내부 helper의 개별 호출 횟수는 계측하지 않았다.

요청 작성에 참여하지 않은 checkpoint reviewer의 [사전 검토](archive/qa/precheck/FINDINGS-PRECHECK-68.v1.ko.md)와 [실제 Got 검토](archive/qa/runtime/FINDINGS-RUNTIME-68.v1.ko.md)도 원본 그대로 보존한다. 이 reviewer는 이전 개발과 사전 자료를 본 nonblind reader이며 원본 API를 재실행하지 않았다. 실제 Got에서 도출한 허용 집합 `[0]`, `[1]`, `[1]`, `[0,2]`가 사전 제안과 같았다. [감독 제안](archive/qa/runtime/PROPOSED-OBSERVED-SUPERVISION-68.v1.json)은10후보의 양성5·음성5와 좁은 scope의 제안 mask만 담는다. 실제 학습 weight는 null, 기존 dataset 변경은0이다. 고유 target API 관측8개를10후보 위치에서 재사용했고 보조 관측7개를 더해15개를 비교한 것이며 모델 정확도가 아니다. 이 archive copier는 해당 별도 검토를 exact copy할 뿐 새 의미 판단을 하지 않는다.

| 첫 실행 관측 | 기록값 | 해석 |
|---|---:|---|
| Controller elapsed wall |0.366244042초|실행기 시작·대기까지 포함한 경과시간. |
| macOS time real |0.36초|OS 도구가 출력한 반올림된 시간. |
| User CPU / system CPU |0.00 / 0.00초|출력 자릿수로 반올림된 값이다. CPU를 쓰지 않았다는 뜻이 아니다. |
| Peak RSS |7,225,344B ≈6.89MiB|macOS가 보고한 최대 resident set. Go heap 또는 GPU 메모리가 아님. |
| Peak memory footprint |4,735,480B ≈4.52MiB|OS의 별도 footprint 지표이며 RSS와 같은 값이 아님. |

한 cold worker의 전체 수명에는 package init·metadata guard·API·observer·출력이 포함된다. 이 수치는 개별 API latency, 학습, 모델 추론, GPU 사용량 또는 일반 성능 분포를 보여 주지 않는다. 실행 전 외부 예약을 하고300초 timeout을 적용했으며, RSS256MiB는 관측 상한으로 기록했다. RSS hard limit은 강제하지 않았다. GOMAXPROCS1과 Go heap soft limit256MiB는 설정값이며 측정값·hard RSS 보장이 아니다.

## 자료의 범위와 라이선스

모든 원본 복사는 [manifest](archive-manifest-68.json)에 개별 SHA·바이트 수·원래 snapshot 이름을 기록한다. 이전 receipt는 이름과 의미를 수정하지 않았다. `.go`와 `go.mod`는 `.go.txt`/`.mod.txt`로만 보관했으며 public runtime port가 아니다. Historical 실패 source와 장부는 보존하고 raw native/synthetic 로그는 제외했다. 원래 receipt가 가리키는 제외 로그의 hash는 목록에 남는다. 모델·binary·로컬 절대경로를 포함하는 controller 원문은 게시하지 않는다.

Upstream15개 원문99035B는 이미 저장된 [MIT manifest](../../../internal/publicbehavior/upstream-manifest.json)와 [Semver 전체 고지](../../../internal/publicbehavior/testdata/upstream/semver/LICENSE.txt), [doublestar 전체 고지](../../../internal/publicbehavior/testdata/upstream/doublestar/LICENSE)를 참조한다. 원문을 다시 복제하지 않는다. [원문 catalog63](../upstream-wording/quote-catalog-63.json)의 출처·revision·raw quote를 유지한다. 새 maintainer Go 코드는 Apache-2.0이다.

이번68 실험과 archive copy에서 Features·role·fit·model/judge/paid trial·protected-final 실행0이며 training_ready=false다. 별도 model/API0은 AI-assisted Codex 협업의 사용량이나 비용0을 뜻하지 않는다. 협업 비용은 측정하지 않았다. 원문 설명과 작은 행동 관측의 연결을 얻었지만, 아직 학습 자격·광범위한 정확도·토큰 절감·최저 메모리 모델의 효용을 입증하지 않았다.
