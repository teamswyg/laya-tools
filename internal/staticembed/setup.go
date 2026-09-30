package staticembed

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Setup downloads only immutable public assets. Ordinary loading is offline.
func Setup(dir string) error {
	if e := os.MkdirAll(dir, 0700); e != nil {
		return e
	}
	client := http.Client{Timeout: 60 * time.Second}
	for _, a := range []struct {
		name, sha string
		limit     int64
	}{{"tokenizer.json", VocabularySHA256, 1 << 20}, {"model.safetensors", ModelSHA, 8 << 20}, {"README.md", "38af170acce93b5eb0c4e18659ab0b30581b2f6860f3e12d8ba8b8751e249363", 1 << 16}} {
		path := filepath.Join(dir, a.name)
		if _, e := readVerified(path, a.sha, a.limit); e == nil {
			continue
		}
		resp, e := client.Get("https://huggingface.co/minishlab/potion-base-2M/resolve/" + Revision + "/" + a.name)
		if e != nil {
			return e
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("download status %d", resp.StatusCode)
		}
		b, e := io.ReadAll(io.LimitReader(resp.Body, a.limit+1))
		resp.Body.Close()
		if e != nil {
			return e
		}
		if int64(len(b)) > a.limit {
			return fmt.Errorf("asset exceeds byte limit")
		}
		f, e := os.CreateTemp(dir, ".asset-*")
		if e != nil {
			return e
		}
		tmp := f.Name()
		_, e = f.Write(b)
		ce := f.Close()
		if e == nil {
			e = ce
		}
		if e == nil {
			_, e = readVerified(tmp, a.sha, a.limit)
		}
		if e == nil {
			e = os.Rename(tmp, path)
		}
		os.Remove(tmp)
		if e != nil {
			return e
		}
	}
	return nil
}
