package response

import (
	"gin-crud/models"
	"github.com/google/uuid"
	"time"
)

type UserResponse struct {
	ID           uuid.UUID `json:"id,omitempty"`
	Name         string    `json:"name,omitempty"`
	Email        string    `json:"email,omitempty"`
	Gender       string    `json:"gender,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	Dob          time.Time `json:"birth_date,omitempty"`
	Address      string    `json:"address,omitempty"`
	City         string    `json:"city,omitempty"`
	Province     string    `json:"province,omitempty"`
	BusinessName string    `json:"business_name,omitempty"`
	BusinessDesc string    `json:"business_desc,omitempty"`
}

func BindUserToResponse(user *models.UmkmData) UserResponse {
	resp := UserResponse{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		Gender:       user.Gender,
		Phone:        user.Phone,
		Dob:          user.Dob,
		Address:      user.Address,
		City:         user.City,
		Province:     user.Province,
		BusinessName: user.BusinessName,
		BusinessDesc: user.BusinessDesc,
	}
	return resp
}
