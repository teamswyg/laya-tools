# 완료 디렉터리 압축본 저장 감사

별도 Node 구현으로 새2압축본을 메모리에서만 읽었습니다. 단일 gzip의 first-deflate 끝·CRC·ISIZE·전체EOF와 tar334 regular files+91 directories=425 entries의 경로·바이트·SHA·mode가 모두 일치합니다. 실제 PAX metadata header는0이며 USTAR prefix를 처리했습니다. 빈 파일83개도 검증됐습니다.

672 journal events는2COMMIT,334개 intent/ACK쌍,2cleanup ACK의 정확한 순서이며 rawJSON+LF 해시 연결도 맞습니다. COMMIT·marker·ACK·archive·selection pins와 FINAL counters가 일치합니다. 현재 원래 두 root는 같은inode의 marker-only 디렉터리이고 restore sibling은 없습니다. 원payload1,804,852B, archive438,302B, control reserve131,072B를 뺀 conservative reclaim1,235,478B가 FINAL과 같습니다.

기존5binaryarchive의 현재 SHA·선택 membership·commit/cleanup/삭제상태를 재확인했습니다. 이전20은 변경 없는 account/helper census로 검증됐습니다. 합계27archive proofs는 legacy mapping20을 늘리거나 roots/512MiB 한도를 초기화하지 않습니다. 이 수치는 디스크 논리 바이트이며 모델 RAM과 무관합니다.

과거 실제 디스크 복원·mode·Sync-before-unlink는 Root 실행 영수증에서 승계합니다. 이번 독립 검증이 그 물리 실행을 재관측한 것은 아닙니다. 검토자는 이전 archive 소스 peer라 비눈가림입니다. 디스크복원·삭제·재압축·Go·Reader·원본API·native·모델·Fit·네트워크0입니다.

첫 소스 조립16,508B가 임의15,000B 하위 guard에 걸려 pre-write파일0이었습니다. 최종 source13,360B와 report3,953B를 유지하며 전체24KiB 한도에서 봉인합니다. 처음 inventory 요약은 큰 nested list를 출력해 잘렸으므로 이후 좁은 count/pointer만 읽었습니다. 원본 body·전체입력 복사나 scientific값 해석은 없습니다.

재현: `node audit-archives.cjs ROOT REPORT` (새 output만 허용). Private pin 파일은 공개 제외입니다.
