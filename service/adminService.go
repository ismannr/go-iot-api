package service

import (
	"errors"
	"fmt"
	"gin-crud/initializers"
	model "gin-crud/models"
	"gin-crud/request"
	"gin-crud/response"
	"gin-crud/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
)

func CreateParticipant(c *gin.Context) {
	var req request.UmkmRequest
	var s strings.Builder
	var userDB model.UmkmData
	var isSatisfied bool = true

	if err := c.Bind(&req); err != nil {
		response.GlobalResponse(c, "Failed to retrieve user request", http.StatusBadRequest, nil)
		return
	}
	if len(req.Name) < 3 {
		s.WriteString("Name, ")
		isSatisfied = false
	}
	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		s.WriteString("Email (wrong format), ")
		isSatisfied = false
	}
	if exist := initializers.DB.Where("email = ?", req.Email).First(&userDB).Error; exist == nil {
		s.WriteString("Email already exist, ")
		isSatisfied = false
	}
	if strings.ToLower(req.Gender) != "male" && strings.ToLower(req.Gender) != "female" {
		s.WriteString("Gender, ")
		isSatisfied = false
	}
	dob, err := utils.ParseDate(req.Dob)
	if err != nil {
		s.WriteString("DOB (wrong date format), ")
		isSatisfied = false
	} else {
		if !utils.IsAdult(req.Dob) {
			s.WriteString("DOB (Must be over 17), ")
			isSatisfied = false
		}
	}

	if len(req.Password) < 8 {
		s.WriteString("Password min. 8 char, ")
		isSatisfied = false
	} else {
		uppercaseRegex := regexp.MustCompile(`[A-Z]`)
		numberRegex := regexp.MustCompile(`[0-9]`)
		if !uppercaseRegex.MatchString(req.Password) || !numberRegex.MatchString(req.Password) {
			s.WriteString("Password (Must contain at least one uppercase letter and one number), ")
			isSatisfied = false
		} else if req.ConfirmPass != req.Password {
			s.WriteString("Confirmation Password doesn't match, ")
			isSatisfied = false
		}
	}

	if !isSatisfied {
		message := "User data requirements not satisfied: " + s.String()
		response.GlobalResponse(c, message, http.StatusBadRequest, nil)
		return
	}

	password, err := utils.HashEncoder(req.Password)
	if err != nil {
		message := "Error encoding the password"
		response.GlobalResponse(c, message, http.StatusBadRequest, nil)
		return
	}
	userId := uuid.New()
	systemUser := model.SystemData{
		ID:       userId,
		Email:    req.Email,
		Password: password,
		Role:     model.RoleUMKM,
		Level:    model.LevelUser,
	}
	user := model.UmkmData{
		ID:           userId,
		Name:         req.Name,
		Gender:       strings.ToUpper(req.Gender),
		Dob:          dob,
		Email:        req.Email,
		Address:      req.Address,
		Province:     req.Province,
		City:         req.City,
		Phone:        req.PhoneNumber,
		SystemDataID: &systemUser.ID,
		SystemData:   &systemUser,
	}

	result := initializers.DB.Create(&user)
	if result.Error != nil {
		response.GlobalResponse(c, "Failed to save user data", http.StatusBadRequest, nil)
		return
	}
	initializers.DB.Save(&user)
	response.GlobalResponse(c, "Participant_data created successfully", http.StatusOK, user)
}

func GetParticipantList(c *gin.Context) {
	var users []model.UmkmData
	if result := initializers.DB.Preload("SystemData").Omit("password").Find(&users); result.Error != nil {
		response.GlobalResponse(c, "Error retrieving data from database", http.StatusInternalServerError, result.Error)
		return
	}

	if len(users) == 0 {
		response.GlobalResponse(c, "No users data found", http.StatusOK, users)
		return
	}

	response.GlobalResponse(c, "Successfully retrieving users", http.StatusOK, users)
}

func getParticipantByIdentifier(identifier string) (*model.UmkmData, error) {
	var user model.UmkmData

	if identifier != "" {
		if _, err := uuid.Parse(identifier); err == nil {
			if result := initializers.DB.Preload("SystemData").Preload("SystemData.RecoveryToken").Where("id = ?", identifier).First(&user); result.Error != nil {
				return nil, result.Error
			}
		} else {
			if result := initializers.DB.Where("email = ?", identifier).Preload("SystemData").First(&user); result.Error != nil {
				return nil, result.Error
			}
		}
	} else {
		return nil, errors.New("identifier is empty")
	}
	return &user, nil
}

func GetParticipantById(c *gin.Context) {
	id := c.Param("id")
	user, err := getParticipantByIdentifier(id)
	if err != nil {
		response.GlobalResponse(c, fmt.Sprintf("Participant_data with ID: %s not found", id), http.StatusBadRequest, err)
		return
	}

	response.GlobalResponse(c, fmt.Sprintf("Successfully retrieving user %s data", id), http.StatusOK, *user)
}

func GetParticipantByEmail(c *gin.Context) {
	email := c.Param("email")

	user, err := getParticipantByIdentifier(email)
	if err != nil {
		response.GlobalResponse(c, fmt.Sprintf("Participant_data with email %s not found", email), http.StatusBadRequest, err)
		return
	}

	response.GlobalResponse(c, fmt.Sprintf("Successfully retrieving user with email %s", email), http.StatusOK, user)
}

func DeleteUserById(c *gin.Context) {
	id := c.Param("id")
	user, err := getParticipantByIdentifier(id)
	if err != nil {
		response.GlobalResponse(c, fmt.Sprintf("Participant_data with email %s not found", id), http.StatusBadRequest, err)
		return
	}
	fmt.Println(user)
	if err := initializers.DB.Preload("RecoveryToken").Unscoped().Delete(&user.SystemData).Error; err != nil {
		response.GlobalResponse(c, fmt.Sprintf("Error deleting participant with ID %s", id), http.StatusInternalServerError, err)
		return
	}

	response.GlobalResponse(c, fmt.Sprintf("Successfully deleted user with ID %s", id), http.StatusOK, nil)
}

func UpdateParticipantById(c *gin.Context) {
	id := c.Param("id")
	participant, err := getParticipantByIdentifier(id)
	if err != nil {
		response.GlobalResponse(c, err.Error(), http.StatusUnauthorized, nil)
		return
	}
	var req request.UmkmRequest
	if err := c.Bind(&req); err != nil {
		response.GlobalResponse(c, "Error binding the requested data", http.StatusBadRequest, err)
		return
	}

	message, err, status, participant := validateParticipantRequest(req, participant)
	if err != nil || status != 200 {
		response.GlobalResponse(c, message, status, nil)
		return
	}

	if len(req.Level) != 0 {
		if req.Level != model.LevelUser && req.Level != model.LevelAdmin {
			response.GlobalResponse(c, "Invalid level(only admin and user)", 400, nil)
		}
		participant.SystemData.Level = req.Level
	}

	if err := initializers.DB.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&participant).Error; err != nil {
		response.GlobalResponse(c, "Failed to update participant data", http.StatusInternalServerError, nil)
		return
	}
	response.GlobalResponse(c, message, http.StatusOK, participant)

}

func AddDevice(c *gin.Context) {
	device := model.Device{
		ID: uuid.New(),
	}

	if err := model.CreateDevice(&device); err != nil {
		response.GlobalResponse(c, "Failed to create device", http.StatusInternalServerError, nil)
		return
	}

	response.GlobalResponse(c, "Device created successfully", http.StatusOK, device)
}
