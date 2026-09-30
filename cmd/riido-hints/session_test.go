package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

const sessionFixture = `{"snapshot_id":"public-demo","query":"read config","documents":[{"id":"a","text":"read"},{"id":"b","text":"config"},{"id":"c","text":"func ReadConfig() {}"}]}`

func TestSessionMatchesStateless(t *testing.T) {
	for _, identifiers := range []bool{false, true} {
		var expected bytes.Buffer
		if err := runPage(strings.NewReader(sessionFixture), &expected, nil, "", identifiers, pageOptions{Limit: 1}); err != nil {
			t.Fatal(err)
		}
		var first response
		if err := json.Unmarshal(expected.Bytes(), &first); err != nil {
			t.Fatal(err)
		}
		next, _ := json.Marshal(struct {
			Cursor string `json:"cursor"`
			Limit  int    `json:"limit"`
		}{first.Page.NextCursor, 2})
		var got bytes.Buffer
		if err := runSession(strings.NewReader(sessionFixture+"\n"+string(next)+"\n"), &got, nil, "", identifiers, 1); err != nil {
			t.Fatal(err)
		}
		dec := json.NewDecoder(&got)
		var a, b response
		if err := dec.Decode(&a); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, first) {
			t.Fatal("first page drift")
		}
		if err := dec.Decode(&b); err != nil {
			t.Fatal(err)
		}
		expected.Reset()
		if err := runPage(strings.NewReader(sessionFixture), &expected, nil, "", identifiers, pageOptions{Limit: 2, Cursor: first.Page.NextCursor}); err != nil {
			t.Fatal(err)
		}
		var last response
		if err := json.Unmarshal(expected.Bytes(), &last); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(b, last) || b.Page.NextCursor != "" {
			t.Fatal("continuation drift")
		}
		if err := dec.Decode(new(any)); err != io.EOF {
			t.Fatal("extra response")
		}
	}
}

func TestSessionRejectsContinuationMutation(t *testing.T) {
	for _, line := range []string{`{}`, `{"cursor":"bad","limit":1}`, `{"cursor":"bad","limit":1,"query":"new"}`, `{} {}`, strings.Repeat("x", 1025)} {
		var out bytes.Buffer
		if err := runSession(strings.NewReader(sessionFixture+"\n"+line+"\n"), &out, nil, "", false, 1); err == nil {
			t.Fatal("accepted malformed continuation")
		}
		if bytes.Count(out.Bytes(), []byte("\n")) != 1 {
			t.Fatal("partial bad response")
		}
	}
	for _, limit := range []int{0, -1, 4097} {
		if err := runSession(strings.NewReader(sessionFixture), io.Discard, nil, "", false, limit); err == nil {
			t.Fatal("invalid limit")
		}
	}
}

func TestSessionLineBounds(t *testing.T) {
	for _, n := range []int{0, 4095, 4096, 4097, 8192} {
		data := strings.Repeat("x", n)
		got, err := readSessionLine(bufio.NewReaderSize(strings.NewReader(data), 4096), 8192)
		if n == 0 {
			if err != io.EOF {
				t.Fatal(err)
			}
		} else if err != nil || string(got) != data {
			t.Fatalf("length %d: %v", n, err)
		}
	}
	if _, err := readSessionLine(bufio.NewReader(strings.NewReader(strings.Repeat("x", 8193))), 8192); err == nil {
		t.Fatal("oversize")
	}
}

// Optional local pinned public fixture; never embed source text in the repo.
func TestSessionRealCatalog(t *testing.T) {
	path := os.Getenv("RIIDO_PUBLIC_SESSION_FIXTURE")
	if path == "" {
		t.Skip("local public catalog not supplied")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, data); err != nil {
		t.Fatal(err)
	}
	data = append(compact.Bytes(), '\n')
	for _, identifiers := range []bool{false, true} {
		var full bytes.Buffer
		if err := runPolicy(bytes.NewReader(data), &full, nil, "", identifiers); err != nil {
			t.Fatal(err)
		}
		var want response
		if err := json.Unmarshal(full.Bytes(), &want); err != nil {
			t.Fatal(err)
		}
		input, writeInput := io.Pipe()
		readOutput, output := io.Pipe()
		done := make(chan error, 1)
		go func() {
			err := runSession(input, output, nil, "", identifiers, 20)
			input.CloseWithError(err)
			output.CloseWithError(err)
			done <- err
		}()
		if _, err := writeInput.Write(data); err != nil {
			t.Fatal(err)
		}
		dec := json.NewDecoder(readOutput)
		var all []candidate
		for {
			var page response
			if err := dec.Decode(&page); err != nil {
				t.Fatal(err)
			}
			all = append(all, page.Candidates...)
			if page.Page.NextCursor == "" {
				break
			}
			next, _ := json.Marshal(struct {
				Cursor string `json:"cursor"`
				Limit  int    `json:"limit"`
			}{page.Page.NextCursor, 20})
			if _, err := writeInput.Write(append(next, '\n')); err != nil {
				t.Fatal(err)
			}
		}
		writeInput.Close()
		if err := <-done; err != nil {
			t.Fatal(err)
		}
		readOutput.Close()
		if len(all) != 3009 || !reflect.DeepEqual(all, want.Candidates) {
			t.Fatal("full catalog mismatch")
		}
	}
}
