package harmony

// 和音記号（度数・種類・転回位置）の判定。
//
// 調の音階（durは長音階、mollは和声的短音階）上の三度堆積から候補テンプレートを
// 作り、実施4音のピッチクラス集合と照合する。完全一致を最優先とし、第5音省略形も
// 許容する（表示には出さない）。VIIの三和音・七の和音は芸大和声の慣例に従い
// V7・V9の根音省略形として扱う。借用和音は V調（属音を主音とする同モードの調）の
// V・V7・V9系のみを対象とする。

// 和音構成音の役割。値は転回位置の番号（0=基本位置の基準となる根音）を兼ねる。
const (
	roleRoot = iota
	roleThird
	roleFifth
	roleSeventh
	roleNinth
)

var inversionNames = [5]string{"基本位置", "第1転回位置", "第2転回位置", "第3転回位置", "第4転回位置"}

var degreeNumerals = [7]string{"I", "II", "III", "IV", "V", "VI", "VII"}

// chordTemplate は照合対象となる和音の雛形。
type chordTemplate struct {
	label       string      // 例: "I", "V7", "V調のV9"
	rootOmitted bool        // 根音省略形
	borrowed    bool        // 借用和音（V調）
	sizeRank    int         // 0=三和音, 1=七の和音, 2=九の和音
	degree      int         // 度数（0..6）。優先順位のタイブレーク用
	pcByRole    map[int]int // 役割 → ピッチクラス（存在する構成音のみ）
}

// scalePitchClasses は調の音階のピッチクラスを返す（mollは和声的短音階）。
func scalePitchClasses(key Key) [7]int {
	steps := [6]int{2, 2, 1, 2, 2, 2}
	if key.mode == "moll" {
		steps = [6]int{2, 1, 2, 2, 1, 3}
	}
	var s [7]int
	s[0] = mod12(basePitchClass[key.tonic.letter] + key.tonic.acc)
	for i, st := range steps {
		s[i+1] = mod12(s[i] + st)
	}
	return s
}

// buildTemplate は音階のdegree度上にnumTones個の三度堆積和音を作る。
func buildTemplate(scale [7]int, degree, numTones int, label string) chordTemplate {
	pcs := make(map[int]int, numTones)
	for r := range numTones {
		pcs[r] = scale[(degree+2*r)%7]
	}
	return chordTemplate{label: label, sizeRank: numTones - 3, degree: degree, pcByRole: pcs}
}

// omitRoot は根音省略形を作る。
func omitRoot(t chordTemplate) chordTemplate {
	pcs := make(map[int]int, len(t.pcByRole)-1)
	for r, pc := range t.pcByRole {
		if r != roleRoot {
			pcs[r] = pc
		}
	}
	t.pcByRole = pcs
	t.rootOmitted = true
	return t
}

// borrowFromDominant はテンプレートをV調（属音を主音とする同モードの調）へ移す。
func borrowFromDominant(t chordTemplate) chordTemplate {
	pcs := make(map[int]int, len(t.pcByRole))
	for r, pc := range t.pcByRole {
		pcs[r] = mod12(pc + 7)
	}
	t.pcByRole = pcs
	t.label = "V調の" + t.label
	t.borrowed = true
	return t
}

// chordTemplates は調に対する候補和音テンプレートの一覧を作る。
func chordTemplates(key Key) []chordTemplate {
	scale := scalePitchClasses(key)
	var ts []chordTemplate
	for d := range 6 { // I〜VI。VIIの三和音はV7の根音省略形として扱う
		ts = append(ts, buildTemplate(scale, d, 3, degreeNumerals[d]))
	}
	ii7 := buildTemplate(scale, 1, 4, "II7")
	v7 := buildTemplate(scale, 4, 4, "V7")
	v9 := buildTemplate(scale, 4, 5, "V9")
	ts = append(ts, ii7, v7, omitRoot(v7), v9, omitRoot(v9))
	v := buildTemplate(scale, 4, 3, "V")
	for _, t := range []chordTemplate{v, v7, omitRoot(v7), v9, omitRoot(v9)} {
		ts = append(ts, borrowFromDominant(t))
	}
	return ts
}

// chordAnalysis は和音記号の判定結果。
type chordAnalysis struct {
	template     chordTemplate
	fifthOmitted bool
	bassRole     int // バスが担う構成音の役割。転回位置の番号
}

func (a chordAnalysis) String() string {
	inv := inversionNames[a.bassRole]
	if a.template.rootOmitted {
		return a.template.label + "(根音省略形・" + inv + ")"
	}
	return a.template.label + "(" + inv + ")"
}

// matchTemplate は実施のピッチクラス集合をテンプレートと照合する。
// 完全形が一致しなければ第5音省略形も試す。
func matchTemplate(t chordTemplate, pcSet [12]bool, bassPC int) (chordAnalysis, bool) {
	match := func(omitFifth bool) (chordAnalysis, bool) {
		var want [12]bool
		bassRole := -1
		for r, pc := range t.pcByRole {
			if omitFifth && r == roleFifth {
				continue
			}
			want[pc] = true
			if pc == bassPC {
				bassRole = r
			}
		}
		if want != pcSet || bassRole < 0 {
			return chordAnalysis{}, false
		}
		return chordAnalysis{template: t, fifthOmitted: omitFifth, bassRole: bassRole}, true
	}
	if a, ok := match(false); ok {
		return a, true
	}
	if _, hasFifth := t.pcByRole[roleFifth]; hasFifth {
		return match(true)
	}
	return chordAnalysis{}, false
}

// betterAnalysis は a が b より妥当な解釈かを返す。
// 優先順位: 省略が少ない > 固有和音 > 単純な和音 > 低い度数。
func betterAnalysis(a, b chordAnalysis) bool {
	ak := [4]int{b2i(a.fifthOmitted), b2i(a.template.borrowed), a.template.sizeRank, a.template.degree}
	bk := [4]int{b2i(b.fifthOmitted), b2i(b.template.borrowed), b.template.sizeRank, b.template.degree}
	for i := range ak {
		if ak[i] != bk[i] {
			return ak[i] < bk[i]
		}
	}
	return false
}

// analyzeChord は4音の和音記号を判定する。どの候補にも一致しなければ ok=false。
func analyzeChord(notes [4]uint8, templates []chordTemplate) (chordAnalysis, bool) {
	var pcSet [12]bool
	bass := notes[0]
	for _, n := range notes {
		pcSet[n%12] = true
		if n < bass {
			bass = n
		}
	}
	var best chordAnalysis
	found := false
	for _, t := range templates {
		a, ok := matchTemplate(t, pcSet, int(bass%12))
		if !ok {
			continue
		}
		if !found || betterAnalysis(a, best) {
			best, found = a, true
		}
	}
	return best, found
}

// chordSymbol は和音記号の表示文字列を返す。判定できない場合は "?"。
func chordSymbol(notes [4]uint8, templates []chordTemplate) string {
	a, ok := analyzeChord(notes, templates)
	if !ok {
		return "?"
	}
	return a.String()
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// chordQuality はコードネームの種類。intervals は根音からの半音数の集合（0を含む）。
type chordQuality struct {
	intervals []int
	suffix    string
}

// chordQualities は優先順位順（上ほど優先）。三和音・七の和音の完全形を先に、
// 第5音省略形と2音の形を後に置く。
var chordQualities = []chordQuality{
	{[]int{0, 4, 7}, ""},
	{[]int{0, 3, 7}, "m"},
	{[]int{0, 3, 6}, "dim"},
	{[]int{0, 4, 8}, "aug"},
	{[]int{0, 4, 7, 10}, "7"},
	{[]int{0, 3, 7, 10}, "m7"},
	{[]int{0, 4, 7, 11}, "maj7"},
	{[]int{0, 3, 7, 11}, "mM7"},
	{[]int{0, 3, 6, 10}, "m7-5"},
	{[]int{0, 3, 6, 9}, "dim7"},
	// 第5音省略の七の和音
	{[]int{0, 4, 10}, "7"},
	{[]int{0, 3, 10}, "m7"},
	{[]int{0, 4, 11}, "maj7"},
	// 2音
	{[]int{0, 4}, ""},
	{[]int{0, 3}, "m"},
	{[]int{0, 7}, "5"},
}

// pitchClassName はピッチクラスの音名（オクターブなし）を返す。
func pitchClassName(pc int, table [12]noteSpelling) string {
	sp := table[pc]
	return noteLetters[sp.letter] + accidentalStr[sp.acc]
}

// chordName は4音のコードネーム（例: "C", "Dm7", "G7/F"）を返す。
// バスが根音でない場合はスラッシュでバス音を付ける。調に依存せず判定でき、
// 綴りは渡された表に従う。どの形にも一致しなければ "?"。
func chordName(notes [4]uint8, table [12]noteSpelling) string {
	var pcSet [12]bool
	bass := notes[0]
	for _, n := range notes {
		pcSet[n%12] = true
		if n < bass {
			bass = n
		}
	}
	bassPC := int(bass % 12)
	// 根音候補は実施に含まれる音。バスを最優先し、残りはピッチクラス昇順
	roots := []int{bassPC}
	for pc := range 12 {
		if pcSet[pc] && pc != bassPC {
			roots = append(roots, pc)
		}
	}
	for _, q := range chordQualities {
		for _, root := range roots {
			var want [12]bool
			for _, iv := range q.intervals {
				want[mod12(root+iv)] = true
			}
			if want != pcSet {
				continue
			}
			name := pitchClassName(root, table) + q.suffix
			if root != bassPC {
				name += "/" + pitchClassName(bassPC, table)
			}
			return name
		}
	}
	return "?"
}
