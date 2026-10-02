# 소스 관찰 81: 공개 읽기 자료

이 폴더는 mapstructure·pflag의 네 행동 목표를 실행 전에 읽은 기록이다. 두 family의 여러 입력을 독립 데이터로 부풀리지 않는다. 새 원문·라이선스12개와 역사적 Go 라이선스1개, 입력만 있는32개 초안을 보존한다. Want·정답·학습 역할·가중치·실제 parent는 아직 만들지 않았다.

[쉽게 읽는 설명](SOURCE-FIRST-81.ko.md), [원문과 입력 범위](SOURCE-PROPOSAL-81.v1.json), [고지](NOTICE-81.md)를 먼저 읽으면 된다. Go 원문은 .go.txt, module 원문은 go.mod.txt로 저장돼 이 폴더에서 import되거나 실행되지 않는다.

타입 조건·오류 wrapper·nil/empty·첫 Set과 반복 Set·Replace 이후 상태를 구분하는 것이 목적이다. 최소 취득 목록은 컴파일 closure가 아니다. mapstructure5개/pflag38개 runtime 원문 후보의 body가 아직 없고, 전체 package 초기화·표준 파서·복사 원전의 코드 동일성은 후속 범위다. 현재 pflag 원문상 startup NewFlagSet 지점1개는 실제 측정값이 아니다.

초기11개 계획·결과를 고치지 않고 고지용 Go1.23.4 LICENSE만12번째로 추가했다. 전체HTTP19회 모두성공·재시도0·177173B다. private 보완 helper와 metadata assembler는 각각첫컴파일1회실패후성공했고, 두 실패원문해시도 장부에 보존했다. 이것은 upstream compile/API 실패가 아니며 원래 compile/import/init/API/tests0이다.

새 Go LICENSE는2009 저작권, 역사적 LICENSE는2012 저작권이므로 같은1479B여도 바이트가 다르다. BSD/MIT 전문과 원래 고지를 모두 보존했다. Go1.23.4 src/flag/flag.go 본문 비교는 아직하지 않았다.

개인 경로가 있는 helper·바이너리·HTTP 응답 원문·세부HTTP장부는 공개 복사에서 제외하고 SHA/크기만 남겼다. 원래 acquisition 결과의 http/... 상대 저장 경로는 역사적 private receipt 위치이며 이 공개 폴더에 body가 있다는 뜻이 아니다. manifest의 omitted 목록이 그 경계를 명시한다. 초기 결과의 unacquired_full_package_runtime_candidates는 획득 파일도 포함하는 전체 root 후보 목록이라는 원래 이름 한계를 proposal에서 따로 설명했다.

AI 보조 읽기이고 협업 비용은 측정하지 않았다. 별도 모델/paid trial0은 협업 비용0이 아니다. 선행 source53·다른 source/caption 읽기 노출이 있어 blind 자료가 아니며, 원문 저자 흐름·학습 준비·독립 일반화·2400 final 합격을 증명하지 않는다. 공유 저장소·외부 게시와 실제 관찰은 root의 후속 작업이다.
