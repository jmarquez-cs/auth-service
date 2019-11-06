package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/dgryski/dgoogauth"
	"github.com/sfreiberg/gotwilio"
)

var smsConfig dgoogauth.OTPConfig

func InitTwilio() {
	// Generate random secret
	secret := make([]byte, 10)
	_, err := rand.Read(secret)
	CheckError(err)

	secretBase32 := base32.StdEncoding.EncodeToString(secret)

	// The OTPConfig gets modified by otpc.Authenticate() to prevent passcode replay, etc.,
	// so allocate it once and reuse it for multiple calls.
	smsConfig = dgoogauth.OTPConfig{
		Secret:      secretBase32,
		WindowSize:  5,
		HotpCounter: 1,
		UTC:         true,
	}
}

func HandleSendSMSCode(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		var user User
		_ = json.NewDecoder(r.Body).Decode(&user)

		ctx := r.Context()
		if err := db.PingContext(ctx); err != nil {
			CheckError(err)
		}

		err := db.QueryRowContext(ctx,
			`SELECT
			phone_number
			FROM auth.public.users WHERE email = $1`, user.Email,
		).Scan(
			&user.PhoneNumber,
		)

		if err != nil {
			CheckError(err)
			// Return error here if user wasn't found
			http.Error(w, "User not found", http.StatusBadRequest)
		}

		codeSent := SendSMSCode(user)

		if codeSent {
			w.Write([]byte("SMS verification code sent"))
		} else {
			http.Error(w, "Error sending SMS", http.StatusInternalServerError)
		}
	}
}

func SendSMSCode(user User) bool {
	// Value parameter for ComputeCode must match window size in OTPConfig
	smsVerificationCode := dgoogauth.ComputeCode(smsConfig.Secret, int64(smsConfig.HotpCounter))
	smsMessage := "Your Test verification code is: " + strconv.Itoa(smsVerificationCode)
	toNumber := "+1" + user.PhoneNumber
	return SendSMS(smsMessage, toNumber)
}

func getTwilioAccount() string {
	twilioAccount, found := os.LookupEnv("TWILIO_ACCOUNT_SID")
	if !found {
		twilioAccountError := errors.New("Error finding Twilio account SID")
		CheckError(twilioAccountError)
		return ""
	}

	return twilioAccount
}

func getTwilioToken() string {
	twilioToken, found := os.LookupEnv("TWILIO_AUTH_TOKEN")
	if !found {
		twilioAccountError := errors.New("Error finding Twilio auth token")
		CheckError(twilioAccountError)
		return ""
	}

	return twilioToken
}

func getTwilioPhoneNumber() string {
	twilioPhoneNumber, found := os.LookupEnv("TWILIO_PHONE_NUMBER")
	if !found {
		twilioAccountError := errors.New("Error finding Twilio phone number")
		CheckError(twilioAccountError)
		return ""
	}

	return twilioPhoneNumber
}

func SendSMS(message string, toNumber string) bool {
	twilioAccount := getTwilioAccount()
	twilioToken := getTwilioToken()
	twilioPhoneNumber := getTwilioPhoneNumber()

	twilioClient := gotwilio.NewTwilioClient(twilioAccount, twilioToken)

	response, exception, err := twilioClient.SendSMS(twilioPhoneNumber, toNumber, message, "", "")

	if response == nil {
		CheckError(err)
		return false
	}

	if exception != nil {
		CheckError(err)
		return false
	}

	return true
}

func HandleVerifySMS(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		var code VerificationCode
		_ = json.NewDecoder(r.Body).Decode(&code)
		verified := verifySMSCode(code.Code)

		if verified {
			// Find user in database to pass to CreateJWT
			ctx := r.Context()
			var user User

			if err := db.PingContext(ctx); err != nil {
				CheckError(err)
			}

			err := db.QueryRowContext(ctx,
				`SELECT id,
				username
				FROM auth.public.users WHERE email = $1`, code.Email,
			).Scan(
				&user.ID,
				&user.Username,
			)

			if err != nil {
				CheckError(err)
				// Return error here if user wasn't found
				http.Error(w, "User not found", http.StatusBadRequest)
			}

			userToken := CustomToken{}
			userToken.Token = CreateJWT(user)
			json.NewEncoder(w).Encode(userToken)
		} else {
			http.Error(w, "SMS code is invalid", http.StatusBadRequest)
		}
	}
}

func verifySMSCode(code string) bool {
	val, err := smsConfig.Authenticate(code)
	CheckError(err)

	if !val {
		return false
	}

	return true
}