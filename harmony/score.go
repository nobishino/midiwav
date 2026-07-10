package harmony

// Score は楽譜出力用の中間表現。MusicXMLなどの楽譜形式への変換に必要な
// 音高（綴り付き）・音価・調号・拍子を保持する。
//
// 音価は「次の和音のオンセットまで」（最終和音は小節末まで）の近似で、
// 休符や声部ごとの独立したリズムは表現しない（4声が同時に打鍵する
// ホモフォニックな実施を前提とする）。
type Score struct {
	Fifths          int  // 調号の五度圏値（正がシャープ、負がフラット）。調が不明なら0
	HasKey          bool // 調が判明しているか
	MeterNum        int  // 拍子の分子
	MeterDenom      int  // 拍子の分母
	TicksPerQuarter int  // 4分音符あたりのティック数（音価の単位）
	Chords          []ScoreChord
}

// ScoreChord は楽譜上の1和音（4声の同時打鍵）。Report.Chords と同順で対応する。
type ScoreChord struct {
	OnsetTicks    int64        // 曲頭からのオンセット（ティック）
	DurationTicks int64        // 音価（ティック）
	Notes         [4]ScoreNote // S, A, T, B の順
}

// ScoreNote は綴り付きの音高。
type ScoreNote struct {
	Step   string // 幹音 "C".."B"
	Alter  int    // 変化（-2〜+2。0は変化なし）
	Octave int    // 科学的ピッチ表記のオクターブ（C4が中央ハ）
}

// baseFifths[i] は noteLetters[i] を主音とする長調の調号（五度圏値）。
var baseFifths = [7]int{0, 2, 4, -1, 1, 3, 5}

// keyFifths は調号のシャープ/フラット数を返す（正がシャープ、負がフラット。
// 例: C-dur=0, e-moll=1, Es-moll=-6）。
func keyFifths(key Key) int {
	f := baseFifths[key.tonic.letter] + 7*key.tonic.acc
	if key.mode == "moll" {
		f -= 3
	}
	return f
}

// buildScore はchoraleから楽譜用中間表現を作る。
func buildScore(ch *chorale, key *Key, table [12]noteSpelling) *Score {
	sc := &Score{
		MeterNum:        ch.meterNum,
		MeterDenom:      ch.meterDenom,
		TicksPerQuarter: ch.ticksPerQuarter,
	}
	if key != nil {
		sc.HasKey = true
		sc.Fifths = keyFifths(*key)
	}
	for _, c := range ch.chords {
		var notes [4]ScoreNote
		for i, n := range c.notes {
			sp, oct := spellNote(n, table)
			notes[i] = ScoreNote{Step: noteLetters[sp.letter], Alter: sp.acc, Octave: oct}
		}
		sc.Chords = append(sc.Chords, ScoreChord{
			OnsetTicks:    c.ticks,
			DurationTicks: c.durTicks,
			Notes:         notes,
		})
	}
	return sc
}
