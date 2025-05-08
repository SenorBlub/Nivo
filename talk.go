package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const groqTalkAPI = "https://api.groq.com/openai/v1/chat/completions"
const languageModel = "llama3-70b-8192"

// Talk takes a writing prompt or plan and uses LLaMA 3 70B to generate the actual content.
func Talk(prompt string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY environment variable is not set")
	}

	request := map[string]interface{}{
		"model": languageModel,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant and skilled writer. Generate content based on the prompt or structure you're given."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.85,
		"max_tokens":  4096,
		"top_p":       1,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", groqTalkAPI, bytes.NewBuffer(body))
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
		return "", fmt.Errorf("no response from Groq language model")
	}

	return result.Choices[0].Message.Content, nil
}

func TalkDocumentation(plan string) (string, error) {
	return Talk(fmt.Sprintf(`You're writing documentation based on the following detailed plan:

%s

Please write clear, concise, and well-structured documentation according to the above. Use markdown formatting where appropriate.`, plan))
}
