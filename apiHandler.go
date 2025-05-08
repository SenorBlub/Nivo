package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func postJSON(url string, payload interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("POST %s failed: %s", url, string(body))
	}

	return io.ReadAll(resp.Body)
}

func postToChiselChunk(text, origin, collection string) ([]byte, error) {
	payload := map[string]interface{}{
		"text":       text,
		"origin":     origin,
		"collection": collection,
	}
	return postJSON("http://localhost:8080/chunk", payload)
}

func callGlintView(image, origin, name string) (string, error) {
	payload := map[string]string{
		"data":   image,
		"origin": origin,
		"name":   name,
	}
	data, err := postJSON("http://localhost:8060/view", payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func callResonoListen(audio, origin, name string) (string, error) {
	payload := map[string]string{
		"data":   audio,
		"origin": origin,
		"name":   name,
	}

	data, err := postJSON("http://localhost:8040/transcribe", payload)
	if err != nil {
		return "", err
	}

	var response struct {
		Transcription string `json:"transcription"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return "", fmt.Errorf("failed to parse transcription response: %w", err)
	}

	return response.Transcription, nil
}
