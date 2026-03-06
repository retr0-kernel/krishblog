package sections

import (
	"context"
	"fmt"
	"strings"

	"krishblog/ent"
	"krishblog/ent/section"
	"krishblog/pkg/slug"
	"krishblog/pkg/uuidutil"
)

// Repository handles all section database operations.
type Repository struct {
	client *ent.Client
}

func NewRepository(client *ent.Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) ListActive(ctx context.Context) ([]*ent.Section, error) {
	return r.client.Section.
		Query().
		Where(section.IsActiveEQ(true)).
		Order(ent.Asc(section.FieldSortOrder)).
		All(ctx)
}

func (r *Repository) ListAll(ctx context.Context) ([]*ent.Section, error) {
	return r.client.Section.
		Query().
		Order(ent.Asc(section.FieldSortOrder)).
		All(ctx)
}

func (r *Repository) GetBySlug(ctx context.Context, s string) (*ent.Section, error) {
	sec, err := r.client.Section.
		Query().
		Where(section.SlugEQ(s)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get section by slug: %w", err)
	}
	return sec, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*ent.Section, error) {
	uid, err := uuidutil.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	sec, err := r.client.Section.Get(ctx, uid)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get section by id: %w", err)
	}
	return sec, nil
}

func (r *Repository) Create(ctx context.Context, req CreateRequest) (*ent.Section, error) {
	sl := req.Slug
	if sl == "" {
		sl = slug.Generate(req.Name)
	}
	sl = strings.ToLower(sl)

	// Ensure slug uniqueness
	if err := r.assertSlugFree(ctx, sl, ""); err != nil {
		return nil, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	layout := section.LayoutFeed
	if req.Layout != "" {
		layout = section.Layout(req.Layout)
	}

	q := r.client.Section.Create().
		SetName(req.Name).
		SetSlug(sl).
		SetLayout(layout).
		SetIsActive(isActive).
		SetSortOrder(req.SortOrder)

	if req.Description != "" {
		q = q.SetDescription(req.Description)
	}
	if req.ThemeColor != "" {
		q = q.SetThemeColor(req.ThemeColor)
	}

	sec, err := q.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create section: %w", err)
	}
	return sec, nil
}

func (r *Repository) Update(ctx context.Context, id string, req UpdateRequest) (*ent.Section, error) {
	uid, err := uuidutil.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}

	q := r.client.Section.UpdateOneID(uid)

	if req.Name != "" {
		q = q.SetName(req.Name)
	}
	if req.Slug != "" {
		sl := strings.ToLower(req.Slug)
		if err := r.assertSlugFree(ctx, sl, id); err != nil {
			return nil, err
		}
		q = q.SetSlug(sl)
	}
	if req.Description != "" {
		q = q.SetDescription(req.Description)
	}
	if req.ThemeColor != "" {
		q = q.SetThemeColor(req.ThemeColor)
	}
	if req.Layout != "" {
		q = q.SetLayout(section.Layout(req.Layout))
	}
	if req.IsActive != nil {
		q = q.SetIsActive(*req.IsActive)
	}
	if req.SortOrder != nil {
		q = q.SetSortOrder(*req.SortOrder)
	}

	sec, err := q.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update section: %w", err)
	}
	return sec, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	uid, err := uuidutil.Parse(id)
	if err != nil {
		return ErrNotFound
	}
	err = r.client.Section.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("delete section: %w", err)
	}
	return nil
}

func (r *Repository) assertSlugFree(ctx context.Context, sl, excludeID string) error {
	q := r.client.Section.Query().Where(section.SlugEQ(sl))
	exists, err := q.Exist(ctx)
	if err != nil {
		return fmt.Errorf("check slug: %w", err)
	}
	if exists {
		return ErrSlugTaken
	}
	return nil
}
