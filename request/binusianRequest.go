package request

type MentorRequest struct {
	MentorId          int64  `json:"mentor_id"`
	MentorName        string `json:"mentor_name"`
	MentorPicture     []byte `json:"mentor_picture"`
	MentorEmployeeId  string `json:"mentor_employee_id"`
	Email             string `json:"email"`
	Password          string `json:"-" json:"password"`
	MentorPhoneNumber string `json:"mentor_phone_number"`
	MentorPosition    string `json:"mentor_position"`
}
