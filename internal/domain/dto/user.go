package dto

type SignUpUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type SignInUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
