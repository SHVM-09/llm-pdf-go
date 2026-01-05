package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
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
		log.Fatal("Usage: go run approach2/main.go <pdf-file>")
	}

	pdfPath := os.Args[1]
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		log.Fatalf("Error: PDF file not found: %s", pdfPath)
	}

	fmt.Printf("📄 Processing PDF: %s\n", pdfPath)
	fmt.Println("=====================================")

	// Get total pages
	totalPages, err := getPageCount(pdfPath)
	if err != nil {
		log.Fatalf("Error getting page count: %v", err)
	}

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

	// Create temp directory for single-page PDFs
	tempDir, err := os.MkdirTemp("", "pdf_pages_*")
	if err != nil {
		log.Fatalf("Error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

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

				// Extract single page PDF
				singlePagePDF, err := extractPagePDF(pdfPath, pageNum, tempDir)
				if err != nil {
					pageResult.Error = fmt.Errorf("error extracting page PDF: %v", err)
					fmt.Printf("  ❌ Page %d: Error extracting PDF: %v\n", pageNum, err)
				} else {
					// Check if file exists
					if _, err := os.Stat(singlePagePDF); os.IsNotExist(err) {
						pageResult.Error = fmt.Errorf("extracted PDF file not found: %s", singlePagePDF)
						fmt.Printf("  ❌ Page %d: Extracted PDF file not found\n", pageNum)
					} else {
						// Read PDF bytes
						pdfBytes, err := os.ReadFile(singlePagePDF)
						if err != nil {
							pageResult.Error = fmt.Errorf("error reading PDF: %v", err)
							fmt.Printf("  ❌ Page %d: Error reading PDF\n", pageNum)
						} else {
							fmt.Printf("  📄 Page %d: PDF extracted (%d bytes)\n", pageNum, len(pdfBytes))

							prompt := fmt.Sprintf("Please provide a concise 2-3 sentence summary of this PDF page %d.", pageNum)

							// Send PDF to Gemini
							content := []*genai.Content{
								{
									Parts: []*genai.Part{
										{
											InlineData: &genai.Blob{
												MIMEType: "application/pdf",
												Data:     pdfBytes,
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

// getPageCount gets the total number of pages using pdfcpu
func getPageCount(pdfPath string) (int, error) {
	file, err := os.Open(pdfPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	conf := model.NewDefaultConfiguration()
	pages, err := api.PageCount(file, conf)
	if err != nil {
		return 0, fmt.Errorf("error getting page count: %v", err)
	}
	return pages, nil
}

// extractPagePDF extracts a single page from PDF using pdfcpu
func extractPagePDF(pdfPath string, pageNum int, tempDir string) (string, error) {
	// pdfcpu creates files with pattern based on input filename
	// Use a simple base name
	baseName := "page"

	// Open input file
	inFile, err := os.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("error opening PDF: %v", err)
	}

	// Create page selection: "1" means page 1, "2" means page 2, etc.
	pageSelection := []string{fmt.Sprintf("%d", pageNum)}

	// Extract the page - pdfcpu creates: {baseName}_{pageNum}.pdf
	conf := model.NewDefaultConfiguration()
	err = api.ExtractPages(inFile, tempDir, baseName, pageSelection, conf)

	// Close file after extraction
	inFile.Close()

	if err != nil {
		return "", fmt.Errorf("pdfcpu ExtractPages error: %v", err)
	}

	// pdfcpu creates file as: {baseName}_page_{pageNum}.pdf
	actualFileName := fmt.Sprintf("%s_page_%d.pdf", baseName, pageNum)
	outputPath := filepath.Join(tempDir, actualFileName)

	// Verify file was created - if not, list directory to see what was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		files, _ := os.ReadDir(tempDir)
		fileList := []string{}
		for _, f := range files {
			fileList = append(fileList, f.Name())
		}
		return "", fmt.Errorf("extracted file not found at %s. Created files: %v", outputPath, fileList)
	}

	return outputPath, nil
}
