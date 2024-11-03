package utils

import (
	"crypto/rand"
	"log"
	"os"
	"time"

	"gopkg.in/gomail.v2"
)

type OTPData struct {
	OTP        string
	ExpiryTime time.Time
}

func GenerateOTP(length int, expiryDuration time.Duration) (*OTPData, error) {
	const otpChars = "1234567890"
	otp := make([]byte, length)
	_, err := rand.Read(otp)
	if err != nil {
		return nil, err
	}

	for i := 0; i < length; i++ {
		otp[i] = otpChars[otp[i]%byte(len(otpChars))]
	}
	otpData := &OTPData{
		OTP:        string(otp),
		ExpiryTime: time.Now().Add(expiryDuration),
	}
	return otpData, nil
}

func SendEmailOTP(to string, otp string) error {
	mail := gomail.NewMessage()
	mail.SetHeader("From", "gfreshproject2024@gmail.com")
	mail.SetHeader("To", to)
	mail.SetHeader("Subject", "Your OTP Code")
	mail.SetBody("text/html", "Your OTP is: "+otp)
	dialermail := os.Getenv("Mail")

	dialer := gomail.NewDialer("smtp.gmail.com", 587, "gfreshproject2024@gmail.com", dialermail)

	err := dialer.DialAndSend(mail)
	if err != nil {
		log.Printf("Failed to send OTP email to %s: %v\n", to, err)
		return err
	}
	log.Printf("OTP email successfully sent to %s\n", to)
	return nil
}
