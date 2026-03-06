package posts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"krishblog/ent"
	"krishblog/ent/post"
	"krishblog/ent/section"
	"krishblog/pkg/slug"
	"krishblog/pkg/uuidutil"
)

// Repository handles all post database operations.
type Repository struct {
	client *ent.Client
}

func NewRepository(client *ent.Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) ListPublished(ctx context.Context, f ListFilter) ([]*ent.Post, int, error) {
	q := r.client.Post.
		Query().
		Where(post.StatusEQ(post.StatusPublished))

	if f.SectionSlug != "" {
		q = q.Where(post.HasSectionWith(section.SlugEQ(f.SectionSlug)))
	}
	if f.Featured {
		q = q.Where(post.IsFeaturedEQ(true))
	}
	if f.Search != "" {
		term := "%" + strings.ToLower(f.Search) + "%"
		q = q.Where(post.Or(
			post.TitleContainsFold(f.Search),
			post.SummaryContainsFold(term),
		))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count posts: %w", err)
	}

	page, limit := normalizePage(f.Page, f.Limit)
	posts, err := q.
		Order(ent.Desc(post.FieldPublishedAt)).
		WithSection().
		WithAuthor().
		Limit(limit).
		Offset((page - 1) * limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list posts: %w", err)
	}

	return posts, total, nil
}

func (r *Repository) GetBySlug(ctx context.Context, s string) (*ent.Post, error) {
	p, err := r.client.Post.
		Query().
		Where(
			post.SlugEQ(s),
			post.StatusEQ(post.StatusPublished),
		).
		WithSection().
		WithAuthor().
		WithBlocks(func(q *ent.PostBlockQuery) {
			q.Order(ent.Asc("position"))
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get post by slug: %w", err)
	}
	return p, nil
}

func (r *Repository) GetByIDAdmin(ctx context.Context, id string) (*ent.Post, error) {
	uid, err := uuidutil.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	p, err := r.client.Post.
		Query().
		Where(post.IDEQ(uid)).
		WithSection().
		WithAuthor().
		WithBlocks(func(q *ent.PostBlockQuery) {
			q.Order(ent.Asc("position"))
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get post by id: %w", err)
	}
	return p, nil
}

func (r *Repository) AdminList(ctx context.Context, f ListFilter) ([]*ent.Post, int, error) {
	q := r.client.Post.Query()

	if f.Status != "" {
		q = q.Where(post.StatusEQ(post.Status(f.Status)))
	}
	if f.SectionSlug != "" {
		q = q.Where(post.HasSectionWith(section.SlugEQ(f.SectionSlug)))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count posts: %w", err)
	}

	page, limit := normalizePage(f.Page, f.Limit)
	posts, err := q.
		Order(ent.Desc(post.FieldCreatedAt)).
		WithSection().
		WithAuthor().
		Limit(limit).
		Offset((page - 1) * limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("admin list posts: %w", err)
	}

	return posts, total, nil
}

func (r *Repository) Create(ctx context.Context, authorID string, req CreateRequest) (*ent.Post, error) {
	authorUID, err := uuidutil.Parse(authorID)
	if err != nil {
		return nil, fmt.Errorf("invalid author id: %w", err)
	}
	sectionUID, err := uuidutil.Parse(req.SectionID)
	if err != nil {
		return nil, ErrInvalidSection
	}

	sl := req.Slug
	if sl == "" {
		sl = slug.Generate(req.Title)
	}
	sl = strings.ToLower(sl)

	if err := r.assertSlugFree(ctx, sl, ""); err != nil {
		return nil, err
	}

	status := post.StatusDraft
	if req.Status != "" {
		status = post.Status(req.Status)
	}

	q := r.client.Post.Create().
		SetTitle(req.Title).
		SetSlug(sl).
		SetStatus(status).
		SetPublished(status == post.StatusPublished).
		SetIsFeatured(req.IsFeatured).
		SetAuthorID(authorUID).
		SetSectionID(sectionUID)

	if req.Summary != "" {
		q = q.SetSummary(req.Summary)
	}
	if req.CoverImage != "" {
		q = q.SetCoverImage(req.CoverImage)
	}
	if req.CoverImageAlt != "" {
		q = q.SetCoverImageAlt(req.CoverImageAlt)
	}
	if req.MetaTitle != "" {
		q = q.SetMetaTitle(req.MetaTitle)
	}
	if req.MetaDesc != "" {
		q = q.SetMetaDesc(req.MetaDesc)
	}
	if status == post.StatusPublished {
		now := time.Now()
		q = q.SetPublishedAt(now)
	}
	if req.ScheduledAt != nil {
		q = q.SetScheduledAt(*req.ScheduledAt)
	}

	p, err := q.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}
	return p, nil
}

func (r *Repository) Update(ctx context.Context, id string, req UpdateRequest) (*ent.Post, error) {
	uid, err := uuidutil.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}

	q := r.client.Post.UpdateOneID(uid)

	if req.Title != "" {
		q = q.SetTitle(req.Title)
	}
	if req.Slug != "" {
		sl := strings.ToLower(req.Slug)
		if err := r.assertSlugFree(ctx, sl, id); err != nil {
			return nil, err
		}
		q = q.SetSlug(sl)
	}
	if req.SectionID != "" {
		sid, err := uuidutil.Parse(req.SectionID)
		if err != nil {
			return nil, ErrInvalidSection
		}
		q = q.SetSectionID(sid)
	}
	if req.Summary != "" {
		q = q.SetSummary(req.Summary)
	}
	if req.CoverImage != "" {
		q = q.SetCoverImage(req.CoverImage)
	}
	if req.CoverImageAlt != "" {
		q = q.SetCoverImageAlt(req.CoverImageAlt)
	}
	if req.MetaTitle != "" {
		q = q.SetMetaTitle(req.MetaTitle)
	}
	if req.MetaDesc != "" {
		q = q.SetMetaDesc(req.MetaDesc)
	}
	if req.IsFeatured != nil {
		q = q.SetIsFeatured(*req.IsFeatured)
	}
	if req.Status != "" {
		q = q.SetStatus(post.Status(req.Status))
		q = q.SetPublished(req.Status == StatusPublished)
		if req.Status == StatusPublished {
			q = q.SetPublishedAt(time.Now())
		}
	}
	if req.ScheduledAt != nil {
		q = q.SetScheduledAt(*req.ScheduledAt)
	}

	p, err := q.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update post: %w", err)
	}
	return p, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	uid, err := uuidutil.Parse(id)
	if err != nil {
		return ErrNotFound
	}
	if err := r.client.Post.DeleteOneID(uid).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("delete post: %w", err)
	}
	return nil
}

func (r *Repository) assertSlugFree(ctx context.Context, sl, excludeID string) error {
	q := r.client.Post.Query().Where(post.SlugEQ(sl))
	if excludeID != "" {
		uid, _ := uuidutil.Parse(excludeID)
		q = q.Where(post.IDNEQ(uid))
	}
	exists, err := q.Exist(ctx)
	if err != nil {
		return fmt.Errorf("check slug: %w", err)
	}
	if exists {
		return ErrSlugTaken
	}
	return nil
}

func normalizePage(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
