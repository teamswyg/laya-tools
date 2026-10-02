# 실패79 보관과 다음 데이터 확보

79개 자료의 추가 학습은 완료됐지만 효용 기준을 실패했습니다. 모델은 모의 후보 확인31회·첫 후보 정답2/10으로, 기존 어휘 기준의27회·5/10보다 나빴습니다. 기본 동작은 바꾸지 않습니다.

[PR105](https://github.com/teamswyg/laya-tools/pull/105)는 같은 head의 필수 CI 네 개를 통과해 자동 병합됐습니다. 첫 Ubuntu 실패는 Go 의존성 다운로드의 HTTP/2 오류였습니다. 코드·기준·수치를 바꾸지 않고 실패한 작업을 다시 실행했습니다. [최종 증명](ROOT-CI-FINAL.v1.json)과 원본 공식 응답에 head·merge·run·job·타임스탬프를 보존했습니다. 최종 API의 네 job은 모두 attempt2로 표시되며, 이전 성공 macOS/secrets의 원래 타임스탬프도 남아 있습니다. 이를 네 작업의 실제 재실행으로 해석하지 않습니다.

[Hugging Face 실패79](https://huggingface.co/JooYoon/riidolaya-shortclaim-data-effect-failed-79/tree/5bef215895b69d3f2ef4b82bb3f1970279d67f46)는 고정 commit `5bef215895b69d3f2ef4b82bb3f1970279d67f46`, tag `failed-data-effect-79-v1`에 게시했습니다. 패키지21파일·185,100B를 다시 내려받아 원본의 크기·SHA-256과 모두 대조했고 tag도 같은 commit을 가리킵니다. 모델32,792B는 GitHub에 포함하지 않았습니다. [게시 증명](HF-PUBLICATION-79.v1.json), [파일 목록](HF-FILE-MANIFEST.v2.json), [출처](HF-PROVENANCE.v3.json)를 보세요.

이는 기존 실패71/72와 같은 **비활성 실패 연구 보관**입니다. 원래 과학적 적합성·운영 준비·worker 게시 허용 false는 그대로 두고, 별도의 [보관 정책](ROOT-INACTIVE-ARCHIVE-POLICY.v1.json)으로 공개 자료만 보관했습니다. 모델 활성화, endpoint, 유료 평가나 추가 학습은 하지 않았습니다. 공개 연구 컬렉션에는 항목 하나를 추가해17→18개가 됐고, 기존17개 항목·순서·설명은 바꾸지 않았습니다.

Wiki는 PR104 자료를 commit `51069ff8eb4abe4fc91ad676e6739886d3df1cc5`, PR105의 실제 학습 결과를 `c594f5adb47e956d7f191d1eae0a4b27d403c98a`에 게시했습니다. 각 네 페이지를 Git 원본에서 그대로 복사하고 원격에서 다시 확인했습니다. 한국어·영어 사용 설명과 측정 한계를 함께 유지했습니다. [Wiki105 증명](WIKI-PUBLICATION-105.v1.json)을 참고하세요.

다음은 같은79개 자료의 추가 탐색보다 [새 계약 확보](../next60-acquisition-drafts/README.ko.md)를 우선합니다. 초안20개 중18개는 기존 Source53 논리 요청을 구체화했고2개는 새 native 동작 요청입니다. 적격0개, 실행0회이며96개 입력·59개 후보 설명은 별도 수량입니다. 보호 final 최소2,400개/도메인 목표도 미충족입니다. [독립 PDCA 분석](../next-pdca-after-79/REPORT.ko.md)과 후속 검토 결과를 함께 읽어야 합니다.

준비 과정에서 확정되지 않은 job attempt를 확정 보고로 잘못 적은 미게시 패키지가 있었습니다. 이를 게시하지 않고 새 버전에서 공식값으로 정정했습니다. 원본 버전을 보존한 [정정 기록](HF-PREPARATION-INTERPRETATION-CORRECTION.v3.json)을 남깁니다. 숫자 결과나 모델은 바꾸지 않았습니다.

예산의 기존 학습 소비3/8회·논리 모델3/16개와 경로별 전체 바이트를 이어서 셉니다. 복사본의 해시가 같아도 저장량을 할인하지 않습니다. [게시 전 자원 snapshot](PUBLICATION-BUDGET.v1.json)은 지정한 원장 범위이며 전체 Mac·원격 저장소·과거 삭제 파일·총 쓰기량 조사가 아닙니다. helper 실행 파일과 캐시는 별도 개발 도구로 구분하며, 디스크 크기와 RAM 절감을 혼동하지 않습니다.
