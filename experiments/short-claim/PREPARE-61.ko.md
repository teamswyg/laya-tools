# 61: 남은 기존 문장의 내용 검토 범위 준비

이 초안은 기존 학습 후보의 문장과 코드를 실제로 대조할 다음 범위를 정합니다. 현재 완료된 것은 저장 JSON의 해시 확인과 참조 연결뿐입니다. 모든 내용 판정은 `pending`이며 공개·계획 동결·독립 내용 검토는 부모 작업에서 이어갑니다. 새 표본, 정답, 역할, 학습, 모델 호출은 0입니다.

선택은 결과 점수와 무관하게 59의 전체 가족에서 `stable-odd`, `atomic-commit`, `error-identity`를 제외합니다. 부모와 후보는 원래 순서 그대로입니다. 남은 부모 인덱스는 0–15, 20–47, 56–71입니다. 60에서 이미 연결한 소스를 다시 분석하거나 실행하지 않았습니다.

| 준비 범위 | 실제 저장 메타데이터 수 |
|---|---:|
| 가족 / 기존 부모 / 후보 위치 | 15 / 60 / 180 |
| 후보가 2 / 3 / 4개인 부모 | 11 / 38 / 11 |
| 기존 known / no_answer / unknown 부모 | 28 / 14 / 18 |
| 소스 ID / 고유 후보 문장 SHA | 49 / 60 |
| 계약 / 기존 literal 사례 | 15 / 114 |
| 기존 59 검토 슬롯 | 615 |
| 선택된 소스의 component 관계 / ID / formatted SHA | 172 / 56 / 39 |

615는 요청 60 + 후보 closure 180 + 후보 fidelity 180 + 후보 coverage 180 + 계약 관측 15를 합친 기존 슬롯 수입니다. 후보별 네 축의 pending 항목이나 문장 해시 60개를 독립 요청 수로 세지 않습니다. 전체 기존 코퍼스는 부모 72개·후보 위치 216개·그룹 17개·known/no_answer 포함 그룹 16개·unknown 21개로 유지됩니다. 선택 범위는 기존 그룹 14개에 닿으며 `closed-window`와 `clamp-window`는 같은 그룹 8에 남습니다.

대상 가족은 retain-active, allow-owner, closed-window, quota-total, remove-first, compact-runs, rotate-left, nondecreasing, clamp-window, prefix-balance, transform-order, owned-snapshot, cancellation-lifecycle, quoted-delimiters, ancestor-cycle입니다. 각 가족은 기존 부모 4개와 후보 위치 12개를 갖지만 부모마다 후보 수가 다릅니다. quoted-delimiters의 4개 부모는 모두 원래 unknown이며 바꾸지 않습니다.

고정 입력은 다음 네 파일입니다. 정확히 읽은 바이트의 SHA를 확인한 뒤 그 바이트를 JSON으로 해석했습니다.

| 입력 | SHA-256 |
|---|---|
| caption-coverage-59.json | `7c1bd449d533d3d46a9211dea0a436ce338fe8f16ee3197d554744c3c24b0a29` |
| source-inventory-60.json | `d99d635929555d4123bbaa84aac99d0fb29db345f8b813ed70074dd3e09cdbd4` |
| content-review-protocol-60.json | `6f4504ba04e0b5813e4dc58ec0f6dbe5101a1127e09a379172dd57d7ed11ff19` |
| content-review-recipe-59.json | `17f0abc9843e082d53e78f0eff79842d592db89f2532c0ece0e3a44814065cff` |

네 입력의 저장소 경로는 모두 `experiments/short-claim/` 아래입니다. 새 scope는 요청·후보 원문을 복사하지 않고 원래 JSON pointer·텍스트 SHA·바이트 수를 참조합니다. 후보마다 기존 formatted root SHA와 typed bundle SHA를 저장60의 원래 root에 연결합니다. component 관계의 기존 순서와 ID도 유지합니다. 같은 enum의 raw 선언을 공유하는 여러 ID를 하나로 합치지 않습니다. 이는 metadata 일치 확인이며 object binding이나 closure 완전성 증명은 아닙니다.

검토 축과 상태 어휘는 59/60과 같습니다. 축은 `source_fidelity`, `request_contract_coverage`, `observation_fields`, `explicit_negative_boundaries`입니다. 상태는 `pending`, `consistent_with_scoped_evidence`, `omits_required_scope`, `contradicts_scoped_evidence`, `unsupported_or_uncertain`입니다. 잘못 동작하는 코드의 설명도 충실할 수 있으므로 fidelity와 요청 충족을 별도로 기록합니다.

다음 읽기는 원래 순서로 진행합니다. 각 주장에 원문 바이트 구간, 구현 root와 모든 관련 helper/type/method/sentinel, observer, literal Input/Want, 관측 필드와 제외 범위를 연결합니다. material한 누락·반대 근거·여러 해석이 남으면 해당 축은 consistent가 될 수 없습니다. 원래 truth의 unknown과 내용 검토의 불확실성은 다른 상태입니다. 부분적으로 분명한 설명을 찾더라도 원래 unknown을 새 정답으로 승격하지 않습니다. 관측이 없는 Got이나 일반적 보장을 만들어 넣지 않습니다.

`boundary-checklists-61.json`은 이 읽기를 돕는 질문 목록입니다. 예를 들어 inclusive 경계, 인접 중복의 필수 축약, 연산 순서, snapshot의 nil/empty와 순차 양방향 쓰기, cancellation의 결과→release 순서, quoted tokenizer의 전체 error/zero 배열, ancestor cycle과 shared descendant의 차이를 확인하게 합니다. 레거시 observer의 panic/input mutation은 unknown이고 typed 후보의 panic은 관측된 mismatch입니다. 이 차이도 보존합니다. 표준 라이브러리·인프라 가정과 지원 범위를 읽지 않은 채 성공 판정을 하지 않습니다.

준비 Go 도구는 저장 메타데이터만 읽습니다. 합성 race/vet 검사가 통과했으며 순서·root/bundle pin·중복 ID·잘못된 상태·literal count 변조 거절을 확인했습니다. private 메타데이터 조인 1회는 입력 4개와 후보 180개, 계약 15개 연결을 완료했고 실패 0입니다. ledger는 실제 시도/완료 수와 고정 실패 코드를 분리합니다. 오류 거절용 합성 테스트는 별도이며 공식 내용 검토의 재시도가 아닙니다. 원본 AST/formatter/Rebind/SourcePins/Bind/Generate/후보 API/순위/유료·모델 호출/fit/protected-final 읽기는 모두 0입니다. 저장소 파일도 수정하지 않았습니다.

이 큰 메타데이터 파일은 유지보수 도구의 참조이며 runtime hint의 메모리·속도 개선을 주장하지 않습니다. 이전 개발 점수와 정답은 이미 관측되었으므로 blind 검증이라고 부르지 않습니다. English Go 범위를 유지하며 이 한국어 문서는 언어 성능 시험이 아닙니다. `synthetic_single_pipeline`, `no_roles_plan`, 역할·작성 흐름 다양성은 아직 해결되지 않았고 `training_ready=false`입니다. CI는 파일·참조의 무결성을 확인할 수 있지만 내용 충실성을 대신 판정하지 않습니다. 새 사람 승인 흐름, 15개 저장소 요건, 모든 개발 fit마다 2,400개 요건을 추가하지 않습니다. 도메인별 2,400개 이상의 보호된 독립 최종평가 목표는 별도로 유지합니다.
