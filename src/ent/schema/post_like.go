package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PostLike struct {
	ent.Schema
}

func (PostLike) Fields() []ent.Field {
	return []ent.Field{
		field.Int("post_id"),
		field.Int("user_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (PostLike) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("post_id", "user_id").Unique(),
		index.Fields("user_id"),
	}
}

func (PostLike) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).
			Ref("likes").
			Unique().
			Field("post_id").
			Required(),
	}
}
