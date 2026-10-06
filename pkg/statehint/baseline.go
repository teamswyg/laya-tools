package statehint

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// RuleBaseline is an unlearned comparison, not a pretrained or fine-tuned
// model. Its one-hot scores indicate deterministic rule selection only and
// must not be interpreted as calibrated empirical probabilities.
func RuleBaseline(text string) (Prediction, error) {
	if len(text) > MaxTextBytes || !utf8.ValidString(text) || strings.ContainsRune(text, 0) {
		return Prediction{}, ErrInput
	}
	text = strings.ToLower(text)
	var matches [IntentCount]bool
	contains := func(phrases ...string) bool {
		for _, phrase := range phrases {
			for from := 0; from <= len(text)-len(phrase); {
				index := strings.Index(text[from:], phrase)
				if index < 0 {
					break
				}
				index += from
				// Non-ASCII phrases use substring matching. ASCII vocabulary
				// is matched at letter/number boundaries, avoiding substrings.
				if phrase[0] >= 0x80 {
					return true
				}
				leftOK, rightOK := index == 0, index+len(phrase) == len(text)
				if !leftOK {
					r, _ := utf8.DecodeLastRuneInString(text[:index])
					leftOK = !unicode.IsLetter(r) && !unicode.IsNumber(r)
				}
				if !rightOK {
					r, _ := utf8.DecodeRuneInString(text[index+len(phrase):])
					rightOK = !unicode.IsLetter(r) && !unicode.IsNumber(r)
				}
				if leftOK && rightOK {
					return true
				}
				from = index + len(phrase)
			}
		}
		return false
	}
	set := func(intent Intent, value bool) { index, _ := IntentIndex(intent); matches[index] = value }
	set(Question, contains("how", "why", "which", "could you", "can you", "어떻게", "왜", "질문", "알려줘", "나요", "까요") || strings.Contains(text, "?"))
	set(Blocker, contains("blocked", "blocker", "cannot", "can't", "error", "failure", "broken", "stuck", "막혔", "막혀", "오류", "실패", "불가능", "대기 중", "기다리고"))
	set(Reference, contains("reference", "documentation", "manual", "attachment", "attached", "https://", "http://", "참고", "자료", "문서", "첨부", "링크"))
	set(Progress, contains("started", "working", "in progress", "implementing", "investigating", "시작했", "진행 중", "진행중", "작업 중", "처리 중", "확인 중"))
	completionNegated := contains("not done", "not completed", "not finished", "incomplete", "아직 완료", "완료되지", "완료하지", "미완료", "끝나지")
	set(CompletionReport, !completionNegated && contains("completed", "finished", "done", "resolved", "완료했", "완료됐", "완료되었", "해결했", "끝났", "마쳤"))
	set(CancelRequest, contains("cancel", "cancelled", "stop this", "abandon", "취소", "중단해", "그만해"))
	set(Planned, contains("plan", "planned", "will", "tomorrow", "next week", "예정", "계획", "내일", "다음 주", "할게", "하겠"))
	winner := Unclear
	count := 0
	for index, match := range matches {
		if match {
			winner = Intents()[index]
			count++
		}
	}
	// Overlapping intent evidence and completion negation abstain explicitly.
	if count != 1 || completionNegated {
		winner = Unclear
	}
	var probabilities [IntentCount]float64
	index, _ := IntentIndex(winner)
	probabilities[index] = 1
	return prediction(probabilities, RuleSource, 0), nil
}
