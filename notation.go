package main

// 楽譜ファイルの生成。和声分析結果からMusicXMLを作り、外部コマンドで
// 画像化する: Verovio でSVGへ、rsvg-convert でPNGへ変換する。
// 外部コマンドが見つからない場合は、その段階までのファイルにとどめる
// （Verovioなし → .musicxml のみ、rsvg-convertなし → .svg まで）。

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nobishino/midiwav/harmony"
)

// renderSVG は MusicXML を Verovio でSVGに変換する。
// 全小節が1ページに収まるよう、ページ高さを大きく取って内容に合わせて切り詰める。
func renderSVG(verovioPath string, musicXML []byte) ([]byte, error) {
	cmd := exec.Command(verovioPath, "--page-height", "60000", "--adjust-page-height", "-o", "-", "-")
	cmd.Stdin = bytes.NewReader(musicXML)
	var out, errBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errBuf
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("verovio: %w: %s", err, errBuf.String())
	}
	return out.Bytes(), nil
}

// svgToPNG は SVG を rsvg-convert 互換コマンドでPNGに変換する。
func svgToPNG(converterPath string, svg []byte) ([]byte, error) {
	cmd := exec.Command(converterPath, "--background-color=white")
	cmd.Stdin = bytes.NewReader(svg)
	var out, errBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errBuf
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", filepath.Base(converterPath), err, errBuf.String())
	}
	return out.Bytes(), nil
}

// notationAttachments は分析結果から楽譜ファイル（.musicxml と PNG/SVG画像）を
// 生成して srcPath の隣に保存し、Discord添付用のファイル一覧を返す。
func notationAttachments(r *harmony.Report, srcPath string, cfg Notation) []discordFile {
	musicXML, err := r.MusicXML()
	if err != nil {
		log.Println("failed to generate MusicXML:", err)
		return nil
	}
	base := strings.TrimSuffix(srcPath, ".mid")
	var files []discordFile
	add := func(ext string, data []byte) {
		if err := os.WriteFile(base+ext, data, 0644); err != nil {
			log.Println(err)
		}
		files = append(files, discordFile{
			name: filepath.Base(base) + ext,
			r:    bytes.NewReader(data),
		})
	}
	add(".musicxml", musicXML)

	verovio, err := exec.LookPath(cfg.VerovioPath)
	if err != nil {
		log.Printf("%s が見つからないため楽譜画像の生成をスキップします", cfg.VerovioPath)
		return files
	}
	svg, err := renderSVG(verovio, musicXML)
	if err != nil {
		log.Println(err)
		return files
	}
	if converter, err := exec.LookPath(cfg.SVG2PNGPath); err == nil {
		png, err := svgToPNG(converter, svg)
		if err == nil {
			add(".png", png)
			return files
		}
		log.Println(err)
	} else {
		log.Printf("%s が見つからないため楽譜はSVGのまま添付します", cfg.SVG2PNGPath)
	}
	add(".svg", svg)
	return files
}
