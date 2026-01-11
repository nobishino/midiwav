package main

import (
	"os"
	"path/filepath"
	"strings"
)

// hoge.midであって、次のいずれかに該当するMIDIファイルを探し、返却する。
//   - hoge.wavが存在しない
//   - hoge.midよりhoge.wavの方が古い
func findUnprocessedMIDIFiles(dir string, recursive bool) ([]string, error) {
	var unprocessed []string
	if recursive {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(path) == ".mid" {
				wavPath := strings.TrimSuffix(path, ".mid") + ".wav"
				if wavInfo, err := os.Stat(wavPath); os.IsNotExist(err) {
					unprocessed = append(unprocessed, path)
				} else if info.ModTime().After(wavInfo.ModTime()) {
					unprocessed = append(unprocessed, path)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if filepath.Ext(entry.Name()) == ".mid" {
				path := filepath.Join(dir, entry.Name())
				wavPath := strings.TrimSuffix(path, ".mid") + ".wav"
				info, err := entry.Info()
				if err != nil {
					return nil, err
				}
				if wavInfo, err := os.Stat(wavPath); os.IsNotExist(err) {
					unprocessed = append(unprocessed, path)
				} else if info.ModTime().After(wavInfo.ModTime()) {
					unprocessed = append(unprocessed, path)
				}
			}
		}
	}
	return unprocessed, nil
}
