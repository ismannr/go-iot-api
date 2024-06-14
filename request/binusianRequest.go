package request

type BinusianRequest struct {
	ID         int64  `json:"mentor_id"`
	Name       string `json:"mentor_name"`
	BinusianID string `json:"mentor_employee_id"`
	Email      string `json:"email"`
	Password   string `json:"-" json:"password"`
	Phone      string `json:"mentor_phone_number"`
}
