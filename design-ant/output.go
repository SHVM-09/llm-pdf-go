package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// encodeBase64 encodes bytes to base64 string
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// generateOutputFilename creates an output filename based on input PDF
func generateOutputFilename(pdfPath, format string) string {
	base := filepath.Base(pdfPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	return fmt.Sprintf("%s_analysis.%s", name, format)
}

// saveJSONOutput saves results to JSON file
func saveJSONOutput(filename string, result FullAnalysisResult) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

