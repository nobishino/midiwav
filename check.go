package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"gitlab.com/gomidi/midi/v2/smf"

	"github.com/nobishino/midiwav/harmony"
)

// runCheck は check サブコマンドを実行する。watch モードと違い副作用を持たず
// （WAV合成・楽譜生成・Discord投稿を行わず）、4声体和声の分析だけを行って
// 結果を stdout に書く。
//
// 終了コード: 0=禁則なし, 1=禁則(⚠)が1件以上, 2=実行エラー。
func runCheck(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(stderr)
	keyName := fs.String("key", "", `調のドイツ語音名。ファイル名の規約より優先する（例: "es-moll"）`)
	format := fs.String("format", "text", `出力形式: "text" または "json"`)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: midiwav check [-key <調>] [-format text|json] <file.mid>...")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() == 0 {
		fs.Usage()
		return 2
	}
	if *format != "text" && *format != "json" {
		fmt.Fprintf(stderr, "midiwav check: -format %q は不正です（\"text\" か \"json\"）\n", *format)
		return 2
	}
	var keyOverride *harmony.Key
	if *keyName != "" {
		k, ok := harmony.ParseKey(*keyName)
		if !ok {
			fmt.Fprintf(stderr, "midiwav check: -key %q は不正です（ドイツ語音名で例: \"es-moll\"）\n", *keyName)
			return 2
		}
		keyOverride = &k
	}

	code := 0
	var results []checkResult
	for _, path := range fs.Args() {
		report, err := analyzeFile(path, keyOverride)
		if err != nil {
			fmt.Fprintf(stderr, "midiwav check: %s: %v\n", path, err)
			return 2
		}
		if countIssues(report, true) > 0 {
			code = 1
		}
		if *format == "json" {
			results = append(results, newCheckResult(path, report))
			continue
		}
		if fs.NArg() > 1 {
			fmt.Fprintf(stdout, "=== %s ===\n", path)
		}
		fmt.Fprint(stdout, report.Format())
	}
	if *format == "json" {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		if err := enc.Encode(results); err != nil {
			fmt.Fprintf(stderr, "midiwav check: %v\n", err)
			return 2
		}
	}
	return code
}

// analyzeFile はMIDIファイル1つを読み、4声体和声として分析する。
// keyOverride が nil の場合はファイル名から調を読み取る。
func analyzeFile(path string, keyOverride *harmony.Key) (*harmony.Report, error) {
	smfData, err := smf.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read MIDI file: %w", err)
	}
	key := keyOverride
	if key == nil {
		if k, ok := harmony.ParseKeyFromFilename(path); ok {
			key = &k
		}
	}
	report, ok := harmony.Analyze(smfData, key)
	if !ok {
		return nil, fmt.Errorf("not a four-part chorale (音符のあるトラックがちょうど4本必要です)")
	}
	return report, nil
}

func countIssues(r *harmony.Report, warn bool) int {
	n := 0
	for _, is := range r.Issues {
		if is.Warn == warn {
			n++
		}
	}
	return n
}

// checkResult は -format json の1ファイル分の出力。トップレベルは
// ファイル数によらず checkResult の配列として出力される。
type checkResult struct {
	File string `json:"file"`
	// Key は "Eb-moll" のような調名。ファイル名からも -key からも
	// 調が得られなかった場合は null（このとき綴り依存の検査はスキップされる）。
	Key     *string      `json:"key"`
	Chords  []checkChord `json:"chords"`
	Issues  []checkIssue `json:"issues"`
	Summary checkSummary `json:"summary"`
}

type checkChord struct {
	Pos string `json:"pos"` // 例: "1小節1拍目"
	// Symbol は和音記号（例: "V₉(根省)の第2転回位置"）。調が不明なら省略。
	Symbol string     `json:"symbol,omitempty"`
	Name   string     `json:"name"` // コードネーム（例: "G7/F"）
	Notes  checkNotes `json:"notes"`
}

type checkNotes struct {
	S string `json:"S"`
	A string `json:"A"`
	T string `json:"T"`
	B string `json:"B"`
}

type checkIssue struct {
	Level   string `json:"level"` // "warning"(⚠ 禁則) | "info"(ℹ 参考)
	Message string `json:"message"`
}

type checkSummary struct {
	Warnings int `json:"warnings"`
	Infos    int `json:"infos"`
}

func newCheckResult(path string, r *harmony.Report) checkResult {
	res := checkResult{File: path, Chords: []checkChord{}, Issues: []checkIssue{}}
	if r.Key != nil {
		s := r.Key.String()
		res.Key = &s
	}
	for _, c := range r.Chords {
		res.Chords = append(res.Chords, checkChord{
			Pos:    c.Pos,
			Symbol: c.Symbol,
			Name:   c.Name,
			Notes:  checkNotes{S: c.Notes[0], A: c.Notes[1], T: c.Notes[2], B: c.Notes[3]},
		})
	}
	for _, is := range r.Issues {
		level := "info"
		if is.Warn {
			level = "warning"
		}
		res.Issues = append(res.Issues, checkIssue{Level: level, Message: is.Msg})
	}
	res.Summary = checkSummary{
		Warnings: countIssues(r, true),
		Infos:    countIssues(r, false),
	}
	return res
}
