package service

import (
	"crypto/tls"
	"fmt"
	"gin-crud/request"
	gomail "gopkg.in/mail.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func mailSender(req request.EmailRequest) (string, error) {
	mail := gomail.NewMessage()
	mail.SetHeader("From", os.Getenv("MAIL_SENDER"))
	mail.SetHeader("To", req.EmailAddressToSend)
	mail.SetHeader("Subject", req.Subject)
	mail.Embed(req.ImagePath)
	mail.SetBody("text/html", req.HtmlBody)

	dialer := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("MAIL_SENDER"), os.Getenv("MAIL_PASSWORD"))
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := dialer.DialAndSend(mail); err != nil {
		log.Println(err)
		return "Failed to send email", err
	}
	message := fmt.Sprintf("Email has been sent to %s", req.EmailAddressToSend)
	return message, nil
}
func htmlRenderer(template string) ([]byte, string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Println("Error getting image:", err)
	}
	htmlDir := filepath.Join(currentDir, "html")
	htmlFilePath := filepath.Join(htmlDir, template)

	htmlContent, err := ioutil.ReadFile(htmlFilePath)
	if err != nil {
		log.Println("Error reading HTML file:", err)
		return nil, "", err
	}

	directoryPath := filepath.Join(currentDir, "file")
	filePath := filepath.Join(directoryPath, os.Getenv("LOGO_PUTIH"))
	return htmlContent, filePath, nil
}
func RegistrationMail(emailAddress string, name string) (string, error) {
	template := "registration_template.html"
	htmlContent, filePath, err := htmlRenderer(template)
	if err != nil {
		log.Println("Error reading HTML file:", err)
		return "Failed reading HTML file", err
	}
	htmlBody := fmt.Sprintf(string(htmlContent), filepath.Base(filePath), name)
	mailRequest := request.EmailRequest{
		EmailAddressToSend: emailAddress,
		Subject:            "Account Registration",
		ImagePath:          filePath,
		HtmlBody:           htmlBody,
	}
	_, err = mailSender(mailRequest)
	if err != nil {
		log.Println("Failed to send mail: " + err.Error())
		return "Failed to send the email", err
	}
	return "Successfully sending registration confirmation to your email", nil
}

func ForgotPasswordMail(emailAddress string, name string, url string) (string, error) {
	template := "reset_password_template.html"
	htmlContent, filePath, err := htmlRenderer(template)
	if err != nil {
		log.Println("Error reading HTML file:", err)
		return "Failed reading HTML file", err
	}
	htmlBody := fmt.Sprintf(string(htmlContent), filepath.Base(filePath), name, url)
	mailRequest := request.EmailRequest{
		EmailAddressToSend: emailAddress,
		Subject:            "Password Recovery",
		ImagePath:          filePath,
		HtmlBody:           htmlBody,
	}
	_, err = mailSender(mailRequest)
	if err != nil {
		log.Println("Failed to send mail: " + err.Error())
		return "Failed to send the email", err
	}
	return "Successfully sending reset password code to your email", nil
}
