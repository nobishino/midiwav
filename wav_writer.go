package main

import (
	"encoding/binary"
	"io"
)

const (
	PCM           = 1
	fmtChunkSize  = 16 // fixed size for PCM
	sampleRate    = 44100 / 2
	channelNum    = 1 // mono
	BitsPerSample = 16
	ByteRate      = sampleRate * channelNum * BitsPerSample / 8
	BlockAlign    = channelNum * BitsPerSample / 8
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
