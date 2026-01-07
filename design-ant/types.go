package main

import "time"

// Config holds application configuration
type Config struct {
	APIKey    string
	ModelName string
	PDFPath   string
}

// ChunkAnalysis represents analysis result for a PDF chunk
type ChunkAnalysis struct {
	ChunkNumber    int       `json:"chunk_number"`
	StartPage      int       `json:"start_page"`
	EndPage        int       `json:"end_page"`
	Analysis       string    `json:"analysis"`
	InputTokens    int       `json:"input_tokens"`
	OutputTokens   int       `json:"output_tokens"`
	InputCost      float64   `json:"input_cost"`
	OutputCost     float64   `json:"output_cost"`
	TotalCost      float64   `json:"total_cost"`
	ProcessingTime string    `json:"processing_time"`
	Error          string    `json:"error,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

// ConsolidatedAnalysis represents the final consolidated analysis
type ConsolidatedAnalysis struct {
	Analysis       string    `json:"analysis"`
	InputTokens    int       `json:"input_tokens"`
	OutputTokens   int       `json:"output_tokens"`
	InputCost      float64   `json:"input_cost"`
	OutputCost     float64   `json:"output_cost"`
	TotalCost      float64   `json:"total_cost"`
	ProcessingTime string    `json:"processing_time"`
	Timestamp      time.Time `json:"timestamp"`
}

// FullAnalysisResult represents the complete analysis result
type FullAnalysisResult struct {
	PDFPath           string                `json:"pdf_path"`
	TotalPages        int                   `json:"total_pages"`
	TotalChunks       int                   `json:"total_chunks"`
	Chunks            []ChunkAnalysis       `json:"chunks"`
	Consolidated      *ConsolidatedAnalysis `json:"consolidated_analysis,omitempty"`
	TotalInputTokens  int                   `json:"total_input_tokens"`
	TotalOutputTokens int                   `json:"total_output_tokens"`
	TotalInputCost    float64               `json:"total_input_cost"`
	TotalOutputCost   float64               `json:"total_output_cost"`
	TotalCost         float64               `json:"total_cost"`
	ProcessingTime    string                `json:"processing_time"`
	GeneratedAt       time.Time             `json:"generated_at"`
}

// AnthropicPricing holds pricing information for different models
type AnthropicPricing struct {
	InputPricePerMTokens  float64 // Price per million input tokens
	OutputPricePerMTokens float64 // Price per million output tokens
}

// ChunkInfo holds information about a PDF chunk
type ChunkInfo struct {
	Path      string
	StartPage int
	EndPage   int
}

