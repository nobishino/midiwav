package synth

import (
	"bufio"
	"encoding/binary"
	"io"
)

const (
	formatPCM     = 1
	fmtChunkSize  = 16 // fixed size for PCM
	sampleRate    = 44100 / 2
	channelNum    = 1 // mono
	bitsPerSample = 16
	byteRate      = sampleRate * channelNum * bitsPerSample / 8
	blockAlign    = channelNum * bitsPerSample / 8
)

func writeWAVE(w io.Writer, normalizedSamples []int16) error {
	// サンプルごとにos.Fileへ直接システムコールが発生するのを避けるためバッファする。
	bw := bufio.NewWriterSize(w, 1<<16)
	numSamples := uint32(len(normalizedSamples)) // モノラル前提
	lew := newLittleEndianWriter(bw)
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
	if err := lew.WriteUint16(formatPCM); err != nil {
		return err
	}
	if err := lew.WriteUint16(channelNum); err != nil {
		return err
	}
	if err := lew.WriteUint32(sampleRate); err != nil {
		return err
	}
	if err := lew.WriteUint32(byteRate); err != nil {
		return err
	}
	if err := lew.WriteUint16(blockAlign); err != nil {
		return err
	}
	if err := lew.WriteUint16(bitsPerSample); err != nil {
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
	return bw.Flush()
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
