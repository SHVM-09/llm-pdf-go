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
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("Warning: Could not load .env file. Using environment variables.")
		}
	}

	// Parse command line arguments
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <pdf-file>\n" +
			"Example: go run main.go ../design-analysis/v6truboEngine.pdf")
	}

	config := &Config{
		APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
		ModelName: "claude-3-5-haiku-20241022", // Using cheapest model
		PDFPath:   os.Args[1],
	}

	if config.APIKey == "" {
		log.Fatal("Error: ANTHROPIC_API_KEY not found in environment variables")
	}

	// Validate PDF file
	if _, err := os.Stat(config.PDFPath); os.IsNotExist(err) {
		log.Fatalf("Error: PDF file not found: %s", config.PDFPath)
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  DESIGN PDF ANALYSIS TOOL (ANTHROPIC)")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("\n📄 Processing: %s\n", filepath.Base(config.PDFPath))
	fmt.Printf("🤖 Model: %s\n", config.ModelName)
	pricing := GetPricing(config.ModelName)
	fmt.Printf("💰 Model Pricing: $%.2f/M input, $%.2f/M output\n\n",
		pricing.InputPricePerMTokens,
		pricing.OutputPricePerMTokens)

	startTime := time.Now()

	// Get total page count
	totalPages, err := getPageCount(config.PDFPath)
	if err != nil {
		log.Fatalf("Error getting page count: %v", err)
	}

	fmt.Printf("📊 Total pages: %d\n", totalPages)

	// Process each page individually for maximum detail extraction
	chunkSize := 1
	fmt.Printf("📦 Processing each page individually for complete data extraction\n\n")

	// Create temporary directory for chunk PDFs
	tempDir, err := os.MkdirTemp("", "pdf-chunks-*")
	if err != nil {
		log.Fatalf("Error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Split PDF into chunks
	chunks, err := splitPDFIntoChunks(config.PDFPath, tempDir, chunkSize, totalPages)
	if err != nil {
		log.Fatalf("Error splitting PDF: %v", err)
	}

	if chunkSize == 1 {
		fmt.Printf("✅ Created %d single-page PDF(s) for processing\n\n", len(chunks))
	} else {
		fmt.Printf("✅ Created %d chunk(s)\n\n", len(chunks))
	}

	// Process chunks with rate limiting
	// Rate limit: 400,000 input tokens per minute
	// Conservative estimate: ~80k tokens per single-page PDF (PDF + prompt)
	// Safe concurrent limit: 4-5 pages at a time to stay well under limit
	maxConcurrent := 4
	fmt.Printf("🚀 Processing pages with rate limiting (max %d concurrent requests)...\n", maxConcurrent)
	fmt.Println(strings.Repeat("-", 70))

	ctx := context.Background()
	results := make([]ChunkAnalysis, len(chunks))

	// Create a semaphore to limit concurrent requests
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, chunk := range chunks {
		wg.Add(1)
		go func(index int, path string, startPage, endPage int) {
			defer wg.Done()

			// Acquire semaphore (blocks if maxConcurrent requests are running)
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Release semaphore when done

			chunkStartTime := time.Now()
			if startPage == endPage {
				fmt.Printf("  🔄 Processing page %d...\n", startPage+1)
			} else {
				fmt.Printf("  🔄 Processing chunk %d (pages %d-%d)...\n", index+1, startPage+1, endPage+1)
			}

			// Retry logic for rate limit errors
			var analysis string
			var inputTokens, outputTokens int
			var err error
			maxRetries := 3
			retryDelay := 2 * time.Second

			for attempt := 0; attempt < maxRetries; attempt++ {
				analysis, inputTokens, outputTokens, err = analyzeChunk(ctx, config.APIKey, config.ModelName, path, startPage+1)

				if err == nil {
					break // Success
				}

				// Check if it's a rate limit error
				if strings.Contains(err.Error(), "rate_limit") || strings.Contains(err.Error(), "429") {
					if attempt < maxRetries-1 {
						waitTime := retryDelay * time.Duration(1<<attempt) // Exponential backoff
						fmt.Printf("  ⚠️  Rate limit hit for page %d, retrying in %v...\n", startPage+1, waitTime)
						time.Sleep(waitTime)
						continue
					}
				} else {
					break // Non-rate-limit error, don't retry
				}
			}

			chunkDuration := time.Since(chunkStartTime)

			mu.Lock()
			pricing := GetPricing(config.ModelName)
			inputCost := float64(inputTokens) / 1_000_000 * pricing.InputPricePerMTokens
			outputCost := float64(outputTokens) / 1_000_000 * pricing.OutputPricePerMTokens

			results[index] = ChunkAnalysis{
				ChunkNumber:    index + 1,
				StartPage:      startPage + 1,
				EndPage:        endPage + 1,
				Analysis:       analysis,
				InputTokens:    inputTokens,
				OutputTokens:   outputTokens,
				InputCost:      inputCost,
				OutputCost:     outputCost,
				TotalCost:      inputCost + outputCost,
				ProcessingTime: chunkDuration.String(),
				Timestamp:      time.Now(),
			}

			if err != nil {
				results[index].Error = err.Error()
				if startPage == endPage {
					fmt.Printf("  ❌ Page %d failed: %v\n", startPage+1, err)
				} else {
					fmt.Printf("  ❌ Chunk %d failed: %v\n", index+1, err)
				}
			} else {
				if startPage == endPage {
					fmt.Printf("  ✅ Page %d completed: %d input tokens, %d output tokens, $%.6f\n",
						startPage+1, inputTokens, outputTokens, results[index].TotalCost)
				} else {
					fmt.Printf("  ✅ Chunk %d completed: %d input tokens, %d output tokens, $%.6f\n",
						index+1, inputTokens, outputTokens, results[index].TotalCost)
				}
			}
			mu.Unlock()
		}(i, chunk.Path, chunk.StartPage, chunk.EndPage)
	}

	wg.Wait()

	// Calculate chunk totals
	var chunkInputTokens, chunkOutputTokens int
	var chunkInputCost, chunkOutputCost float64

	for _, result := range results {
		chunkInputTokens += result.InputTokens
		chunkOutputTokens += result.OutputTokens
		chunkInputCost += result.InputCost
		chunkOutputCost += result.OutputCost
	}

	// Skip consolidation - use individual page analyses directly
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  FINALIZING RESULTS")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("✅ Using individual page analyses (no consolidation needed)")
	fmt.Println("   All page-by-page details are preserved in the output")

	totalDuration := time.Since(startTime)

	// Calculate final totals (no consolidation costs)
	totalInputTokens := chunkInputTokens
	totalOutputTokens := chunkOutputTokens
	totalInputCost := chunkInputCost
	totalOutputCost := chunkOutputCost

	// Create full result (no consolidated analysis - using individual page analyses)
	fullResult := FullAnalysisResult{
		PDFPath:           config.PDFPath,
		TotalPages:        totalPages,
		TotalChunks:       len(chunks),
		Chunks:            results,
		Consolidated:      nil, // No consolidation - all details in individual page analyses
		TotalInputTokens:  totalInputTokens,
		TotalOutputTokens: totalOutputTokens,
		TotalInputCost:    totalInputCost,
		TotalOutputCost:   totalOutputCost,
		TotalCost:         totalInputCost + totalOutputCost,
		ProcessingTime:    totalDuration.String(),
		GeneratedAt:       time.Now(),
	}

	// Output results
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  FINAL ANALYSIS SUMMARY")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Page-by-Page Analysis:\n")
	fmt.Printf("  - Input Tokens:  %d\n", chunkInputTokens)
	fmt.Printf("  - Output Tokens: %d\n", chunkOutputTokens)
	fmt.Printf("  - Cost:          $%.6f\n", chunkInputCost+chunkOutputCost)
	fmt.Printf("TOTAL:\n")
	fmt.Printf("  - Input Tokens:  %d\n", totalInputTokens)
	fmt.Printf("  - Output Tokens: %d\n", totalOutputTokens)
	fmt.Printf("  - Total Cost:    $%.6f\n", totalInputCost+totalOutputCost)
	fmt.Printf("  - Processing Time: %s\n", totalDuration)
	fmt.Println(strings.Repeat("=", 70))

	// Save JSON output
	jsonFile := generateOutputFilename(config.PDFPath, "json")
	if err := saveJSONOutput(jsonFile, fullResult); err != nil {
		log.Printf("Warning: Could not save JSON output: %v", err)
	} else {
		fmt.Printf("\n💾 JSON results saved to: %s\n", jsonFile)
	}

	// Suggest HTML viewer
	fmt.Printf("\n🌐 View results in HTML: Open viewer.html in your browser and load %s\n", jsonFile)
}
