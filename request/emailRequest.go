package request

type EmailRequest struct {
	EmailAddressToSend string
	Subject            string
	ImagePath          string
	HtmlBody           string
}
