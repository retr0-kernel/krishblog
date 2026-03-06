package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Comment("Primary key"),

		field.String("email").
			Unique().
			NotEmpty().
			MaxLen(255).
			Comment("Unique email address used for login"),

		field.String("password_hash").
			NotEmpty().
			Sensitive().
			Comment("bcrypt hash — never returned in API responses"),

		field.String("full_name").
			NotEmpty().
			MaxLen(200).
			Comment("Display name"),

		field.String("avatar_url").
			Optional().
			Nillable().
			MaxLen(2048).
			Comment("R2 URL for avatar image"),

		field.Enum("role").
			Values("superadmin", "admin", "editor", "viewer").
			Default("editor").
			Comment("RBAC role"),

		field.Bool("is_active").
			Default(true).
			Comment("Soft-disable without deleting"),

		field.Time("last_login_at").
			Optional().
			Nillable().
			Comment("Timestamp of most recent successful login"),

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Row creation time"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Row last-update time"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("posts", Post.Type).
			Comment("Posts authored by this user"),

		edge.To("media_assets", MediaAsset.Type).
			Comment("Media uploaded by this user"),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").Unique(),
		index.Fields("role"),
		index.Fields("is_active"),
	}
}
