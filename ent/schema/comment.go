package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Comment struct {
	ent.Schema
}

func (Comment) Fields() []ent.Field {
	return []ent.Field{
		field.Int("post_id"),
		field.Int("author_id"),
		field.Int("parent_id").Optional().Nillable(),
		field.Text("content"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Comment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("post_id", "parent_id"),
		index.Fields("author_id"),
	}
}

func (Comment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).
			Ref("comments").
			Unique().
			Field("post_id").
			Required(),
		edge.From("parent", Comment.Type).
			Ref("replies").
			Unique().
			Field("parent_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("replies", Comment.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
