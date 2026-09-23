package constants

type UserRole int

const (
	ADMIN UserRole = iota
	USER  UserRole
)

func (r UserRole) String() string {
	return []string{"admin", "user"}[r]
}

type UserState int

const (
	PEDNING  UserState = iota
	ACTIVE   UserState
	INACTIVE UserState
	BANNED   UserState
)

func (s UserState) String() string {
	return []string{"pending", "active", "inactive", "banned"}[s]
}
