package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Section holds the schema definition for the Section entity.
type Section struct {
	ent.Schema
}

// Fields of the Section.
func (Section) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),

		field.String("name").
			NotEmpty().
			MaxLen(100).
			Comment("Human-readable section name, e.g. 'Technology'"),

		field.String("slug").
			Unique().
			NotEmpty().
			MaxLen(100).
			//Match(slugRegexp).
			Comment("URL-safe identifier, e.g. 'technology'"),

		field.String("description").
			Optional().
			MaxLen(500).
			Comment("Short description shown on section listing pages"),

		field.String("theme_color").
			Optional().
			MaxLen(7).
			Comment("Hex colour code for UI theming, e.g. '#3B82F6'"),

		field.String("cover_image").
			Optional().
			Nillable().
			MaxLen(2048).
			Comment("R2 URL for section cover image"),

		field.String("meta_title").
			Optional().
			MaxLen(70).
			Comment("SEO <title> override"),

		field.String("meta_desc").
			Optional().
			MaxLen(160).
			Comment("SEO meta description override"),

		field.Enum("layout").
			Values("feed", "grid", "featured", "minimal", "magazine").
			Default("feed").
			Comment("Display layout used by the reader frontend"),

		field.Bool("is_active").
			Default(true).
			Comment("Hide section without deleting it"),

		field.Int("sort_order").
			Default(0).
			Comment("Ascending sort order in navigation"),

		field.JSON("settings", map[string]interface{}{}).
			Optional().
			Comment("Layout-specific configuration blob"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Section.
func (Section) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("posts", Post.Type).
			Comment("All posts belonging to this section"),
	}
}

// Indexes of the Section.
func (Section) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("is_active"),
		index.Fields("sort_order"),
	}
}
