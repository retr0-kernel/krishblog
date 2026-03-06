package pagination

import (
	"math"
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

type Params struct {
	Page   int
	Limit  int
	Offset int
}

type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

func Parse(c echo.Context) Params {
	page := parseInt(c.QueryParam("page"), DefaultPage)
	limit := parseInt(c.QueryParam("limit"), DefaultLimit)
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return Params{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

func NewMeta(p Params, total int64) Meta {
	totalPages := int(math.Ceil(float64(total) / float64(p.Limit)))
	if totalPages < 1 {
		totalPages = 1
	}
	return Meta{
		Page:       p.Page,
		Limit:      p.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    p.Page < totalPages,
		HasPrev:    p.Page > 1,
	}
}

func parseInt(s string, fallback int) int {
	n, err := strconv.Atoi(s)
	if err != nil || s == "" {
		return fallback
	}
	return n
}
