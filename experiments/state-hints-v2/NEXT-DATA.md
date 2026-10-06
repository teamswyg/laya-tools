# 다음 외부 자료 검토 / Next external-data review

현재 원본 합성 개발 사례는 실제 제품 정확도를 검증하지 못합니다. 별도 공개 사람이 작성한 자료를 검토하되, 공개 문장을 바로 현재8의미의 정답으로 바꾸지 않습니다.

[Amazon MASSIVE 공식 설명](https://github.com/alexa/massive/blob/main/README.md)은51언어의 virtual-assistant utterance와60intent를 다룹니다. [공식 NOTICE](https://github.com/alexa/massive/blob/main/NOTICE.md)는 데이터CC BY4.0과SLURP 원천을 고지합니다. [공식 Dataset card](https://huggingface.co/datasets/AmazonScience/massive/blob/c8ca39d49d2c06ee13daeb137f26b302e596510d/README.md?code=true)도 별도 확인했습니다. 코드와 데이터 라이선스는 구분해야 하며 재사용 시 원문 terms·NOTICE·attribution·수정 기록을 보존해야 합니다.

MASSIVE의 intent는 현재 업무 완료/진행/질문/자료 ontology와 같지 않습니다. 따라서 지금 모델에 학습·평가 라벨로 넣지 않았습니다. 먼저 질문·행동요청·범위밖 입력을 검토하는 별도 proxy로 정의하거나, 명시적 새 annotation rubric과 실제 업무 데이터의 local-only 검증을 준비해야 합니다. 여러 언어의 같은 원천문장은 같은 source group으로 취급해야 합니다.

This is a source/license/domain-fit screen only. No dataset download, model fit, inference, label remapping or publication took place. MASSIVE's virtual-assistant intents do not establish ground truth for Riido work-state hints. Any future use needs a separate frozen mapping/annotation protocol, source-group partition and full attribution. Existing synthetic final cases must not select later models.
