package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindUnprocessedMIDIFiles_Recursive(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	
	// Create subdirectories
	subDir1 := filepath.Join(tmpDir, "subdir1")
	subDir2 := filepath.Join(tmpDir, "subdir2")
	if err := os.Mkdir(subDir1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(subDir2, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create MIDI files in root
	midiRoot := filepath.Join(tmpDir, "root.mid")
	if err := os.WriteFile(midiRoot, []byte("dummy midi"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create MIDI file in subdir1
	midiSub1 := filepath.Join(subDir1, "sub1.mid")
	if err := os.WriteFile(midiSub1, []byte("dummy midi"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create MIDI file in subdir2
	midiSub2 := filepath.Join(subDir2, "sub2.mid")
	if err := os.WriteFile(midiSub2, []byte("dummy midi"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create a WAV file for root.mid (older than MIDI)
	wavRoot := filepath.Join(tmpDir, "root.wav")
	if err := os.WriteFile(wavRoot, []byte("dummy wav"), 0644); err != nil {
		t.Fatal(err)
	}
	// Make WAV older
	oldTime := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(wavRoot, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	
	t.Run("non-recursive", func(t *testing.T) {
		files, err := findUnprocessedMIDIFiles(tmpDir, false)
		if err != nil {
			t.Fatalf("findUnprocessedMIDIFiles failed: %v", err)
		}
		
		// Should only find root.mid (because its WAV is older)
		// Should NOT find files in subdirectories
		if len(files) != 1 {
			t.Errorf("Expected 1 file, got %d: %v", len(files), files)
		}
		if len(files) > 0 && filepath.Base(files[0]) != "root.mid" {
			t.Errorf("Expected root.mid, got %s", filepath.Base(files[0]))
		}
	})
	
	t.Run("recursive", func(t *testing.T) {
		files, err := findUnprocessedMIDIFiles(tmpDir, true)
		if err != nil {
			t.Fatalf("findUnprocessedMIDIFiles failed: %v", err)
		}
		
		// Should find all MIDI files without WAV files, plus root.mid (WAV is older)
		// root.mid (has old WAV), sub1.mid (no WAV), sub2.mid (no WAV)
		if len(files) != 3 {
			t.Errorf("Expected 3 files, got %d: %v", len(files), files)
		}
		
		// Verify all expected files are found
		found := make(map[string]bool)
		for _, f := range files {
			found[filepath.Base(f)] = true
		}
		
		expected := []string{"root.mid", "sub1.mid", "sub2.mid"}
		for _, exp := range expected {
			if !found[exp] {
				t.Errorf("Expected to find %s", exp)
			}
		}
	})
}

func TestFindUnprocessedMIDIFiles_NonRecursive(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	
	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create MIDI file in root
	midiRoot := filepath.Join(tmpDir, "root.mid")
	if err := os.WriteFile(midiRoot, []byte("dummy midi"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create MIDI file in subdirectory
	midiSub := filepath.Join(subDir, "sub.mid")
	if err := os.WriteFile(midiSub, []byte("dummy midi"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Find files with recursive=false
	files, err := findUnprocessedMIDIFiles(tmpDir, false)
	if err != nil {
		t.Fatalf("findUnprocessedMIDIFiles failed: %v", err)
	}
	
	// Should only find root.mid, not sub.mid
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d: %v", len(files), files)
	}
	
	if len(files) > 0 && filepath.Base(files[0]) != "root.mid" {
		t.Errorf("Expected root.mid, got %s", filepath.Base(files[0]))
	}
}
