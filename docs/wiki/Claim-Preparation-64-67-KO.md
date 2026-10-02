# 작은 힌트 모델의 학습 준비64–67

목표는 LLM 작업에 앞서 후보를 확인할 순서에 작은 힌트를 주는 모델이다. 예를 들어 “이 코드는 중첩 디렉터리까지 찾을 가능성이 있다”라는 점수가 확인 순서를 도울 수 있다. 점수가 틀려도 원래 검사와 fallback을 유지해야 한다. 이 단계에서 새 모델의 추론 속도나 Codex 비용 절감은 아직 입증하지 않았다.

현재는 **학습에 넣을 근거가 있는 자료를 골라내고, 서로 비슷한 코드가 학습·검증 양쪽으로 새지 않도록 준비하는 단계**다. 요청72개와 후보216개를 새 표본으로 부풀리지 않고 그대로 보존했다.

| 준비 | 하는 일 | 현재 상태 |
|---|---|---|
| Go 배열 투영64 | 정답·제외 mask를 보존하며 학습 자료 배열을 만드는 내부 유지보수 모듈 | 합성·포트 검증; 실제 데이터 투영 전 |
| 전체 그룹 연결65 | 같은 요청·후보·helper·타입을 공유하는 관계를 묶기 | 원본17그룹의 metadata 결속 검증 |
| 역할 실행 도구66 | 전체 그룹 단위로 train/validation/calibration을 배정하기 | seed와 실행 계획 고정; 실제 원본 배정 전 |
| 감독 범위 검사67 | 저장된 정답과 설명 검토를 대조해 학습 loss에 쓸 후보 제안하기 | 첫 실제 metadata 검사1회 성공 |

67에서는 양성35개·음성95개에 사용할 근거가 남았다. known 후보23개는 라벨을 바꾸지 않고 학습 가중치를0으로 두어야 한다. unknown 후보63개는 라벨이 null이다. 합계는216개다. 이런 후보를 삭제해서 표본이나 평가 분모를 바꾸지 않는다.

원래 저장 정답을 가진 그룹은16개이며, mask를 적용한 뒤 적격 후보가 하나 이상인 그룹은15개다. 그룹52의 known 후보9개는 모두 제외 대상이다. 역할별로 실제 몇 그룹이 남는지는 아직 확인하지 않았다. 기존9/3/3 조건이 부족하면 그 사실을 기록하며, 유리한 seed를 찾거나 mask를 느슨하게 바꾸지 않는다.

67의 첫 metadata 검사 프로세스는0.71초, 최대 RSS 약26.2MiB였다. 이는8.59MB의 JSON 자료를 검증한 한 번의 검사다. 모델 추론·GPU·Go heap·상주 라우터 속도 측정과 구분한다. 공개 Go 포트는 합성 검사와 빌드를 통과했으며, 이 실제 수치는 이전 private 원본 바이너리의 관측이다.

## 직접 사용하기

저장소 루트에서 Go1.27.1로 실행한다. 모델 다운로드나 Python은 필요하지 않다.

```sh
go run ./cmd/riido-supervision --input-root . \
  --output supervision-new.json \
  --plan-sha256 c9a07943f8047b69e32d93354e812ca5b158b1dc7476264bd5b8505a86242408
```

출력은 새 파일이어야 한다. 사람은 안내와 실패 코드를 읽고, 에이전트는 종료 코드와 JSON을 읽을 수 있다. 성공은 `passed_metadata_proposed_supervision_only`이며 학습 준비 완료를 뜻하지 않는다. 파일당2MiB·합계16MiB의 읽기 한도는 RSS 상한이 아니다.

다음은 source/input/계획/실제 binary를 고정하고 자동 CI가 병합한 뒤 역할을 한 번 실행하는 것이다. 새로운 작성 원천의 설명도 실제 후보로 연결해 제한된 의미 범위를 검토해야 한다. 그 뒤 같은 범위에서 기존 기준보다 확인 횟수나 전체 비용이 줄어드는지 시험한다. 개발 실험과 별도로 충분한 고유 요청을 가진 보호 최종 평가를 준비한다.

[67 사용법·실제 기록](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/supervision-scope-67/README.ko.md) · [독립 참조 대조](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/supervision-scope-67/QA-67.v1.ko.md) · [66 실행 조건](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/role-execution-preparation/README.ko.md) · [개발 이슈](https://github.com/teamswyg/laya-tools/issues/19)
