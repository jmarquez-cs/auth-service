package main

import (
	"gopkg.in/square/go-jose.v2/jwt"
)

// User Struct (model)
type User struct {
	ID                 string `json:"id"`
	Username           string `json:"username"`
	Email              string `json:"email"`
	Password           string `json:"password,omitempty"`
	PhoneNumber        string `json:"phone_number,omitempty"`
	EmailVerification  bool   `json:"email_verification"`
	SMSVerification    bool   `json:"sms_verification"`
	GoogleVerification bool   `json:"google_verification"`
}

// If a user has 2FA enabled this will hold those settings to be returned to the front end
type UserVerification struct {
	User User `json:"user"`
}

// CustomClaim is a struct for holding both public and private claims
type CustomClaim struct {
	Claims        *jwt.Claims
	PrivateClaim1 string                 `json:"privateClaim1,omitempty"`
	PrivateClaim2 string                 `json:"privateClaim2,omitempty"`
	AnyClaim      map[string]interface{} `json:"anyClaim"`
}

// CustomToken is a struct to hold the generated JWT token
type CustomToken struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// params is a struct for key the derivative function  
type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}