# 저장 byte 기준선 비교

같은 고정 원문 8390462B → 공백만 제거한 JSON 4697570B → 기존 compact 1978694B입니다. Formatting만의 감소 44.0130%, minified 대비 binary 저장 감소 57.8783%입니다. 기존 pretty 대비 76.4173%는 두 효과가 합쳐진 수치입니다. 기존 roundtrip 결과는 수정하거나 반복하지 않았습니다.

Gzip DefaultCompression 각1회: pretty 594236B, minified 400256B, compact 328206B. 각각의 SHA·원문 pin·사전 계획·독립 카운터는 SAVED-BYTES-CONTROL.v1.json에 있습니다. Gzip을 풀어 비교하거나 속도/RSS를 측정하지 않았습니다. 압축 전송/캐시 비교는 decoded JSON64MiB cap·owned payload·RSS를 자동 해결하지 않습니다.

이번 helper1 성공·실패0, json.Compact1·gzip3·retry0. 원래 roundtrip 반복·compact decoder·gzip decoder·원본 API/모델/학습/역할/새 label·최종 자료 읽기·공유 수정·HF는0. Private raw 출력·host 경로·실행 source는 공개하지 않습니다. 학습/성능 승인/최종2,400 성공은 false입니다.
