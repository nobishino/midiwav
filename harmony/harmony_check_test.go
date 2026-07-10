package harmony

import (
	"strings"
	"testing"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

func TestParseKeyFromFilename(t *testing.T) {
	tests := []struct {
		path string
		want string
		ok   bool
	}{
		{"es-moll.mid", "Eb-moll", true},
		{"As-dur.mid", "Ab-dur", true},
		{"/path/to/課題1_h-moll.mid", "B-moll", true},
		{"fis_dur.mid", "F#-dur", true},
		{"b-dur.mid", "Bb-dur", true},
		{"C-Dur.mid", "C-dur", true},
		{"nokey.mid", "", false},
		{"melody.mid", "", false},
	}
	for _, tt := range tests {
		key, ok := ParseKeyFromFilename(tt.path)
		if ok != tt.ok {
			t.Errorf("ParseKeyFromFilename(%q) ok = %v, want %v", tt.path, ok, tt.ok)
			continue
		}
		if ok && key.String() != tt.want {
			t.Errorf("ParseKeyFromFilename(%q) = %s, want %s", tt.path, key, tt.want)
		}
	}
}

func TestBuildSpellingTable(t *testing.T) {
	mustKey := func(name string) Key {
		t.Helper()
		key, ok := ParseKeyFromFilename(name)
		if !ok {
			t.Fatalf("failed to parse key from %q", name)
		}
		return key
	}

	// C-dur: 半音階音は下の音階音の上方変位（C# D# F# G# A#）
	cdur := buildSpellingTable(mustKey("c-dur.mid"))
	for pc, want := range map[int]noteSpelling{
		0: {0, 0}, 1: {0, 1}, 3: {1, 1}, 4: {2, 0}, 6: {3, 1}, 10: {5, 1}, 11: {6, 0},
	} {
		if cdur[pc] != want {
			t.Errorf("C-dur table[%d] = %v, want %v", pc, cdur[pc], want)
		}
	}

	// es-moll: 音階音は Eb F Gb Ab Bb Cb Db
	esmoll := buildSpellingTable(mustKey("es-moll.mid"))
	for pc, want := range map[int]noteSpelling{
		3: {2, -1}, 5: {3, 0}, 6: {4, -1}, 8: {5, -1}, 10: {6, -1}, 11: {0, -1}, 1: {1, -1},
	} {
		if esmoll[pc] != want {
			t.Errorf("es-moll table[%d] = %v, want %v", pc, esmoll[pc], want)
		}
	}
}

func TestNoteName(t *testing.T) {
	if got := noteName(60, flatSpellingTable); got != "C4" {
		t.Errorf("noteName(60) = %s, want C4", got)
	}
	if got := noteName(61, flatSpellingTable); got != "Db4" {
		t.Errorf("noteName(61) = %s, want Db4", got)
	}
	key, _ := ParseKeyFromFilename("c-dur.mid")
	if got := noteName(61, buildSpellingTable(key)); got != "C#4" {
		t.Errorf("noteName(61, C-dur) = %s, want C#4", got)
	}
}

func TestIntervalQuality(t *testing.T) {
	key, _ := ParseKeyFromFilename("c-dur.mid")
	table := buildSpellingTable(key)
	tests := []struct {
		n1, n2   uint8
		wantSize int
		wantName string
	}{
		{65, 71, 4, "増"},  // F4 → B4: 増4度
		{71, 77, 5, "減"},  // B4 → F5: 減5度
		{60, 64, 3, "長"},  // C4 → E4: 長3度
		{64, 60, 3, "長"},  // 下行でも同じ
		{60, 67, 5, "完全"}, // C4 → G4: 完全5度
	}
	for _, tt := range tests {
		size, _, name := intervalQuality(tt.n1, tt.n2, table)
		if size != tt.wantSize || name != tt.wantName {
			t.Errorf("intervalQuality(%d, %d) = %d, %s, want %d, %s",
				tt.n1, tt.n2, size, name, tt.wantSize, tt.wantName)
		}
	}
	// a-moll における F4 → G#4 は増2度
	amoll, _ := ParseKeyFromFilename("a-moll.mid")
	size, dev, name := intervalQuality(65, 68, buildSpellingTable(amoll))
	if size != 2 || dev != 1 || name != "増" {
		t.Errorf("intervalQuality(F4, G#4, a-moll) = %d, %d, %s, want 2, 1, 増", size, dev, name)
	}
}

// buildChoraleSMF は各和音が全音符で並ぶSATB 4トラックのSMFを作る。
func buildChoraleSMF(t *testing.T, chords ...[4]uint8) *smf.SMF {
	t.Helper()
	s := smf.New()
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
	return s
}

func TestExtractChorale(t *testing.T) {
	s := buildChoraleSMF(t,
		[4]uint8{72, 67, 64, 60},
		[4]uint8{71, 67, 62, 55},
	)
	ch, ok := extractChorale(s)
	if !ok {
		t.Fatal("extractChorale: not recognized as a 4-voice chorale")
	}
	if len(ch.chords) != 2 {
		t.Fatalf("got %d chords, want 2", len(ch.chords))
	}
	if ch.chords[1].beat != 4 {
		t.Errorf("second chord beat = %f, want 4", ch.chords[1].beat)
	}
	if ch.pos(ch.chords[1].beat) != "2小節1拍目" {
		t.Errorf("pos = %s, want 2小節1拍目", ch.pos(ch.chords[1].beat))
	}

	// トラックが3本しかない場合は4声体とみなさない
	s3 := smf.New()
	for v := range 3 {
		var tr smf.Track
		tr.Add(0, midi.NoteOn(0, uint8(60+v), 100))
		tr.Add(960, midi.NoteOff(0, uint8(60+v)))
		tr.Close(0)
		if err := s3.Add(tr); err != nil {
			t.Fatal(err)
		}
	}
	if _, ok := extractChorale(s3); ok {
		t.Error("extractChorale: 3-track SMF should not be recognized as a chorale")
	}
}

func TestHarmonyReportClean(t *testing.T) {
	// C-dur: I → V。禁則なし。
	s := buildChoraleSMF(t,
		[4]uint8{72, 67, 64, 60}, // C5 G4 E4 C4
		[4]uint8{71, 67, 62, 55}, // B4 G4 D4 G3
	)
	report, ok := harmonyReport(s, "c-dur.mid")
	if !ok {
		t.Fatal("harmonyReport: not recognized as a 4-voice chorale")
	}
	if !strings.Contains(report, "調: C-dur") {
		t.Errorf("report should contain key, got:\n%s", report)
	}
	if !strings.Contains(report, "✓ 禁則違反は検出されませんでした") {
		t.Errorf("report should be clean, got:\n%s", report)
	}
}

func TestHarmonyReportParallelFifth(t *testing.T) {
	// S-A 間の連続5度: C5/F4 → D5/G4
	s := buildChoraleSMF(t,
		[4]uint8{72, 65, 57, 53},
		[4]uint8{74, 67, 59, 55},
	)
	report, ok := harmonyReport(s, "c-dur.mid")
	if !ok {
		t.Fatal("harmonyReport: not recognized as a 4-voice chorale")
	}
	if !strings.Contains(report, "[連続5度] 1小節1拍目→2小節1拍目 S-A") {
		t.Errorf("report should contain parallel fifth S-A, got:\n%s", report)
	}
	if strings.Contains(report, "✓") {
		t.Errorf("report should not be clean, got:\n%s", report)
	}
}

func TestHarmonyReportCrossingAndSpacing(t *testing.T) {
	// 1和音目: A(65) が S(64) より高い → 交差。2和音目: S-A が 8度超。
	s := buildChoraleSMF(t,
		[4]uint8{64, 65, 57, 48},
		[4]uint8{76, 62, 55, 48},
	)
	report, ok := harmonyReport(s, "c-dur.mid")
	if !ok {
		t.Fatal("harmonyReport: not recognized as a 4-voice chorale")
	}
	if !strings.Contains(report, "[交差] 1小節1拍目") {
		t.Errorf("report should contain voice crossing, got:\n%s", report)
	}
	if !strings.Contains(report, "[間隔] 2小節1拍目: S-A が 14 半音") {
		t.Errorf("report should contain spacing violation, got:\n%s", report)
	}
}

func TestCheckSpelledAugmentedAndCrossRelation(t *testing.T) {
	amoll, _ := ParseKeyFromFilename("a-moll.mid")
	table := buildSpellingTable(amoll)

	// A(声部2番目)が F4 → G#4 と増2度で進行
	aug := &chorale{
		chords: []chord{
			{beat: 0, notes: [4]uint8{72, 65, 60, 45}},
			{beat: 4, notes: [4]uint8{71, 68, 59, 52}},
		},
		beatsPerMeasure: 4, beatUnit: 1,
	}
	issues := checkSpelled(aug, table)
	found := false
	for _, is := range issues {
		if is.Warn && strings.Contains(is.Msg, "[増音程]") && strings.Contains(is.Msg, "F4 ↑ G#4（増2度）") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected augmented second issue, got: %+v", issues)
	}

	// 対斜: T の G3 → 次の和音で S に G#4
	cross := &chorale{
		chords: []chord{
			{beat: 0, notes: [4]uint8{71, 64, 55, 52}},
			{beat: 4, notes: [4]uint8{68, 64, 57, 52}},
		},
		beatsPerMeasure: 4, beatUnit: 1,
	}
	issues = checkSpelled(cross, table)
	found = false
	for _, is := range issues {
		if !is.Warn && strings.Contains(is.Msg, "[対斜]") && strings.Contains(is.Msg, "T の G3 と S の G#4") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected cross relation issue, got: %+v", issues)
	}
}
