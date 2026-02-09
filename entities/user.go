package entities

type ForgotPassword struct {
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
}
