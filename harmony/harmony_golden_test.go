package harmony

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/gomidi/midi/v2/smf"
)

var update = flag.Bool("update", false, "update golden files")

// harmonyReport はファイル名から調を読み、分析して整形するテストヘルパー。
// main パッケージでの利用と同じ組み立てを再現する。
func harmonyReport(s *smf.SMF, path string) (string, bool) {
	var key *Key
	if k, ok := ParseKeyFromFilename(path); ok {
		key = &k
	}
	r, ok := Analyze(s, key)
	if !ok {
		return "", false
	}
	return r.Format(), true
}

// TestHarmonyReportGolden は testdata/*.mid に対する添削レポートを
// ゴールデンファイル（<case>.golden.txt）と照合する。調はファイル名で指定する
// （例: parallel-fifth_c-dur.mid）。バグ再現用のMIDIファイルをこのディレクトリに
// 置き、go test -update でゴールデンファイルを生成・更新できる。
func TestHarmonyReportGolden(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("testdata", "*.mid"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no MIDI files found in testdata")
	}
	for _, path := range paths {
		name := strings.TrimSuffix(filepath.Base(path), ".mid")
		t.Run(name, func(t *testing.T) {
			smfData, err := smf.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read SMF: %v", err)
			}
			report, ok := harmonyReport(smfData, path)
			if !ok {
				t.Fatal("harmonyReport: not recognized as a 4-voice chorale")
			}

			goldenPath := strings.TrimSuffix(path, ".mid") + ".golden.txt"
			if *update {
				if err := os.WriteFile(goldenPath, []byte(report), 0644); err != nil {
					t.Fatalf("failed to update golden file: %v", err)
				}
				t.Logf("updated golden file: %s", goldenPath)
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("failed to read golden file: %v", err)
			}
			if report != string(want) {
				t.Errorf("report differs from golden file %s\n--- got ---\n%s--- want ---\n%s", goldenPath, report, want)
			}
		})
	}
}
