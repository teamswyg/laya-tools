# 기존 Go 모델의 로컬 추론 비용 측정

이미 보유한 SHA 고정 버전1 `.rsh` 모델에 원본 가상 문장 네 개를 반복 입력합니다.
학습·정확도 평가·다운로드·업무 변경을 하지 않습니다. 부모와 새 모델을 같은 명령으로
측정하기 위한 유지보수 도구이며, 이 문장은 학습·선택·골든셋에 넣지 않습니다.

```sh
go build -trimpath -o .cache/statehint-bench ./cmd/riido-statehint-bench
.cache/statehint-bench --model MODEL.rsh --model-sha256 SHA --iterations 500000
.cache/statehint-bench --model MODEL.rsh --model-sha256 SHA --iterations 500000 --profiles-dir NEW_PRIVATE_DIRECTORY
```

모델 해시·형식·학습 단계가 유효한지 먼저 확인합니다. 작업 공간을 재사용하며,
모델 검증 시간과 반복 추론 시간을 별도로 출력합니다. 반복 횟수는 4–1,000,000회입니다.
프로파일 경로는 새 폴더여야 하며 CPU·heap 파일은 로컬에만 보관합니다.

JSON은 평균 추론 시간과 루프 중 Go 할당량을 기록합니다. 개별 문장·예측값·정답·
호스트 경로를 출력하지 않습니다. 프로파일 실행에는 프로파일러의 할당과 비용이 포함됩니다.
프로파일 실행이 더 빠르게 관측돼도 코드 개선 효과로 해석하지 마세요. 시작·파일 읽기·
JSON 출력·종료를 포함한 전체 프로세스 비용은 별도의 OS 측정입니다.

이 측정은 네 짧은 문장, 한 실행·장치 조건의 개발 자료입니다. p95 지연, 실사용 처리량,
실제 업무 정확도, 전체 앱 RSS나 GPU 비용을 증명하지 않습니다. Go pprof는 Go CPU/heap
정보이며 GPU를 측정하지 않습니다. 원본 pprof나 모델 본체는 Git에 올리지 않습니다.

[기존 v0.2 모델의 개발 측정](COSTS.development.json)은 20,000회 비프로파일 실행과
20,000회·500,000회 프로파일 실행을 따로 보존합니다. 각 숫자는 단일 관측이며
프로파일의 유무에 따른 시간 차이를 성능 개선으로 해석하지 않습니다.

[English](README.en.md)
