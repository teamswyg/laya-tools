# 읽기 전용 주장 근거 shadow

로컬Go **미학습 연구 필터**입니다. 답변·활동·종료 문구 후보와 입력으로 제공한
유형별 이벤트 사실을 독립적으로 기록합니다. 모델·Fit·확률·앱/DB클라이언트·태스크
상태 권한은 없습니다. 의미 품질,정밀도90%나Bloom필터 오탐 한도를 입증하지
않습니다. 기존 분류기 confidence.9/margin.05기준을 바꾸지 않으며 이 별도 근거
계약이 그 기준을 통과하거나 대체하지 않습니다.

로컬JSON바이트SHA를 명시하며 최대1MiB/100관측,문장은각4,096UTF8바이트입니다.
일반 schema예시는 English README의 직접 작성한 공개 가상 자료입니다.
`riido-assertion-shadow-input-v1`의 observations는 다음을 구분합니다.

- `text_observation`:role은`comment`또는`body_snapshot`,문장과 actor/unit/timestamp/version/source.
- `observed_event`:kind와 actor/unit/timestamp/version/source. 문장을 넣지 않습니다.

태스크 본문의 여러 문단/조각도 본문 관측이며 댓글이나 독립 표본으로 세지 않습니다.
actor/unit/time/version이 없으면 불확실성으로 기록합니다. 본문만 있는 입력에서는
댓글 입력이 없음을 보고하며 현재 작업 상태의 진실을 만들어내지 않습니다.

이벤트 종류는 CommentObserved,CommentCreated,ProgressObserved,CommandSucceeded,
TurnEnded입니다. 호출자가 제공한 관측 진술만 기록하며 인증·멤버십·운영 이벤트나
DBcommit을 확인하지 않습니다. 기존 댓글을 관측한 것과 만든 것은 다릅니다.
어느 이벤트도 태스크 완료를 뜻하지 않습니다. 승인된 native owner fixture의 읽기
검토는 관측 순서·제외·무변경 경계만 뒷받침합니다. 공개 테스트는 별도로 작성한
일반 자료이며 합성 replay는 운영 이벤트나 사람의 의미 정답이 아닙니다.

필터는 원본UTF8 **바이트** 오프셋과 그대로의 부분문장을 유지합니다.
HTML태그/속성이나URL물음표를 요청으로 다루지 않습니다. 인용/코드·미래/조건·
부정·필수 잔여 작업·문맥 미해결에는 명시적 불확실성을 붙입니다. 좁은 직접
답변/활동 패턴은 실험 표시 후보가 될 수 있지만 미학습 의견이며 신뢰할 만한
의미 라벨이 아닙니다. 한영 표현의 적용 범위가 제한되고 놓치는 표현도 있습니다.
HTMLentity도 바꾸지 않아 원본 위치를 보존합니다. 확신도를 꾸미지 않으며
**완료 표시는 언제나 보류**합니다. 명확해 보이는 종료 문구도 예외가 아니고 목표
완료를 주장하지 않습니다.

```sh
go run ./cmd/riido-statehint-assertion-shadow \
  --input OBSERVATIONS.json --input-sha256 INPUT_SHA --check

go run ./cmd/riido-statehint-assertion-shadow \
  --input OBSERVATIONS.json --input-sha256 INPUT_SHA \
  --out .cache/statehint-assertion-shadow/NEW-RUN
```

check는 결과를 쓰지 않습니다. 새 폴더0700,report.json0600입니다. 입력SHA,
출처/actor/unit/time/version,원문 위치,불확실성,입력 이벤트 사실,본문/댓글 수와
입력 공백을 보존합니다. 처리시간·Go힙을 기록하되 OS최대RSS와 다르므로 전체
명령 RSS는 외부 프로세스 wrapper로 별도 측정합니다. 비공개 입력/근거 보고서는
무시되는 로컬 공간에만 두고Git/HF에 넣지 않습니다.

직접 만든 테스트로 생성/관측≠완료,Unicode/HTML바이트 위치,인용/코드,부정/미래/
잔여 작업,본문만 있는 공백,독립 후보,입력 상한,SHA/종류 및 비공개 독점 출력을
확인합니다. 개인 소스코드·fixture원문·태스크 식별자·실제prompt는 포함하지 않습니다.
Native댓글 pipeline을 켜거나 앱기본값·주석·반응·라벨·상태·DB·분류 모델을 바꾸지 않습니다.

실제 로컬 관측의 집계만 아래에 공개합니다. 승인된 비공개 태스크 본문 한 건의
54문단은 본문 관측54개이며, 댓글0개·입력 이벤트0개였습니다. 종료 문구 후보2개는
모두 보류됐고 표시 후보0개, confidence는 null이었습니다. 본문 원문·식별자·주소·
근거 보고서는 공개하지 않습니다. 이는 입력 공백과 읽기 전용 동작의 확인이며
의미 품질의 평가나54개의 독립 표본이 아닙니다.

별도로 직접 작성한 가상 이벤트 replay는 댓글 문장3개·유형별 이벤트5개에서
답변/활동 후보2개를 기록했습니다. 완료 표시는 계속 보류됐고 이벤트에는 완료
권한이 없었습니다. 이 합성 자료는 인증·운영·native PostgreSQL fixture의 정답
근거가 아닙니다. 두 실행 모두 모델·Fit·앱 상태 쓰기는0회였습니다.

| 로컬 실행 | 내부 처리시간 | 전체 명령시간 | 외부 측정 최대RSS |
| --- | ---: | ---: | ---: |
| 비공개 본문 입력1건 | 6.416ms | 0.60s | 10,846,208B |
| 직접 작성한 합성 이벤트 replay | 0.358584ms | 타이머 해상도 미만 | 6,455,296B |

각각 한 번의 관측이며 성능이나 정확도 보장이 아닙니다. 합성 실행의 전체시간
표시값0은 실제0초를 뜻하지 않습니다. 진행·완료·질문 세 힌트의 목표는 아직
달성되지 않았습니다. 운영 댓글 pipeline은 기본 OFF이고 활성화나 배포를 하지
않았습니다.
