package main

import (
	"github.com/teamswyg/laya-tools/internal/app"
	"testing"
)

func TestRepoPreviewDoesNotNeedRuntime(t *testing.T) {
	a := &app.App{}
	if err := runRepos(a, "../../examples/repositories/catalog.json", "refund invoices", false, false, true, .9); err != nil {
		t.Fatal(err)
	}
	if err := runRepos(a, "../../examples/repositories/catalog.json", "로그인 오류", true, false, true, .9); err != nil {
		t.Fatal(err)
	}
}
