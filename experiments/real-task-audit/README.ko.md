# 실제 GitHub 작업 원천 점검 — 실험28

**명목2,594개 행을 그대로 세지 않고 확인한 결과, 정규화된 고유 요청문은2,575개였다.** 출처는53개 저장소다. 숫자로는2,400을 넘지만 아직 독립 최종 평가나 모델 성능 검증을 완료한 것은 아니다.

## 무엇을 확보했나

[사전 계획](plan-28.json)에 따라 [SWE-bench Full 고정 버전](https://huggingface.co/datasets/SWE-bench/SWE-bench/tree/c6fe717fd7a4c3ac1daa4055a4fd082c6a1d28a2)과 [Multilingual 고정 버전](https://huggingface.co/datasets/SWE-bench/SWE-bench_Multilingual/tree/846e647b9f33c0b51b739d005d13d85493c9af09)의 test Parquet 및 README만 로컬로 받았다. 개발용 split·컨테이너·저장소 본체는 받지 않았다. 두 파일씩 원격 체크섬 검증을 통과했다. 선택 다운로드이므로 전체 HF 저장소 복제 검증은 아니다.

| 자료 | 행 / 고유 식별자 | 고유 요청문 | 저장소 | 저장소+수정 전 커밋 그룹 |
|---|---:|---:|---:|---:|
| Full | 2,294 | 2,275 | 12 | 2,174 |
| Multilingual | 300 | 300 | 41 | 300 |
| 합계 | 2,594 | 2,575 | 53 | 2,474 |

고유 요청문은 소문자화·연속 공백 정리 후 같은 문자열을 묶은 값이다. Full에서19개의 추가 행이 같은 정규화 요청문을 반복한다는 뜻이며, 중복 그룹 수가19라는 뜻은 아니다. 두 자료 사이에는 동일 식별자나 정규화 요청문이 없었다. 빈 식별자·요청문, 잘못된 저장소 형식·커밋 형식도 없었다.

이전 CodeSearchNet 전체14,291행의 고유 정규화 요청문11,678개와 비교해 일치 요청문은0개, 동일 저장소도0개였다. 이는 정확한 문자열/저장소 겹침 검사이며 의미가 비슷한 질문·코드 복제·모델 사전학습 오염까지 없다는 증명은 아니다.

## 숫자만으로 승인하지 않는 이유

Django가850행을 차지하고 Full2,294행은 Python 프로젝트 자료다. 전체 수가 많아도 영역별 균형은 다르다. 같은 수정 전 스냅샷을 공유하는 작업도 있다. 다음 분할에서는 요청문 중복과 스냅샷 연결을 함께 묶고, 저장소별 결과와 집계 방법을 정해야 한다. 관련된 행을 학습과 시험 양쪽으로 나누면 성능이 과장될 수 있다.

이 자료는 실제 이슈 해결 과제다. 기존 함수 설명 검색보다 작업 맥락이 풍부하지만 **패치에 등장하는 파일이 모든 관련 파일의 정답은 아니며, 패치 크기가 모델 난이도나 작업 분할의 정답도 아니다.** 공개 벤치마크이므로 큰 모델이 이미 보았을 가능성도 배제하지 않는다. 우리 프로젝트에서 새로 확보했다는 사실과 모델에 완전히 새로운 문제라는 주장은 구분한다.

## 정답 정보 분리와 구현

관리용 변환기는 SHA-256·행 수·열 자료형을 확인하고128행씩 네 열만 읽는다: `instance_id`, `repo`, `base_commit`, `problem_statement`. 패치·테스트 패치·힌트·난이도·테스트 목록·실행 스크립트·이미지 정보는 요청 projection에 넣지 않는다. 원본 Parquet에는 이런 열이 있지만 실행하거나 모델에 전달하지 않았다.

Parquet 디코딩에 기존 PyArrow25.0.1을 사용하는 작은 관리용 스크립트가 있다. 집계·입력 검증·중복·이전 자료 겹침·구성 해시는 Go의 정렬 배열로 계산하며 런타임 Python 의존성이나 새 Go 의존성을 추가하지 않았다. 원문 요청과 개별 해시는 출력하지 않고 집계와 전체 구성 해시만 공개한다.

## 라이선스와 원천 불일치

[벤치마크 코드의 고정 LICENSE](https://github.com/SWE-bench/SWE-bench/blob/02e7a74ffd0b707aab73d203fe87bdc7c76afc8e/LICENSE)는 MIT다. Multilingual 카드에도 MIT 표시가 있지만 Full의 해당 카드 메타데이터에는 license 필드가 없다. 이것만으로 모든 원천 저장소 코드·이슈가 같은 조건으로 재배포 가능하다고 판단하지 않는다. **개별 원천 조건 검토는 미완료**이며 원문 재배포나 새 학습 모델 공개의 근거로 사용하지 않았다. Git에는 자체 분석 코드와 집계·해시·설명만 올린다.

[공식 FAQ](https://www.swebench.com/SWE-bench/faq/)는 Multilingual42저장소로 안내하지만 이 고정 파일은41개다. 또 Full 카드 본문은 테스트 목록을 JSON 문자열로 설명하지만 실제 두 Parquet 파일은 모두 `list<string>`이다. 변환기는 실제 스키마를 검사한다. [원천 메모와 해시](source-notes-28.json)에 차이를 기록했다.

## 검증·자원·다음 단계

전체 보고서는 재실행에서 바이트 일치했다. 합성 테스트로 문자열 정규화·중복·출처 간 겹침·잘못된 필드·추가 정답 열 거부·원문 비공개·행 순서와 무관한 구성 해시를 검사했다. 변환기 테스트는 답안/실행 열 제외 및 틀린 원천 해시 거부를 확인했다. 전체 Go race/vet·형식 검사를 통과했다.

다운로드한 Parquet 둘은35,169,743바이트, 로컬 요청 projection 둘은5,433,482바이트다. 기존 자료를 포함한 Go 집계1회는 M4 Pro CPU wall0.64초,user0.28초,peak RSS75,776,000바이트(약72.3MiB)였다. 다운로드·Python 변환·학습·추론·GPU 비용은 포함하지 않는다.

```sh
# 고정 Parquet 다운로드 후, 관리용 변환을 각각 한 번 수행한다.
python scripts/training/project_real_task_audit.py --source full --input .cache/real-task-full-28/data/test-00000-of-00001.parquet --output .cache/real-task-full-28.jsonl
python scripts/training/project_real_task_audit.py --source multilingual --input .cache/real-task-multilingual-28/data/test-00000-of-00001.parquet --output .cache/real-task-multilingual-28.jsonl
go run ./cmd/riido-taskaudit --out .cache/real-task-audit-new
```

[전체 집계](results-28.json). 다음은 연결 그룹·원천 조건·수정 전 스냅샷 접근 비용을 확인하고 평가 전용 목록과 입력 계약을 별도로 고정하는 단계다. 이번에는 학습·모델 추론·에이전트 실행을 하지 않았다. 기존 최종 reserve는 사용하지 않았고 새 모델/HF 배포도 없다. **2,575개 후보 요청 확보와 실제 독립 평가 완료를 혼동하지 않는다.**
