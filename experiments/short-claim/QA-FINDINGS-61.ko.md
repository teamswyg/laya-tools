# 61 판정 기록의 독립 바이트·참조 QA

검토한 판정 파일을 바꾸지 않고, 975개 기록의 저장 연결과 해시를 독립적으로 확인했다. 615개는 기존 59 검토 항목이고, 360개는 같은 후보의 관측 필드·명시적 부정 경계에 대한 추가 축이다. 새 요청이나 학습 정답 975개를 확보한 뜻은 아니다.

나는 61 준비용 참조 생성기를 작성했지만 실제 판정 작성자는 아니다. 이 검토는 그 판정 작성과 독립인 **저장 형식·바이트·참조 대조**다. 이미 원천과 metadata를 읽었으므로 블라인드 검토가 아니며, 독립 데이터 원천이나 문장 의미의 정확도를 새로 증명하지 않는다.

고정 manifest는 `f8f0b2320bb214f102575384cace477067a0f075b98ee09c071e778d9285224b`, 결과 receipt는 `report.json`이다. 31개 JSON의 SHA와 크기, 원본 Go 4개 파일, raw span 98개, 선언 연결 105개, 49개 root와 172개 component 관계를 확인했다. 전체 enum 선언을 공유하는 여러 ID도 저장된 관계 순서대로 유지했다. 기존 formatted·normalized·bundle 해시는 저장된 pointer와 대조했고, Go AST·소스 formatter·후보 함수는 실행하지 않았다.

975개의 record digest, 975개의 전체 인용과 9개의 부분 인용 범위, 1,860개의 request/caption 참조, 6,680개의 evidence 연결이 맞았다. 615개 원래 항목 및 360개 추가 항목이 shard와 manifest 사이에 일대일로 대응한다. 60개 부모, 180개 후보 위치, 15개 계약과 후보 수 분포 11개×2·38개×3·11개×4가 유지된다. 마지막에 입력 31개와 원본 소스 4개를 다시 읽어 변경 없음도 확인했다.

| 기존 항목 | scoped consistent | contradicts | omits | uncertain |
|---|---:|---:|---:|---:|
| 615개 전체 | 446 | 94 | 10 | 65 |
| 후보 설명의 source fidelity | 177 | 0 | 0 | 3 |
| 후보 설명의 request coverage | 32 | 94 | 7 | 47 |

이 표는 작성자의 판정 상태를 다시 집계한 결과다. 원천 충실성과 요청 충족은 다른 축이며, consistent가 후보 정답이라는 뜻도 아니다. 선택 범위의 unknown 18개와 전체 unknown 21개, 전체 그룹 17개와 known 포함 그룹 16개는 그대로다. 기존 60 overlay SHA `1beb5c06d3fb70c0a5e0727de793f1890b4ab4a47f1d392a6666b068f096c526`도 바뀌지 않았다. 판정·그룹·source ID·기존 truth는 provenance이며 scorer feature로 만들지 않았다.

실제 metadata QA는 두 번 실행했다. 첫 실행은 verifier가 legacy `Components:null`을 배열로 가정해 `array_shape`로 실패했다. 기존 저장 형식의 빈 관계 표현을 helper가 잘못 취급한 것이며 원문 판정의 문제를 발견한 것은 아니다. 실패 report를 보존하고 그 표현만 좁게 지원한 뒤 두 번째 실행에서 통과했다. 준비용 합성 race 테스트는 세 번 모두 통과했고, 마지막은 테스트·하위 테스트 32개와 package pass를 확인했다. vet도 세 번 통과했다. 장부는 `ledger.json`에 있다.

실제 판정 생성 1회·실패 0회, 판정 writer의 준비 빌드 2회·실패 1회 및 writer 자신의 바이트 QA 2회·실패 1회는 별도 이력으로 기록했다. 이 verifier의 실패와 섞지 않았다. 원본 판정 수정·재생성, 새 의미판정·truth·라벨·원본 graph·seed·roles·fit·weights·모델 또는 후보/API 실행·protected final 열람·공유 레포 수정·게시 모두 0이다. 별도 judge 프로세스 0은 이 Codex AI 협업의 비용이 0이라는 뜻이 아니며 비용은 측정하지 않았다. `training_ready`는 false다.
