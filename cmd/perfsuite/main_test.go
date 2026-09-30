package main

import "testing"

func TestNativeFallbackIsNotPerformanceSuccess(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		ok         bool
	}{
		{"route-hard", `{"abstained":true,"reason":"Laya unavailable"}`, false},
		{"route-hard", `{"result":{"abstained":true,"probabilities":[0.2,0.3,0.5]},"ms":1}`, true},
		{"search-laya", `{"results":[{"relevance":0.2}]}`, true},
		{"search-laya", `{"results":[{"score":1}]}`, false},
		{"search-laya", `{"results":[{"relevance":0.2}],"warnings":["fallback"]}`, false},
		{"repositories-laya", `{"model_judgments":0}`, false},
		{"repositories-laya", `{"model_judgments":15}`, true},
		{"native-warm", `not JSON`, false},
	} {
		t.Run(tc.name+tc.body, func(t *testing.T) {
			if err := validate(scenario{name: tc.name, native: true}, []byte(tc.body)); (err == nil) != tc.ok {
				t.Fatalf("err=%v expected success=%v", err, tc.ok)
			}
		})
	}
}
