package main

import (
	"strings"
	"testing"
)

func templatesFor(t *testing.T, filename string) []chordTemplate {
	t.Helper()
	key, ok := parseKeyFromFilename(filename)
	if !ok {
		t.Fatalf("failed to parse key from %q", filename)
	}
	return chordTemplates(key)
}

func TestChordSymbolCDur(t *testing.T) {
	ts := templatesFor(t, "c-dur.mid")
	tests := []struct {
		name  string
		notes [4]uint8 // S A T B
		want  string
	}{
		{"I 基本位置", [4]uint8{72, 67, 64, 60}, "I(基本位置)"},                  // C5 G4 E4 C4
		{"I 第1転回位置", [4]uint8{72, 67, 60, 52}, "I(第1転回位置)"},              // C5 G4 C4 E3
		{"IV 第1転回位置", [4]uint8{77, 72, 65, 57}, "IV(第1転回位置)"},            // F5 C5 F4 A3
		{"V 基本位置", [4]uint8{74, 71, 67, 55}, "V(基本位置)"},                  // D5 B4 G4 G3
		{"V7 基本位置", [4]uint8{77, 71, 62, 55}, "V7(基本位置)"},                // F5 B4 D4 G3
		{"V7 第5音省略も V7", [4]uint8{77, 71, 67, 55}, "V7(基本位置)"},           // F5 B4 G4 G3
		{"V7 第3転回位置", [4]uint8{74, 71, 67, 53}, "V7(第3転回位置)"},            // D5 B4 G4 F3
		{"V7 根音省略形", [4]uint8{74, 65, 62, 59}, "V7(根音省略形・第1転回位置)"},       // D5 F4 D4 B3
		{"V9 根音省略形", [4]uint8{69, 65, 62, 59}, "V9(根音省略形・第1転回位置)"},       // A4 F4 D4 B3
		{"II7 基本位置", [4]uint8{72, 69, 65, 50}, "II7(基本位置)"},              // C5 A4 F4 D3
		{"V調のV7 基本位置", [4]uint8{72, 69, 66, 50}, "V調のV7(基本位置)"},          // C5 A4 F#4 D3
		{"V調のV9 根音省略形", [4]uint8{76, 72, 69, 54}, "V調のV9(根音省略形・第1転回位置)"}, // E5 C5 A4 F#3
		{"判定不能", [4]uint8{72, 71, 61, 60}, "?"},                          // C5 B4 C#4 C4
	}
	for _, tt := range tests {
		if got := chordSymbol(tt.notes, ts); got != tt.want {
			t.Errorf("%s: chordSymbol(%v) = %s, want %s", tt.name, tt.notes, got, tt.want)
		}
	}
}

func TestChordSymbolEsMoll(t *testing.T) {
	ts := templatesFor(t, "es-moll.mid")
	tests := []struct {
		name  string
		notes [4]uint8 // S A T B
		want  string
	}{
		{"I 基本位置", [4]uint8{75, 70, 66, 51}, "I(基本位置)"},                  // Eb5 Bb4 Gb4 Eb3
		{"IV 基本位置", [4]uint8{75, 71, 68, 56}, "IV(基本位置)"},                // Eb5 Cb5 Ab4 Ab3
		{"V 第1転回位置", [4]uint8{70, 65, 62, 50}, "V(第1転回位置)"},              // Bb4 F4 D4 D3
		{"V9 根音省略形", [4]uint8{74, 65, 59, 44}, "V9(根音省略形・第3転回位置)"},       // D5 F4 Cb4 Ab2
		{"V調のV9 根音省略形", [4]uint8{72, 69, 66, 51}, "V調のV9(根音省略形・第3転回位置)"}, // C5 A4 Gb4 Eb3
	}
	for _, tt := range tests {
		if got := chordSymbol(tt.notes, ts); got != tt.want {
			t.Errorf("%s: chordSymbol(%v) = %s, want %s", tt.name, tt.notes, got, tt.want)
		}
	}
}

func TestChordSymbolAMollDiminishedII(t *testing.T) {
	// a-moll の II（減三和音）は第1転回位置で頻出
	ts := templatesFor(t, "a-moll.mid")
	if got := chordSymbol([4]uint8{71, 65, 62, 50}, ts); got != "II(第1転回位置)" { // B4 F4 D4 D3
		t.Errorf("chordSymbol = %s, want II(第1転回位置)", got)
	}
}

func TestHarmonyReportChordSymbols(t *testing.T) {
	s := buildChoraleSMF(t,
		[4]uint8{72, 67, 64, 60}, // I 基本位置
		[4]uint8{74, 71, 67, 53}, // V7 第3転回位置
	)
	report, ok := harmonyReport(s, "c-dur.mid")
	if !ok {
		t.Fatal("harmonyReport: not recognized as a 4-voice chorale")
	}
	if !strings.Contains(report, "I(基本位置) S:C5") {
		t.Errorf("report should contain chord symbol for the first chord, got:\n%s", report)
	}
	if !strings.Contains(report, "V7(第3転回位置) S:D5") {
		t.Errorf("report should contain chord symbol for the second chord, got:\n%s", report)
	}

	// 調が読み取れない場合は和音記号を付けない
	noKey, ok := harmonyReport(s, "nokey.mid")
	if !ok {
		t.Fatal("harmonyReport: not recognized as a 4-voice chorale")
	}
	if strings.Contains(noKey, "基本位置") {
		t.Errorf("report without key should not contain chord symbols, got:\n%s", noKey)
	}
}
