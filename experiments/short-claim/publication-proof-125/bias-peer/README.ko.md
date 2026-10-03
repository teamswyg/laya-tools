# 후보 순서 편향 점검 소스 검토

현재 순위 계산 핵심의 필수 수정은0건입니다. 작성자가 아닌 검토자가 비맹검으로 핵심3개 Go snapshot과 기존 baseline 정렬/Prepared 검증을 읽었습니다. 실제33 데이터·Reader·Baselines·Go 실행은0입니다.

33개 부모를 한 번 읽고 primary33개와 같은 부모의 전체 permutation186개를 구분합니다. Text와 normalized text, label 순서가 같은 map으로 이동하며 동점은 현재 display 순서를 보존하고 selected index로 되돌립니다. Baselines219회/순위876개는 산술 계획이며 새 표본 수가 아닙니다.

작은 권고1건: CLI는 output fresh 예약보다 Run이 먼저입니다. Root는 실행 전에 exclusive attempt를 예약하거나 해당 순서를 옮겨 기존 output 때문에 계산 뒤 기록이 사라지는 상황을 막아야 합니다. 실패 결과를 Baselines0으로 기록하거나 자동 재시도하지 않습니다.

후보2/3개에서 Top3는 자명합니다. permutation 합계는3후보 부모를6회,2후보 부모를2회 가중합니다. 방문 감소는 모든 후보를 끝까지 확인하는 비용 대비 시뮬레이션이며5% best-baseline 효용이나 모델·메모리 성능이 아닙니다.
