package request

type CreateUserRequest struct {
	Username `json:"username"`
	Email    `json:"email"`
	Password `json:"password"`
}
