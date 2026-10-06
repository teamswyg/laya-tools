// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehint

import (
	"strings"
	"unicode/utf8"
)

// RuleSpeechAct is a new unlearned control, frozen before the new locked test.
// It prioritizes explicit questions and scoped uncertainty over keyword counts.
// Its one-hot result describes rule selection, never calibrated probability.
func RuleSpeechAct(text string) (Prediction, error) {
	if len(text) > MaxTextBytes || !utf8.ValidString(text) || strings.ContainsRune(text, 0) {
		return Prediction{}, ErrInput
	}
	s := strings.ToLower(strings.TrimSpace(text))
	one := func(intent Intent) Prediction {
		var p [8]float64
		i, _ := IntentIndex(intent)
		p[i] = 1
		return prediction(p, RuleSource, 0)
	}
	if s == "" {
		return one(Unclear), nil
	}
	has := func(terms ...string) bool {
		for _, t := range terms {
			if strings.Contains(s, t) {
				return true
			}
		}
		return false
	}
	// Reporting somebody else's assertion and explicit hypothetical framing do
	// not establish the speaker's current speech act. This is conservative.
	if has("hypothetical", "suppose ", "if we ", "if the ", "would be", "만약", "가정하", "예를 들", "라고 썼", "라고 했", "인용", "quoted", "the message says", "someone said") {
		return one(Unclear), nil
	}
	question := has("?", "？", "please explain", "could you explain", "can you explain", "how do ", "how can ", "why does ", "what is ", "what does ", "is it ", "알려 주", "알려주", "설명해", "설명해 주", "어떻게", "무엇", "왜 ", "언제 ", "인가요", "있나요", "맞나요", "할까요")
	if question {
		return one(Question), nil
	}
	if has("please cancel", "cancel this", "cancel the ", "withdraw this", "withdraw the ", "취소해", "취소하겠", "취소 부탁", "철회해", "중단해 주세요", "중단해주세요") {
		return one(CancelRequest), nil
	}
	blocked := has("blocked", "cannot proceed", "can't proceed", "unable to continue", "waiting for access", "missing permission", "진행할 수 없", "진행이 불가능", "막혔", "권한이 없", "권한을 기다", "응답을 기다", "승인을 기다", "접근할 수 없")
	active := has("working on", "currently implementing", "currently reviewing", "still working", "in progress", "진행 중", "진행중", "작업 중", "작업중", "검토 중", "검토중", "구현 중", "처리 중")
	negativeDone := has("not done", "not finished", "not completed", "isn't finished", "hasn't finished", "완료되지", "완료하지 못", "끝나지", "마무리하지 못", "아직 남", "remaining work", "still pending")
	done := has("completed", "finished", "done now", "all done", "delivered", "submitted", "완료했", "완료됐", "완료되었", "끝냈", "끝났", "마쳤", "마무리했", "모두 처리했")
	plan := has("will start", "plan to ", "planning to ", "scheduled for", "tomorrow i ", "next week i ", "할 예정", "시작할 예정", "계획입니다", "계획이야", "계획 중", "아직 시작하지", "내일 시작", "다음 주 시작")
	reference := has("for reference", "reference material", "here is the ", "attached is", "sharing the ", "sharing a ", "documentation link", "참고자료", "참고 자료", "참고용", "자료를 공유", "링크를 공유", "첨부합니다", "문서를 공유")
	if done && (negativeDone || active || blocked || plan) {
		if negativeDone && active && !blocked && !plan {
			return one(Progress), nil
		}
		return one(Unclear), nil
	}
	count := 0
	var intent Intent
	for _, x := range []struct {
		flag bool
		i    Intent
	}{{blocked, Blocker}, {active, Progress}, {done && !negativeDone, CompletionReport}, {plan, Planned}, {reference, Reference}} {
		if x.flag {
			count++
			intent = x.i
		}
	}
	if count == 1 {
		return one(intent), nil
	}
	return one(Unclear), nil
}
