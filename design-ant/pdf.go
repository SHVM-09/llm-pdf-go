package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// splitPDFIntoChunks splits PDF into chunks and returns chunk file paths
func splitPDFIntoChunks(pdfPath, tempDir string, chunkSize, totalPages int) ([]ChunkInfo, error) {
	var chunks []ChunkInfo

	for startPage := 0; startPage < totalPages; startPage += chunkSize {
		endPage := startPage + chunkSize
		if endPage > totalPages {
			endPage = totalPages
		}

		// Extract pages using pdfcpu
		file, err := os.Open(pdfPath)
		if err != nil {
			return nil, fmt.Errorf("error opening PDF: %v", err)
		}

		pageSelection := []string{}
		for p := startPage + 1; p <= endPage; p++ {
			pageSelection = append(pageSelection, fmt.Sprintf("%d", p))
		}

		conf := model.NewDefaultConfiguration()
		err = api.ExtractPages(file, tempDir, fmt.Sprintf("chunk_%d", startPage+1), pageSelection, conf)
		file.Close()

		if err != nil {
			return nil, fmt.Errorf("error extracting pages %d-%d: %v", startPage+1, endPage, err)
		}

		// Find the created file
		actualFileName := fmt.Sprintf("chunk_%d_page_%s.pdf", startPage+1, strings.Join(pageSelection, "_"))
		actualPath := filepath.Join(tempDir, actualFileName)

		// pdfcpu might create files with different naming, try to find it
		if _, err := os.Stat(actualPath); os.IsNotExist(err) {
			// Try alternative naming
			files, _ := os.ReadDir(tempDir)
			for _, f := range files {
				if strings.Contains(f.Name(), fmt.Sprintf("chunk_%d", startPage+1)) {
					actualPath = filepath.Join(tempDir, f.Name())
					break
				}
			}
		}

		chunks = append(chunks, ChunkInfo{
			Path:      actualPath,
			StartPage: startPage,
			EndPage:   endPage - 1, // 0-indexed
		})
	}

	return chunks, nil
}

// getPageCount returns the total number of pages in a PDF
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

