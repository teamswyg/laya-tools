# 작은 주장 모델에 사용할 실제 동작 확인

우리가 원하는 모델은 비슷한 후보 중 어떤 것이 요구에 맞을지 싼 힌트를 주는 모델입니다. 따라서 먼저 실제 동작을 확인한 자료가 필요합니다. 이번에는 기존 20개 초안에 있던 요청 2개에 입력 9개를 고정하고 원본 Go 후보 5개의 동작을 한 번 실행해 23개 관측을 얻었습니다. 새로운 소스 가족은 추가하지 않았습니다.

CIDR 네트워크 파싱은 IPNet 후보가5개 입력을 모두 만족했습니다. 단순 IP·마스크 후보는 각각0/5였습니다. 대안 함수를 조합하는 요청은 OrCompose가4/4, Compose가0/4였습니다. 서로 이름이 비슷해도 오류에서 멈추는 방식과 `nil` 처리 방식이 달랐습니다. 오류 메시지가 빈 문자열이어도 실제 오류가 있는 경우를 함께 확인했습니다.

패닉과 확인 불가는0이었습니다. 원본 실행 자식의 OS 최대 메모리는약9.11MiB, 시작·입력 검증을 포함한 시간은약0.37초였습니다. 이는 원본 코드 관측이며 Laya 추론이나 모델의 정확도·Codex 비용 절감을 측정한 결과가 아닙니다. 실행 도구의 크기·Go 힙·OS 메모리도 서로 다른 수치입니다.

이23개를 고유 과제23개로 늘려 세지 않습니다. 관측 당시 학습 적격은 0개였습니다. 이후 학습/배포 권리·고유 요청 중복·관련 그룹·역할을 따로 확인한 [개발용 요청 2개 편입](Native2-Training-KO)을 보세요. 다음 모델이 단순 정렬보다 실제로 적은 확인으로 같은 정답을 찾는지도 별도로 검증합니다. 효용을 실패한 기존 모델은 비활성입니다.

[상세 결과와 측정](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-native2-actual-observations) · [원문 대조](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-native2-actual-semantic-review) · [English](https://github.com/teamswyg/laya-tools/wiki/Native2-Observation-EN)
