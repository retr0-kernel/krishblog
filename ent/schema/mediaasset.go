package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type MediaAsset struct {
	ent.Schema
}

func (MediaAsset) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),

		field.UUID("uploader_id", uuid.UUID{}).
			Optional().
			Nillable(),

		field.String("url").
			NotEmpty().
			MaxLen(2048),

		field.String("storage_key").
			Unique().
			NotEmpty().
			MaxLen(1024),

		field.String("filename").
			NotEmpty().
			MaxLen(255),

		field.String("original_name").
			NotEmpty().
			MaxLen(255),

		field.String("mime_type").
			NotEmpty().
			MaxLen(100),

		field.Enum("asset_type").
			Values("image", "video", "document", "audio"),

		field.Int64("size_bytes").
			Positive(),

		field.Int("width").
			Optional().
			Nillable().
			Min(0),

		field.Int("height").
			Optional().
			Nillable().
			Min(0),

		field.Float("duration_sec").
			Optional().
			Nillable().
			Min(0),

		field.String("thumbnail_url").
			Optional().
			Nillable().
			MaxLen(2048),

		field.String("alt_text").
			Optional().
			MaxLen(300),

		field.String("caption").
			Optional().
			MaxLen(500),

		field.JSON("metadata", map[string]interface{}{}).
			Optional(),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (MediaAsset) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("uploader", User.Type).
			Ref("media_assets").
			Field("uploader_id").
			Unique(),
	}
}

func (MediaAsset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("storage_key").Unique(),
		index.Fields("uploader_id"),
		index.Fields("asset_type"),
		index.Fields("mime_type"),
		index.Fields("created_at"),
	}
}
