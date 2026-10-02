# 공개 연구 보관과 읽기 확인

첫 FP32 모델은 효용에 실패했지만, 같은 실패를 반복하지 않도록 [Hugging Face 연구 보관소](https://huggingface.co/JooYoon/riidolaya-shortclaim-fp32-failed-71/tree/32b8f4579065247be0f71c83e1e7c143845e7b33)에 원본 계수·수치 결과·한영 설명·라이선스 고지·checksum을 공개했다. 모델은 Git에 올리지 않았다. 모델 본체32,792B를 포함한 프로젝트 파일9개135,328B를 새로 내려받아 SHA를 대조했고 모두 일치했다. `failed-fp32-71-v1` 태그가 같은 commit으로 해석됨을 확인했다. 공개 연구 컬렉션의 기존15개 항목은 그대로 유지하고 이 실패 자료 하나만 추가했다.

연구 보관은 성능 승인과 다르다. 첫 모델의27→31회 확인 악화, Top1 5→2, `PublicationQualified=false`·`ProductionReady=false`·원래 `publication_approved=false`를 유지한다. 자동 pipeline·widget·기본 실행 연결·유료 배포는 없다. [실제 게시·다운로드 검증 기록](HF-PUBLICATION-71.v1.json)이 고정 commit과 파일 해시를 연결한다.

PR94의 [한국어 Wiki](https://github.com/teamswyg/laya-tools/wiki/First-Claim-Fit-71-KO)와 [영문 Wiki](https://github.com/teamswyg/laya-tools/wiki/First-Claim-Fit-71-EN)도 게시했고, 원격 Git 내용을 CI를 통과한 문서와 대조했다. [Wiki 검증 기록](WIKI-PUBLICATION-94.v1.json)은 성공한 helper 검증을 보존한다. Root 인계에서는 최초 원격 읽기 확인에 잘못 적은 Wiki commit을 사용하여 안전하게 거절된1회가 있었고, 실제 commit을 읽어 수정한 두 번째 확인은 통과했다. Wiki push는1회다. 이 앞선 실패는 성공 helper 기록에 합쳐 지우지 않는다.

Issue19 댓글의 첫 읽기 비교도 CLI가 붙인 마지막 줄바꿈 때문에 실패했고, 원문 바이트를 다시 읽어 일치함을 확인했다. 댓글은 한 번 게시했다. HF 게시 전 첫 CI JSON assertion은 다른 응답 형태의 필드명 때문에 실패했으며, 어떠한 원격 변경도 수행하기 전에 올바른 형태로 검증해 통과했다. 이런 metadata 인계 실패는 학습 재시도나 모델 효용 변화가 아니다.

두 번째 scratch 형제 모델도 별도 [실패 연구 보관소](https://huggingface.co/JooYoon/riidolaya-shortclaim-rank-bce-failed-72/tree/d090b00e9a5dab00d5372dfd6412c9aee0c60b7b)에 공개했다. 프로젝트 파일9개157,923B를 새로 내려받아 모두 대조했고 `failed-rank-bce-72-v1` 태그는 같은 commit으로 해석됐다. 컬렉션의 앞선16항목을 그대로 보존해17개가 됐다. [게시 검증](HF-PUBLICATION-72.v1.json)과 [PR95 최종 CI·병합 기록](CI-PR95.v1.json)을 연결한다. 실제 fit의 source는d506이고 이후851의 Go 소스는 같으며, 원격 공개는851 CI·자동 병합을 확인한 뒤 진행했다. 27→33 악화와 미승인 상태는 그대로다.

[English](README.en.md)
