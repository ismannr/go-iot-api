package request

type RecoveryRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	NewPassword          string `json:"new_password"`
	PasswordConfirmation string `json:"password_confirmation"`
}
