package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// analyzeChunk sends a PDF chunk to Anthropic API and returns analysis
func analyzeChunk(ctx context.Context, apiKey, modelName, chunkPath string, pageNumber int) (string, int, int, error) {
	// Read PDF chunk file directly
	pdfBytes, err := os.ReadFile(chunkPath)
	if err != nil {
		return "", 0, 0, fmt.Errorf("error reading PDF chunk: %v", err)
	}

	// Encode PDF to base64
	pdfBase64 := encodeBase64(pdfBytes)

	// Create request payload with PDF as document
	requestBody := map[string]interface{}{
		"model":      modelName,
		"max_tokens": 8192, // Increased to allow comprehensive analysis without truncation
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "document",
						"source": map[string]interface{}{
							"type":       "base64",
							"media_type": "application/pdf",
							"data":       pdfBase64,
						},
					},
					{
						"type": "text",
						"text": generateAnalysisPrompt(pageNumber),
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", 0, 0, fmt.Errorf("error marshaling request: %v", err)
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, 0, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, 0, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, 0, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", 0, 0, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResponse struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", 0, 0, fmt.Errorf("error parsing response: %v", err)
	}

	analysis := ""
	if len(apiResponse.Content) > 0 {
		analysis = apiResponse.Content[0].Text
	}

	return analysis, apiResponse.Usage.InputTokens, apiResponse.Usage.OutputTokens, nil
}
