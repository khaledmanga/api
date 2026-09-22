package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field",
	"entgo.io/ent/schema/mixin"
	"time"

	"api/src/constants"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Mixin() 

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field {
		field.String("email"),
		field.String("username"),
		field.String("role").GoType(constants.UserRole(0)).Default(int(constants.USER)),
		field.String("state").GoType(constants.UserState(0)).Default(int(constants.PENDING))
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
