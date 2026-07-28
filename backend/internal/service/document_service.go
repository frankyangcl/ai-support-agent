package service

import "github.com/frankyangcl/ai-support-agent/backend/internal/repository"

type DocumentService struct {
	Repo *repository.DocumentRepository
}

func NewDocumentService(repo *repository.DocumentRepository) *DocumentService {
	return &DocumentService{
		Repo: repo,
	}
}

func (s *DocumentService) CreateDocument(filename, content string) (int, error) {
	return s.Repo.Create(filename, content)
}

func (s *DocumentService) ListDocuments() ([]repository.Document, error) {
	return s.Repo.List()
}