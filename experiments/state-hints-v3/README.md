# 세 표시 모델의 실제 이어 학습 V3

[English](README.en.md) · [원본 계획](PLAN.original.json) · [재현용 경로 계획](REPRODUCE-PLAN.json) · [전체 결과](RESULTS.first.json)

목표는 개발 작업의 진행·완료 보고·질문을 작은 Go 모델로 제안하는 것입니다. 기존 v0.2 가중치에서 원본 합성400문장으로 실제 CE+AdamW 이어 학습을 했습니다. 모델 구조는 같은1,024특징·8개 의미·FP32이고 파일도32,960바이트입니다. Laya 파생/LoRA/증류 모델은 아닙니다. optimizer는 새로 시작했고 선택 모델은520번 갱신돼 누적step1,920→2,440입니다.

## 사전에 정한 선택

| 후보 | 학습률 | validation 정답/80 | 세 표시+none NLL |
|---|---:|---:|---:|
| 선택: v0.2 이어 학습 |0.005|68|0.42385|
| v0.2 이어 학습 |0.01|68|0.42690|
| v0.2 이어 학습 |0.02|69|0.43852|
| 처음부터 학습 |0.02|56|0.73647|

모든 후보는40epoch·batch32·seed1729·decay0.001입니다. validation NLL만으로 선택했고 더 높은 정답 수 때문에 기준을 바꾸지 않았습니다. 학습은8개 의미 CE, 선택 손실은 진행/완료보고/질문과 나머지5개의확률합인none입니다. 원래 예측확률과argmax를재정규화하지않습니다. 별도calibration80에서0.5–5.0 grid로보정한온도는1.0입니다. confidence0.9·margin0.05는유지했습니다. 파일·온도·조건을저장한뒤새test200을처음parse했습니다.

## 새200개에서의 실제 결과와 손실

| 같은 새test200 | 부모v0.2 | 선택v3 |
|---|---:|---:|
| 전체 정답 |152/200 (76%)|158/200 (79%)|
| 진행 보고 정답 |45/50|44/50|
| 완료 보고 정답 |40/50|46/50|
| 질문 정답 |24/50|36/50|
| 제안 중 정답 |57/57|90/94|
| 잘못된 완료 보고 제안 |0|4|
| 올바른표시coverage(정답대상150개) |38%|60%|

**정답과적용범위가늘었지만 완료오표시도늘었습니다.** 그래서 기본모델·정책을교체하지않고 연구후보로만 기록합니다. 더많이제안하는것과잘못된표시를줄이는것은다른목표입니다. 한국어82/100·영어76/100이며이모델의정답률입니다. 이는 실제제품정답이나운영완료판단근거가아닙니다. 규칙control은96/200·제안75/88정답·완료오표시13개였습니다. 규칙점수는확률이아닙니다.

전체8개confusion·확률·보류·진단을보고서에보존합니다. 상태변경·실제조회는0입니다. 저장된파일을다시읽어SHA를확인하고200문장전체확률이일치함을검증했습니다. 원본첫stdout/stderr와모델가중치는privatecache에보관하고Git에는올리지않았습니다. 선택모델SHA는`6fafd22cb5af6f3c9a5ab51ae0022e67e82d7d753702a770aab5559e1804869d`입니다.

## 데이터와 재현

학습400/validation80/calibration80/test200은두AI작성자가만든원본합성760개입니다. primary3각100학습·50test와다른5의미각20학습·10test이며언어를반씩구성했습니다. 실사용분포나통계적독립성·인간검수정답을보장하지않습니다. TestEN100의초기병렬draft는예측전에다른업무문맥으로바꿨고초기drafthash와상관관계메모를private에보존했습니다. 결과를보고재주석하거나삭제한것이아닙니다. 이제test결과가노출돼다음모델선택의새test로재사용하지않습니다.

Go1.27.1에서부모공개모델을명시적으로내려받고새출력경로로재현합니다. 아래재현plan은원본plan의private상대cache경로만공개자료파일명에맞췄습니다. 원본plan·lock·driver몸체hash는별도로보존합니다.

```sh
hf download JooYoon/riidolaya-statehint-go-v0.2 statehint.rsh --revision 81335eadd9f2753d3b86c932e9b00c636714e3b4 --local-dir ./models/statehint-v0.2
go run ./cmd/riido-statehint-tune --plan experiments/state-hints-v3/REPRODUCE-PLAN.json --parent ./models/statehint-v0.2/statehint.rsh --driver cmd/riido-statehint-tune/main.go --out ./new-v3-run
```

재현은노출된개발결과검증이지새blind평가가아닙니다. 플랫폼간동일한소수점/속도를보장하지않습니다. 원래macOS4fit·검증·보정·test·reload단일Go프로세스OSpeakRSS는15,974,400B(15.23MiB)였습니다. GPU·실제조회·장기운영비용이아니며Goheap과더하지않습니다.

게시소스CI와HF새가중치게시검증은아직진행전입니다. 가중치게시도운영승격을뜻하지않습니다. 실제privateTask조회는opaque업무revision·현재catalog·본문연결을별도로검증해야합니다. 모델의완료보고는문장의주장이고업무완료사실이아닙니다. 공개학습자료와코드의라이선스는Apache-2.0입니다.
