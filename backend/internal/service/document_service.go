package service

import (
	"fmt"

	"github.com/frankyangcl/ai-support-agent/backend/internal/chunker"
	"github.com/frankyangcl/ai-support-agent/backend/internal/parser"
	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
)

type DocumentService struct {
	Repo        *repository.DocumentRepository
	ChunkRepo   *repository.ChunkRepository
	PDFParser   *parser.PDFParser
	TextChunker *chunker.TextChunker
}

func NewDocumentService(
	repo *repository.DocumentRepository,
	chunkRepo *repository.ChunkRepository,
	pdfParser *parser.PDFParser,
	textChunker *chunker.TextChunker,
) *DocumentService {
	return &DocumentService{
		Repo:        repo,
		ChunkRepo:   chunkRepo,
		PDFParser:   pdfParser,
		TextChunker: textChunker,
	}
}

func (s *DocumentService) CreateDocument(
	filename string,
	content string,
) (int, error) {
	return s.Repo.Create(filename, content)
}

func (s *DocumentService) ListDocuments() (
	[]repository.Document,
	error,
) {
	return s.Repo.List()
}

func (s *DocumentService) ExtractPDFText(path string) (string, error) {
	return s.PDFParser.ExtractText(path)
}

func (s *DocumentService) CreateDocumentFromPDF(
	filename string,
	path string,
) (int, int, int, error) {
	text, err := s.PDFParser.ExtractText(path)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("extract PDF text: %w", err)
	}

	chunks := s.TextChunker.Split(text)
	if len(chunks) == 0 {
		return 0, 0, 0, fmt.Errorf("PDF produced no text chunks")
	}

	documentID, err := s.Repo.Create(filename, text)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("create document: %w", err)
	}

	chunkInputs := make(
		[]repository.CreateChunkInput,
		0,
		len(chunks),
	)

	for _, textChunk := range chunks {
		chunkInputs = append(
			chunkInputs,
			repository.CreateChunkInput{
				ChunkIndex:     textChunk.Index,
				Content:        textChunk.Content,
				CharacterCount: textChunk.CharacterCount,
			},
		)
	}

	if err := s.ChunkRepo.CreateBatch(
		documentID,
		chunkInputs,
	); err != nil {
		deleteErr := s.Repo.Delete(documentID)
		if deleteErr != nil {
			return 0, 0, 0, fmt.Errorf(
				"create chunks: %v; rollback document: %w",
				err,
				deleteErr,
			)
		}

		return 0, 0, 0, fmt.Errorf(
			"create document chunks: %w",
			err,
		)
	}

	characterCount := len([]rune(text))

	return documentID, characterCount, len(chunks), nil
}
