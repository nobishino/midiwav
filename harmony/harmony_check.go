// Package harmony は四声体和声のMIDI禁則チェック（芸大和声ベース）を提供する。
package harmony

// 四声体和声のMIDI禁則チェッカー（芸大和声ベース）
//
// 前提:
//   - SATBが4トラックに分かれていること（上から S, A, T, B の順）。
//   - ファイル名に調をドイツ語音名で含めること（例: es-moll.mid, As-dur.mid）。
//     転調しない課題向け。調が読み取れない場合、綴り依存の検査はスキップする。
//
// チェック項目:
//   - 連続1度・5度・8度（全声部ペア、反行連続を含む）
//   - 並達1度・5度・8度（外声間、ソプラノが跳躍のとき）
//   - 声部の交差（同一和音内）
//   - 隣接上3声の間隔（S-A, A-T が1オクターブ以内）
//   - 音域（芸大和声の声域表ベース・参考警告）
//   - 増音程の旋律進行（要・調）。増1度（半音階的変位）は許容として除外
//   - 減音程の旋律跳躍（要・調）。参考表示（跳躍後は反行が望ましい）
//   - 対斜（要・調）。参考表示（教科書の許容例外と照合すること）

import (
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gitlab.com/gomidi/midi/v2/smf"
)

var voiceNames = [4]string{"S", "A", "T", "B"}

// 参考: 芸大和声の声域（要教科書照合）
var voiceRanges = [4][2]uint8{
	{60, 81}, // S
	{53, 74}, // A
	{48, 69}, // T
	{40, 62}, // B
}

var noteLetters = [7]string{"C", "D", "E", "F", "G", "A", "B"}

// basePitchClass[i] は noteLetters[i] のピッチクラス。
// 度数(g%7)を添字にすると、その度数の完全/長音程の半音数にもなる。
var basePitchClass = [7]int{0, 2, 4, 5, 7, 9, 11}

var accidentalStr = map[int]string{-2: "bb", -1: "b", 0: "", 1: "#", 2: "##"}

var qualityNames = map[int]string{-2: "重減", -1: "減", 1: "増", 2: "重増"}

// noteSpelling は音名の綴り（幹音と変化）。
type noteSpelling struct {
	letter int // noteLetters の添字
	acc    int // 変化（-2〜+2）
}

// germanNames はドイツ語音名 → 綴り。b はドイツ語の B♭、h が B♮。
var germanNames = map[string]noteSpelling{
	"c": {0, 0}, "cis": {0, 1}, "ces": {0, -1},
	"d": {1, 0}, "dis": {1, 1}, "des": {1, -1},
	"e": {2, 0}, "eis": {2, 1}, "es": {2, -1},
	"f": {3, 0}, "fis": {3, 1}, "fes": {3, -1},
	"g": {4, 0}, "gis": {4, 1}, "ges": {4, -1},
	"a": {5, 0}, "ais": {5, 1}, "as": {5, -1},
	"h": {6, 0}, "his": {6, 1}, "b": {6, -1},
}

// Key は調（主音とdur/moll）。
type Key struct {
	tonic noteSpelling
	mode  string // "dur" または "moll"
}

func (k Key) String() string {
	return noteLetters[k.tonic.letter] + accidentalStr[k.tonic.acc] + "-" + k.mode
}

var keyPattern = regexp.MustCompile(`(?:^|[^a-z])([a-h](?:is|es|s)?)[-_ ]?(dur|moll)`)

// ParseKeyFromFilename はファイル名のドイツ語音名から調を読み取る（例: es-moll.mid）。
func ParseKeyFromFilename(path string) (Key, bool) {
	base := filepath.Base(path)
	stem := strings.ToLower(strings.TrimSuffix(base, filepath.Ext(base)))
	m := keyPattern.FindStringSubmatch(stem)
	if m == nil {
		return Key{}, false
	}
	tonic, ok := germanNames[m[1]]
	if !ok {
		return Key{}, false
	}
	return Key{tonic: tonic, mode: m[2]}, true
}

// flatSpellingTable は調が不明なときの代用（黒鍵はフラット表記）。
var flatSpellingTable = [12]noteSpelling{
	{0, 0}, {1, -1}, {1, 0}, {2, -1}, {2, 0}, {3, 0},
	{4, -1}, {4, 0}, {5, -1}, {5, 0}, {6, -1}, {6, 0},
}

// buildSpellingTable はピッチクラス → 綴りの12音表を作る。全音階音は音階の綴り、
// 半音階音は「下の音階音の上方変位」を優先し、無理なら上の音の下方変位。
func buildSpellingTable(key Key) [12]noteSpelling {
	steps := [7]int{2, 2, 1, 2, 2, 2, 1}
	if key.mode == "moll" {
		steps = [7]int{2, 1, 2, 2, 1, 2, 2}
	}
	var table [12]noteSpelling
	var isScaleTone [12]bool
	li := key.tonic.letter
	pc := mod12(basePitchClass[li] + key.tonic.acc)
	for i := range 7 {
		if i > 0 {
			pc = mod12(pc + steps[i-1])
			li = (li + 1) % 7
		}
		table[pc] = noteSpelling{li, mod12(pc-basePitchClass[li]+6) - 6}
		isScaleTone[pc] = true
	}
	for pc := range 12 {
		if isScaleTone[pc] {
			continue
		}
		lo, hi := table[mod12(pc-1)], table[mod12(pc+1)]
		if abs(lo.acc+1) <= 2 {
			table[pc] = noteSpelling{lo.letter, lo.acc + 1}
		} else {
			table[pc] = noteSpelling{hi.letter, hi.acc - 1}
		}
	}
	return table
}

// spellNote はMIDIノート番号を綴りとオクターブに分解する。
func spellNote(note uint8, table [12]noteSpelling) (noteSpelling, int) {
	sp := table[note%12]
	octave := floorDiv(int(note)-basePitchClass[sp.letter]-sp.acc, 12) - 1
	return sp, octave
}

func noteName(note uint8, table [12]noteSpelling) string {
	sp, octave := spellNote(note, table)
	return fmt.Sprintf("%s%s%d", noteLetters[sp.letter], accidentalStr[sp.acc], octave)
}

// intervalQuality は旋律音程を (度数, 変位, 質の名前) で返す。
// 変位: 0=完全/長/短, +1=増, -1=減 など。
func intervalQuality(n1, n2 uint8, table [12]noteSpelling) (int, int, string) {
	sp1, o1 := spellNote(n1, table)
	sp2, o2 := spellNote(n2, table)
	g := abs((o2*7 + sp2.letter) - (o1*7 + sp1.letter)) // 度数-1
	c := abs(int(n2) - int(n1))                         // 半音数
	dev := c - (basePitchClass[g%7] + 12*(g/7))
	perfect := g%7 == 0 || g%7 == 3 || g%7 == 4
	if !perfect && dev == -1 {
		return g + 1, 0, "短"
	}
	if dev == 0 {
		if perfect {
			return g + 1, 0, "完全"
		}
		return g + 1, 0, "長"
	}
	if !perfect && dev < -1 {
		dev++
	}
	name, ok := qualityNames[dev]
	if !ok {
		name = fmt.Sprintf("dev%d", dev)
	}
	return g + 1, dev, name
}

// chord は同時に発音される4声の音（S, A, T, B の順）。
type chord struct {
	beat     float64 // 曲頭からの拍数（4分音符換算）
	ticks    int64   // 曲頭からのオンセット（ティック）
	durTicks int64   // 音価（ティック）。次の和音のオンセットまで（最終和音は小節末まで）の近似
	notes    [4]uint8
}

// chorale は4声体とみなしたMIDIから抽出した和音列。
type chorale struct {
	chords          []chord
	beatsPerMeasure float64 // 1小節の拍数（4分音符換算）
	beatUnit        float64 // 1拍の長さ（4分音符換算）
	meterNum        int     // 拍子の分子
	meterDenom      int     // 拍子の分母
	ticksPerQuarter int     // 4分音符あたりのティック数
}

func (c *chorale) pos(beat float64) string {
	measure := int(beat/c.beatsPerMeasure) + 1
	beatInMeasure := int(math.Mod(beat, c.beatsPerMeasure)/c.beatUnit) + 1
	return fmt.Sprintf("%d小節%d拍目", measure, beatInMeasure)
}

// extractChorale はSMFから4声体の和音列を抽出する。音符のあるトラックが
// ちょうど4本（上から S, A, T, B）でない場合は4声体でないとみなす。
// 和音は4声すべてが同時に発音するオンセットから作る。
func extractChorale(s *smf.SMF) (*chorale, bool) {
	metricTicks, ok := s.TimeFormat.(smf.MetricTicks)
	if !ok {
		return nil, false
	}
	num, denom := uint8(4), uint8(4)
	var parts []map[int64]uint8
	for _, track := range s.Tracks {
		var t int64
		notes := make(map[int64]uint8)
		for _, ev := range track {
			t += int64(ev.Delta)
			var key, velocity uint8
			if ev.Message.GetNoteOn(nil, &key, &velocity) && velocity > 0 {
				notes[t] = key
			} else {
				ev.Message.GetMetaMeter(&num, &denom)
			}
		}
		if len(notes) > 0 {
			parts = append(parts, notes)
		}
	}
	if len(parts) != 4 {
		return nil, false
	}
	onsetSet := make(map[int64]struct{})
	for _, p := range parts {
		for t := range p {
			onsetSet[t] = struct{}{}
		}
	}
	onsets := make([]int64, 0, len(onsetSet))
	for t := range onsetSet {
		onsets = append(onsets, t)
	}
	slices.Sort(onsets)

	var chords []chord
	for _, o := range onsets {
		var notes [4]uint8
		all := true
		for i, p := range parts {
			n, ok := p[o]
			if !ok {
				all = false
				break
			}
			notes[i] = n
		}
		if all {
			chords = append(chords, chord{beat: float64(o) / float64(metricTicks), ticks: o, notes: notes})
		}
	}
	if len(chords) == 0 {
		return nil, false
	}

	// 音価はnote-offを読まず「次の和音のオンセットまで」とする近似。
	// 最終和音はその小節の末尾まで伸ばす。
	tpq := int64(metricTicks)
	measureTicks := tpq * 4 * int64(num) / int64(denom)
	for i := range chords {
		if i+1 < len(chords) {
			chords[i].durTicks = chords[i+1].ticks - chords[i].ticks
		} else if measureTicks > 0 {
			chords[i].durTicks = measureTicks - chords[i].ticks%measureTicks
		} else {
			chords[i].durTicks = tpq
		}
	}

	return &chorale{
		chords:          chords,
		beatsPerMeasure: float64(num) * 4 / float64(denom),
		beatUnit:        4 / float64(denom),
		meterNum:        int(num),
		meterDenom:      int(denom),
		ticksPerQuarter: int(metricTicks),
	}, true
}

// Issue は検出した禁則・参考警告。
type Issue struct {
	Warn bool // true: ⚠ 禁則, false: ℹ 参考
	Msg  string
}

// checkBasic は綴りに依存しない検査: 交差・間隔・音域・連続・並達。
func checkBasic(ch *chorale, table [12]noteSpelling) []Issue {
	var issues []Issue
	nm := func(n uint8) string { return noteName(n, table) }
	for _, c := range ch.chords {
		s, a, t, b := c.notes[0], c.notes[1], c.notes[2], c.notes[3]
		if !(s >= a && a >= t && t >= b) {
			names := make([]string, 4)
			for i, n := range c.notes {
				names[i] = nm(n)
			}
			issues = append(issues, Issue{true,
				fmt.Sprintf("[交差] %s: %s", ch.pos(c.beat), strings.Join(names, " "))})
		}
		if int(s)-int(a) > 12 {
			issues = append(issues, Issue{true,
				fmt.Sprintf("[間隔] %s: S-A が %d 半音（8度超）", ch.pos(c.beat), int(s)-int(a))})
		}
		if int(a)-int(t) > 12 {
			issues = append(issues, Issue{true,
				fmt.Sprintf("[間隔] %s: A-T が %d 半音（8度超）", ch.pos(c.beat), int(a)-int(t))})
		}
		for i, n := range c.notes {
			if n < voiceRanges[i][0] || n > voiceRanges[i][1] {
				issues = append(issues, Issue{false,
					fmt.Sprintf("[音域?] %s: %s の %s が声域外の可能性（要照合）", ch.pos(c.beat), voiceNames[i], nm(n))})
			}
		}
	}
	for k := 0; k+1 < len(ch.chords); k++ {
		c1, c2 := ch.chords[k], ch.chords[k+1]
		for i := range 4 {
			for j := i + 1; j < 4; j++ {
				hi1, lo1, hi2, lo2 := c1.notes[i], c1.notes[j], c2.notes[i], c2.notes[j]
				int1, int2 := mod12(int(hi1)-int(lo1)), mod12(int(hi2)-int(lo2))
				moved := hi1 != hi2 && lo1 != lo2
				pair := voiceNames[i] + "-" + voiceNames[j]
				if moved && int1 == int2 && (int1 == 0 || int1 == 7) {
					kind := "5度"
					if int1 == 0 {
						kind = "8度/1度"
					}
					issues = append(issues, Issue{true,
						fmt.Sprintf("[連続%s] %s→%s %s: %s/%s → %s/%s",
							kind, ch.pos(c1.beat), ch.pos(c2.beat), pair,
							nm(hi1), nm(lo1), nm(hi2), nm(lo2))})
				}
				// 並達は外声間のみ、ソプラノが跳躍して同方向のとき
				if i == 0 && j == 3 && moved && (int2 == 0 || int2 == 7) && int1 != int2 {
					sameDir := (int(hi2)-int(hi1))*(int(lo2)-int(lo1)) > 0
					if sameDir && abs(int(hi2)-int(hi1)) > 2 {
						kind := "5度"
						if int2 == 0 {
							kind = "8度"
						}
						issues = append(issues, Issue{true,
							fmt.Sprintf("[並達%s] %s→%s 外声: %s/%s → %s/%s",
								kind, ch.pos(c1.beat), ch.pos(c2.beat),
								nm(hi1), nm(lo1), nm(hi2), nm(lo2))})
					}
				}
			}
		}
	}
	return issues
}

// checkSpelled は調の綴りに依存する検査: 増・減音程の旋律進行、対斜。
func checkSpelled(ch *chorale, table [12]noteSpelling) []Issue {
	var issues []Issue
	nm := func(n uint8) string { return noteName(n, table) }
	for k := 0; k+1 < len(ch.chords); k++ {
		c1, c2 := ch.chords[k], ch.chords[k+1]
		for v := range 4 {
			n1, n2 := c1.notes[v], c2.notes[v]
			if n1 == n2 {
				continue
			}
			size, dev, q := intervalQuality(n1, n2, table)
			arrow := "↓"
			if n2 > n1 {
				arrow = "↑"
			}
			if dev >= 1 && size >= 2 {
				issues = append(issues, Issue{true,
					fmt.Sprintf("[増音程] %s→%s %s: %s %s %s（%s%d度）",
						ch.pos(c1.beat), ch.pos(c2.beat), voiceNames[v], nm(n1), arrow, nm(n2), q, size)})
			} else if dev <= -1 {
				issues = append(issues, Issue{false,
					fmt.Sprintf("[減音程] %s→%s %s: %s %s %s（%s%d度・跳躍後は反行が望ましい）",
						ch.pos(c1.beat), ch.pos(c2.beat), voiceNames[v], nm(n1), arrow, nm(n2), q, size)})
			}
		}
		// 対斜: 同じ幹音が隣接和音で異なる変化を持ち、その半音変化が同一声部で起きていない
		var sp1, sp2 [4]noteSpelling
		for i := range 4 {
			sp1[i], _ = spellNote(c1.notes[i], table)
			sp2[i], _ = spellNote(c2.notes[i], table)
		}
		for i := range 4 {
			for j := range 4 {
				if sp1[i].letter != sp2[j].letter || sp1[i].acc == sp2[j].acc || i == j {
					continue
				}
				if sp2[i] == sp2[j] { // 同声部でも同じ変化に到達
					continue
				}
				if sp1[j] == sp1[i] {
					continue
				}
				issues = append(issues, Issue{false,
					fmt.Sprintf("[対斜] %s→%s: %s の %s と %s の %s（要・教科書の許容例外と照合）",
						ch.pos(c1.beat), ch.pos(c2.beat),
						voiceNames[i], nm(c1.notes[i]), voiceNames[j], nm(c2.notes[j]))})
			}
		}
	}
	return issues
}

// Chord は和音一覧の1項目。
type Chord struct {
	Pos         string    // 位置（例: "1小節1拍目"）
	Symbol      string    // 和音記号（度数・種類・転回位置）。調が不明の場合は空
	ShortSymbol string    // 楽譜表示用の短い和音記号（例: "V₉(r-)"）。調が不明の場合は空
	Name        string    // コードネーム（例: "C", "G7/F"）
	Notes       [4]string // S, A, T, B の順の音名
}

// Report は4声体和声の分析結果。
type Report struct {
	Key    *Key // 調。不明なら nil
	Chords []Chord
	Issues []Issue
	Score  *Score // 楽譜出力用の中間表現。Chords と同順で対応する
}

// Analyze は4声体和声の禁則チェックを実行し、分析結果を返す。
// key が nil の場合、綴りはフラット表記で代用し、綴りに依存する検査は
// スキップする。4声体とみなせないMIDIの場合は ok=false を返す。
func Analyze(s *smf.SMF, key *Key) (*Report, bool) {
	ch, ok := extractChorale(s)
	if !ok {
		return nil, false
	}
	table := flatSpellingTable
	var templates []chordTemplate
	if key != nil {
		table = buildSpellingTable(*key)
		templates = chordTemplates(*key)
	}

	r := &Report{Key: key, Score: buildScore(ch, key, table)}
	for _, c := range ch.chords {
		chord := Chord{Pos: ch.pos(c.beat), Name: chordName(c.notes, table)}
		if key != nil {
			if a, ok := analyzeChord(c.notes, templates); ok {
				chord.Symbol = a.String()
				chord.ShortSymbol = a.ShortString()
			} else {
				chord.Symbol = "?"
				chord.ShortSymbol = "?"
			}
		}
		for i, n := range c.notes {
			chord.Notes[i] = noteName(n, table)
		}
		r.Chords = append(r.Chords, chord)
	}
	r.Issues = checkBasic(ch, table)
	if key != nil {
		r.Issues = append(r.Issues, checkSpelled(ch, table)...)
	}
	return r, true
}

// Format は分析結果を添削レポートのテキストに整形する。
// 調は ParseKeyFromFilename でファイル名から得ている前提の文言を含む。
func (r *Report) Format() string {
	var b strings.Builder
	if r.Key != nil {
		fmt.Fprintf(&b, "調: %s（ファイル名から）\n", r.Key)
	} else {
		b.WriteString("注意: ファイル名から調を読み取れず。和音記号の判定と増音程・減音程・対斜の検査はスキップし、綴りはフラット表記で代用。\n")
	}
	fmt.Fprintf(&b, "%d 和音を検査:\n", len(r.Chords))
	for _, c := range r.Chords {
		fmt.Fprintf(&b, "  %s ", c.Pos)
		if r.Key != nil {
			fmt.Fprintf(&b, " %s", c.Symbol)
		}
		fmt.Fprintf(&b, " [%s]", c.Name)
		for i := 3; i >= 0; i-- { // 下からB, T, A, Sの順で表示する
			fmt.Fprintf(&b, " %s:%s", voiceNames[i], c.Notes[i])
		}
		b.WriteByte('\n')
	}

	b.WriteByte('\n')
	var warns, infos int
	for _, is := range r.Issues {
		if is.Warn {
			b.WriteString("⚠ " + is.Msg + "\n")
			warns++
		}
	}
	for _, is := range r.Issues {
		if !is.Warn {
			b.WriteString("ℹ " + is.Msg + "\n")
			infos++
		}
	}
	if warns == 0 {
		if infos > 0 {
			b.WriteString("✓ 禁則違反は検出されませんでした（参考警告あり）\n")
		} else {
			b.WriteString("✓ 禁則違反は検出されませんでした\n")
		}
	}
	return b.String()
}

func mod12(x int) int {
	return ((x % 12) + 12) % 12
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func floorDiv(a, b int) int {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}
