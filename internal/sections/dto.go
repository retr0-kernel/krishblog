package sections

import "time"

type SectionLayout string

const (
	LayoutFeed     SectionLayout = "feed"
	LayoutGrid     SectionLayout = "grid"
	LayoutFeatured SectionLayout = "featured"
	LayoutMinimal  SectionLayout = "minimal"
	LayoutMagazine SectionLayout = "magazine"
)

type CreateRequest struct {
	Name        string        `json:"name"        validate:"required,min=2,max=100"`
	Slug        string        `json:"slug"        validate:"omitempty,min=2,max=100"`
	Description string        `json:"description" validate:"omitempty,max=500"`
	ThemeColor  string        `json:"theme_color" validate:"omitempty,max=7"`
	Layout      SectionLayout `json:"layout"      validate:"omitempty,oneof=feed grid featured minimal magazine"`
	IsActive    *bool         `json:"is_active"`
	SortOrder   int           `json:"sort_order"`
}

type UpdateRequest struct {
	Name        string        `json:"name"        validate:"omitempty,min=2,max=100"`
	Slug        string        `json:"slug"        validate:"omitempty,min=2,max=100"`
	Description string        `json:"description" validate:"omitempty,max=500"`
	ThemeColor  string        `json:"theme_color" validate:"omitempty,max=7"`
	Layout      SectionLayout `json:"layout"      validate:"omitempty,oneof=feed grid featured minimal magazine"`
	IsActive    *bool         `json:"is_active"`
	SortOrder   *int          `json:"sort_order"`
}

type SectionResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	ThemeColor  string    `json:"theme_color,omitempty"`
	Layout      string    `json:"layout"`
	IsActive    bool      `json:"is_active"`
	SortOrder   int       `json:"sort_order"`
	PostCount   int       `json:"post_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
