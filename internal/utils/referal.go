package utils

import "github.com/google/uuid"

func GenerateReferralCode() string {
	uuidCode := uuid.New()
	return uuidCode.String()
}
