package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Post struct {
	ent.Schema
}

func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").Default(""),
		field.Text("content"),
		field.Int("author_id"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Post) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("author_id"),
		index.Fields("created_at"),
	}
}

func (Post) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("comments", Comment.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("likes", PostLike.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
