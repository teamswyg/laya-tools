# Read-only catalog bridge / 조회 catalog 연결

`SnapshotReader.Read` is implemented by an application's trusted boundary. `Adapter.Propose` maps caller content to current opaque label/status references using explicit configuration and an injected predictor. It exposes no mutation, notification, event or user-JSON trust API.

실제 ID·이름·인증정보는 private reader 안에 둡니다. 공개 코드에는 opaque 참조만 제공하고, 라벨의 의미 연결은 설정으로 정합니다. 그룹 container·inactive·중복·누락 연결을 제외합니다. 같은 상태 유형에 후보가 여러 개이면 설정된 preferred 참조가 없을 때 보류합니다. 별도 조회 시점은 그대로 표시하고 `catalog_consistency:unqualified`를 유지합니다.

Canonical Work 버전이 없는 항목은 라벨·이모지만 제안하며 임의 버전0/1을 만들지 않습니다. 초기 bridge에는 verified event가 없어 상태 변경이 항상false입니다. 앱이 실제 조회 권한·현재 catalog·버전·명령 중복을 확인해야 하며, 이 interface만으로 production 통합이나 권한을 증명하지 않습니다. 테스트는 명시된 test double이며 실제 서비스 호출 결과가 아닙니다.

The classifier hot path owns arrays and has no shared locks. This bounded catalog boundary uses per-call maps for matching at most64work references/128labels/32statuses; it is not a claim of zero allocation or a substitute for measuring actual reader/JSON/application costs.
