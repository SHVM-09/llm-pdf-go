# LLM PDF App

A Go-based demo application that processes PDF files by splitting them into individual pages and using concurrent goroutines to summarize each page using Google's Gemini API.

## Features

- 📄 **PDF Processing**: Extracts text from PDF files page by page
- 🚀 **Concurrent Processing**: Uses Go routines to process multiple pages in parallel
- 🤖 **Gemini Integration**: Summarizes each page using Google's Gemini API
- ⚡ **Fast Performance**: Parallel processing significantly reduces total processing time

## Prerequisites

- Go 1.21 or higher
- A Gemini API key (get one from [Google AI Studio](https://makersuite.google.com/app/apikey))

## Setup

1. **Clone/Navigate to the project directory**:
   ```bash
   cd llm-pdf-app
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Set up environment variables**:
   - Copy `.env.example` to `.env`:
     ```bash
     cp .env.example .env
     ```
   - Edit `.env` and add your Gemini API key:
     ```
     GEMINI_API_KEY=your_actual_api_key_here
     ```

## Usage

Run the application with a PDF file path:

```bash
go run main.go <path-to-pdf-file>
```

Example:
```bash
go run main.go document.pdf
```

Or if the PDF is in a different directory:
```bash
go run main.go /path/to/your/document.pdf
```

## How It Works

1. **PDF Selection**: The app takes a PDF file path as a command-line argument
2. **PDF Splitting**: Extracts text from each page of the PDF
3. **Concurrent Processing**: Launches a goroutine for each page
4. **Gemini API Calls**: Each goroutine makes an independent API call to Gemini
5. **Summary Collection**: Collects all summaries and displays them in order

## Output

The application will display:
- Total number of pages in the PDF
- Progress updates as each page is processed
- Total processing time
- A summary for each page (Page 1, Page 2, etc.)

## Example Output

```
📄 Processing PDF: document.pdf
=====================================
📊 Total pages: 5

🚀 Processing pages with Gemini API (concurrent)...
=====================================

✅ Page 1: Processed successfully
✅ Page 2: Processed successfully
✅ Page 3: Processed successfully
✅ Page 4: Processed successfully
✅ Page 5: Processed successfully

⏱️  Total processing time: 2.3s

==================================================
📋 PAGE SUMMARIES
==================================================

📄 Page 1:
   This page introduces the main concepts and provides an overview...

📄 Page 2:
   The second page delves into specific implementation details...
```

## Dependencies

- `github.com/gen2brain/go-fitz`: PDF text extraction
- `github.com/joho/godotenv`: Environment variable management

## Notes

- The app uses the `gemini-pro` model by default
- Each page is processed independently, so failures on one page don't affect others
- The API key is read from the `.env` file or environment variables
- Processing time is significantly reduced due to concurrent execution

