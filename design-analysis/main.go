package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

// Config holds application configuration
type Config struct {
	APIKey      string
	ModelName   string
	PDFPath     string
	OutputLevel string // executive, technical, detailed
}

// DesignAnalysisResult holds the structured analysis result
type DesignAnalysisResult struct {
	ExecutiveSummary string
	TechnicalDetails string
	BillOfMaterials  string
	Specifications   string
	Drawings         string
	Recommendations  string
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		// Try loading from parent directory
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("Warning: Could not load .env file. Using environment variables.")
		}
	}

	// Parse command line arguments
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <pdf-file> [output-level]\n" +
			"Output levels: executive (default), technical, detailed")
	}

	config := &Config{
		APIKey:      os.Getenv("GEMINI_API_KEY"),
		ModelName:   "gemini-2.5-flash-lite", // Using stable, free-tier compatible model
		PDFPath:     os.Args[1],
		OutputLevel: "executive",
	}

	if config.APIKey == "" {
		log.Fatal("Error: GEMINI_API_KEY not found in environment variables")
	}

	if len(os.Args) >= 3 {
		config.OutputLevel = os.Args[2]
	}

	// Validate PDF file
	if _, err := os.Stat(config.PDFPath); os.IsNotExist(err) {
		log.Fatalf("Error: PDF file not found: %s", config.PDFPath)
	}

	fmt.Println(strings.Repeat("=", 62))
	fmt.Println("  CAD/DESIGN PDF ANALYSIS TOOL")
	fmt.Println(strings.Repeat("=", 62))
	fmt.Printf("\n📄 Processing: %s\n", filepath.Base(config.PDFPath))
	fmt.Printf("📊 Output Level: %s\n", config.OutputLevel)
	fmt.Printf("🤖 Model: %s\n\n", config.ModelName)

	startTime := time.Now()

	// Read entire PDF file
	fmt.Println("📖 Reading PDF file...")
	pdfBytes, err := os.ReadFile(config.PDFPath)
	if err != nil {
		log.Fatalf("Error reading PDF file: %v", err)
	}
	fmt.Printf("✅ PDF loaded: %d bytes\n\n", len(pdfBytes))

	// Initialize Gemini client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: config.APIKey})
	if err != nil {
		log.Fatalf("Error creating Gemini client: %v", err)
	}

	// Generate comprehensive prompt based on output level
	prompt := GeneratePrompt(config.OutputLevel)

	// Send entire PDF to LLM
	fmt.Println("🚀 Sending PDF to LLM for analysis...")
	fmt.Println("   (This may take a moment for large PDFs)")
	fmt.Println()

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

	apiStartTime := time.Now()
	result, err := client.Models.GenerateContent(ctx, config.ModelName, content, nil)
	if err != nil {
		log.Fatalf("❌ API Error: %v", err)
	}

	apiDuration := time.Since(apiStartTime)
	totalDuration := time.Since(startTime)

	fmt.Printf("✅ Analysis completed in: %v\n", apiDuration)
	fmt.Printf("⏱️  Total time: %v\n\n", totalDuration)

	// Format and display results
	analysis := result.Text()
	formattedOutput := FormatOutput(analysis, config.OutputLevel)

	fmt.Println(strings.Repeat("=", 62))
	fmt.Println("  ANALYSIS RESULTS")
	fmt.Println(strings.Repeat("=", 62))
	fmt.Println()
	fmt.Println(formattedOutput)

	// Save to file
	outputFile := generateOutputFilename(config.PDFPath, config.OutputLevel)
	if err := os.WriteFile(outputFile, []byte(formattedOutput), 0644); err != nil {
		log.Printf("Warning: Could not save output to file: %v", err)
	} else {
		fmt.Printf("\n💾 Results saved to: %s\n", outputFile)
	}
}

// generateOutputFilename creates an output filename based on input PDF
func generateOutputFilename(pdfPath, outputLevel string) string {
	base := filepath.Base(pdfPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	return fmt.Sprintf("%s_analysis_%s.txt", name, outputLevel)
}
