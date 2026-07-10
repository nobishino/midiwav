package main

import (
	"os"
	"strings"
	"testing"
)

func TestBuildDiscordPost(t *testing.T) {
	newTempFile := func(t *testing.T, name string) *os.File {
		t.Helper()
		f, err := os.CreateTemp(t.TempDir(), name)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close() })
		return f
	}

	t.Run("no report", func(t *testing.T) {
		srcMIDI := newTempFile(t, "song-*.mid")
		dstWAV := newTempFile(t, "song-*.wav")
		content, files := buildDiscordPost(srcMIDI, dstWAV, "")
		if strings.Contains(content, "添削") {
			t.Errorf("content should not mention harmony report, got: %s", content)
		}
		if len(files) != 2 {
			t.Errorf("got %d files, want 2 (midi + wav)", len(files))
		}
	})

	t.Run("short report is inlined", func(t *testing.T) {
		srcMIDI := newTempFile(t, "song-*.mid")
		dstWAV := newTempFile(t, "song-*.wav")
		content, files := buildDiscordPost(srcMIDI, dstWAV, "✓ 禁則違反は検出されませんでした\n")
		if !strings.Contains(content, "✓ 禁則違反は検出されませんでした") {
			t.Errorf("content should contain the report inline, got: %s", content)
		}
		if len(files) != 2 {
			t.Errorf("got %d files, want 2 (no extra report file)", len(files))
		}
	})

	t.Run("long report is attached as a file", func(t *testing.T) {
		srcMIDI := newTempFile(t, "song-*.mid")
		dstWAV := newTempFile(t, "song-*.wav")
		longReport := strings.Repeat("あ", discordContentLimit)
		content, files := buildDiscordPost(srcMIDI, dstWAV, longReport)
		if strings.Contains(content, longReport) {
			t.Errorf("content should not inline the long report, got: %s", content)
		}
		if len(files) != 3 {
			t.Fatalf("got %d files, want 3 (midi + wav + report)", len(files))
		}
		reportFile := files[2]
		if !strings.HasSuffix(reportFile.name, "-check.txt") {
			t.Errorf("report file name = %s, want suffix -check.txt", reportFile.name)
		}
	})
}
