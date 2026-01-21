package main

import (
	"errors"
	"fmt"
	"io"
	"math"

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
	trackNum := len(smfData.Tracks)
	sampleTracks := make([][]float64, 0, trackNum)
	var bpm float64
	// TODO: 複数トラック対応
	for _, track := range smfData.Tracks {
		var time float64    // 経過時間（秒）
		var sampleIndex int // サンプルの番号
		keyToVelocity := make(map[uint8]uint8)
		attack := make(map[uint8]uint16)             // 発音直後の音かどうか 音の連打を表現するために使ってみる
		samples := make([]float64, 0, sampleRate*60) // heuristic: 1分間のサンプル分のcapを事前確保
		for _, ev := range track {
			var key, velocity uint8
			switch msg := ev.Message; {
			case msg.GetMetaTempo(&bpm):
				fmt.Printf("Set Tempo to %f BPM(MetaTempo)\n", bpm)
			case msg.GetNoteEnd(nil, &key):
			case msg.GetNoteOn(nil, &key, &velocity):
				attack[key] = 1000 // 発音直後の音としてマーク
			case bpm == 0:
				continue // bpm == 0 の状態では時計を進めないし、音も鳴らさない
			}
			deltaTime := float64(ev.Delta) * 60 / (bpm * float64(metricTicks))
			for sampleTime(sampleIndex) < time+deltaTime {
				// サンプルの値を出力する
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
			// player stateの更新
			keyToVelocity[key] = velocity
			time += deltaTime
		}
		sampleTracks = append(sampleTracks, samples)
	}
	return summarizeSamples(sampleTracks), nil
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
		if sums[i] > maxDeviation {
			maxDeviation = sums[i]
		}
	}
	return normalizeToInt16(sums, maxDeviation)
}

func normalizeToInt16(samples []float64, maxDeviation float64) []int16 {
	normalized := make([]int16, len(samples))
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
