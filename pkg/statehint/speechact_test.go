package statehint

import "testing"

func TestSpeechActOriginalContrasts(t *testing.T) {
	cases := []struct {
		text string
		want Intent
	}{{"Please explain why this was completed.", Question}, {"I have completed the change.", CompletionReport}, {"The change is not completed; I am still working on it.", Progress}, {"Suppose the work were completed.", Unclear}, {"Someone said completed, but the current facts are missing.", Unclear}, {"The work is completed and still in progress.", Unclear}, {"내일 시작할 예정입니다.", Planned}, {"권한이 없어 진행할 수 없습니다.", Blocker}, {"자료를 공유합니다. 참고용입니다.", Reference}}
	for _, c := range cases {
		p, e := RuleSpeechAct(c.text)
		if e != nil || p.Intent != c.want || p.Source != RuleSource {
			t.Fatalf("contrast %q: %+v %v", c.text, p, e)
		}
	}
}
