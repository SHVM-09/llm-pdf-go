# CAD/Design PDF Analysis Tool

A Go-based LLM application that analyzes CAD and design PDF documents and generates formatted, understandable output for different corporate levels.

## Features

- **Direct PDF Processing**: Sends entire PDF directly to LLM (no page-by-page extraction)
- **Multiple Output Levels**: 
  - **Executive**: High-level summary for management
  - **Technical**: Detailed technical information for engineers
  - **Detailed**: Comprehensive analysis with all details
- **Cost-Effective**: Uses cheaper LLM models (Gemini Flash)
- **Structured Output**: Well-formatted, markdown-based results
- **Generic Solution**: Works with various CAD/design PDF formats

## Prerequisites

- Go 1.24 or higher
- Gemini API key (get from [Google AI Studio](https://makersuite.google.com/app/apikey))

## Setup

1. **Install Dependencies**:
   ```bash
   cd llm-pdf-app/design-analysis
   go mod init design-analysis
   go get google.golang.org/genai
   go get github.com/joho/godotenv
   ```

2. **Set Environment Variable**:
   Create a `.env` file in the parent directory or set environment variable:
   ```bash
   export GEMINI_API_KEY="your-api-key-here"
   ```
   
   Or create `.env` file:
   ```
   GEMINI_API_KEY=your-api-key-here
   ```

## Usage

### Basic Usage
```bash
go run main.go prompts.go formatter.go <path-to-pdf>
```

### With Output Level
```bash
# Executive summary (default)
go run main.go prompts.go formatter.go v6truboEngine.pdf executive

# Technical details
go run main.go prompts.go formatter.go v6truboEngine.pdf technical

# Detailed analysis
go run main.go prompts.go formatter.go v6truboEngine.pdf detailed
```

### Example
```bash
go run main.go prompts.go formatter.go v6truboEngine.pdf technical
```

## Output

The tool generates:
1. **Console Output**: Formatted analysis displayed in terminal
2. **Text File**: Analysis saved as `{pdf-name}_analysis_{level}.txt`

## Cost Optimization Tips

### 1. **Model Selection**
- Currently using: `gemini-2.0-flash-exp` (cheapest option)
- Alternative cheap models:
  - `gemini-1.5-flash` - Fast and cost-effective
  - `gemini-1.5-flash-8b` - Even cheaper, smaller model
  - `gemini-2.0-flash-exp` - Latest experimental, good balance

### 2. **API Pricing** (as of 2024)
- **Gemini Flash**: ~$0.075 per 1M input tokens, ~$0.30 per 1M output tokens
- **Gemini Pro**: ~$0.50 per 1M input tokens, ~$1.50 per 1M output tokens
- **Flash is 6-7x cheaper** than Pro models

### 3. **Best Practices for Cost Efficiency**
- ✅ Use Flash models (already implemented)
- ✅ Send entire PDF at once (avoids multiple API calls)
- ✅ Use appropriate output level (executive is shorter/cheaper)
- ✅ Cache results for repeated analyses
- ❌ Avoid page-by-page processing (increases API calls)
- ❌ Avoid using Pro models unless necessary

### 4. **Estimated Costs**
For a typical CAD PDF (50 pages, ~5MB):
- **Flash Model**: ~$0.01 - $0.05 per analysis
- **Pro Model**: ~$0.10 - $0.30 per analysis

## Architecture

```
design-analysis/
├── main.go          # Main application logic
├── prompts.go       # LLM prompt templates
├── formatter.go     # Output formatting
└── README.md        # This file
```

## How It Works

1. **PDF Loading**: Reads entire PDF file into memory
2. **LLM Processing**: Sends PDF + structured prompt to Gemini API
3. **Analysis**: LLM extracts and organizes information
4. **Formatting**: Results are formatted based on output level
5. **Output**: Displays and saves formatted results

## Prompt Engineering Best Practices

The tool uses structured prompts that:
- ✅ Clearly define the task and expected output
- ✅ Use examples and format specifications
- ✅ Adapt to different audience levels
- ✅ Request structured, markdown-formatted responses
- ✅ Include context about CAD/design documents

## Customization

### Change Model
Edit `main.go`:
```go
ModelName: "gemini-1.5-flash", // Change to your preferred model
```

### Customize Prompts
Edit `prompts.go` to modify extraction requirements or add new output levels.

### Adjust Formatting
Edit `formatter.go` to change output formatting.

## Troubleshooting

### API Key Issues
- Ensure `GEMINI_API_KEY` is set correctly
- Check API key has proper permissions
- Verify billing is enabled on Google Cloud

### PDF Processing Issues
- Ensure PDF is not corrupted
- Check file size (very large PDFs may hit token limits)
- Verify PDF is readable (not password-protected)

### Model Availability
- Some models may be region-restricted
- Check [Gemini API documentation](https://ai.google.dev/docs) for availability

## Future Enhancements

- [ ] Support for multiple LLM providers (OpenAI, Anthropic)
- [ ] Batch processing for multiple PDFs
- [ ] JSON output format option
- [ ] Interactive mode for follow-up questions
- [ ] Cost tracking and reporting
- [ ] PDF image extraction for better visual analysis

## License

MIT

## Contributing

Feel free to submit issues and enhancement requests!

