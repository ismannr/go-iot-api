package service

import (
	"errors"
	"fmt"
	"gin-crud/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"gin-crud/initializers"
	model "gin-crud/models"
	"gin-crud/request"
	"gin-crud/response"
	"github.com/gin-gonic/gin"
)

func getUmkmByAuth(c *gin.Context) (*model.UmkmData, error) {
	user, err := getUserByAuth(c)
	if err != nil {
		return nil, err
	}

	umkm, ok := user.(*model.UmkmData)
	if !ok {
		return nil, errors.New("invalid user data")
	}
	return umkm, nil
}

func GetUserData(c *gin.Context) {
	user, err := getUmkmByAuth(c)
	if err != nil {
		response.GlobalResponse(c, err.Error(), http.StatusUnauthorized, nil)
		return
	}
	resp := response.BindUserToResponse(user)
	response.GlobalResponse(c, "Successfully retrieving user data", http.StatusOK, resp)
}

func UpdateData(c *gin.Context) {
	user, err := getUmkmByAuth(c)
	if err != nil {
		response.GlobalResponse(c, err.Error(), http.StatusUnauthorized, nil)
		return
	}
	var req request.UmkmRequest
	if err := c.Bind(&req); err != nil {
		response.GlobalResponse(c, "Error binding the requested data", http.StatusBadRequest, err)
		return
	}

	message, err, status, user := validateParticipantRequest(req, user)
	if err != nil || status != 200 {
		response.GlobalResponse(c, message, status, nil)
		return
	}

	if err := initializers.DB.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&user).Error; err != nil {
		response.GlobalResponse(c, "Failed to update user data", http.StatusInternalServerError, nil)
		return
	}
	resp := response.BindUserToResponse(user)
	response.GlobalResponse(c, message, http.StatusOK, resp)

}

func validateParticipantRequest(req request.UmkmRequest, participant *model.UmkmData) (string, error, int, *model.UmkmData) {
	var invalid []string
	var valid []string
	var isSatisfied bool = true

	if len(req.Password) != 0 {
		if len(req.Password) < 8 {
			invalid = append(invalid, "Password must be at least 8 characters")
			isSatisfied = false
		} else if utils.HashIsMatched(participant.SystemData.Password, req.Password) == true {
			invalid = append(invalid, "New password cannot be the same as the previous password")
			isSatisfied = false
		} else if req.Password != req.ConfirmPass {
			invalid = append(invalid, "Password and Confirm Password do not match")
			isSatisfied = false
		} else {
			uppercaseRegex := regexp.MustCompile(`[A-Z]`)
			numberRegex := regexp.MustCompile(`[0-9]`)
			if !uppercaseRegex.MatchString(req.Password) || !numberRegex.MatchString(req.Password) {
				invalid = append(invalid, "Password must contain at least one uppercase letter and one number")
				isSatisfied = false
			} else {
				hashedPassword, err := utils.HashEncoder(req.Password)
				if err != nil {
					return "Error encoding the password", err, 500, nil
				}
				valid = append(valid, "Password")
				participant.SystemData.Password = hashedPassword
			}
		}
	}

	if len(req.PhoneNumber) != 0 && req.PhoneNumber != participant.Phone {
		if !regexp.MustCompile(`^\d{10,14}$`).MatchString(req.PhoneNumber) {
			invalid = append(invalid, "Phone number (must consist of 10-14 digits)")
			isSatisfied = false
		} else {
			valid = append(valid, "Phone Number")
			participant.Phone = req.PhoneNumber
		}
	}

	if len(req.Address) != 0 && req.Address != participant.Address {
		if len(req.Address) < 5 {
			invalid = append(invalid, "Address (min. 5 characters)")
			isSatisfied = false
		} else {
			valid = append(valid, "Address")
			participant.Address = req.Address
		}
	}

	if len(req.Province) != 0 && req.Province != participant.Province {
		valid = append(valid, "Province")
		participant.Province = req.Province
	}

	if len(req.City) != 0 && req.City != participant.City {
		valid = append(valid, "City")
		participant.City = req.City
	}
	if len(req.Dob) != 0 {
		dob, err := utils.ParseDate(req.Dob)
		if err != nil {
			invalid = append(invalid, "Wrong date format!")
			isSatisfied = false
		} else {
			dob = time.Date(dob.Year(), dob.Month(), dob.Day(), 0, 0, 0, 0, time.UTC)
			participantDob := time.Date(participant.Dob.Year(), participant.Dob.Month(), participant.Dob.Day(), 0, 0, 0, 0, time.UTC)

			if !dob.Equal(participantDob) {
				if !utils.IsAdult(req.Dob) {
					invalid = append(invalid, "User must be over 17!")
					isSatisfied = false
				}
				log.Println(dob)
				log.Println(participantDob)
				valid = append(valid, "Date of Birth")
				participant.Dob = dob
			}
		}
	}

	if len(req.BusinessName) != 0 && req.BusinessName != participant.BusinessName {
		valid = append(valid, "Business Name")
		participant.BusinessName = req.BusinessName
	}

	if len(req.BusinessDesc) != 0 && req.BusinessDesc != participant.BusinessDesc {
		valid = append(valid, "Business Description")
		participant.BusinessDesc = req.BusinessDesc
	}

	if !isSatisfied {
		return "Invalid fields: " + strings.Join(invalid, ", "), nil, 400, nil
	}
	return "Updated fields: " + strings.Join(valid, ", "), nil, 200, participant
}

func GetAllUserDevices(c *gin.Context) {
	user, err := getUmkmByAuth(c)
	if err != nil {
		response.GlobalResponse(c, "Invalid user", http.StatusUnauthorized, nil)
		return
	}

	devices, err := model.GetAllUserDevices(initializers.DB, user.ID)
	if err != nil {
		response.GlobalResponse(c, "Failed to retrieve user devices", http.StatusInternalServerError, nil)
		return
	}

	var responseData []response.DeviceResponse

	for _, device := range devices {
		respDevice := response.DeviceResponse{
			ID:          device.ID,
			Name:        device.Name,
			IsActivated: device.IsActivated,
			UmkmDataId:  device.UmkmDataId,
		}
		responseData = append(responseData, respDevice)
	}

	response.GlobalResponse(c, "Successfully retrieved user devices", http.StatusOK, responseData)
}

func GetDeviceById(c *gin.Context) {
	id := c.Param("id")

	uuId, err := uuid.Parse(id)
	if err != nil {
		response.GlobalResponse(c, "Invalid device ID format", http.StatusBadRequest, nil)
		return
	}

	participant, err := getUmkmByAuth(c)
	if err != nil {
		response.GlobalResponse(c, err.Error(), http.StatusUnauthorized, nil)
		return
	}
	device, err := model.GetUserDeviceById(initializers.DB, participant.ID, uuId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.GlobalResponse(c, "Device not found", http.StatusOK, nil)
		} else {
			response.GlobalResponse(c, "Failed to retrieve device", http.StatusInternalServerError, nil)
		}
		return
	}
	resp := request.DeviceRequest{
		ID:          device.ID,
		Name:        device.Name,
		IsActivated: device.IsActivated,
		UmkmDataId:  *device.UmkmDataId,
	}
	response.GlobalResponse(c, "Successfully retrieved device", http.StatusOK, resp)
}

func RegisterDeviceById(c *gin.Context) {
	id := c.Param("id")
	var req request.DeviceRequest

	if req.Name == "" {
		response.GlobalResponse(c, "Device name cannot be empty", http.StatusBadRequest, nil)
		return
	}

	uuId, err := uuid.Parse(id)
	if err != nil {
		response.GlobalResponse(c, "Invalid device ID format", http.StatusBadRequest, nil)
		return
	}

	participant, err := getUmkmByAuth(c)
	if err != nil {
		response.GlobalResponse(c, err.Error(), http.StatusUnauthorized, nil)
		return
	}

	err = model.RegisterDeviceById(initializers.DB, participant.ID, uuId, req.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.GlobalResponse(c, "Device not found or already registered", http.StatusNotFound, nil)
		} else if errors.Is(err, utils.ErrDeviceAlreadyRegistered) {
			response.GlobalResponse(c, "Device already registered", http.StatusNotFound, nil)
		} else {
			response.GlobalResponse(c, "Failed to register device", http.StatusInternalServerError, nil)
		}
		return
	}

	response.GlobalResponse(c, "Successfully registering device", http.StatusOK, nil)
}

func UpdateDeviceName(c *gin.Context) {
	var req request.UmkmRequest
	if err := c.Bind(&req); err != nil {
		response.GlobalResponse(c, "Error binding the requested data", http.StatusBadRequest, err)
		return
	}

	id := c.Param("id")

	uuId, err := uuid.Parse(id)
	if err != nil {
		response.GlobalResponse(c, "Invalid device ID format", http.StatusBadRequest, nil)
		return
	}

	participant, err := getUmkmByAuth(c)
	if err != nil {
		response.GlobalResponse(c, "", http.StatusUnauthorized, nil)
		return
	}

	err = model.UpdateDeviceName(initializers.DB, participant.ID, uuId, req.Name)
	if err != nil {
		response.GlobalResponse(c, "Device not found", http.StatusBadRequest, nil)
		log.Println(err.Error())
		return
	}

	response.GlobalResponse(c, "successfully updating device name", http.StatusOK, nil)
}

func CreateDeviceGroup(c *gin.Context) {
	var req request.UmkmRequest
	var message string
	var status int
	if err := c.Bind(&req); err != nil {
		response.GlobalResponse(c, "Error binding the requested data", http.StatusBadRequest, err)
		return
	}

	user, err := getUmkmByAuth(c)
	if err != nil {
		response.GlobalResponse(c, "", http.StatusUnauthorized, nil)
		return
	}

	err, message, status = model.CreateGrouping(initializers.DB, user.ID, req.GroupName)
	if err != nil {
		response.GlobalResponse(c, message, status, nil)
		log.Println(err.Error())
		return
	}

	response.GlobalResponse(c, message, status, nil)
}

func RenameGroup(c *gin.Context) {
	var req request.GroupRequest
	var message string
	var status int

	id := c.Param("id")

	uuId, err := uuid.Parse(id)
	if err != nil {
		response.GlobalResponse(c, "Invalid device ID format", http.StatusBadRequest, nil)
		return
	}

	if err := c.Bind(&req); err != nil {
		response.GlobalResponse(c, "Error binding the requested data", http.StatusBadRequest, err)
		return
	}

	user, err := getUmkmByAuth(c)
	if err != nil {
		response.GlobalResponse(c, "", http.StatusUnauthorized, nil)
		return
	}

	err, message, status = model.RenameGrouping(initializers.DB, user.ID, uuId, req.NewGroupName)
	if err != nil {
		response.GlobalResponse(c, message, status, nil)
		log.Println(err.Error())
		return
	}

	response.GlobalResponse(c, message, status, nil)
}

func AddDeviceToGroup(c *gin.Context) {
	var req request.UmkmRequest
	var message string
	var status int
	if err := c.Bind(&req); err != nil {
		response.GlobalResponse(c, "Error binding the requested data", http.StatusBadRequest, err)
		return
	}

	id := c.Param("id")

	uuId, err := uuid.Parse(id)
	if err != nil {
		response.GlobalResponse(c, "Invalid device ID format", http.StatusBadRequest, nil)
		return
	}

	user, err := getUmkmByAuth(c)
	if err != nil {
		response.GlobalResponse(c, "", http.StatusUnauthorized, nil)
		return
	}

	err, message, status = model.AssignDeviceToGroup(initializers.DB, uuId, user.ID, req.GroupName)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.GlobalResponse(c, message, status, nil)
		return
	}
	if err != nil {
		response.GlobalResponse(c, message, status, nil)
		log.Println(err.Error())
		return
	}

	response.GlobalResponse(c, message, status, nil)
}

func RemoveDeviceFromGroup(c *gin.Context) {
	var message string
	var status int
	id := c.Param("id")
	uuId, err := uuid.Parse(id)
	if err != nil {
		response.GlobalResponse(c, "Invalid device ID format", http.StatusBadRequest, nil)
		return
	}
	user, err := getUmkmByAuth(c)
	if err != nil {
		response.GlobalResponse(c, "", http.StatusUnauthorized, nil)
		return
	}
	err, message, status = model.UnassignDeviceFromGroup(initializers.DB, uuId, user.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.GlobalResponse(c, message, status, nil)
		return
	}
	if err != nil {
		response.GlobalResponse(c, message, status, nil)
		log.Println(err.Error())
		return
	}
	response.GlobalResponse(c, message, status, nil)
}

func DeleteDeviceById(c *gin.Context) {
	var message string
	var status int

	id := c.Param("id")
	uuId, err := uuid.Parse(id)
	if err != nil {
		response.GlobalResponse(c, "Invalid device ID format", http.StatusBadRequest, nil)
		return
	}

	user, err := getUmkmByAuth(c)
	if err != nil {
		response.GlobalResponse(c, err.Error(), http.StatusUnauthorized, nil)
		return
	}

	err, message, status = model.UnassignDeviceFromGroup(initializers.DB, uuId, user.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.GlobalResponse(c, message, status, nil)
		return
	}
	if err != nil {
		response.GlobalResponse(c, message, status, nil)
		log.Println(err.Error())
		return
	}

	err = model.DeleteDeviceById(initializers.DB, user.ID, uuId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.GlobalResponse(c, "Device not found", http.StatusNotFound, nil)
		} else if errors.Is(err, utils.ErrDeviceAlreadyDeleted) {
			response.GlobalResponse(c, "Device not found", http.StatusNotFound, nil)
		} else {
			response.GlobalResponse(c, "Failed to delete device", http.StatusInternalServerError, nil)
		}
		return
	}

	response.GlobalResponse(c, "Successfully deleted device relationship", http.StatusOK, nil)
}

func GetAllGroup(c *gin.Context) {
	var groups []model.DeviceGrouping

	participant, err := getUmkmByAuth(c)
	if err != nil {
		fmt.Println(1)
		response.GlobalResponse(c, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	if err := initializers.DB.Where("umkm_data_id = ?", participant.ID).Find(&groups).Error; err != nil {
		fmt.Println(2)
		response.GlobalResponse(c, "Failed to retrieve groups", http.StatusInternalServerError, nil)
		return
	}
	fmt.Println(3)
	response.GlobalResponse(c, "Successfully retrieved all groups", http.StatusOK, groups)
}

func GetGroupById(c *gin.Context) {
	var devices []model.Device

	id := c.Param("id")

	groupID, err := uuid.Parse(id)
	if err != nil {
		response.GlobalResponse(c, "Invalid device ID format", http.StatusBadRequest, nil)
		return
	}

	user, err := getUmkmByAuth(c)
	if err != nil {
		response.GlobalResponse(c, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	if err := initializers.DB.Select("id, name, is_activated, group_name, group_id, umkm_data_id").Where("umkm_data_id = ? AND group_id = ?", user.ID, groupID).Find(&devices).Error; err != nil {
		response.GlobalResponse(c, "Failed to retrieve devices", http.StatusInternalServerError, nil)
		return
	}

	if len(devices) == 0 {
		response.GlobalResponse(c, "No devices found", http.StatusNotFound, nil)
		return
	}

	response.GlobalResponse(c, "Successfully retrieved all groups", http.StatusOK, devices)
}

func DeleteGroupById(c *gin.Context) {
	var group model.DeviceGrouping
	var message string
	var status int

	id := c.Param("id")

	groupID, err := uuid.Parse(id)
	if err != nil {
		response.GlobalResponse(c, "Invalid device ID format", http.StatusBadRequest, nil)
		return
	}

	user, err := getUmkmByAuth(c)
	if err != nil {
		fmt.Println(1)
		response.GlobalResponse(c, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	err = initializers.DB.Where("id = ? AND umkm_data_id = ?", groupID, user.ID).
		First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.GlobalResponse(c, "Group not found", http.StatusNotFound, nil)
		return
	} else if err != nil {
		response.GlobalResponse(c, "Failed to retrieve group", http.StatusInternalServerError, nil)
		log.Println(err.Error())
		return
	}

	err, message, status = model.UnassignDevicesFromGroup(initializers.DB, groupID, user.ID)
	if err != nil {
		response.GlobalResponse(c, message, status, nil)
		return
	}

	err = initializers.DB.Unscoped().Delete(&group).Error
	if err != nil {
		response.GlobalResponse(c, "Failed deleting group", http.StatusInternalServerError, nil)
		log.Println(err.Error())
	}
	response.GlobalResponse(c, "Succesfully deleting group", 200, nil)
}
