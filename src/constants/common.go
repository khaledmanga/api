package constants

type UserRole int

const (
	ADMIN UserRole = iota
	USER
)

func (r UserRole) String() string {
	return []string{"admin", "user"}[r]
}

type UserState int

const (
	PENDING UserState = iota
	ACTIVE
	INACTIVE
	BANNED
)

func (s UserState) String() string {
	return []string{"pending", "active", "inactive", "banned"}[s]
}
