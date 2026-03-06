package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PostBlock holds the schema definition for the PostBlock entity.
// Each block is one atomic unit of rich content within a post
// (heading, paragraph, image, code, quote, embed, etc.)
type PostBlock struct {
	ent.Schema
}

// Fields of the PostBlock.
func (PostBlock) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),

		field.UUID("post_id", uuid.UUID{}).
			Comment("Owning post"),

		field.Enum("type").
			Values(
				"heading",
				"paragraph",
				"image",
				"code",
				"quote",
				"divider",
				"embed",
				"callout",
				"list",
				"table",
			).
			Comment("Block renderer type"),

		field.Text("content").
			Optional().
			Comment("Primary text content (markdown or plain text depending on type)"),

		field.JSON("attrs", map[string]interface{}{}).
			Optional().
			Comment("Type-specific attributes — e.g. {lang: 'go'} for code, {level: 2} for heading"),

		field.Int("position").
			NonNegative().
			Comment("Zero-based display order within the post"),
	}
}

// Edges of the PostBlock.
func (PostBlock) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).
			Ref("blocks").
			Field("post_id").
			Required().
			Unique().
			Comment("Post this block belongs to"),
	}
}

// Indexes of the PostBlock.
func (PostBlock) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("post_id"),
		// Enforce block order uniqueness per post
		index.Fields("post_id", "position").Unique(),
	}
}
