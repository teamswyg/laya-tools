package inference

import (
	"os"
	"strings"
	"sync"
	"testing"
)

func TestPredictClosedEngine(t *testing.T) {
	e := &Engine{}
	if _, err := e.Predict("test", "choice", "Choose", []string{"a", "b"}); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("got %v", err)
	}
}
func TestConcurrentPredictAndClose(t *testing.T) {
	dir, lib := os.Getenv("LAYA_MODEL_DIR"), os.Getenv("LAYA_RUNTIME")
	if dir == "" || lib == "" {
		t.Skip("native fixture required")
	}
	e, err := New(Options{ModelDir: dir, Runtime: lib, Provider: "cpu", Threads: 4, MaxTokens: 512})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := e.Predict("A typo in a comment.", "choice", "Choose difficulty", []string{"easy", "hard"})
			if err != nil && !strings.Contains(err.Error(), "closed") {
				t.Error(err)
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		if err := e.Close(); err != nil {
			t.Error(err)
		}
	}()
	close(start)
	wg.Wait()
	if _, err := e.Predict("test", "choice", "Choose", []string{"a", "b"}); err == nil {
		t.Fatal("predict after close succeeded")
	}
}
