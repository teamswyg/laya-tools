package filelabels

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestProjectionIsolationAndIdentity(t *testing.T) {
	cases := []struct {
		text string
		bad  bool
	}{
		{`{"instance_id":"x","patch":"diff --git a/x b/x\nold mode 100644\nnew mode 100755\n"}`, false},
		{`{"instance_id":"x","patch":"bad patch"}`, false},
		{`{"instance_id":"x","patch":"bad","problem_statement":"must remain separate"}`, true},
		{"{\"instance_id\":\"x\",\"patch\":\"bad\"}\n{\"instance_id\":\"x\",\"patch\":\"bad\"}", true},
	}
	for _, c := range cases {
		file := filepath.Join(t.TempDir(), "labels.jsonl")
		if e := os.WriteFile(file, []byte(c.text), 0600); e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256([]byte(c.text))
		got, e := Read(file, hex.EncodeToString(h[:]))
		if (e != nil) != c.bad {
			t.Fatalf("schema result %v", e)
		}
		if !c.bad && len(got) != 1 {
			t.Fatal("lost row")
		}
		if _, e := Read(file, "wrong"); e == nil {
			t.Fatal("hash unchecked")
		}
	}
}
