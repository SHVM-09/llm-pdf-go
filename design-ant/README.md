# Design PDF Analysis Tool (Anthropic)

A Go-based tool that analyzes CAD and design PDF documents using Anthropic's Claude API with concurrent processing, cost tracking, and structured output.

## Features

- **Smart PDF Splitting**: Automatically splits PDFs into 5-page chunks if the document exceeds 20 pages
- **Concurrent Processing**: Processes all chunks simultaneously using goroutines for maximum efficiency
- **Cost Tracking**: Tracks token usage and calculates costs for each chunk and total analysis
- **Structured Output**: Generates results in both JSON and CSV formats
- **Cheap Model**: Uses Claude 3.5 Haiku (cheapest Anthropic model) for cost efficiency
- **Detailed Analysis**: Provides comprehensive design document analysis similar to design-analysis tool

## Prerequisites

- Go 1.24 or higher
- Anthropic API key (get from [Anthropic Console](https://console.anthropic.com/))

## Setup

1. **Install Dependencies**:
   ```bash
   cd llm-pdf-app/design-ant
   go mod download
   ```

2. **Set Environment Variable**:
   Create a `.env` file in the parent directory or set environment variable:
   ```bash
   export ANTHROPIC_API_KEY="your-api-key-here"
   ```
   
   Or create `.env` file in `llm-pdf-app/`:
   ```
   ANTHROPIC_API_KEY=your-api-key-here
   ```

## Usage

### Basic Usage
```bash
go run main.go <path-to-pdf>
```

### Example
```bash
go run main.go ../design-analysis/v6truboEngine.pdf
```

## How It Works

1. **PDF Analysis**: Reads the PDF and determines total page count
2. **Chunking**: If PDF > 20 pages, splits into 5-page chunks; otherwise processes as single chunk
3. **Image Conversion**: Converts each PDF page to PNG images for better visual analysis
4. **Concurrent Processing**: Sends all chunks to Anthropic API simultaneously
5. **Analysis**: Each chunk is analyzed for design information, BOM, specifications, etc.
6. **Cost Tracking**: Tracks input/output tokens and calculates costs
7. **Output Generation**: Creates JSON and CSV files with complete analysis

## Output Files

The tool generates two output files:

1. **JSON File** (`{pdf-name}_analysis.json`):
   - Complete structured analysis with all chunks
   - Token usage and cost breakdown per chunk
   - Total costs and processing time
   - Full analysis text for each chunk

2. **CSV File** (`{pdf-name}_analysis.csv`):
   - Summary table with chunk information
   - Token counts and costs per chunk
   - Processing times
   - Error information if any

## Viewing Results

### HTML Viewer

A beautiful HTML viewer is included to visualize your analysis results:

1. **Open the viewer:**
   ```bash
   # Open viewer.html in your browser
   open viewer.html  # macOS
   # or
   xdg-open viewer.html  # Linux
   # or just double-click viewer.html
   ```

2. **Load your JSON file:**
   - Click "Choose File" and select your `*_analysis.json` file
   - Or open directly: `viewer.html?file=v6truboEngine_analysis.json`

3. **Features:**
   - 📊 Summary dashboard with key metrics
   - 📑 Tabbed interface to navigate between chunks
   - 📝 Formatted analysis content with markdown rendering
   - 💰 Cost breakdown per chunk
   - ⏱️ Processing time information
   - 📱 Responsive design for mobile and desktop

### Direct JSON Viewing

You can also view the JSON file directly in any text editor or JSON viewer.

## Cost Information

### Model Pricing (Claude 3.5 Haiku)
- **Input**: $0.25 per million tokens
- **Output**: $1.25 per million tokens

### Estimated Costs
For a typical CAD PDF (50 pages, split into 10 chunks of 5 pages):
- **Per Chunk**: ~$0.01 - $0.05 (depending on content complexity)
- **Total**: ~$0.10 - $0.50 for complete analysis

### Cost Optimization Tips
- ✅ Uses cheapest Anthropic model (Haiku)
- ✅ Concurrent processing reduces total time
- ✅ Chunking allows parallel processing
- ✅ Tracks costs for budget management

## Analysis Output Structure

Each chunk analysis includes:

1. **Document Metadata**: Drawn By, Checked By, Approved By, dates, revision info
2. **Project Overview**: Product name, drawing numbers, dimensions, weight, materials
3. **Bill of Materials (BOM)**: Complete parts list with quantities and part numbers
4. **Technical Specifications**: Dimensions, tolerances, materials, finishes
5. **Drawing Information**: Views, standards, scales, geometric features
6. **Assembly Information**: Assembly sequence, relationships, exploded views
7. **Quality & Manufacturing Notes**: Special instructions, quality requirements

## Example Output

### Console Output
```
======================================================================
  DESIGN PDF ANALYSIS TOOL (ANTHROPIC)
======================================================================

📄 Processing: v6truboEngine.pdf
🤖 Model: claude-3-5-haiku-20241022
💰 Model Pricing: $0.25/M input, $1.25/M output

📊 Total pages: 25
📦 Splitting into chunks of 5 pages each

✅ Created 5 chunk(s)

🚀 Processing chunks concurrently...
----------------------------------------------------------------------
  🔄 Processing chunk 1 (pages 1-5)...
  🔄 Processing chunk 2 (pages 6-10)...
  ...
  ✅ Chunk 1 completed: 1250 input tokens, 850 output tokens, $0.001344
  ✅ Chunk 2 completed: 1180 input tokens, 920 output tokens, $0.001415
  ...

======================================================================
  ANALYSIS SUMMARY
======================================================================
Total Input Tokens:  6250
Total Output Tokens: 4250
Total Input Cost:    $0.001563
Total Output Cost:   $0.005313
Total Cost:          $0.006875
Processing Time:     45.2s
======================================================================

💾 JSON results saved to: v6truboEngine_analysis.json
💾 CSV results saved to: v6truboEngine_analysis.csv
```

## JSON Output Structure

```json
{
  "pdf_path": "v6truboEngine.pdf",
  "total_pages": 25,
  "total_chunks": 5,
  "chunks": [
    {
      "chunk_number": 1,
      "start_page": 1,
      "end_page": 5,
      "analysis": "...",
      "input_tokens": 1250,
      "output_tokens": 850,
      "input_cost": 0.000313,
      "output_cost": 0.001063,
      "total_cost": 0.001375,
      "processing_time": "8.5s",
      "timestamp": "2024-01-15T10:30:00Z"
    }
  ],
  "total_input_tokens": 6250,
  "total_output_tokens": 4250,
  "total_input_cost": 0.001563,
  "total_output_cost": 0.005313,
  "total_cost": 0.006875,
  "processing_time": "45.2s",
  "generated_at": "2024-01-15T10:30:45Z"
}
```

## Troubleshooting

### API Key Issues
- Ensure `ANTHROPIC_API_KEY` is set correctly
- Check API key has proper permissions
- Verify billing is enabled on Anthropic account

### PDF Processing Issues
- Ensure PDF is not corrupted
- Check file size (very large PDFs may hit token limits)
- Verify PDF is readable (not password-protected)

### Image Conversion Issues
- Some PDFs may have pages that can't be converted to images
- Tool will skip problematic pages and continue with others
- Check logs for warnings about specific pages

## Comparison with design-analysis

| Feature | design-analysis | design-ant |
|---------|----------------|------------|
| API Provider | Google Gemini | Anthropic Claude |
| Processing | Single request | Concurrent chunks |
| PDF Size | Entire PDF at once | Split into chunks |
| Cost Tracking | No | Yes (detailed) |
| Output Format | Text only | JSON + CSV |
| Model | Gemini Flash | Claude Haiku |
| Best For | Small PDFs | Large PDFs (>20 pages) |

## Future Enhancements

- [ ] Support for other Anthropic models (Sonnet, Opus)
- [ ] Configurable chunk size
- [ ] Progress bar for long-running analyses
- [ ] Resume capability for interrupted analyses
- [ ] Batch processing for multiple PDFs
- [ ] Cost estimation before processing

## License

MIT

## Contributing

Feel free to submit issues and enhancement requests!

