# 공개 근거74 읽기 전용 검토

검토 범위의 게시 blocker는 남아 있지 않다. 사본 장부48개, 네 원천 closure19파일과 README4개, 설명 작성자5파일·별도 검토자6파일이 원본 바이트와 일치했다. 원문 source19파일은 고정 Git blob SHA1도 같았다. 전체 MIT 고지4개와 Go BSD 전문1479B(dd26a7…)·출처 고지가 보존되며 공식 비교 revision을 실제 복사 당시 release로 단정하지 않는다. 이 검토는 소스의 조건부 고지 보존을 확인하며 학습 데이터·가중치의 포괄적 권리를 승인하지 않는다.

저장된 PR97/98 API 기록은 각 고정 head/merge와 필수 CI4개 성공을 연결한다. 로컬 Git 객체에서 두 head/merge tree가 같았고 Wiki97 여섯 파일59,678B도 고정 원래 head와 저장 Wiki commit152ff088…의 byte/SHA/blob이 같았다. 과거 remote-pending 사본 기록과 이후 actual published 기록을 혼동하지 않았다. 새 GitHub API 조회나 fetch는 하지 않았다.

root가 추가한 설명은 같은 부모3개·후보9개·소스 구간26개, 설명8개 지원과 query의 제한된 조건부 지원, 정답/역할/가중치 null을 보존한다. 정보 부족을 음성 정답으로 바꾸지 않는다. root 초안의 “BCE 가중치0으로 pair 손실이 차단되지 않는다”는 오류는 게시 전에 정정됐다. 고정 ranking 소스531ce807…에서 endpoint SampleWeights의 곱이 pair weight이고,0이면 gradient/NLL에서 skip함을 읽기 전용으로 확인했다. 감사 쌍의 보존과 실제 손실 기여는 다르다. 원래72 결과·모델·소스는 변경되지 않았다. 원래 caller 바인딩 상세는 별도 작성자/peer 기록의 근거이며 여기서 다시 실행하지 않았다.

공개 fixture 특징 시간의 정렬 중앙값은4240→3214ns(24.198%)와351653→282633ns(19.627%)로 root의 약24%/20% 설명과 같다. 할당량은2216B/19회와70792B/72회 그대로다. 단독 reader 저장 기록의 RSS는19,120,128B/19,300,352B(18.23/18.41MiB), 마지막 heap4,660,664B/4,765,512B이므로 compact RAM 절감은 입증되지 않았다. 누적 할당44.84/21.61MB는 별도 지표다. 고정 순서·warm 파일·한 child씩이며 이전82.3MiB 전체 왕복과 직접 비교하지 않는다.

검토한 root 문서15개의 로컬 링크348곳이 존재한다. 외부 URL의 현재 응답이나 anchor 의미는 확인하지 않았다. source archive에 한정한 -text와 정확한3파일의 whitespace 설정은 고정 MIT EOF/README CRLF 바이트를 유지하며 자체 문서·보안·CI 검사를 끄지 않는다. 공개 선택 범위에서 host 경로와 일반 HF/GitHub 토큰 형태의 일치를 찾지 못했다. 전체 secret scan은 root 실행 보고에 귀속한다.

검토자는 reader74 작성자이며 observer75 작성자와 prior source/projection exposure도 있다. 자신의 reader stage32개 사본이 같다는 확인은 독립 구현·품질 검증이 아니다. reader24열 등의 실제 비교는 저장된 root/별도 runtime QA 기록의 reported comparison으로만 다룬다. 새 원 API·init·reader·codec·controller·특징·역할·fit·모델 실행, 공유 수정, 외부 게시는0이다. 앞선 잘못된 경로 검색과 예상 부재 디렉터리 조회의 비정상 종료2개, 출력 truncation2개를 장부에 보존했다. 새 수치 gate·학습 준비·다양성·일반화·생산 승인0이다.
