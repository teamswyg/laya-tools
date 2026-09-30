# 성능과 문제 해결

[English](https://github.com/teamswyg/laya-tools/wiki/Performance-and-Troubleshooting-EN) · [홈](https://github.com/teamswyg/laya-tools/wiki)

## Laya를 켤지 판단하기

M4 Pro/24GiB 초기 측정이며 모든 기기에서의 보장값이 아닙니다.

| 실행 경로 | 관찰 결과 |
|---|---|
| 저장소 6개 예제의 키워드 검색만 | 최대 프로세스 RSS 약 11.5MiB |
| Laya를 추가한 저장소 preview | RSS 약 1.40GiB, 모델 시작 약 942ms |
| Laya 옵션을 켠 혼합 24요청 | 로딩 제외 중앙값 39.669ms, p95 54.040ms. 추론 생략 요청 포함 |
| 가상 저장소 1,000개의 순수 Go 후보 검색 | 요청당 약 0.17ms. 인덱스 생성·추론 제외 |
| 개발 검색 평가의 후보 8개 재정렬 | 약 1.6–1.8초 추가 |

쉬운 가상 저장소 평가에서 키워드 Top-1은 정답 있는 질문 18/18, Laya 원시 선택은 판단한 15건 중 8건 정답, 기준 통과 추천은 0건이었습니다. 추천 coverage가 0이면 정밀도가 높다는 증거가 아닙니다. 실제 Codex 절감도 아직 입증하지 않았습니다. 키워드부터 사용하고 실제 업무를 대표하는 정답 있는 질문으로 비교하세요.

## 자주 생기는 문제

| 증상 | 확인할 것 |
|---|---|
| `riidolaya: command not found` | `./riidolaya` 또는 `./bin/riidolaya`로 실행하거나 설치 폴더를 PATH에 추가합니다. |
| 예제 카탈로그·설정 파일 없음 | laya-tools clone 폴더에서 실행합니다. 바이너리 압축에는 예제가 없습니다. |
| 모델·런타임 없음 | `riidolaya doctor` 확인 후 `riidolaya setup`. 다운로드 무결성도 검사합니다. |
| 다운로드·체크섬 실패 | 네트워크·프록시 설정을 확인하고 setup을 다시 시도합니다. 체크섬·TLS 검증을 우회하지 않습니다. |
| 검색 결과 없음 | 정확한 식별자, root, 무시 규칙을 확인합니다. 없는 후보를 재정렬로 되살릴 수 없습니다. |
| 하위 폴더 검색 실패 | `rg`를 설치합니다. Git 루트와 일반 폴더의 파일 탐색 방식이 다릅니다. |
| 모델 라우터가 계속 strong 선택 | `abstained`·`reason`을 봅니다. 기본 확신도는 0.9이고 한국어 분류는 미검증입니다. 기준을 낮추면 좋아진다고 가정하지 않습니다. |
| `--laya`인데 repo 결과가 candidate | 미지원 문자·설명 또는 추론 실패 사유로 키워드 후보만 남을 수 있습니다. |
| repo 결과가 abstain | 모호함·여러 저장소·낮은 확신도·none·잘림 등이 원인입니다. 설명을 줄이고 작업 의도를 확인합니다. |
| 상주 서비스의 높은 메모리 | 프로세스마다 모델을 보유할 수 있습니다. 필요하면 재사용하고 끝나면 종료합니다. |
| Core ML 초기화 실패 | 시험한 동적 모델은 실패했습니다. 기본 CPU를 사용합니다. GPU·ANE 효과는 미검증입니다. |
| 에이전트 출력 해석 실패 | `--json`·JSONL을 사용하고 stderr를 분리합니다. 종료 성공을 추천 성공으로 해석하지 않습니다. |

## 내 컴퓨터에서 측정하기

```sh
riidolaya bench --iterations 30 --threads 4
riidolaya bench --cpu-profile cpu.pprof --heap-profile heap.pprof --ort-profile ort-trace
go tool pprof -top cpu.pprof
go tool pprof -top heap.pprof
/usr/bin/time -l riidolaya bench --iterations 30
```

마지막 명령은 macOS용입니다. Go pprof는 네이티브·GPU 메모리를 제외하므로 Go heap을 전체 RAM으로 보고하지 않습니다. 네이티브 호출은 이름이 불분명할 수 있고 ORT 프로파일링 자체도 오버헤드가 있습니다. 프로파일·비공개 카탈로그를 공개 이슈에 올리지 말고 버전·플랫폼·명령·민감정보를 제거한 재현 예제로 설명합니다.

[전체 측정](https://github.com/teamswyg/laya-tools/blob/main/docs/measurements.ko.md) · [저장소 평가](https://github.com/teamswyg/laya-tools/blob/main/docs/repository-routing-preview.ko.md) · [문제 제보](https://github.com/teamswyg/laya-tools/issues)
