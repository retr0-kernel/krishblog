package posts

import (
	"context"
	"errors"

	"krishblog/ent"
	"krishblog/pkg/pagination"
)

var (
	ErrNotFound       = errors.New("post not found")
	ErrSlugTaken      = errors.New("slug already in use")
	ErrInvalidSection = errors.New("invalid section id")
)

// Service implements post business logic.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListPublished(ctx context.Context, f ListFilter, p pagination.Params) ([]PostResponse, int64, error) {
	f.Page = p.Page
	f.Limit = p.Limit
	posts, total, err := s.repo.ListPublished(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	return toPostResponses(posts), int64(total), nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*PostResponse, error) {
	p, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	r := toPostResponse(p)
	return &r, nil
}

func (s *Service) AdminList(ctx context.Context, f ListFilter, p pagination.Params) ([]PostResponse, int64, error) {
	f.Page = p.Page
	f.Limit = p.Limit
	posts, total, err := s.repo.AdminList(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	return toPostResponses(posts), int64(total), nil
}

func (s *Service) Create(ctx context.Context, authorID string, req CreateRequest) (*PostResponse, error) {
	p, err := s.repo.Create(ctx, authorID, req)
	if err != nil {
		return nil, err
	}
	r := toPostResponse(p)
	return &r, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*PostResponse, error) {
	p, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	r := toPostResponse(p)
	return &r, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ── mapping ───────────────────────────────────────────────────────────────────

func toPostResponse(p *ent.Post) PostResponse {
	r := PostResponse{
		ID:          p.ID.String(),
		AuthorID:    p.AuthorID.String(),
		SectionID:   p.SectionID.String(),
		Title:       p.Title,
		Slug:        p.Slug,
		Summary:     p.Summary,
		CoverImage:  p.CoverImage,
		Status:      PostStatus(p.Status),
		Published:   p.Published,
		IsFeatured:  p.IsFeatured,
		ReadTime:    p.ReadTime,
		WordCount:   p.WordCount,
		MetaTitle:   p.MetaTitle,
		MetaDesc:    p.MetaDesc,
		PublishedAt: p.PublishedAt,
		ScheduledAt: p.ScheduledAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	if p.Edges.Section != nil {
		r.SectionSlug = p.Edges.Section.Slug
	}
	return r
}

func toPostResponses(posts []*ent.Post) []PostResponse {
	out := make([]PostResponse, len(posts))
	for i, p := range posts {
		out[i] = toPostResponse(p)
	}
	return out
}
