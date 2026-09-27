package request

type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
