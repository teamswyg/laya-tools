# Native2 저장 관찰의 원문 대조

구체적인 blocker는 없다. 저장된 23행은 고정 순서·Want 및 후보별 IPNet/IP/IPMask 상태를 보존하며 원문과 일치한다. 유한 입력 만족은 IPNet 5/5, IP 0/5, IPMask 0/5, OrCompose 4/4, Compose 0/4다. 합계 9 만족·14 알려진 불일치·unknown 0·panic 0이다.

IP와 IPMask는 마지막 bare-IP 입력에서 성공한다. 그래도 각자의 타입·값은 CIDR 네트워크 계약과 다르므로 불일치가 맞다. OrCompose는 nil/nil 성공에도 즉시 멈추며 실패 메시지를 LF로 합친다. 빈 hook은 nonnil 빈 오류다. Compose는 E1에서 멈추고, nil 성공 뒤에는 두 번째 hook에 invalid reflect source/null 값을 넘기며, 빈 hook은 원래 4,nil을 반환한다. 이는 source fidelity와 요청 만족이 서로 다른 판단임을 보여 준다.

Flag.Changed=false 15개는 direct Value.Set의 진단이다. FlagSet.Set 호출이나 Changed 갱신을 주장하지 않는다. 오류 type/identity도 Want에 없던 기준으로 추가하지 않았다. 명시 upstream entrypoint 76회와 자체 callback 9회는 초기화·내부·stdlib 전체 호출 수가 아니다.

저장 바깥 기록은 native Start 1·Wait 1·exit 0·재시도 0, child RSS 9,551,872 B, Go heap snapshot 3,053,392 B, startup/pin 포함 wall 0.37042475 s를 보인다. heap snapshot과 lifetime peak RSS는 다르며 RSS는 관찰값이다. Go native 유한 관찰이며 Laya/GPU나 모델 효과 측정이 아니다.

검토자는 v3 비교식 작성자이고 관찰기·Want·실행 저자는 아니다. 비맹검 AI 원문 판독과 저장 결과 검토이며 독립 native 재현은 아니다. 원 실행/jq/Go/모델 재실행 0, 원문·Want·봉인 변경 0이다. truth/role/weight는 null을 유지하며 권리·학습 자격·qualified count를 부여하지 않는다. [영수증](RECEIPT.v1.json)과 [장부](LEDGER.v1.json)에 범위와 핀이 있다.
