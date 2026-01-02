package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/go-fitz"
	"github.com/joho/godotenv"
)

// This struct holds the text from one page of the PDF
type PageData struct {
	PageNumber int
	Text       string
}

// These structs match what the Gemini API expects
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

// These structs match what the Gemini API returns
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

func main() {
	// Step 1: Load the API key from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error: Could not load .env file. Make sure it exists!")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: GEMINI_API_KEY not found in .env file")
	}

	// Step 2: Get the PDF filename from command line
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <pdf-file>")
	}

	pdfPath := os.Args[1]

	// Check if the file exists
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		log.Fatalf("Error: PDF file not found: %s", pdfPath)
	}

	fmt.Printf("📄 Processing PDF: %s\n", pdfPath)
	fmt.Println("=====================================")

	// Step 3: Open the PDF file
	doc, err := fitz.New(pdfPath)
	if err != nil {
		log.Fatalf("Error opening PDF: %v", err)
	}
	defer doc.Close() // Make sure to close the file when we're done

	totalPages := doc.NumPage()
	fmt.Printf("📊 Total pages: %d\n\n", totalPages)

	// Step 4: Extract text from all pages using goroutines
	// This is where we demonstrate Go's concurrency!
	fmt.Println("🔄 Extracting text from pages (using goroutines)...")
	startTime := time.Now()

	// Create a slice to store all page texts
	pages := make([]PageData, totalPages)

	// WaitGroup helps us wait for all goroutines to finish
	var wg sync.WaitGroup

	// Mutex protects our shared data (the pages slice) from race conditions
	var mu sync.Mutex

	// Loop through each page and extract text in parallel
	for i := 0; i < totalPages; i++ {
		// Tell WaitGroup we're starting a new goroutine
		wg.Add(1)

		// Launch a goroutine for this page
		// This means all pages are processed at the same time!
		go func(pageIndex int) {
			// Tell WaitGroup this goroutine is done when we exit
			defer wg.Done()

			// Extract text from this specific page
			text, err := doc.Text(pageIndex)
			if err != nil {
				log.Printf("Warning: Error on page %d: %v", pageIndex+1, err)
				text = ""
			}

			// Lock the mutex before updating shared data
			// This prevents multiple goroutines from writing at the same time
			mu.Lock()
			pages[pageIndex] = PageData{
				PageNumber: pageIndex + 1,
				Text:       strings.TrimSpace(text),
			}
			mu.Unlock() // Unlock when done

			fmt.Printf("✅ Page %d: Text extracted\n", pageIndex+1)
		}(i) // Pass the page index to the goroutine
	}

	// Wait for ALL goroutines to finish before continuing
	// This is important - we need all pages extracted before we can continue
	wg.Wait()

	extractTime := time.Since(startTime)
	fmt.Printf("\n⏱️  Text extraction completed in: %v\n", extractTime)
	fmt.Printf("📝 All %d pages extracted concurrently using goroutines!\n\n", totalPages)

	// Step 5: Build one big prompt with ALL pages
	fmt.Println("🚀 Preparing to send all pages to Gemini API in ONE request...")
	fmt.Println("=====================================")

	var promptBuilder strings.Builder
	promptBuilder.WriteString("Please provide concise summaries for each page of this PDF document. For each page, provide a 2-3 sentence summary.\n\n")

	// Add each page to the prompt
	for _, page := range pages {
		if page.Text != "" {
			promptBuilder.WriteString(fmt.Sprintf("=== PAGE %d ===\n%s\n\n", page.PageNumber, page.Text))
		}
	}

	promptBuilder.WriteString("Please format your response as:\nPage 1: [summary]\nPage 2: [summary]\n...")

	prompt := promptBuilder.String()

	// Step 6: Make ONE API call with all pages
	apiStartTime := time.Now()
	summary, err := callGeminiAPI(apiKey, prompt)
	apiTime := time.Since(apiStartTime)

	if err != nil {
		log.Fatalf("❌ API Error: %v", err)
	}

	fmt.Printf("✅ API call completed in: %v\n\n", apiTime)

	// Step 7: Save the summary to a file
	outputFile := strings.TrimSuffix(pdfPath, ".pdf") + "_summary.txt"
	err = saveToFile(outputFile, summary, pdfPath, totalPages)
	if err != nil {
		log.Printf("Warning: Could not save to file: %v", err)
	} else {
		fmt.Printf("💾 Summary saved to: %s\n\n", outputFile)
	}

	// Step 8: Display the summary
	fmt.Println("==================================================")
	fmt.Println("📋 SUMMARY")
	fmt.Println("==================================================")
	fmt.Println(summary)
}

// callGeminiAPI sends our prompt to Gemini and gets back a summary
func callGeminiAPI(apiKey, prompt string) (string, error) {
	// Use a simple model name that works
	modelName := "gemini-2.5-flash"
	apiVersion := "v1"

	// Build the request body in the format Gemini expects
	requestBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	// Convert our struct to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating JSON: %v", err)
	}

	// Build the API URL
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/%s/models/%s:generateContent?key=%s", apiVersion, modelName, apiKey)

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Set the content type header
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client with a long timeout (in case the PDF is large)
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse the JSON response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("error parsing response: %v", err)
	}

	// Extract the text from the response
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in API response")
	}

	return strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text), nil
}

// saveToFile writes the summary to a text file
func saveToFile(filename, summary, pdfPath string, totalPages int) error {
	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header information
	fmt.Fprintf(file, "PDF Summary Report\n")
	fmt.Fprintf(file, "==================\n\n")
	fmt.Fprintf(file, "Source PDF: %s\n", pdfPath)
	fmt.Fprintf(file, "Total Pages: %d\n", totalPages)
	fmt.Fprintf(file, "Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "%s\n\n", strings.Repeat("=", 80))

	// Write the actual summary
	fmt.Fprintf(file, "%s\n", summary)

	return nil
}
