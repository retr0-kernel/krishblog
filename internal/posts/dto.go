package posts

import "time"

type PostStatus string

const (
	StatusDraft     PostStatus = "draft"
	StatusScheduled PostStatus = "scheduled"
	StatusPublished PostStatus = "published"
	StatusArchived  PostStatus = "archived"
)

type CreateRequest struct {
	SectionID     string     `json:"section_id"      validate:"required,uuid4"`
	Title         string     `json:"title"           validate:"required,min=3,max=300"`
	Slug          string     `json:"slug"            validate:"omitempty,min=3,max=300"`
	Summary       string     `json:"summary"         validate:"omitempty,max=500"`
	CoverImage    string     `json:"cover_image"     validate:"omitempty,url"`
	CoverImageAlt string     `json:"cover_image_alt" validate:"omitempty,max=300"`
	Status        PostStatus `json:"status"          validate:"omitempty,oneof=draft scheduled published archived"`
	IsFeatured    bool       `json:"is_featured"`
	MetaTitle     string     `json:"meta_title"      validate:"omitempty,max=70"`
	MetaDesc      string     `json:"meta_desc"       validate:"omitempty,max=160"`
	ScheduledAt   *time.Time `json:"scheduled_at"`
}

type UpdateRequest struct {
	SectionID     string     `json:"section_id"      validate:"omitempty,uuid4"`
	Title         string     `json:"title"           validate:"omitempty,min=3,max=300"`
	Slug          string     `json:"slug"            validate:"omitempty,min=3,max=300"`
	Summary       string     `json:"summary"         validate:"omitempty,max=500"`
	CoverImage    string     `json:"cover_image"     validate:"omitempty,url"`
	CoverImageAlt string     `json:"cover_image_alt" validate:"omitempty,max=300"`
	Status        PostStatus `json:"status"          validate:"omitempty,oneof=draft scheduled published archived"`
	IsFeatured    *bool      `json:"is_featured"`
	MetaTitle     string     `json:"meta_title"      validate:"omitempty,max=70"`
	MetaDesc      string     `json:"meta_desc"       validate:"omitempty,max=160"`
	ScheduledAt   *time.Time `json:"scheduled_at"`
}

type ListFilter struct {
	SectionSlug string
	Status      string
	Featured    bool
	Search      string
	Page        int
	Limit       int
}

type PostResponse struct {
	ID            string     `json:"id"`
	SectionID     string     `json:"section_id"`
	SectionSlug   string     `json:"section_slug,omitempty"`
	AuthorID      string     `json:"author_id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	Summary       string     `json:"summary,omitempty"`
	CoverImage    *string    `json:"cover_image,omitempty"`
	CoverImageAlt string     `json:"cover_image_alt,omitempty"`
	Status        PostStatus `json:"status"`
	Published     bool       `json:"published"`
	IsFeatured    bool       `json:"is_featured"`
	ReadTime      int        `json:"read_time"`
	WordCount     int        `json:"word_count"`
	MetaTitle     string     `json:"meta_title,omitempty"`
	MetaDesc      string     `json:"meta_desc,omitempty"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	ScheduledAt   *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
