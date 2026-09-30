// riido-hubcheck validates a local release without uploading or requiring HF.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/researchbundle"
	"os"
	"os/exec"
)

func run() error {
	dir := flag.String("package", "", "explicit staged public package")
	ci := flag.Bool("require-ci", false, "also require successful GitHub CI on exact source revision")
	flag.Parse()
	m, e := researchbundle.Verify(*dir)
	if e != nil {
		return e
	}
	if *ci {
		b, e := exec.Command("gh", "run", "list", "--repo", "teamswyg/laya-tools", "--workflow", "ci.yml", "--commit", m.SourceRevision, "--limit", "30", "--json", "headSha,status,conclusion").Output()
		if e != nil {
			return fmt.Errorf("CI lookup failed: %w", e)
		}
		var runs []struct{ HeadSha, Status, Conclusion string }
		if e = json.Unmarshal(b, &runs); e != nil {
			return e
		}
		ok := false
		for _, r := range runs {
			if r.HeadSha == m.SourceRevision && r.Status == "completed" && r.Conclusion == "success" {
				ok = true
			}
		}
		if !ok {
			return fmt.Errorf("successful CI required for exact source revision")
		}
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Repository string `json:"repository"`
		Source     string `json:"source_revision"`
		Integrity  bool   `json:"integrity_verified"`
		CI         bool   `json:"ci_verified"`
	}{m.Repository, m.SourceRevision, true, *ci})
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
