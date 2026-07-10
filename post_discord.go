package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// discordFile is a file attached to a Discord webhook post.
type discordFile struct {
	name string
	r    io.Reader
}

func postToDiscord(webhookURL string, content string, files ...discordFile) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL is empty")
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// payload_json
	payload := map[string]any{
		"content": content,
	}
	pb, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := w.WriteField("payload_json", string(pb)); err != nil {
		return err
	}

	for i, f := range files {
		fw, err := w.CreateFormFile(fmt.Sprintf("file[%d]", i), f.name)
		if err != nil {
			return err
		}
		if _, err := io.Copy(fw, f.r); err != nil {
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
