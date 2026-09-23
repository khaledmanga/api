package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"time"

	"api/src/constants"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email"),
		field.String("username"),
		field.String("password"),
		field.String("role").GoType(constants.UserRole(0)).Default(int(constants.USER)),
		field.String("state").GoType(constants.UserState(0)).Default(int(constants.PENDING)),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
