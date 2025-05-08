package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const groqThinkAPI = "https://api.groq.com/openai/v1/chat/completions"
const reasoningModel = "deepseek-r1-distill-llama-70b"

// Think takes a goal or content description and asks the reasoning model to explain what should be written and how.
func Think(prompt string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY environment variable is not set")
	}

	systemPrompt := `You are a strategic assistant. Your job is to reason through how a piece of content should be written.
First, explain the purpose of the output. Then, explain how it should be structured or styled.
Focus on providing a clear and detailed explanation, not writing the content itself.`

	request := map[string]interface{}{
		"model": reasoningModel,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  1024,
		"top_p":       1,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", groqThinkAPI, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("Groq API error: %s", result.Error)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from Groq reasoning model")
	}

	return result.Choices[0].Message.Content, nil
}

func ThinkDocumentation(prompt string) (string, error) {
	return Think(fmt.Sprintf(`You're tasked with planning technical documentation.
The user provided this context: %q

Your job is to describe:
- What the documentation should include
- What order it should be in
- Why this structure makes sense
- Any examples or special formatting worth including

Do not write the documentation, just explain how it should be written and why.`, prompt))
}
