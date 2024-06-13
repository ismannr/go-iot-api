package request

import "gin-crud/models"

type UmkmRequest struct {
	Name         string       `json:"name"`
	Email        string       `json:"email" gorm:"unique"`
	Password     string       `json:"password"`
	ConfirmPass  string       `json:"confirm_password"`
	Gender       string       `json:"gender"`
	PhoneNumber  string       `json:"phone" gorm:"unique"`
	Dob          string       `json:"dob"`
	Address      string       `json:"address"`
	City         string       `json:"city"`
	Province     string       `json:"province"`
	BusinessName string       `json:"business_name"`
	BusinessDesc string       `json:"business_desc"`
	Role         models.Role  `json:"role"`
	Level        models.Level `json:"level"`
}
