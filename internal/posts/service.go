package posts

import (
	"context"
	"errors"

	"krishblog/internal/database"
	"krishblog/pkg/pagination"
)

type Service struct {
	db    *database.Postgres
	redis *database.Redis
}

func NewService(db *database.Postgres, redis *database.Redis) *Service {
	return &Service{db: db, redis: redis}
}

func (s *Service) ListPublished(ctx context.Context, section, tag, query string, p pagination.Params) ([]PostResponse, int64, error) {
	return []PostResponse{}, 0, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*PostResponse, error) {
	return nil, errors.New("not found")
}

func (s *Service) AdminList(ctx context.Context, status, section string, p pagination.Params) ([]PostResponse, int64, error) {
	return []PostResponse{}, 0, nil
}

func (s *Service) Create(ctx context.Context, authorID string, req CreateRequest) (*PostResponse, error) {
	return nil, errors.New("not implemented: wire Ent in step 2")
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*PostResponse, error) {
	return nil, errors.New("not implemented: wire Ent in step 2")
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status PostStatus) (*PostResponse, error) {
	return nil, errors.New("not implemented: wire Ent in step 2")
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return errors.New("not implemented: wire Ent in step 2")
}
