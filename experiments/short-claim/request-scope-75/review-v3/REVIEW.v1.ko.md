# 범위가 보이는 요청 75 독립 검토

v3 요청·소스 만족 제안·캡션별 BCE 마스크의 범위는 읽은 근거와 맞는다. **pair 학습을 설명하는 한 문구는 사실 정정이 필요하다.** 기존 v1/v2/v3나 실제 학습 결과를 수정하지 않고 별도 correction version으로 남겨야 한다.

검토자는 request75/mask 작성자가 아니지만 caption74 작성자다. 따라서 요청과 supervision 제안의 저자 독립성만 있으며, B 캡션 자체의 작성 독립성이나 전체 맹검은 없다. AI 보조 source/text 검토다. 원본 API·Normalize·Features·역할·학습은 실행하지 않았다.

datasize 요청은 nonnil receiver, ASCII 정수·이진 바이트 단위와 syntax/bits 오류의 receiver0, range 오류의 uint64 최대값을 직접 드러낸다. 원문의 UnmarshalText:203–216이 이를 뒷받침한다. Parse는 새 지역 값을 쓰는 함수, MustParse는 오류 시 panic하므로 요청의 receiver/반환 오류 조건과 다르다. 예전 'distinct states'의 세 상태라는 모호함을 소급 고치지 않고 새 요청 범위로 분리했다.

query 요청의32단어에는 exported/nonignored/default primitive, custom 제외, 그리고 **included named nonnil non-time** 조건이 보인다. 원문은 omitempty/isEmptyValue 검사(:185–187, :321–343)를 재귀(:264–268)보다 먼저 한다. 'included'는 그 생략 검사를 통과한 필드만 가리킨다. anonymous embedding의 flattening(:168–175), nil pointer, time 특별 처리(:259–262), custom callback(:194–205)은 중첩 대괄호 주장의 범위 밖이다. default primitive 표현에서는 slice 인덱스 순서 Add(:248–254), 빈 slice 생략(:216–220), omitempty 없는 빈 string Add(:271,:316)가 근거다. 이 해석으로 조건부 코드 만족 제안은 타당하다. 다만 Values 캡션을 단독으로 읽었을 때의 더 넓은 문장을 무조건 충실한 계약으로 승인하지 않는다. request와 함께 읽는 범위이고 기존 caption74 fidelity/coverage 상태는 그대로다.

shlex 요청은 그대로이며 Split:403–415의 누적 목록·오류 접두 반환과 Lexer/Tokenizer의 단일 반환은 구별된다. 성공한 빈 WordToken도 append한다. Next가 같은 helper를 사용한다는 사실은 원문으로 읽을 수 있지만, 직접 선택된 Next 자체를 목록 반환 API로 바꾸지 않는다.

| 학습 정보 충분성 제안 | A | B | 공통 |
| --- | ---: | ---: | ---: |
| 긍정 | 0 | 3 | 0 |
| 부정 | 4 | 6 | 4 |
| 비제로 긍정×부정 쌍 | 0 | 6 | 0 |

A의 충분한 부정은 Parse·MustParse·Encoder·Lexer.Next다. 요청과 직접 충돌하는 API 형태/정책이 보인다. A의 다른 문장은 필요한 동작을 생략하거나 부정 근거가 불충분하므로 보수적으로 마스크한다. 생략을 코드의 불만족이나 unknown의 자동 negative로 바꾸지 않았다. B의3P6N은 좁아진 요청과 조건부 source reading 안에서 가능한 supervision **제안**이며 실제 라벨/가중치가 아니다. 서로 다른 arm 마스크를 쓰면 캡션과 학습 범위가 함께 바뀌므로 캡션 효과만의 비교라고 할 수 없다. 공통 긍정0으로 이 묶음만의 ready paired 학습도 주장할 수 없다.

정정 근거는 실제 두 번째 학습에 묶인 d506 source다. ranking.go:267은 두 SampleWeights를 곱한다. :347–348은 pair weight0의 gradient, :378–379는 NLL을 건너뛴다. saved driver fit.go:118은 projection의 같은 train/validation을 넘기고, main.go:373은 그대로 FitWithRankingTrace에 전달한다. fit.go:205–213은 LossEligible에 따른1/0 가중치가 그대로인지 확인한다. 별도로 pair 가중치를 unmask한 경로는 없다. 따라서 **BCE sampleweight0이면 그 endpoint를 포함한 쌍의 학습 기여도0**다. audit에 행·쌍을 물리적으로 보존하는 것과 loss/gradient 기여는 다르다. 기존 코드를 추가 수정해야만 차단된다는 설명은 부정확하다. 향후 BCE와 다른 pair 정책을 원할 때에만 그 정책을 별도 설계하는 것이 맞다. 기존72 숫자·가중치·결과는 그대로다.

stdlib metadata 검사1회가 성공했다. 세 버전의 요청9개, 같은9후보·두 캡션, 기존 evidence26/3source/라이선스 참조의 보존, 모든 actual truth/역할/가중치/pair 필드의 null을 대조했다. 단어 수는 v1=27/30/21, v2=27/32/21, v3=27/32/21이며 각각512B/32word 이하다. 실제 normalizer 검증은 하지 않았다. 코드·metadata·관측값·mask는 model feature가 아니며 request+선택 caption만 입력이라는 경계가 유지된다. 이 준비는 코퍼스 편입·원천 다양성·학습 준비·2400/domain final 승인이 아니다.
