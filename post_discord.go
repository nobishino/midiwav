package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func postToDiscord(webhookURL string, fs ...*os.File) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL is empty")
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// payload_json
	payload := map[string]any{
		"content": "MIDIファイルとWAVファイルを保存しました",
	}
	pb, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := w.WriteField("payload_json", string(pb)); err != nil {
		return err
	}

	for i, f := range fs {
		// ファイル先頭に戻す
		if _, err := f.Seek(0, 0); err != nil {
			return err
		}

		// file
		fw, err := w.CreateFormFile(fmt.Sprintf("file[%d]", i), filepath.Base(f.Name()))
		if err != nil {
			return err
		}
		if _, err := io.Copy(fw, f); err != nil {
			return err
		}
	}

	if err := w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", webhookURL, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("discord error: %s\n%s", resp.Status, string(body))
	}

	fmt.Println("OK:", resp.Status)
	return nil
}
