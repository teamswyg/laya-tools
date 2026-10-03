# CI125 Linux 저장 결과 감사

Mac·Linux 저장 결과의 고정 SHA를 확인하고 별도 Node 구현으로219 cases/876 rankings를 한 번 재계산했습니다. 모든 실제 순서·정확한 동률·매핑·방문·집계와 sidecar 전체가 일치했습니다. 기존 score bound 및1,868 robust pairs는 모두 통과했고688 overlap pairs를 기록했습니다.

Changed ranks250 = score-only236 + behavior14. 동률14, 순서6, selected 순서6, 방문6, rank metadata0 차이입니다. Primary 집계 차이는0입니다. Permuted BM25/narrow_rule은 각각 visits258→255, Top1 132→135, reduction0.532608695652174→0.5380434782608695로 달라졌고 나머지는 같습니다. REPORT에 전체 actual 집계·차이·변화 위치가 있습니다.

노출된 development33 감사이며 순열186은 같은 부모를 반복합니다. 독립 표본 수나 실제 비용 절감·모델 품질 검증이 아닙니다. 검토자는 이전 bias 준비/계약 제안 저자로 비눈가림이며 현재 portable checker는 작성하지 않았습니다. 원본·Reader·Baselines·Go·모델·학습·네트워크 실행, 새 정답·역할·적격 결정0입니다.

보존된 실패: 임의10KiB source guard의 pre-write 파일0; .mjs/CommonJS loader 오류의 입력 읽기0; 최종 문서 조립24KiB guard 초과203B의 추가 파일0. 같은 저장 JSON의 실제 비교는1회입니다. 원소스는 그대로이며 아래 방식으로 성공했습니다.
`node --input-type=commonjs - FROZEN ACTUAL DIFF REPORT < audit-saved.mjs`

GitHub 출처는 Root 다운로드 영수증 승계입니다. Private pin 파일은 공개 제외입니다.
