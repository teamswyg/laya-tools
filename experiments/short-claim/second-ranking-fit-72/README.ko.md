# 두 번째 학습: 후보 정렬 목표를 추가했지만 효용 실패

작은 주장 모델의 목적은 사람이든 에이전트든 확인할 후보를 먼저 찾도록 힌트를 주는 것이다. 이번 실험은 기존 후보별 정답 학습(BCE)에 같은 요청의 정답 후보를 다른 후보보다 앞에 두는 목표를 더했다. 계수 8,192개·FP32·seed1729·학습률0.1·L2 0.0001·batch128·최대50epoch는 유지하고, 추가 목표의 비중은 미리 정한 λ=1 한 번만 실행했다. 첫 모델의 압축 자식이 아닌 별도 학습 방식이다.

| 같은 개발 validation에서 관측 | 가장 강한 lexical_ordered | 첫 BCE 모델 | 두 번째 BCE+정렬 모델 |
|---|---:|---:|---:|
| 전체 후보 확인 횟수: 적을수록 좋음 | 27 | 31 | 33 |
| Top1 정답: 정답 있는 요청10개 | 5 | 2 | 1 |
| Top3 정답: 같은10개 | 10 | 10 | 10 |

기본 방식 대비 확인 횟수 감소율은 **−22.2222%**다. 확인 횟수와 Top1 조건이 실패했으므로 기본 설정에 적용하지 않는다. `PublicationQualified=false`, `ProductionReady=false`를 그대로 유지한다. 재학습·seed/λ 탐색·기준값 완화는 하지 않았다.

선택 epoch50에서 train BCE는0.6311540918774127, validation BCE는0.6885500655576609다. 요청별 후보 쌍 평균 손실은 train0.593613195205774, validation0.6927276831485621이다. 학습 손실 감소가 실제 확인 작업 절감으로 이어지지 않았다. 기록된50epoch에서 기존의 가장 이른 최소 validation BCE 선택 규칙을 유지했다.

원래76요청의 저장 projection·역할·정답·mask를 재사용했다. Fit은 알려진 후보91/45행이며, 그 안의0가중치18/1행도 보존한다. 후보 쌍은 train56/validation28개이고 그중0가중치16/2개를 남긴다. 실제 정렬 손실을 공급하는 부모는16/10개다. Utility는 알려진 validation15요청(정답 있음10·no-answer5)의 모든 후보에 적용하며 unknown5요청은 제외한다. Calibration은 fit에 들어가지 않는다. 이번 검증 자료는 첫 실패에 따라 다음 실험을 설계하는 데도 사용했으므로 독립 최종 시험이 아니다. 별도 도메인별2,400개 새 최종 요청과 원천·작성 다양성은 아직 확보하지 못했다.

실제 corpus fit은 누적 두 번째이며 이번 실행은1회·재시도0회다. Go1.27.1 CPU1·Go heap soft target256MiB에서 외부 관측 peak RSS28,999,680B, footprint26,444,304B, controller wall2.969253125초다. 이 시간은 학습·검증·저장을 포함하며 같은 Mac의 전체 Go 검사도 진행 중이었다. 순수 추론 지연이나 이전 모델 대비 속도 비교로 사용하지 않는다. GPU·Laya encoder·유료 LLM 호출을 사용하지 않았다.

새 계수 파일은32,792B, SHA `539bd0de1c4b08af99645ebc113eeaa7a7aeaef8dcf4282d10e3b3336e737c6f`다. 참조 decoder는 계수를65,536B float64로 펼치므로 파일 크기와 전체 메모리는 다르다. 모델 본체는 Git에 포함하지 않는다. 원본 결과 SHA는 `e01d7daf8a5f26cd5578674e87f996a578d747f6622221308da4624a71995220`다.

독립 수치 검토는 저장된 계수·projection만으로 점수·가중 BCE·선택 epoch의 후보 쌍 손실·후보 순서·확인 횟수를 재계산해 일치를 확인했다. 원본 trainer·feature·baseline·role API나 실제 학습은 다시 실행하지 않았다. 중간 epoch의 계수를 복원하거나 재현성·일반화까지 검증한 것은 아니다.

[원본 결과](results.json) · [외부 실행 기록](ROOT-ACTUAL-LEDGER.v1.json) · [독립 수치 검토](RECEIPT-SECOND-FIT-NUMERIC.v1.json) · [소스 CI](SOURCE-CI.v1.json) · [English](README.en.md)
