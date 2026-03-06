package sections

import (
	"context"
	"errors"

	"krishblog/ent"
)

// Sentinel errors
var (
	ErrNotFound  = errors.New("section not found")
	ErrSlugTaken = errors.New("slug already in use")
)

// Service implements section business logic.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListActive(ctx context.Context) ([]SectionResponse, error) {
	secs, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	return toSectionResponses(secs), nil
}

func (s *Service) ListAll(ctx context.Context) ([]SectionResponse, error) {
	secs, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return toSectionResponses(secs), nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*SectionResponse, error) {
	sec, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	r := toSectionResponse(sec)
	return &r, nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*SectionResponse, error) {
	sec, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	r := toSectionResponse(sec)
	return &r, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*SectionResponse, error) {
	sec, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	r := toSectionResponse(sec)
	return &r, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ── mapping ───────────────────────────────────────────────────────────────────

func toSectionResponse(s *ent.Section) SectionResponse {
	return SectionResponse{
		ID:          s.ID.String(),
		Name:        s.Name,
		Slug:        s.Slug,
		Description: s.Description,
		ThemeColor:  s.ThemeColor,
		Layout:      string(s.Layout),
		IsActive:    s.IsActive,
		SortOrder:   s.SortOrder,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func toSectionResponses(secs []*ent.Section) []SectionResponse {
	out := make([]SectionResponse, len(secs))
	for i, s := range secs {
		out[i] = toSectionResponse(s)
	}
	return out
}
