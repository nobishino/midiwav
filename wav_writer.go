package main

import (
	"encoding/binary"
	"io"
	"math"
	"time"
)

const (
	PCM           = 1
	fmtChunkSize  = 16 // fixed size for PCM
	sampleRate    = 44100 / 2
	duration      = 1.0 // 発音する秒数
	freq          = 440.0
	freq2         = freq * 1.5
	channelNum    = 1 // mono
	BitsPerSample = 16
	ByteRate      = sampleRate * channelNum * BitsPerSample / 8
	BlockAlign    = channelNum * BitsPerSample / 8
	numSamples    = sampleRate * duration
)

func writeWAVE(w io.Writer, normalizedSamples []int16) error {
	numSamples := uint32(len(normalizedSamples)) // モノラル前提
	lew := newLittleEndianWriter(w)
	// === WAV header ===
	if err := lew.WriteString("RIFF"); err != nil {
		return err
	}
	if err := lew.WriteUint32(36 + numSamples*2*channelNum); err != nil {
		return err
	}
	if err := lew.WriteString("WAVE"); err != nil {
		return err
	}

	if err := lew.WriteString("fmt "); err != nil {
		return err
	}
	if err := lew.WriteUint32(fmtChunkSize); err != nil {
		return err
	}
	if err := lew.WriteUint16(PCM); err != nil {
		return err
	}
	if err := lew.WriteUint16(channelNum); err != nil {
		return err
	}
	if err := lew.WriteUint32(sampleRate); err != nil {
		return err
	}
	if err := lew.WriteUint32(ByteRate); err != nil {
		return err
	}
	if err := lew.WriteUint16(BlockAlign); err != nil {
		return err
	}
	if err := lew.WriteUint16(BitsPerSample); err != nil {
		return err
	}

	if err := lew.WriteString("data"); err != nil {
		return err
	}
	if err := lew.WriteUint32(numSamples * channelNum * 2); err != nil {
		return err
	}

	// === PCM data ===
	for _, value := range normalizedSamples {
		if err := lew.WriteInt16(value); err != nil {
			return err
		}
	}
	return nil
}

// createWavFile generates a stereo WAV file with a test tone and writes it to the "samples" directory.
func createWavFile(f io.Writer) error {

	// f, err := os.Create(filepath.Join("samples", fmt.Sprintf("test%s.wav", dateTimeString())))
	// if err != nil {
	// 	return err
	// }
	// defer f.Close()

	w := newLittleEndianWriter(f)

	// === WAV header ===
	if err := w.WriteString("RIFF"); err != nil {
		return err
	}
	if err := w.WriteUint32(36 + numSamples*2*channelNum); err != nil {
		return err
	}
	if err := w.WriteString("WAVE"); err != nil {
		return err
	}

	if err := w.WriteString("fmt "); err != nil {
		return err
	}
	if err := w.WriteUint32(fmtChunkSize); err != nil {
		return err
	}
	if err := w.WriteUint16(PCM); err != nil {
		return err
	}
	if err := w.WriteUint16(channelNum); err != nil {
		return err
	}
	if err := w.WriteUint32(sampleRate); err != nil {
		return err
	}
	if err := w.WriteUint32(ByteRate); err != nil {
		return err
	}
	if err := w.WriteUint16(BlockAlign); err != nil {
		return err
	}
	if err := w.WriteUint16(BitsPerSample); err != nil {
		return err
	}

	if err := w.WriteString("data"); err != nil {
		return err
	}
	if err := w.WriteUint32(numSamples * channelNum * 2); err != nil {
		return err
	}

	// === PCM data ===
	for i := 0; i < numSamples; i++ {
		t := float64(i) / sampleRate
		vLeft := math.Sin(2 * math.Pi * freq * t)
		sampleLeft := int16(vLeft * 32767)
		if err := w.WriteInt16(sampleLeft); err != nil {
			return err
		}
	}
	return nil
}

func createSamplePCMData(numSamples int) []int16 {
	samples := make([]int16, 0, numSamples)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / sampleRate
		v := math.Sin(2 * math.Pi * freq * t)
		sampleValue := int16(v * 32767)
		samples = append(samples, sampleValue)
	}
	return samples
}

type littleEndianWriter struct {
	w io.Writer
}

func (lew *littleEndianWriter) WriteString(s string) error {
	_, err := lew.w.Write([]byte(s))
	return err
}

func (lew *littleEndianWriter) WriteUint16(v uint16) error {
	return binary.Write(lew.w, binary.LittleEndian, v)
}

func (lew *littleEndianWriter) WriteUint32(v uint32) error {
	return binary.Write(lew.w, binary.LittleEndian, v)
}

func (lew *littleEndianWriter) WriteInt16(v int16) error {
	return binary.Write(lew.w, binary.LittleEndian, v)
}

func newLittleEndianWriter(w io.Writer) *littleEndianWriter {
	return &littleEndianWriter{w: w}
}

func dateTimeString() string {
	// 現在日時を "YYYYMMDD_HHMMSS" 形式で取得
	return time.Now().Format("20060102_150405")
}
