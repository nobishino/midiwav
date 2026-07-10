package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"

	"gitlab.com/gomidi/midi/v2/smf"
)

func midiToWAVE(dst io.Writer, src io.Reader) error {
	smfData, err := smf.ReadFrom(src)
	if err != nil {
		return err
	}
	samples, err := smfToPCMArray(smfData)
	if err != nil {
		return err
	}
	if err := writeWAVE(dst, samples); err != nil {
		return err
	}
	return nil
}

func smfToPCMArray(smfData *smf.SMF) ([]int16, error) {
	metricTicks, ok := smfData.TimeFormat.(smf.MetricTicks)
	if !ok {
		return nil, errors.New("only MetricTicks time format is supported")
	}
	fmt.Println("MetricTicks:", metricTicks)
	tempi := newTempoMap(smfData, metricTicks)
	trackNum := len(smfData.Tracks)
	sampleTracks := make([][]float64, 0, trackNum)
	for _, track := range smfData.Tracks {
		var tick int64      // 曲頭からの絶対tick
		var sampleIndex int // サンプルの番号
		keyToVelocity := make(map[uint8]uint8)
		attack := make(map[uint8]uint16)             // 発音直後の音かどうか 音の連打を表現するために使ってみる
		samples := make([]float64, 0, sampleRate*60) // heuristic: 1分間のサンプル分のcapを事前確保
		for _, ev := range track {
			tick += int64(ev.Delta)
			var key, velocity uint8
			isNote := false
			switch msg := ev.Message; {
			case msg.GetNoteEnd(nil, &key):
				isNote = true
			case msg.GetNoteOn(nil, &key, &velocity):
				attack[key] = 1000 // 発音直後の音としてマーク
				isNote = true
			}
			// このイベントの時刻までのサンプルを出力する
			eventTime := tempi.seconds(tick)
			for sampleTime(sampleIndex) < eventTime {
				var deviation float64
				// すべての音符について、正弦波を重ね合わせ
				for k, velocity := range keyToVelocity {
					// TODO: Sine波以外の波形も選べるようにする
					amplitude := 10000.0 + float64(10*attack[k]) // 発音直後のサンプルは音量を上げる
					if attack[k] > 0 {
						attack[k]--
					}
					deviation += amplitude * float64(velocity) * squareWave(2*math.Pi*sampleTime(sampleIndex)*frequencyFromKey(k))
				}
				samples = append(samples, deviation)
				sampleIndex++ // sampleIndex == すでに出力し終わったサンプル数になる
			}
			// player stateの更新（音符以外のイベントは発音状態を変えない）
			if isNote {
				keyToVelocity[key] = velocity
			}
		}
		sampleTracks = append(sampleTracks, samples)
	}
	return summarizeSamples(sampleTracks), nil
}

// tempoMap は全トラックのテンポイベントから作る、絶対tick → 秒の変換表。
// SMF仕様に従い、テンポ未指定の区間は 120 BPM とする。
type tempoMap struct {
	ticks []int64   // テンポ変化点の絶対tick（昇順・重複なし）
	secs  []float64 // 各変化点の時刻（秒）
	spt   []float64 // 各変化点以降の 1 tick あたりの秒数
}

func newTempoMap(smfData *smf.SMF, metricTicks smf.MetricTicks) *tempoMap {
	type change struct {
		tick int64
		bpm  float64
	}
	changes := []change{{0, 120}} // SMFのデフォルトテンポ
	for _, track := range smfData.Tracks {
		var tick int64
		for _, ev := range track {
			tick += int64(ev.Delta)
			var bpm float64
			if ev.Message.GetMetaTempo(&bpm) && bpm > 0 {
				changes = append(changes, change{tick, bpm})
			}
		}
	}
	slices.SortStableFunc(changes, func(a, b change) int { return cmp.Compare(a.tick, b.tick) })

	m := &tempoMap{}
	var sec float64
	for i, c := range changes {
		if i > 0 {
			sec += float64(c.tick-changes[i-1].tick) * m.spt[len(m.spt)-1]
		}
		spt := 60 / (c.bpm * float64(metricTicks))
		if last := len(m.ticks) - 1; last >= 0 && m.ticks[last] == c.tick {
			// 同一tickの変化は後勝ち（tick 0 のデフォルトが上書きされるケースを含む）
			m.secs[last], m.spt[last] = sec, spt
		} else {
			m.ticks = append(m.ticks, c.tick)
			m.secs = append(m.secs, sec)
			m.spt = append(m.spt, spt)
		}
	}
	return m
}

// seconds は曲頭からの絶対tickを秒に変換する。
func (m *tempoMap) seconds(tick int64) float64 {
	i, found := slices.BinarySearch(m.ticks, tick)
	if !found {
		i-- // 直近の変化点（ticks[i] <= tick の最大の i）。ticks[0] == 0 なので i >= 0
	}
	return m.secs[i] + float64(tick-m.ticks[i])*m.spt[i]
}

func sampleTime(sampleIndex int) float64 {
	return float64(sampleIndex) / sampleRate
}

func frequencyFromKey(key uint8) float64 {
	return 440.0 * math.Pow(2, (float64(key)-69)/12)
}

func summarizeSamples(sampleTracks [][]float64) []int16 {
	maxLength := 0
	for _, samples := range sampleTracks {
		if len(samples) > maxLength {
			maxLength = len(samples)
		}
	}
	sums := make([]float64, maxLength)
	maxDeviation := 0.0
	for i := range maxLength {
		var sum float64
		for _, samples := range sampleTracks {
			if i < len(samples) {
				sum += samples[i]
			}
		}
		sums[i] = sum
		if math.Abs(sums[i]) > maxDeviation {
			maxDeviation = math.Abs(sums[i])
		}
	}
	return normalizeToInt16(sums, maxDeviation)
}

func normalizeToInt16(samples []float64, maxDeviation float64) []int16 {
	normalized := make([]int16, len(samples))
	if maxDeviation == 0 {
		// 全サンプルが無音。0除算（NaN）を避け、無音のまま返す。
		return trimSilence(normalized)
	}
	for i, v := range samples {
		normalized[i] = int16((v / maxDeviation) * (1<<15 - 1))
	}
	return trimSilence(normalized)
}

// trimSilence removes leading and trailing consecutive zeros (silence) from the sample slice.
func trimSilence(samples []int16) []int16 {
	if len(samples) == 0 {
		return samples
	}

	// Find the first non-zero sample
	start := 0
	for start < len(samples) && samples[start] == 0 {
		start++
	}

	// If all samples are zero, return empty slice
	if start == len(samples) {
		return []int16{}
	}

	// Find the last non-zero sample
	end := len(samples) - 1
	for end >= 0 && samples[end] == 0 {
		end--
	}

	// Return the trimmed slice (inclusive range)
	return samples[start : end+1]
}

func squareWave(phase float64) float64 {
	if math.Sin(phase) >= 0 {
		return 1.0
	}
	return -1.0
}
