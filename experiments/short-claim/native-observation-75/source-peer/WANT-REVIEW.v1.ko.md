# Wrap / Marshal 사전 Want 독립 검토

고정된 24개 사전 기대값에서 실행을 막는 의미론·fixture 문제를 찾지 않았다. 먼저 원문 해석을 SOURCE-ORACLE와 SOURCE-RECEIPT로 고정한 다음 작성자의 Want, compiled `pure/spec.go`, 직접 호출 adapter를 읽었다. 기대값의 실제 호출 관측은 하지 않았다.

Wrap 12개는 빈 문자열, 세 단어, 명시적 LF, 짧은 줄바꿈, 긴 단어, 연속 공백, 탭, 한글 rune, NBSP, 선행·후행 공백, 폭0을 다룬다. 모두 원문의 버퍼/조건을 수작업으로 따라 도출된다. Marshal 12개는 nil·빈 map, 전체 행 정렬, 선행0·음수0·plus·minus·비ASCII 숫자·긴 숫자, 빈 값, 7개 특수문자, space/hash를 다룬다. 모두 ASCII digit 검사와 순서 있는 escape·행 정렬로 설명된다. 각 사례의 근거는 FINITE-PROBE-REVIEW에 있다.

입력과 반환 관측의 의미도 맞는다. nil map의 snapshot은 `Nil:true, Entries:null`, 빈 map은 `Nil:false, Entries:[]`로 구별한다. map snapshot의 정렬은 관찰 adapter의 canonical 표현이며 원본 API가 map을 정렬·변경했다는 뜻이 아니다. 12개 Marshal map의 before/after는 원문에서 쓰기 경로가 없는 것과 일치한다. 모든 고정 사례의 error는 null이다. Wrap의 error/receiver 필드는 원래 API에 없으므로 null이지 '오류 처리 관측'이 아니다. panicfalse는 이 유한 source 기대값이다.

실제 Want JSON과 handoff 내 Specification이 같고, seal/plan/handoff의 Want·binary·14개 closure 핀이 연결된다. 14개 파일의 byte/hash와 두 MIT notice 원본/복사본의 완전한 바이트 일치를 검사했다. 실행하지 않고 binary buildinfo를 읽어 Go1.27.1, darwin/arm64, CGO0, trimpath, 두 정확한 local upstream module 연결을 확인했다. 이것은 소스·빌드 정보 provenance이며 hermetic compiler 증명은 아니다.

원본 init은 main이나 argument 검사보다 먼저 실행된다. 원문 parser의 errors.New3개/regexp.MustCompile3개는 소스에서 확인한 초기화 표현이며 동적 계측 횟수가 아니다. 미래 root는 초기화 전에 바깥 reservation을 저장해야 한다. compiled dispatcher는 12 Wrap+12 Marshal 직접 호출만 연결한다. parser/환경·파일 API 호출은 이 경로에 없다. 예약 카운터는 call 직전 의도이며, 중단 시 실제 진입 횟수와 동일하다고 단정하지 않는다. 부분 기록은 prefix를 유지한다. 차이나 panic도 삭제하지 않는다. 완료 exit0만으로24match를 추정하지 말고 state/returns/matches/differences/records를 읽어야 한다.

독립 metadata 검사 2회 중 첫 번째는 검토자 shell의 `path` 변수명이 zsh PATH에 영향을 줘 exit127로 실패했다. 처음 JSON join은 통과했지만 파일 핀 검사는 완료되지 않았다. 실패한 명령을 private에 보존했고, 작업 전용 변수명을 사용한 두 번째 검사에서 terminal PASS를 확인했다. author의 Want·코드·자료는 수정하지 않았다.

이번 결론은 24개 유한 사전 기대값을 뒷받침한다. 2개 행동 목표/2개 원천 가족이며 24개 독립 부모가 아니다. broad truth·후보 충실성·부모 라벨·원천 다양성·학습 준비를 승인하지 않는다. 원본 import/init/API/관찰기/테스트/Features/Project/Fit/모델/HF 실행은 0회다. AI 보조 협업 비용은 측정하지 않았다. 기존 source-before-Want 메모와 모든 과거 artifact는 그대로다.
