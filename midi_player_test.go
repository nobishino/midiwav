package main

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/gomidi/midi/v2/smf"
)

func TestMIDIToWAVE(t *testing.T) {
	filenames := []string{
		"issue0001_ok1",
		"issue0001_ok2",
		"issue0001_ng1",
		"issue0001_ng2",
	}
	for _, filename := range filenames {
		t.Run(filename, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", filename+".mid"))
			if err != nil {
				t.Fatalf("failed to open MIDI file: %v", err)
			}
			t.Cleanup(func() {
				f.Close()
			})
			dst, err := os.Create(filepath.Join("testdata", filename+".wav"))
			if err != nil {
				t.Fatalf("failed to open WAVE file: %v", err)
			}
			t.Cleanup(func() {
				dst.Close()
			})

			err = midiToWAVE(dst, f)
			if err != nil {
				t.Fatalf("MIDIToWAVE failed: %v", err)
			}
		})
	}
}

func TestViewIssue0001MIDI(t *testing.T) {
	filenames := []string{
		"issue0001_ok2",
		"issue0001_ng2",
	}
	for _, filename := range filenames {
		t.Run(filename, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", filename+".mid"))
			if err != nil {
				t.Fatalf("failed to open MIDI file: %v", err)
			}
			t.Cleanup(func() {
				f.Close()
			})
			smfData, err := smf.ReadFrom(f)
			if err != nil {
				t.Fatalf("failed to read SMF data: %v", err)
			}
			t.Log(smfData.String())

		})
	}
}
