# 원문 목록을 채우고 복사 구간을 확인했다

이전81에서 부족하다고 기록한 mapstructure5개·pflag38개와 Go1.23.4 `src/flag/flag.go` 한 파일을 추가 취득했다. 먼저 고정 계획을 저장했고, 공식 GitHub의 고정 커밋과 blob SHA1·크기를 대조했다. 44개 body 전부 확인됐다. HTTP46회 모두 성공·재시도0, 원문192489B와 메타데이터16781B로 총209270B였다. source81 원문13개도 수정 없이 로컬 복사했다. 현재 원문·module·라이선스57파일은332806B다.

원천 취득에서 기대값·정답·역할을 만들지는 않았다. 기존81의32개 입력 초안도 Want가 없는 제안이며 그대로 유지한다. import·init·원 API·테스트·모델·학습은 모두0회다.

`pflag/text.go`가 복사했다고 밝힌 구간을 확인했다. `type textValue`에서 마지막 String 메서드까지868B는 pflag10–43행과 고정 Go 원전299–332행에서 바이트가 같았다. 전체 파일은 다르다. pflag의 package/import, 복사 문구, Type/GetText/TextVar 등 추가 메서드와 Go의 다른 값·초기화는 이 같은 구간 밖이다. 이 결과는 해당 구간의 원전 연결이며, 두 package 전체의 동작이나 모든 입력에 대한 동등성을 증명하지 않는다. Go LICENSE2009 전문의 확인과도 다른 증거다.

취득한43개 목록은 바꾸지 않았다. mapstructure의 `reflect_go1_19.go`와 `internal/errors/join_go1_19.go`는 `!go1.20`, 상대 파일은 `go1.20` 조건이다. Go1.27용 정적 후보는 mapstructure6개와 pflag42개, 합48개이며 오래된 조건의2개 파일은 그 후보에서 제외하는 제안이다. 현재 컴파일러로 파일 선택을 실행한 것이 아니다. 구버전 원문도 기록·배포에서 삭제하지 않는다. 예전 원문 목록의 ‘missing’은 나중에 그 도구체인에서 모두 실행된다는 뜻이 아니었다.

추가로 `internal/errors/join_go1_19.go`에서 Go Authors2022 BSD 헤더를 발견했다. mapstructure 루트 MIT만으로 그 헤더를 덮으면 안 된다. pflag `golangflag.go`도 Go Authors2009 헤더를 갖는다. 원래 헤더와 rootMIT/rootBSD/GoBSD 전문을 모두 유지한다. Go2022 join 파일의 원전 revision과 복사 동일성은 아직 불명확하며 임의로 채우지 않았다. Go2012 역사적 LICENSE도 새2009 전문과 별개로 남긴다. 자세한 고지는 [NOTICE](NOTICE-82.md)에 있다.

이제 pflag 등록 과정의 `isNoOptBoolValue`도 `bool.go`15–20행에 연결된다. 선택된 모든 runtime 파일의 import 블록을 문자열로 읽었으며, 발견된 외부 모양의 경로는 같은 module의 `internal/errors`뿐이다. 이 파일 body는 보존됐다. 원 parser·Go compiler·dependency resolution을 실행한 것은 아니므로 실제 빌드 성공·hermetic closure·표준 라이브러리 초기화 계측을 주장하지 않는다.

후속 관찰 준비에서 해야 할 일은 작다. 실제 Go1.27.1 빌드가 선택한 원문·adapter·binary를 고정하고, import 전에 별도 실행 예약을 남기며, package 초기화까지 child 전체 수명에 포함한다. pflag의 `CommandLine = NewFlagSet`은 정적 호출 지점이지 실제 반환을 센 값이 아니다. 비교용 Go1.23.4 flag.go는 inert 텍스트이며 나중에 실행되는 Go1.27.1 표준 flag 원문으로 취급하지 않는다. 정확한 날짜·CSV 실패 관측의 Want는 후속 source-supported 범위로 남겨 둔다.

계획 생성과 취득 helper는 각각1회 성공했다. source 봉인 helper는 중첩 map의 Go 타입 표기를 빠뜨려 컴파일에서2회 실패한 후3번째에 성공했다. 두 실패는 HTTP·원문 실행·출력 생성 전에 발생했고 실패 source 해시를 보존했다. HTTP 재시도나 원 API 실패로 바꾸어 기록하지 않았다.

AI 보조 원문 읽기이며 Codex 협업 비용은 측정하지 않았다. 별도 모델/paid trial0은 협업이 무료라는 뜻이 아니다. source53·79·81 등에 노출된 비블라인드 작성자다. 두 가족의 alias/helper·Cobra/pflag·Mitchell/wordwrap·Go 저자 흐름을 독립 데이터로 승인하지 않는다. 실제 parent·labels·roles·weights·훈련 준비·2400 final 검증은 모두 이번 범위 밖이다.
