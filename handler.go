package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Common structure to respond with simple JSON.
func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ImageBase64 string `json:"image"`
		Name        string `json:"name"`
		Origin      string `json:"origin"`
		Collection  string `json:"collection,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.ImageBase64 == "" {
		http.Error(w, "Missing image", http.StatusBadRequest)
		return
	}

	text, err := callGlintView(input.ImageBase64, input.Origin, input.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Image processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	result, err := postToChiselChunk(text, input.Origin, input.Collection)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to chunk result: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func listenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AudioBase64 string `json:"audio"`
		Name        string `json:"name"`
		Origin      string `json:"origin"`
		Collection  string `json:"collection,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.AudioBase64 == "" {
		http.Error(w, "Missing audio", http.StatusBadRequest)
		return
	}

	text, err := callResonoListen(input.AudioBase64, input.Origin, input.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Audio transcription failed: %v", err), http.StatusInternalServerError)
		return
	}

	result, err := postToChiselChunk(text, input.Origin, input.Collection)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to chunk result: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Text       string `json:"text"`
		Origin     string `json:"origin"`
		Collection string `json:"collection,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Text == "" {
		http.Error(w, "Missing text", http.StatusBadRequest)
		return
	}

	result, err := postToChiselChunk(input.Text, input.Origin, input.Collection)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to store in Chisel: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func documentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Collection string `json:"collection,omitempty"`
		Prompt     string `json:"prompt"`
	}
	fmt.Println("handler reached")

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Prompt == "" {
		http.Error(w, "Invalid JSON or missing 'prompt' field", http.StatusBadRequest)
		return
	}

	// Set default collection if not provided.
	if input.Collection == "" {
		input.Collection = "Memory"
	}

	// Phase 1: Plan the documentation.
	plan, err := ThinkDocumentation(input.Prompt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Planning failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Phase 2: Generate the documentation.
	output, err := TalkDocumentation(plan)
	if err != nil {
		http.Error(w, fmt.Sprintf("Generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{
		"plan":          plan,
		"documentation": output,
	})
}

func thinkHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Collection string `json:"collection,omitempty"`
		Prompt     string `json:"prompt"`
	}
	fmt.Println("handler reached")

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Prompt == "" {
		http.Error(w, "Invalid JSON or missing 'prompt' field", http.StatusBadRequest)
		return
	}
	// Set default collection if not provided.
	if input.Collection == "" {
		input.Collection = "Memory"
	}

	response, err := Think(input.Prompt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Thinking failed: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{
		"plan": response,
	})
}

func explainHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Collection string `json:"collection,omitempty"`
		Prompt     string `json:"prompt"`
	}
	fmt.Println("handler reached")

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Prompt == "" {
		http.Error(w, "Invalid JSON or missing 'prompt' field", http.StatusBadRequest)
		return
	}
	// Set default collection if not provided.
	if input.Collection == "" {
		input.Collection = "Memory"
	}

	response, err := Talk(input.Prompt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Explanation failed: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{
		"text": response,
	})
}

func rememberHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Text       string   `json:"text"`
		Subject    string   `json:"subject"`
		Origin     string   `json:"origin,omitempty"`
		Tags       []string `json:"tags,omitempty"`
		Time       string   `json:"timestamp,omitempty"` // ISO 8601
		Collection string   `json:"collection,omitempty"`
	}
	fmt.Println("handler reached")

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Text == "" || input.Subject == "" {
		http.Error(w, "Missing 'text' or 'subject'", http.StatusBadRequest)
		return
	}
	// Set default collection if not provided.
	if input.Collection == "" {
		input.Collection = "Memory"
	}

	// Build chunk data to send to Chisel `/chunk`.
	payload := map[string]interface{}{
		"text":       input.Text,
		"origin":     input.Origin,
		"collection": input.Collection,
		"metadata": map[string]interface{}{
			"subject":         input.Subject,
			"tags":            input.Tags,
			"manually_stored": true,
		},
	}
	if input.Time != "" {
		payload["timestamp"] = input.Time
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:8080/chunk", "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to remember: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	io.Copy(w, resp.Body)
}

func askHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Query      string `json:"query"`
		Collection string `json:"collection,omitempty"`
	}
	fmt.Println("handler reached")

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Query == "" {
		http.Error(w, "Invalid JSON or missing 'query'", http.StatusBadRequest)
		return
	}

	// Set default collection if not provided.
	if input.Collection == "" {
		input.Collection = "Memory"
	}

	answer, err := Ask(input.Query, input.Collection)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ask failed: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{
		"answer": answer,
	})
}
