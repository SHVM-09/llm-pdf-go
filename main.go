package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/go-fitz"
	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

type PageData struct {
	PageNumber int
	Text       string
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error: Could not load .env file. Make sure it exists!")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: GEMINI_API_KEY not found in .env file")
	}

	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <pdf-file>")
	}

	pdfPath := os.Args[1]
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		log.Fatalf("Error: PDF file not found: %s", pdfPath)
	}

	fmt.Printf("📄 Processing PDF: %s\n", pdfPath)
	fmt.Println("=====================================")

	doc, err := fitz.New(pdfPath)
	if err != nil {
		log.Fatalf("Error opening PDF: %v", err)
	}
	defer doc.Close()

	totalPages := doc.NumPage()
	fmt.Printf("📊 Total pages: %d\n\n", totalPages)

	fmt.Println("🔄 Extracting text from pages (using goroutines)...")
	startTime := time.Now()
	pages := make([]PageData, totalPages)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < totalPages; i++ {
		wg.Add(1)
		go func(pageIndex int) {
			defer wg.Done()
			text, err := doc.Text(pageIndex)
			if err != nil {
				log.Printf("Warning: Error on page %d: %v", pageIndex+1, err)
				text = ""
			}
			mu.Lock()
			pages[pageIndex] = PageData{
				PageNumber: pageIndex + 1,
				Text:       strings.TrimSpace(text),
			}
			mu.Unlock()
			fmt.Printf("✅ Page %d: Text extracted\n", pageIndex+1)
		}(i)
	}
	wg.Wait()

	fmt.Printf("\n⏱️  Text extraction completed in: %v\n", time.Since(startTime))
	fmt.Printf("📝 All %d pages extracted concurrently using goroutines!\n\n", totalPages)

	fmt.Println("🚀 Preparing to send all pages to Gemini API in ONE request...")
	fmt.Println("=====================================")

	var promptBuilder strings.Builder
	promptBuilder.WriteString("Please provide concise summaries for each page of this PDF document. For each page, provide a 2-3 sentence summary.\n\n")
	for _, page := range pages {
		if page.Text != "" {
			promptBuilder.WriteString(fmt.Sprintf("=== PAGE %d ===\n%s\n\n", page.PageNumber, page.Text))
		}
	}
	promptBuilder.WriteString("Please format your response as:\nPage 1: [summary]\nPage 2: [summary]\n...")

	apiStartTime := time.Now()
	summary, err := callGeminiAPI(apiKey, promptBuilder.String())
	if err != nil {
		log.Fatalf("❌ API Error: %v", err)
	}
	fmt.Printf("✅ API call completed in: %v\n\n", time.Since(apiStartTime))

	fmt.Println("==================================================")
	fmt.Println("📋 SUMMARY")
	fmt.Println("==================================================")
	fmt.Println(summary)
}

func callGeminiAPI(apiKey, prompt string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return "", fmt.Errorf("error creating Gemini client: %v", err)
	}
	result, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash-lite", genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("error calling Gemini API: %v", err)
	}
	return result.Text(), nil
}
