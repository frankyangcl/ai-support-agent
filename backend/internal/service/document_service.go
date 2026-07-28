package service

import (
	"github.com/frankyangcl/ai-support-agent/backend/internal/parser"
	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
)

type DocumentService struct {
	Repo      *repository.DocumentRepository
	PDFParser *parser.PDFParser
}

func NewDocumentService(
	repo *repository.DocumentRepository,
	pdfParser *parser.PDFParser,
) *DocumentService {
	return &DocumentService{
		Repo:      repo,
		PDFParser: pdfParser,
	}
}

func (s *DocumentService) CreateDocument(filename, content string) (int, error) {
	return s.Repo.Create(filename, content)
}

func (s *DocumentService) ListDocuments() ([]repository.Document, error) {
	return s.Repo.List()
}

func (s *DocumentService) ExtractPDFText(path string) (string, error) {
	return s.PDFParser.ExtractText(path)
}
