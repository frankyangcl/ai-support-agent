package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/frankyangcl/ai-support-agent/backend/internal/chunker"
	"github.com/frankyangcl/ai-support-agent/backend/internal/parser"
	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
)

type DocumentService struct {
	Repo             *repository.DocumentRepository
	ChunkRepo        *repository.ChunkRepository
	PDFParser        *parser.PDFParser
	TextChunker      *chunker.TextChunker
	EmbeddingService *EmbeddingService
}

func NewDocumentService(
	repo *repository.DocumentRepository,
	chunkRepo *repository.ChunkRepository,
	pdfParser *parser.PDFParser,
	textChunker *chunker.TextChunker,
	embeddingService *EmbeddingService,
) *DocumentService {
	return &DocumentService{
		Repo:             repo,
		ChunkRepo:        chunkRepo,
		PDFParser:        pdfParser,
		TextChunker:      textChunker,
		EmbeddingService: embeddingService,
	}
}

func (s *DocumentService) CreateDocument(
	filename string,
	content string,
) (int, error) {
	return s.Repo.Create(filename, content, "")
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
	ctx context.Context,
	filename string,
	path string,
) (int, int, int, error) {
	fileHash, err := calculateFileHash(path)
	if err != nil {
		return 0, 0, 0, err
	}

	exists, err := s.Repo.ExistsByHash(fileHash)
	if err != nil {

		return 0, 0, 0, fmt.Errorf(
			"check duplicate document: %w",
			err,
		)
	}

	if exists {
		return 0, 0, 0, fmt.Errorf(
			"document already exists in knowledge base",
		)
	}
	text, err := s.PDFParser.ExtractText(path)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("extract PDF text: %w", err)
	}

	chunks := s.TextChunker.Split(text)
	if len(chunks) == 0 {
		return 0, 0, 0, fmt.Errorf("PDF produced no text chunks")
	}

	documentID, err := s.Repo.Create(filename, text, fileHash)
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

	embeddedCount, err := s.EmbeddingService.ProcessDocumentChunks(
		ctx,
		documentID,
	)
	if err != nil {
		deleteErr := s.Repo.Delete(documentID)
		if deleteErr != nil {
			return 0, 0, 0, fmt.Errorf(
				"generate embeddings: %v; rollback document: %w",
				err,
				deleteErr,
			)
		}

		return 0, 0, 0, fmt.Errorf(
			"generate document embeddings: %w",
			err,
		)
	}

	if embeddedCount != len(chunks) {
		deleteErr := s.Repo.Delete(documentID)
		if deleteErr != nil {
			return 0, 0, 0, fmt.Errorf(
				"expected %d embeddings, got %d; rollback document: %w",
				len(chunks),
				embeddedCount,
				deleteErr,
			)
		}

		return 0, 0, 0, fmt.Errorf(
			"expected %d embeddings, got %d",
			len(chunks),
			embeddedCount,
		)
	}

	return documentID, characterCount, len(chunks), nil
}

func (s *DocumentService) GetDocument(
	id int,
) (
	*repository.DocumentDetail,
	[]repository.DocumentChunk,
	error,
) {
	doc, err := s.Repo.GetByID(id)
	if err != nil {
		return nil, nil, err
	}

	chunks, err := s.ChunkRepo.ListByDocumentID(id)
	if err != nil {
		return nil, nil, err
	}

	return doc, chunks, nil
}

func (s *DocumentService) DeleteDocument(id int) error {
	return s.Repo.Delete(id)
}
func calculateFileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file for hashing: %w", err)
	}

	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:]), nil
}
