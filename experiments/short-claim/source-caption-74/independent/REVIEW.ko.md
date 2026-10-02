# Caption 74: 독립 원문 검토

9개 새 caption은 API 경계를 구분하는 데 도움이 됩니다. 원문에서 8개 문구의 내용을 직접 확인했고, Values 문구는 exported/nonignored/default primitive field 및 이름 있는 중첩 범위에서 지지됩니다. Wrapper·Next를 새로 호출하지 않았다는 이유만으로 명확한 반환형·위임·오류 처리를 판정 불가로 만들지 않았습니다.

원래 요청 두 개는 아직 의미가 모호합니다. datasize는 syntax와 bits가 receiver 0, range가 최대값이므로 “distinct states”가 세 종류 모두 다른 값을 뜻하면 충돌합니다. query의 기본 반복 값은 순서·중복을 보존하지만 빈 slice는 생략되고, delimiter/numbered 옵션·custom encoder·anonymous 중첩은 다른 정책입니다. 영수증은 정확한 두 상태 및 기본 필드 범위를 명시하는 별도 요청 버전을 제안합니다. 원래 문장과 null 정답을 바꾸지 않았습니다. 나중의 A/B도 동일한 명확한 요청을 양쪽에 써야 하며, 버전은 새로운 독립 과제가 아닙니다.

직접 선택한 API 모양 기준으로 Parse/MustParse는 caller receiver 메서드가 아니고 MustParse는 오류를 panic으로 바꿉니다. valueString은 한 문자열을 반환합니다. 두 Next는 완성 목록을 반환하지 않습니다. 반면 Split은 default lexer의 성공한 단어만 누적해 뒤의 quote/escape 오류와 함께 반환합니다. 이는 source-only 지원/충돌 사실입니다. Encoder interface의 실제 callback 동작과 임의 caller adapter는 unknown이며, 자료 부재를 음성/no_answer로 바꾸지 않습니다.

stdlib 바이트 검사 1회가 3개 원본 SHA, 26개 evidence span, 9개 root span, 9개 원래 caption span, 새 caption 해시·단독 단어/바이트 한도, 기존 요청·번역·후보 순서·binding·라이선스와 null을 대조했습니다. runtime Normalize/Validate는 호출하지 않아 실제 정규화 단어 한도 관측은 아직 아닙니다. 원 API·Features·Project·모델·학습·Got 본문 재독은 0입니다.

검토자는 caption 작성자와 다르지만, 첫 3원천 준비 및 이전 저장 관찰 결과 QA를 했습니다. 따라서 blind 검증이나 새 독립 원천 작성 흐름이라고 주장하지 않습니다. API 원문·Want 노출을 보존하며 현재 검토는 source fidelity/coverage 및 직접 선택 API의 제한된 만족 판단입니다. 원래 truth/허용집합/역할/가중치는 모두 null 그대로이고 실제 승격·새 부모는 0입니다. 같은 가족의 두 caption·별칭·번역을 다른 역할로 쪼개지 않습니다. 영어 target-aware 요약이며 한국어 문서는 언어 시험이 아닙니다. 자세한 사실과 제한은 [영수증](RECEIPT.v1.json)에 있습니다.
