# 다음 네 원천의 실행 범위와 고지

godotenv·wordwrap·logfmt·dataurl의 고정 공개 원문을 읽고, 다음7개 행동 목표의 실행 범위를 검토했다. 목표·입력값을 늘린 제안이며 새 독립 부모나 정답이 추가된 것은 아니다. [한영 상세 조사](SOURCE-AUDIT-74.ko.md)와 [원문 검사 기록](SOURCE-READ-RECEIPT.v1.json)을 보존한다. 조사 단계에서는 원본 API·init·모델·학습 실행이0회다.

첫 실행 대상으로 wordwrap의 정확한 공백 배치와 godotenv Marshal의 정렬·인용·정수 문자열 보존을 선택했다. logfmt는 순서·오류·다음 호출 상태를 함께 기록해야 한다. dataurl은 작은 percent helper를 먼저 다루고, DecodeString의 goroutine 종료와 없는 go.mod의 유지보수 recipe는 별도로 준비한다. 각 원천의 범위를 좁히는 이유와 미확인 사항은 상세 조사에 있다.

MIT 원문 고지를 보존했다. logfmt/dataurl에는 Go 표준 소스에서 가져온 코드가 있으므로 [Go 출처 고지](attribution/NOTICE-Go.txt)와 [BSD 전문](attribution/Go-BSD-LICENSE)을 함께 보관한다. Root도 공식 고정 Go LICENSE 사본의 SHA와 해당 전문의 일치를 확인했다. 공식 비교 revision은 복사 당시 release를 입증하지 않는다. 소스의 조건부 재사용 경로를 확인한 것이며 학습 데이터나 가중치 권리를 자동으로 승인한 것은 아니다.

`upstream`에는 공개 source closure19파일과 README4개를 바이트 그대로 보관한다. Go 소스·go.mod·go.sum은 원래 파일명 뒤 `.txt`를 붙여 연구 원문으로 보관하며 런타임 의존성으로 가져오지 않는다. SHA는 확장자 변경 전 원문 바이트와 같다. [사본 장부](../COPY-74.v1.json)가 실제 게시 대상 파일을 연결한다. 원본 tree·공식 비교 자료는 고정 GitHub 링크로도 확인할 수 있다. 동일 helper·alias·wrapper·번역은 향후 같은 그룹으로 묶으며 단순히 같은 저자명이나 표준 라이브러리를 썼다는 이유로 모두 합치지는 않는다.

이 사본에 한정한 Git attributes는 원래 bytes를 유지한다. 두 MIT 고지의 마지막 빈 줄과 logfmt README의 CRLF를 바꾸지 않고 정확한3파일의 whitespace 조건만 명시했다. 자체 작성 문서의 형식·보안 검사와 CI는 계속 적용한다.

[English](README.en.md)
