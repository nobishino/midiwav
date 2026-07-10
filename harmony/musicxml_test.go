package harmony

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/gomidi/midi/v2/smf"
)

func TestNoteTypeAndDots(t *testing.T) {
	const tpq = 960
	tests := []struct {
		dur      int64
		wantType string
		wantDots int
	}{
		{8 * tpq, "breve", 0},
		{4 * tpq, "whole", 0},
		{6 * tpq, "whole", 1},
		{2 * tpq, "half", 0},
		{3 * tpq, "half", 1},
		{tpq, "quarter", 0},
		{tpq / 2, "eighth", 0},
		{tpq / 4, "16th", 0},
		{tpq/4 + 1, "", 0}, // 不規則な音価は type なし
	}
	for _, tt := range tests {
		typ, dots := noteTypeAndDots(tt.dur, tpq)
		if typ != tt.wantType || dots != tt.wantDots {
			t.Errorf("noteTypeAndDots(%d) = %q, %d, want %q, %d", tt.dur, typ, dots, tt.wantType, tt.wantDots)
		}
	}
}

func TestSplitAtMeasures(t *testing.T) {
	// 小節=3840ティック。2880から1920の音は 960+960 に分かれてタイで結ばれる。
	segs := splitAtMeasures(0, 2880, 1920, 3840)
	if len(segs) != 2 {
		t.Fatalf("got %d segments, want 2", len(segs))
	}
	first, second := segs[0], segs[1]
	if first.start != 2880 || first.dur != 960 || !first.tieStart || first.tieStop {
		t.Errorf("first segment = %+v, want start=2880 dur=960 tieStart", first)
	}
	if second.start != 3840 || second.dur != 960 || second.tieStart || !second.tieStop {
		t.Errorf("second segment = %+v, want start=3840 dur=960 tieStop", second)
	}

	// 休符（chordIdx=-1）はタイで結ばない
	rests := splitAtMeasures(-1, 0, 7680, 3840)
	if len(rests) != 2 {
		t.Fatalf("got %d rest segments, want 2", len(rests))
	}
	for _, s := range rests {
		if s.tieStart || s.tieStop {
			t.Errorf("rest segment should not be tied: %+v", s)
		}
	}
}

func TestMusicXMLBasic(t *testing.T) {
	key, _ := ParseKeyFromFilename("c-dur.mid")
	s := buildChoraleSMF(t,
		[4]uint8{72, 67, 64, 60}, // I
		[4]uint8{74, 71, 67, 55}, // V
	)
	r, ok := Analyze(s, &key)
	if !ok {
		t.Fatal("Analyze failed")
	}
	data, err := r.MusicXML()
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)

	// well-formed であること
	var doc mxlScorePartwise
	if err := xml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("output is not well-formed XML: %v", err)
	}
	if len(doc.Parts) != 1 || len(doc.Parts[0].Measures) != 2 {
		t.Errorf("got %d parts / %d measures, want 1 part / 2 measures",
			len(doc.Parts), len(doc.Parts[0].Measures))
	}

	for _, want := range []string{
		"<fifths>0</fifths>",
		"<beats>4</beats>",
		"<beat-type>4</beat-type>",
		"<divisions>960</divisions>",
		"<sign>G</sign>",
		"<sign>F</sign>",
		"<step>C</step>",
		"<type>whole</type>",
		"<words>I(基本位置)</words>",
		"<words>V(基本位置)</words>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output should contain %s, got:\n%s", want, got)
		}
	}
	// 4声なので各小節に backup が3つ
	if n := strings.Count(got, "<backup>"); n != 6 {
		t.Errorf("got %d backups, want 6 (3 per measure)", n)
	}
}

func TestMusicXMLAccidentalAndNoKey(t *testing.T) {
	// 調なし: fifths を出さず、綴りはフラット。Eb4 は alter -1
	s := buildChoraleSMF(t,
		[4]uint8{75, 67, 63, 60}, // Eb5 G4 Eb4 C4
	)
	r, ok := Analyze(s, nil)
	if !ok {
		t.Fatal("Analyze failed")
	}
	data, err := r.MusicXML()
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if strings.Contains(got, "<fifths>") {
		t.Error("output without key should not contain <fifths>")
	}
	if !strings.Contains(got, "<alter>-1</alter>") {
		t.Errorf("output should contain alter -1 for Eb, got:\n%s", got)
	}
	if strings.Contains(got, "<words>") {
		t.Error("output without key should not contain chord symbols")
	}
}

func TestMusicXMLTieAcrossBarline(t *testing.T) {
	// 4/4（小節=3840ティック）。2つ目の和音は 2880 から 1920 の長さで
	// 小節線をまたぐため、960+960 に分割されタイで結ばれる。
	s := buildTimedChoraleSMF(t,
		[]int64{2880, 1920, 1920},
		[][4]uint8{
			{72, 67, 64, 60},
			{74, 71, 67, 55},
			{72, 67, 64, 48},
		},
	)
	r, ok := Analyze(s, nil)
	if !ok {
		t.Fatal("Analyze failed")
	}
	data, err := r.MusicXML()
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, `<tie type="start">`) || !strings.Contains(got, `<tied type="start">`) {
		t.Errorf("output should contain tie start, got:\n%s", got)
	}
	if !strings.Contains(got, `<tie type="stop">`) || !strings.Contains(got, `<tied type="stop">`) {
		t.Errorf("output should contain tie stop, got:\n%s", got)
	}
}

// TestMusicXMLGolden は testdata/*.mid のMusicXML出力をゴールデンファイル
// （<case>.musicxml）と照合する。go test -update で生成・更新できる。
func TestMusicXMLGolden(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("testdata", "*.mid"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no MIDI files found in testdata")
	}
	for _, path := range paths {
		name := strings.TrimSuffix(filepath.Base(path), ".mid")
		t.Run(name, func(t *testing.T) {
			smfData, err := smf.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read SMF: %v", err)
			}
			var key *Key
			if k, ok := ParseKeyFromFilename(path); ok {
				key = &k
			}
			r, ok := Analyze(smfData, key)
			if !ok {
				t.Fatal("Analyze: not recognized as a 4-voice chorale")
			}
			data, err := r.MusicXML()
			if err != nil {
				t.Fatal(err)
			}

			goldenPath := strings.TrimSuffix(path, ".mid") + ".musicxml"
			if *update {
				if err := os.WriteFile(goldenPath, data, 0644); err != nil {
					t.Fatalf("failed to update golden file: %v", err)
				}
				t.Logf("updated golden file: %s", goldenPath)
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("failed to read golden file: %v", err)
			}
			if string(data) != string(want) {
				t.Errorf("MusicXML differs from golden file %s\n--- got ---\n%s", goldenPath, data)
			}
		})
	}
}
