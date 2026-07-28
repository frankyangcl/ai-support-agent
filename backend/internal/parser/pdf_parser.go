package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

type PDFParser struct{}

func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

func (p *PDFParser) ExtractText(path string) (string, error) {
	file, reader, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("open PDF: %w", err)
	}
	defer file.Close()

	textReader, err := reader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extract PDF text: %w", err)
	}

	content, err := io.ReadAll(textReader)
	if err != nil {
		return "", fmt.Errorf("read extracted text: %w", err)
	}

	text := strings.TrimSpace(string(content))
	if text == "" {
		return "", fmt.Errorf("no extractable text found in PDF")
	}

	return text, nil
}
