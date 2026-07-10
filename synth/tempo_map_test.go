package synth

import (
	"math"
	"testing"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

// newTestSMF は与えたトラックからSMF（960 ticks/四分音符）を作る。
func newTestSMF(t *testing.T, tracks ...smf.Track) *smf.SMF {
	t.Helper()
	s := smf.New()
	for _, tr := range tracks {
		tr.Close(0)
		if err := s.Add(tr); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func TestTempoMap(t *testing.T) {
	almostEqual := func(t *testing.T, got, want float64) {
		t.Helper()
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("got %f, want %f", got, want)
		}
	}

	t.Run("no tempo event defaults to 120 BPM", func(t *testing.T) {
		var tr smf.Track
		tr.Add(0, midi.NoteOn(0, 60, 100))
		tr.Add(960, midi.NoteOff(0, 60))
		s := newTestSMF(t, tr)
		tm := newTempoMap(s, s.TimeFormat.(smf.MetricTicks))
		almostEqual(t, tm.seconds(960), 0.5) // 120BPMの四分音符 = 0.5秒
	})

	t.Run("tempo at tick 0 overrides the default", func(t *testing.T) {
		var tr smf.Track
		tr.Add(0, smf.MetaTempo(60))
		tr.Add(0, midi.NoteOn(0, 60, 100))
		tr.Add(960, midi.NoteOff(0, 60))
		s := newTestSMF(t, tr)
		tm := newTempoMap(s, s.TimeFormat.(smf.MetricTicks))
		almostEqual(t, tm.seconds(960), 1.0)
	})

	t.Run("mid-piece tempo change", func(t *testing.T) {
		var tr smf.Track
		tr.Add(0, smf.MetaTempo(120))
		tr.Add(960, smf.MetaTempo(240))
		s := newTestSMF(t, tr)
		tm := newTempoMap(s, s.TimeFormat.(smf.MetricTicks))
		almostEqual(t, tm.seconds(480), 0.25)
		almostEqual(t, tm.seconds(960), 0.5)
		almostEqual(t, tm.seconds(1920), 0.75) // 変化後は240BPM: 四分音符 = 0.25秒
	})

	t.Run("tempo events in track 0 apply to all tracks", func(t *testing.T) {
		// テンポマップはトラック横断で共有される（tickは絶対値なのでトラックを問わない）
		var tempoTrack smf.Track
		tempoTrack.Add(0, smf.MetaTempo(120))
		tempoTrack.Add(960, smf.MetaTempo(240))
		var noteTrack smf.Track
		noteTrack.Add(0, midi.NoteOn(0, 48, 100))
		noteTrack.Add(1920, midi.NoteOff(0, 48))
		s := newTestSMF(t, tempoTrack, noteTrack)
		tm := newTempoMap(s, s.TimeFormat.(smf.MetricTicks))
		almostEqual(t, tm.seconds(1920), 0.75)
	})
}

// テンポイベントより先にNoteOnが来ても、0除算（Inf/NaNや無限ループ）に
// ならないことのリグレッションテスト。
func TestSMFToPCMArrayNoteBeforeTempo(t *testing.T) {
	var tr smf.Track
	tr.Add(0, midi.NoteOn(0, 60, 100))
	tr.Add(480, smf.MetaTempo(240))
	tr.Add(480, midi.NoteOff(0, 60))
	s := newTestSMF(t, tr)

	samples, err := smfToPCMArray(s)
	if err != nil {
		t.Fatal(err)
	}
	// 0〜480tickは120BPM（0.25秒）、480〜960tickは240BPM（0.125秒）で計0.375秒
	want := int(math.Ceil(0.375 * sampleRate))
	if len(samples) != want {
		t.Errorf("got %d samples, want %d", len(samples), want)
	}
}
