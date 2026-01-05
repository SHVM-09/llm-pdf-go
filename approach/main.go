package main

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/go-fitz"
	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

type PageResult struct {
	PageNumber int
	Summary    string
	Error      error
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
		log.Fatal("Usage: go run approach/main.go <pdf-file>")
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
	maxPages := 10
	if totalPages < maxPages {
		maxPages = totalPages
	}

	fmt.Printf("📊 Total pages: %d (processing first %d pages)\n\n", totalPages, maxPages)

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		log.Fatalf("Error creating Gemini client: %v", err)
	}

	results := make([]PageResult, maxPages)
	batchSize := 5
	startTime := time.Now()
	var mu sync.Mutex

	fmt.Printf("🚀 Processing pages in batches of %d...\n", batchSize)
	fmt.Println("=====================================")

	for batchStart := 0; batchStart < maxPages; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > maxPages {
			batchEnd = maxPages
		}

		fmt.Printf("\n📦 Processing batch: Pages %d-%d\n", batchStart+1, batchEnd)

		var wg sync.WaitGroup
		for i := batchStart; i < batchEnd; i++ {
			wg.Add(1)
			go func(pageIndex int) {
				defer wg.Done()

				pageNum := pageIndex + 1
				fmt.Printf("  🔄 Processing page %d...\n", pageNum)

				var pageResult PageResult
				pageResult.PageNumber = pageNum

				img, err := doc.Image(pageIndex)
				if err != nil {
					pageResult.Error = fmt.Errorf("error rendering page: %v", err)
					fmt.Printf("  ❌ Page %d: Error rendering\n", pageNum)
				} else {
					// Send image as-is without resizing or encoding optimization
					var imgBuf bytes.Buffer
					if err := png.Encode(&imgBuf, img); err != nil {
						pageResult.Error = fmt.Errorf("error encoding image: %v", err)
						fmt.Printf("  ❌ Page %d: Error encoding\n", pageNum)
					} else {
						prompt := fmt.Sprintf("Please provide a concise 2-3 sentence summary of this PDF page %d.", pageNum)

						content := []*genai.Content{
							{
								Parts: []*genai.Part{
									{
										InlineData: &genai.Blob{
											MIMEType: "image/png",
											Data:     imgBuf.Bytes(),
										},
									},
									{
										Text: prompt,
									},
								},
							},
						}

						result, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash-lite", content, nil)
						if err != nil {
							pageResult.Error = fmt.Errorf("API error: %v", err)
							fmt.Printf("  ❌ Page %d: API error\n", pageNum)
						} else {
							pageResult.Summary = result.Text()
							fmt.Printf("  ✅ Page %d: Summary received\n", pageNum)
						}
					}
				}

				// Single lock/unlock for writing result
				mu.Lock()
				results[pageIndex] = pageResult
				mu.Unlock()
			}(i)
		}

		wg.Wait()
		fmt.Printf("✅ Batch complete: Pages %d-%d\n", batchStart+1, batchEnd)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\n⏱️  Total processing time: %v\n", elapsed)
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("📋 SUMMARIES")
	fmt.Println(strings.Repeat("=", 50) + "\n")

	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("Page %d: ❌ Error - %v\n\n", result.PageNumber, result.Error)
		} else {
			fmt.Printf("Page %d:\n%s\n\n", result.PageNumber, result.Summary)
		}
	}
}
