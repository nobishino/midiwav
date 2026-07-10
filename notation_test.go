package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"

	"github.com/nobishino/midiwav/harmony"
)

// testChoraleReport は禁則のない2和音のコラールの分析結果を作る。
func testChoraleReport(t *testing.T) *harmony.Report {
	t.Helper()
	s := smf.New()
	chords := [][4]uint8{
		{72, 67, 64, 60}, // C5 G4 E4 C4
		{71, 67, 62, 55}, // B4 G4 D4 G3
	}
	for v := range 4 {
		var tr smf.Track
		if v == 0 {
			tr.Add(0, smf.MetaTempo(120))
			tr.Add(0, smf.MetaMeter(4, 4))
		}
		for _, c := range chords {
			tr.Add(0, midi.NoteOn(0, c[v], 100))
			tr.Add(4*960, midi.NoteOff(0, c[v]))
		}
		tr.Close(0)
		if err := s.Add(tr); err != nil {
			t.Fatal(err)
		}
	}
	key, _ := harmony.ParseKeyFromFilename("c-dur.mid")
	r, ok := harmony.Analyze(s, &key)
	if !ok {
		t.Fatal("Analyze failed")
	}
	return r
}

func fileNames(files []discordFile) []string {
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.name
	}
	return names
}

func TestNotationAttachmentsWithoutVerovio(t *testing.T) {
	r := testChoraleReport(t)
	srcPath := filepath.Join(t.TempDir(), "c-dur.mid")
	cfg := Notation{VerovioPath: "no-such-verovio-command", SVG2PNGPath: "no-such-converter"}

	files := notationAttachments(r, srcPath, cfg)
	if got := fileNames(files); len(got) != 1 || got[0] != "c-dur.musicxml" {
		t.Fatalf("files = %v, want [c-dur.musicxml]", got)
	}
	// .musicxml はディスクにも保存される
	data, err := os.ReadFile(strings.TrimSuffix(srcPath, ".mid") + ".musicxml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "<score-partwise") {
		t.Error("saved file should contain MusicXML")
	}
}

func TestNotationAttachmentsWithVerovio(t *testing.T) {
	if _, err := exec.LookPath("verovio"); err != nil {
		t.Skip("verovio not installed")
	}
	r := testChoraleReport(t)
	srcPath := filepath.Join(t.TempDir(), "c-dur.mid")

	// SVG→PNG変換が無い場合はSVGを添付する
	cfg := Notation{VerovioPath: "verovio", SVG2PNGPath: "no-such-converter"}
	files := notationAttachments(r, srcPath, cfg)
	if got := fileNames(files); len(got) != 2 || got[1] != "c-dur.svg" {
		t.Fatalf("files = %v, want [c-dur.musicxml c-dur.svg]", got)
	}

	// rsvg-convert があればPNGを添付する
	if _, err := exec.LookPath("rsvg-convert"); err != nil {
		t.Skip("rsvg-convert not installed")
	}
	cfg = Notation{VerovioPath: "verovio", SVG2PNGPath: "rsvg-convert"}
	files = notationAttachments(r, srcPath, cfg)
	if got := fileNames(files); len(got) != 2 || got[1] != "c-dur.png" {
		t.Fatalf("files = %v, want [c-dur.musicxml c-dur.png]", got)
	}
	png, err := os.ReadFile(strings.TrimSuffix(srcPath, ".mid") + ".png")
	if err != nil {
		t.Fatal(err)
	}
	if len(png) < 8 || string(png[1:4]) != "PNG" {
		t.Error("saved file should be a PNG image")
	}
}

func TestLoadConfigNotationDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("[[target]]\ndir = \"/tmp\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	config, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if config.Notation.VerovioPath != "verovio" || config.Notation.SVG2PNGPath != "rsvg-convert" {
		t.Errorf("notation defaults = %+v, want verovio / rsvg-convert", config.Notation)
	}
}
