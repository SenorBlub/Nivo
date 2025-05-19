package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func Ask(query, collection string) (string, error) {
	if collection == "" {
		collection = "Database"
	}

	// 1. Call Chisel `/lookup`
	lookupPayload := map[string]string{
		"query":      query,
		"collection": collection,
	}
	lookupBody, _ := json.Marshal(lookupPayload)
	lookupResp, err := http.Post(ChiselIp+"/lookup", "application/json", bytes.NewBuffer(lookupBody))
	if err != nil {
		return "", fmt.Errorf("failed to lookup context: %w", err)
	}
	defer lookupResp.Body.Close()

	raw, _ := io.ReadAll(lookupResp.Body)
	var parsed struct {
		Chunks []struct {
			Text string `json:"text"`
		} `json:"chunks"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("invalid lookup result: %w", err)
	}

	// 2. Build context
	context := ""
	for i, chunk := range parsed.Chunks {
		if i >= 5 {
			break
		}
		context += fmt.Sprintf("- %s\n", chunk.Text)
	}

	// 3. Generate answer with Groq LLaMA 3 70B
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY not set")
	}

	request := map[string]interface{}{
		"model": "llama3-70b-8192",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant who answers questions based only on the context provided."},
			{"role": "user", "content": fmt.Sprintf("Here is some context:\n%s\n\nQuestion: %s", context, query)},
		},
		"temperature": 0.7,
	}

	reqBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Groq: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("Groq response decode failed: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return result.Choices[0].Message.Content, nil
}
