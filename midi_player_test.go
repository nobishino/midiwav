package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/gomidi/midi/v2/smf"
)

var update = flag.Bool("update", false, "update golden files")

func TestMIDIToWAVE(t *testing.T) {
	filenames := []string{
		"issue0001_ok1",
		"issue0001_ok2",
		"issue0001_ng1",
		"issue0001_ng2",
	}
	// "testdata/issue0002.mid",
	for _, filename := range filenames {
		t.Run(filename, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", filename+".mid"))
			if err != nil {
				t.Fatalf("failed to open MIDI file: %v", err)
			}
			t.Cleanup(func() {
				f.Close()
			})

			// Generate WAV output to a buffer
			var buf bytes.Buffer
			err = midiToWAVE(&buf, f)
			if err != nil {
				t.Fatalf("MIDIToWAVE failed: %v", err)
			}

			// Get the generated output
			got := buf.Bytes()
			goldenPath := filepath.Join("testdata", filename+".wav")

			// Update mode: write the output to the golden file
			if *update {
				if err := os.WriteFile(goldenPath, got, 0644); err != nil {
					t.Fatalf("failed to update golden file: %v", err)
				}
				return
			}

			// Compare mode: read the golden file and compare
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("failed to read golden file: %v", err)
			}

			if !bytes.Equal(got, want) {
				t.Errorf("WAV output differs from golden file %s\ngot %d bytes, want %d bytes", goldenPath, len(got), len(want))
			}
		})
	}
}

func TestViewIssue0001MIDI(t *testing.T) {

	filenames := []string{
		// "issue0001_ok1",
		"issue0001_ok2",
		// "issue0001_ng1",
		"issue0001_ng2",
	}
	// "testdata/issue0002.mid",
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
