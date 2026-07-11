package harmony

// MusicXML (score-partwise) の生成。
//
// 大譜表2段のコラール定番レイアウトで出力する: 上段（ト音記号）に
// S（声部1・符幹上）とA（声部2・符幹下）、下段（ヘ音記号）に
// T（声部3・符幹上）とB（声部4・符幹下）。和音記号は下段の下に
// <direction><words> として付ける。
//
// 音価はScoreの近似（次のオンセットまで）に従う。小節線をまたぐ音は
// 分割してタイで結ぶ。曲頭に空白がある場合は休符で埋める。

import (
	"encoding/xml"
	"fmt"
)

type mxlScorePartwise struct {
	XMLName  xml.Name    `xml:"score-partwise"`
	Version  string      `xml:"version,attr"`
	PartList mxlPartList `xml:"part-list"`
	Parts    []mxlPart   `xml:"part"`
}

type mxlPartList struct {
	ScoreParts []mxlScorePart `xml:"score-part"`
}

type mxlScorePart struct {
	ID   string      `xml:"id,attr"`
	Name mxlPartName `xml:"part-name"`
}

// mxlPartName はパート名。part-name はMusicXMLの必須要素だが、楽譜への
// 表示は不要なため print-object="no" で抑制する（#48）。
type mxlPartName struct {
	PrintObject string `xml:"print-object,attr"`
	Value       string `xml:",chardata"`
}

type mxlPart struct {
	ID       string       `xml:"id,attr"`
	Measures []mxlMeasure `xml:"measure"`
}

type mxlMeasure struct {
	Number int   `xml:"number,attr"`
	Items  []any // mxlAttributes / mxlDirection / mxlNote / mxlBackup の列
}

type mxlAttributes struct {
	XMLName   xml.Name  `xml:"attributes"`
	Divisions int       `xml:"divisions"`
	Key       *mxlKey   `xml:"key"`
	Time      *mxlTime  `xml:"time"`
	Staves    int       `xml:"staves"`
	Clefs     []mxlClef `xml:"clef"`
}

type mxlKey struct {
	Fifths int `xml:"fifths"`
}

type mxlTime struct {
	Beats    int `xml:"beats"`
	BeatType int `xml:"beat-type"`
}

type mxlClef struct {
	Number int    `xml:"number,attr"`
	Sign   string `xml:"sign"`
	Line   int    `xml:"line"`
}

type mxlDirection struct {
	XMLName   xml.Name `xml:"direction"`
	Placement string   `xml:"placement,attr"`
	Words     mxlWords `xml:"direction-type>words"`
	Staff     int      `xml:"staff"`
}

// mxlWords は direction の文字列。default-y で縦位置を固定する
// （指定しないとVerovioの自動配置によりバス音符の高さに追従してしまう）。
// font-style を指定しないとVerovioは斜体で描画する。
type mxlWords struct {
	DefaultY  int    `xml:"default-y,attr"`
	FontStyle string `xml:"font-style,attr"`
	Value     string `xml:",chardata"`
}

// chordNameY はコードネームの縦位置（ヘ音記号譜の上第1線基準・tenths単位）。
// バスの最低音 F2 の全音符と重ならない値（#56）。
const chordNameY = -100

// mxlNote の要素順はMusicXMLのDTDに従う:
// rest/pitch, duration, tie, voice, type, dot, accidental, stem, staff, notations
type mxlNote struct {
	XMLName    xml.Name      `xml:"note"`
	Rest       *struct{}     `xml:"rest"`
	Pitch      *mxlPitch     `xml:"pitch"`
	Duration   int64         `xml:"duration"`
	Ties       []mxlTie      `xml:"tie"`
	Voice      int           `xml:"voice"`
	Type       string        `xml:"type,omitempty"`
	Dots       []struct{}    `xml:"dot"`
	Accidental string        `xml:"accidental,omitempty"`
	Stem       string        `xml:"stem,omitempty"`
	Staff      int           `xml:"staff"`
	Notations  *mxlNotations `xml:"notations"`
}

type mxlPitch struct {
	Step   string `xml:"step"`
	Alter  *int   `xml:"alter"`
	Octave int    `xml:"octave"`
}

type mxlTie struct {
	Type string `xml:"type,attr"`
}

type mxlNotations struct {
	Tied []mxlTied `xml:"tied"`
}

type mxlTied struct {
	Type string `xml:"type,attr"`
}

type mxlBackup struct {
	XMLName  xml.Name `xml:"backup"`
	Duration int64    `xml:"duration"`
}

// segment は小節線で分割した音（またはタイの一部・休符）。
type segment struct {
	chordIdx          int // 対応する Score.Chords の添字。休符は -1
	start, dur        int64
	tieStart, tieStop bool
}

// splitAtMeasures は [start, start+dur) を小節境界で分割する。
func splitAtMeasures(chordIdx int, start, dur, measureTicks int64) []segment {
	var segs []segment
	for dur > 0 {
		remain := measureTicks - start%measureTicks
		d := min(dur, remain)
		segs = append(segs, segment{chordIdx: chordIdx, start: start, dur: d})
		start += d
		dur -= d
	}
	// タイは音符（休符でない）が2つ以上に分かれた場合のみ
	if chordIdx >= 0 && len(segs) > 1 {
		for i := range segs {
			if i > 0 {
				segs[i].tieStop = true
			}
			if i < len(segs)-1 {
				segs[i].tieStart = true
			}
		}
	}
	return segs
}

// noteTypeAndDots は音価（ティック）から音符の種類と付点の数を返す。
// 対応しない音価の場合は空文字列を返す（duration のみで表現する）。
func noteTypeAndDots(dur, tpq int64) (string, int) {
	bases := []struct {
		name  string
		ticks int64
	}{
		{"breve", 8 * tpq},
		{"whole", 4 * tpq},
		{"half", 2 * tpq},
		{"quarter", tpq},
		{"eighth", tpq / 2},
		{"16th", tpq / 4},
	}
	for _, b := range bases {
		if dur == b.ticks {
			return b.name, 0
		}
		if dur == b.ticks*3/2 {
			return b.name, 1
		}
	}
	return "", 0
}

var accidentalNames = map[int]string{
	-2: "flat-flat", -1: "flat", 0: "natural", 1: "sharp", 2: "double-sharp",
}

// keyAlters は調号が各幹音に与える変化を返す（例: fifths=-6 なら B,E,A,D,G,C が -1）。
func keyAlters(fifths int) map[string]int {
	alters := make(map[string]int, 7)
	sharps := []string{"F", "C", "G", "D", "A", "E", "B"}
	for i := 0; i < fifths && i < 7; i++ {
		alters[sharps[i]] = 1
	}
	flats := []string{"B", "E", "A", "D", "G", "C", "F"}
	for i := 0; i < -fifths && i < 7; i++ {
		alters[flats[i]] = -1
	}
	return alters
}

// noteLine は臨時記号の適用範囲（同一小節内の同じ幹音・同じオクターブ）を表すキー。
type noteLine struct {
	step   string
	octave int
}

// computeAccidentals は1小節分のセグメント列に対して、表示すべき臨時記号を
// (セグメント添字, 声部) → 記号名 のマップで返す。臨時記号は五線ごとに
// 「調号＋その小節内で確定した変化」と食い違う音に付け、同じ幹音・オクターブ
// にはその小節内で再度付けない。タイの後半には付けない（前の小節から持続する）。
// セグメント列は時間順である前提（同時刻の音は声部番号順に処理する）。
func computeAccidentals(msegs []segment, sc *Score, base map[string]int) map[[2]int]string {
	state := [2]map[noteLine]int{make(map[noteLine]int), make(map[noteLine]int)}
	result := make(map[[2]int]string)
	for si, s := range msegs {
		if s.chordIdx < 0 {
			continue
		}
		for voice := 1; voice <= 4; voice++ {
			n := sc.Chords[s.chordIdx].Notes[voice-1]
			staffIdx := 0
			if voice >= 3 {
				staffIdx = 1
			}
			line := noteLine{step: n.Step, octave: n.Octave}
			expected, seen := state[staffIdx][line]
			if !seen {
				expected = base[n.Step]
			}
			if n.Alter == expected || s.tieStop {
				continue
			}
			result[[2]int{si, voice}] = accidentalNames[n.Alter]
			state[staffIdx][line] = n.Alter
		}
	}
	return result
}

// 声部（1=S, 2=A, 3=T, 4=B）の五線と符幹の向き。
func voiceStaffStem(voice int) (staff int, stem string) {
	staff = 1
	if voice >= 3 {
		staff = 2
	}
	stem = "up"
	if voice%2 == 0 {
		stem = "down"
	}
	return staff, stem
}

// MusicXML は分析結果を MusicXML (score-partwise) に変換する。
func (r *Report) MusicXML() ([]byte, error) {
	sc := r.Score
	if sc == nil || len(sc.Chords) == 0 {
		return nil, fmt.Errorf("no score data")
	}
	tpq := int64(sc.TicksPerQuarter)
	measureTicks := tpq * 4 * int64(sc.MeterNum) / int64(sc.MeterDenom)
	if measureTicks <= 0 {
		return nil, fmt.Errorf("invalid meter %d/%d", sc.MeterNum, sc.MeterDenom)
	}

	// 曲頭の空白（弱起でない前提だが、位置ずれを防ぐため休符で埋める）と
	// 各和音を小節境界で分割してタイムラインを作る。
	var segs []segment
	if first := sc.Chords[0].OnsetTicks; first > 0 {
		segs = append(segs, splitAtMeasures(-1, 0, first, measureTicks)...)
	}
	for i, c := range sc.Chords {
		segs = append(segs, splitAtMeasures(i, c.OnsetTicks, c.DurationTicks, measureTicks)...)
	}

	// 小節ごとにまとめる
	numMeasures := int((segs[len(segs)-1].start + segs[len(segs)-1].dur + measureTicks - 1) / measureTicks)
	byMeasure := make([][]segment, numMeasures)
	for _, s := range segs {
		m := int(s.start / measureTicks)
		byMeasure[m] = append(byMeasure[m], s)
	}

	baseAlters := keyAlters(sc.Fifths)
	var measures []mxlMeasure
	for m, msegs := range byMeasure {
		measure := mxlMeasure{Number: m + 1}
		accidentals := computeAccidentals(msegs, sc, baseAlters)
		if m == 0 {
			attrs := &mxlAttributes{
				Divisions: sc.TicksPerQuarter,
				Time:      &mxlTime{Beats: sc.MeterNum, BeatType: sc.MeterDenom},
				Staves:    2,
				Clefs: []mxlClef{
					{Number: 1, Sign: "G", Line: 2},
					{Number: 2, Sign: "F", Line: 4},
				},
			}
			if sc.HasKey {
				attrs.Key = &mxlKey{Fifths: sc.Fifths}
			}
			measure.Items = append(measure.Items, attrs)
		}
		var measureLen int64
		for _, s := range msegs {
			measureLen += s.dur
		}
		for voice := 1; voice <= 4; voice++ {
			if voice > 1 {
				measure.Items = append(measure.Items, &mxlBackup{Duration: measureLen})
			}
			staff, stem := voiceStaffStem(voice)
			for si, s := range msegs {
				// コードネームは最初の声部の、和音が始まる位置にだけ付ける（#56）。
				// 調に依存せず判定できるため、調が不明でも表示する
				if voice == 1 && s.chordIdx >= 0 && !s.tieStop {
					measure.Items = append(measure.Items, &mxlDirection{
						Placement: "below",
						Words:     mxlWords{DefaultY: chordNameY, FontStyle: "normal", Value: r.Chords[s.chordIdx].Name},
						Staff:     2,
					})
				}
				note := &mxlNote{Duration: s.dur, Voice: voice, Staff: staff}
				if s.chordIdx < 0 {
					note.Rest = &struct{}{}
				} else {
					n := sc.Chords[s.chordIdx].Notes[voice-1]
					pitch := &mxlPitch{Step: n.Step, Octave: n.Octave}
					if n.Alter != 0 {
						alter := n.Alter
						pitch.Alter = &alter
					}
					note.Pitch = pitch
					note.Accidental = accidentals[[2]int{si, voice}]
					note.Stem = stem
					if s.tieStart || s.tieStop {
						var notations mxlNotations
						if s.tieStop {
							note.Ties = append(note.Ties, mxlTie{Type: "stop"})
							notations.Tied = append(notations.Tied, mxlTied{Type: "stop"})
						}
						if s.tieStart {
							note.Ties = append(note.Ties, mxlTie{Type: "start"})
							notations.Tied = append(notations.Tied, mxlTied{Type: "start"})
						}
						note.Notations = &notations
					}
				}
				if typ, dots := noteTypeAndDots(s.dur, tpq); typ != "" {
					note.Type = typ
					note.Dots = make([]struct{}, dots)
				}
				measure.Items = append(measure.Items, note)
			}
		}
		measures = append(measures, measure)
	}

	doc := mxlScorePartwise{
		Version: "4.0",
		PartList: mxlPartList{ScoreParts: []mxlScorePart{
			{ID: "P1", Name: mxlPartName{PrintObject: "no", Value: "SATB"}},
		}},
		Parts: []mxlPart{{ID: "P1", Measures: measures}},
	}
	body, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	header := xml.Header +
		`<!DOCTYPE score-partwise PUBLIC "-//Recordare//DTD MusicXML 4.0 Partwise//EN" "http://www.musicxml.org/dtds/partwise.dtd">` + "\n"
	return append([]byte(header), append(body, '\n')...), nil
}
