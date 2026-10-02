# 특징 해시 prefix 재사용 독립 읽기 검토

코드 호환성이나 기준선의 공정성에서 blocker는 찾지 않았다. FNV-1a는 각 byte마다 같은 XOR·uint64 곱셈을 적용하므로 query→NUL→document를 한 번에 읽는 것과 query+NUL의 중간 상태 뒤에 document를 이어 읽는 것은 같다. UTF-8/내부 NUL byte도 이 성질을 바꾸지 않는다. Query/document term 순서, sign/index, scale, sort와 충돌 FP64 합산 순서, 마지막 lexical recall 열은 그대로다. 제어 경로의 originalByteHash는 이전 HEAD의 직접 loop이며 pair마다 callback을 추가하지 않는다.

저장 benchmark의 **보고 계산 오류1개**를 발견했다. Bounded32의 새 경로5값을 정렬하면 `281536,281685,282633,282987,288122`이며 중앙값은282633ns이다. 초안 METRICS/KOEN 문서의282987ns·19.53%는282633ns·약19.63%로 정정해야 한다. Root가 오류를 인정했고 원래 초안을 private에 보존하며 고칠 예정이다. 이 v1은 정정 확인 전 기록이며, 이후 확인은 별도 receipt로 남긴다. 기존 benchmark나 Features를 다시 실행할 필요는 없다.

짧은 예제 중앙값4240→3214ns(24.20%)와 할당2216B/19회, bounded32의 할당70792B/72회 동일성은 저장5표본과 맞는다. 두 synthetic 문자열 예제와 이 host/toolchain 실행에 한정된다. 이 수치만으로 전체 요청 latency·학습/추론 효용·RSS·일반적인 모든 입력의 속도 개선을 주장하지 않는다. 배열 layout, 의미 특징과 학습 목표도 바뀌지 않았다.

검토자는 변경 작성자가 아니며 static diff/이전 HEAD/합성 테스트 코드/기존 benchmark text를 읽었다. 코드·테스트·benchmark·Features/Fit/corpus/모델/원본 API는 재실행하지 않았고 공유 수정·외부 게시도0이다. 기존 테스트의 byte-hash 대조와 sparse/score-bit 대조는 타당한 합성 설계이나, 이 검토자의 새 실행 성공으로 세지 않았다. AI 보조 협업 비용은 측정하지 않았다.
