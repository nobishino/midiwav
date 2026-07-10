package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunCheckTextMatchesGolden は check サブコマンドの text 出力が
// harmony パッケージのゴールデンファイルと一致することを確認する。
// 終了コードはゴールデンに ⚠ が含まれるかどうかから決まる。
func TestRunCheckTextMatchesGolden(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("harmony", "testdata", "*.mid"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no MIDI files found in harmony/testdata")
	}
	for _, path := range paths {
		name := strings.TrimSuffix(filepath.Base(path), ".mid")
		t.Run(name, func(t *testing.T) {
			golden, err := os.ReadFile(strings.TrimSuffix(path, ".mid") + ".golden.txt")
			if err != nil {
				t.Fatal(err)
			}
			wantCode := 0
			if strings.Contains(string(golden), "⚠") {
				wantCode = 1
			}

			var stdout, stderr bytes.Buffer
			code := runCheck([]string{path}, &stdout, &stderr)
			if code != wantCode {
				t.Errorf("exit code = %d, want %d (stderr: %s)", code, wantCode, stderr.String())
			}
			if stdout.String() != string(golden) {
				t.Errorf("output differs from golden\n--- got ---\n%s--- want ---\n%s", stdout.String(), golden)
			}
		})
	}
}

func TestRunCheckJSON(t *testing.T) {
	path := filepath.Join("harmony", "testdata", "aug-second_a-moll.mid")
	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"-format", "json", path}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("exit code = %d, want 1 (stderr: %s)", code, stderr.String())
	}

	var results []checkResult
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, stdout.String())
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	r := results[0]
	if r.File != path {
		t.Errorf("file = %q, want %q", r.File, path)
	}
	if r.Key == nil || *r.Key != "A-moll" {
		t.Errorf("key = %v, want A-moll", r.Key)
	}
	if len(r.Chords) == 0 {
		t.Error("chords is empty")
	}
	for _, c := range r.Chords {
		if c.Symbol == "" {
			t.Errorf("chord %s: symbol is empty (key is known)", c.Pos)
		}
		if c.Notes.S == "" || c.Notes.B == "" {
			t.Errorf("chord %s: notes are incomplete: %+v", c.Pos, c.Notes)
		}
	}
	warns := 0
	for _, is := range r.Issues {
		switch is.Level {
		case "warning":
			warns++
		case "info":
		default:
			t.Errorf("unknown issue level %q", is.Level)
		}
	}
	if warns == 0 {
		t.Error("want at least 1 warning (増2度)")
	}
	if r.Summary.Warnings != warns {
		t.Errorf("summary.warnings = %d, want %d", r.Summary.Warnings, warns)
	}
}

// TestRunCheckKeyOverride は -key が調不明のファイル名より優先されることを確認する。
func TestRunCheckKeyOverride(t *testing.T) {
	path := filepath.Join("harmony", "testdata", "nokey.mid")

	var without bytes.Buffer
	if code := runCheck([]string{path}, &without, &bytes.Buffer{}); code == 2 {
		t.Fatal("unexpected error without -key")
	}
	if !strings.Contains(without.String(), "調を読み取れず") {
		t.Errorf("output without -key should note missing key:\n%s", without.String())
	}

	var with bytes.Buffer
	if code := runCheck([]string{"-key", "C-dur", path}, &with, &bytes.Buffer{}); code == 2 {
		t.Fatal("unexpected error with -key")
	}
	if !strings.Contains(with.String(), "調: C-dur") {
		t.Errorf("output with -key should report C-dur:\n%s", with.String())
	}
}

func TestRunCheckErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no args", nil},
		{"missing file", []string{"no-such-file.mid"}},
		{"bad format", []string{"-format", "xml", filepath.Join("harmony", "testdata", "nokey.mid")}},
		{"bad key", []string{"-key", "x-moll", filepath.Join("harmony", "testdata", "nokey.mid")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := runCheck(tt.args, &stdout, &stderr); code != 2 {
				t.Errorf("exit code = %d, want 2", code)
			}
		})
	}
}
