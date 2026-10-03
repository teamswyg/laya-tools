# 다음 사례 실행을 위한 작은 Go 기록기

이 묶음은 riidolaya의 주장 모델 실험을 준비합니다. 작은 모델이 후보를 제안하더라도, 원전 코드가 실제로 무엇을 했는지 남기고 검증해야 학습 근거로 쓸 수 있습니다. 미확인 값은 미확인으로 유지합니다. 이 기록기를 통과했다는 이유만으로 정답이나 학습 허가가 생기지는 않습니다.

지금 검증된 것은 직접 만든 합성 입력과 실패 처리입니다. 실제 추가 원전 worker·새 모델·학습은 아직 실행하지 않았습니다. 기존 개발 요청 35개와 압축 기록 35개는 서로 다른 수치입니다. 보호된 2,400개/domain 평가와 전체 검증 작업량 5% 개선도 남아 있습니다.

| 준비물 | 지금 확인한 것 | 남은 것 |
|---|---|---|
| frame | 39개 순서, 고정 배열과 중첩 span, 호출 전 ACK, 불확실한 호출 재실행 차단. 합성 테스트 14개 | 원전 ABI와 사례별 호출 연결 |
| durable | 독립 순서·counter 검사, 저장→파일 Sync→부모 Sync→ACK, durable/ACK prefix 분리. 합성 테스트 6개 | 실제 파일 예약·Sync, 프로세스 종료, 결과 파일 대응 |
| artifact-seal | 파일 내용·inode·크기·mode·mtime 변경과 링크 거절. 합성 테스트 6개, Darwin/Linux vet | compiler/build identity, 라이선스와 실행 admission |

Go 1.27.1이 설치된 Linux 또는 macOS에서 저장소 루트의 아래 명령으로 CI와 같은 합성 검사를 실행할 수 있습니다. 원전 worker나 모델을 실행하거나 다운로드하지 않습니다. 공개 module 템플릿을 전용 임시 폴더에서 연결하고 끝나면 해당 폴더만 지웁니다.

```sh
bash scripts/verify-next60-union-preparation.sh
```

정상 반환/패닉이 모순되거나 null인 AFTER 기록은 거절합니다. nil-panic 호환 설정에서도 쓰기·ACK 중 패닉은 반드시 중단합니다. 최초 소스의 두 문제와 수정 전 테스트 통과 기록은 history에 보존했습니다. 독립 검토 후 수정본을 별도로 검사했습니다.

한 줄 한도는 LF 포함 2,048B, 전체 frame 한도는 640입니다. 고정 DTO의 모든 필드 폭을 과대평가한 인코딩도 1,505B입니다. 이 수치는 통신 크기의 상한이며 프로세스 전체 메모리나 추론 속도 측정이 아닙니다. 단일 소유 loop와 고정 배열을 써서 map과 lock이 필요하지 않습니다. span은 함께 읽는 작은 구조를 연속 배열로 보관하며, SIMD/SoA 성능 개선은 별도 측정이 필요합니다.

중복 기록 네 개는 실제 디스크 복원·SHA·mode 검증 후 압축했고 독립 saved-only 검사도 통과했습니다. 이번 보수적 회수는 666,800B, 누적은 11,902,227B입니다. 공개 canonical 기록은 유지합니다. compiler-only harness와 일회성 실행 파일은 설계안이며 기존 바이너리의 보관 범위나 자원 한도를 바꾸는 허가가 아닙니다.

새로운 원전 worker는 fixture/Want 봉인, 사례별 호출 범위, source/권리/compiler 확인, 실제 실행 파일 핀, 종료·EOF·결과 대응, 새 자원 admission이 함께 닫힌 뒤에 실행합니다. native/GPU/모델 추론과 학습은 각각 실제 근거로 보고합니다. 기존 Laya CI 추론 PASS는 별도이며 GPU 실행 확인은 없습니다.

기록: [PR128](https://github.com/teamswyg/laya-tools/pull/128), [Issue19](https://github.com/teamswyg/laya-tools/issues/19#issuecomment-5970088870). 이 묶음은 직접 작성한 Apache-2.0 소스이며 upstream 코드·weights·개인 입력을 포함하지 않습니다.
