// Package assets installs only checksum-pinned public model/runtime bundles.
package assets

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//go:embed manifest.json
var manifest []byte

type Artifact struct {
	URL    string            `json:"url"`
	SHA256 string            `json:"sha256"`
	Files  map[string]string `json:"files"`
}

func Cache() string {
	if v := os.Getenv("LAYA_CACHE"); v != "" {
		return v
	}
	p, err := os.UserCacheDir()
	if err != nil {
		return ".laya-cache"
	}
	return filepath.Join(p, "laya-tools")
}
func ModelDir(name string) string { return filepath.Join(Cache(), name) }
func Runtime() string {
	if v := os.Getenv("LAYA_RUNTIME"); v != "" {
		return v
	}
	name := "libonnxruntime.1.30.0.dylib"
	if runtime.GOOS == "linux" {
		name = "libonnxruntime.so.1.30.0"
	}
	return filepath.Join(Cache(), "runtime", name)
}
func Setup(ctx context.Context, model string, w io.Writer) error {
	if model != "base" && model != "code" {
		return fmt.Errorf("model must be base or code")
	}
	var artifacts map[string]Artifact
	if err := json.Unmarshal(manifest, &artifacts); err != nil {
		return err
	}
	for _, item := range []struct{ key, dir string }{{"runtime-" + runtime.GOOS + "-" + runtime.GOARCH, "runtime"}, {model, model}} {
		a, ok := artifacts[item.key]
		if !ok {
			return fmt.Errorf("no pinned artifact for %s", item.key)
		}
		dest := filepath.Join(Cache(), item.dir)
		if Verify(dest, a.Files) == nil {
			fmt.Fprintln(w, item.dir+": verified cache hit")
			continue
		}
		fmt.Fprintln(w, "Downloading and verifying "+item.key)
		if err := install(ctx, a, dest); err != nil {
			return err
		}
	}
	return nil
}
func Verify(dir string, files map[string]string) error {
	if len(files) == 0 {
		return fmt.Errorf("empty manifest")
	}
	for name, want := range files {
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		h := sha256.New()
		_, err = io.Copy(h, f)
		f.Close()
		if err != nil {
			return err
		}
		if hex.EncodeToString(h.Sum(nil)) != want {
			return fmt.Errorf("checksum mismatch: %s", name)
		}
	}
	return nil
}
func install(ctx context.Context, a Artifact, dest string) error {
	if !strings.HasPrefix(a.URL, "https://") || len(a.SHA256) != 64 {
		return fmt.Errorf("invalid pinned artifact")
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Dir(dest), ".install-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	req, err := http.NewRequestWithContext(ctx, "GET", a.URL, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 15 * time.Minute, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if r.URL.Scheme != "https" || len(via) > 8 {
			return fmt.Errorf("unsafe redirect")
		}
		return nil
	}}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}
	archive, err := os.CreateTemp(stage, "archive-")
	if err != nil {
		return err
	}
	h := sha256.New()
	_, err = io.Copy(io.MultiWriter(archive, h), io.LimitReader(resp.Body, 2<<30))
	if err != nil {
		archive.Close()
		return err
	}
	if hex.EncodeToString(h.Sum(nil)) != a.SHA256 {
		archive.Close()
		return fmt.Errorf("archive checksum mismatch")
	}
	if _, err = archive.Seek(0, 0); err != nil {
		archive.Close()
		return err
	}
	defer archive.Close()
	gz, err := gzip.NewReader(archive)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	seen := map[string]bool{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// Official runtime archives also contain dSYM debug files with the same
		// basename as the library. Accept only flat model files or package/lib files.
		parts := strings.Split(strings.TrimPrefix(hdr.Name, "./"), "/")
		if len(parts) > 3 || len(parts) == 3 && parts[1] != "lib" {
			continue
		}
		name := filepath.Base(hdr.Name)
		if _, ok := a.Files[name]; !ok {
			continue
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if seen[name] || hdr.Size > 2<<30 {
			return fmt.Errorf("invalid archive entry")
		}
		seen[name] = true
		f, err := os.OpenFile(filepath.Join(stage, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		_, err = io.CopyN(f, tr, hdr.Size)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	if err = Verify(stage, a.Files); err != nil {
		return err
	}
	archive.Close()
	os.Remove(archive.Name())
	// Existing invalid installs are moved aside only after the replacement verifies.
	backup := dest + ".previous"
	if _, err = os.Stat(backup); err == nil {
		return fmt.Errorf("previous installation backup exists: remove it before retry")
	}
	if _, err = os.Stat(dest); err == nil {
		if err = os.Rename(dest, backup); err != nil {
			return err
		}
	}
	if err = os.Rename(stage, dest); err != nil {
		_ = os.Rename(backup, dest)
		return err
	}
	return os.RemoveAll(backup)
}
