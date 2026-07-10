package harmony

import (
	"testing"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

func TestKeyFifths(t *testing.T) {
	tests := []struct {
		filename string
		want     int
	}{
		{"c-dur.mid", 0},
		{"g-dur.mid", 1},
		{"f-dur.mid", -1},
		{"fis-dur.mid", 6},
		{"b-dur.mid", -2}, // b はドイツ語の B♭
		{"as-dur.mid", -4},
		{"a-moll.mid", 0},
		{"e-moll.mid", 1},
		{"h-moll.mid", 2},
		{"cis-moll.mid", 4},
		{"es-moll.mid", -6},
	}
	for _, tt := range tests {
		key, ok := ParseKeyFromFilename(tt.filename)
		if !ok {
			t.Fatalf("failed to parse key from %q", tt.filename)
		}
		if got := keyFifths(key); got != tt.want {
			t.Errorf("keyFifths(%s) = %d, want %d", key, got, tt.want)
		}
	}
}

// buildTimedChoraleSMF は各和音のオンセット間隔（ティック）を指定して
// SATB 4トラックのSMFを作る。gaps[i] は i番目の和音から次の和音までの間隔。
func buildTimedChoraleSMF(t *testing.T, gaps []int64, chords [][4]uint8) *smf.SMF {
	t.Helper()
	if len(gaps) != len(chords) {
		t.Fatal("gaps and chords must have the same length")
	}
	s := smf.New()
	for v := range 4 {
		var tr smf.Track
		if v == 0 {
			tr.Add(0, smf.MetaTempo(120))
			tr.Add(0, smf.MetaMeter(4, 4))
		}
		for i, c := range chords {
			tr.Add(0, midi.NoteOn(0, c[v], 100))
			tr.Add(uint32(gaps[i]), midi.NoteOff(0, c[v]))
		}
		tr.Close(0)
		if err := s.Add(tr); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func TestAnalyzeScore(t *testing.T) {
	key, _ := ParseKeyFromFilename("es-moll.mid")
	s := buildChoraleSMF(t,
		[4]uint8{75, 70, 66, 51}, // Eb5 Bb4 Gb4 Eb3
		[4]uint8{74, 70, 65, 46}, // D5 Bb4 F4 Bb2
	)
	r, ok := Analyze(s, &key)
	if !ok {
		t.Fatal("Analyze: not recognized as a 4-voice chorale")
	}
	sc := r.Score
	if sc == nil {
		t.Fatal("Report.Score should not be nil")
	}
	if !sc.HasKey || sc.Fifths != -6 {
		t.Errorf("HasKey=%v Fifths=%d, want true, -6", sc.HasKey, sc.Fifths)
	}
	if sc.MeterNum != 4 || sc.MeterDenom != 4 {
		t.Errorf("meter = %d/%d, want 4/4", sc.MeterNum, sc.MeterDenom)
	}
	if sc.TicksPerQuarter != 960 {
		t.Errorf("TicksPerQuarter = %d, want 960", sc.TicksPerQuarter)
	}
	if len(sc.Chords) != len(r.Chords) {
		t.Fatalf("Score.Chords has %d items, want %d (same as Report.Chords)", len(sc.Chords), len(r.Chords))
	}
	// 綴り: Eb5 = {E, -1, 5}、Bb2 = {B, -1, 2}
	if got := sc.Chords[0].Notes[0]; got != (ScoreNote{Step: "E", Alter: -1, Octave: 5}) {
		t.Errorf("Notes[0] = %+v, want E flat 5", got)
	}
	if got := sc.Chords[1].Notes[3]; got != (ScoreNote{Step: "B", Alter: -1, Octave: 2}) {
		t.Errorf("Notes[3] = %+v, want B flat 2", got)
	}
}

func TestScoreDurations(t *testing.T) {
	// 4/4 (小節=3840ティック)。2分音符2つ + 小節頭の最終和音。
	s := buildTimedChoraleSMF(t,
		[]int64{1920, 1920, 960},
		[][4]uint8{
			{72, 67, 64, 60},
			{74, 71, 67, 55},
			{72, 67, 64, 48},
		},
	)
	r, ok := Analyze(s, nil)
	if !ok {
		t.Fatal("Analyze: not recognized as a 4-voice chorale")
	}
	sc := r.Score
	if sc.HasKey {
		t.Error("HasKey should be false without key")
	}
	wantOnsets := []int64{0, 1920, 3840}
	wantDurs := []int64{1920, 1920, 3840} // 最終和音は小節末まで（小節頭なので全小節分）
	if len(sc.Chords) != 3 {
		t.Fatalf("got %d chords, want 3", len(sc.Chords))
	}
	for i, c := range sc.Chords {
		if c.OnsetTicks != wantOnsets[i] || c.DurationTicks != wantDurs[i] {
			t.Errorf("chord %d: onset=%d dur=%d, want onset=%d dur=%d",
				i, c.OnsetTicks, c.DurationTicks, wantOnsets[i], wantDurs[i])
		}
	}
}

func TestScoreLastChordMidMeasure(t *testing.T) {
	// 最終和音が小節の途中（3拍目）から始まる場合、小節末までの長さになる。
	s := buildTimedChoraleSMF(t,
		[]int64{1920, 960},
		[][4]uint8{
			{72, 67, 64, 60},
			{74, 71, 67, 55},
		},
	)
	r, ok := Analyze(s, nil)
	if !ok {
		t.Fatal("Analyze: not recognized as a 4-voice chorale")
	}
	last := r.Score.Chords[1]
	if last.OnsetTicks != 1920 || last.DurationTicks != 1920 {
		t.Errorf("last chord: onset=%d dur=%d, want onset=1920 dur=1920", last.OnsetTicks, last.DurationTicks)
	}
}
