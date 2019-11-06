package main

import (
	"bufio"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/dgryski/dgoogauth"
	"rsc.io/qr"
)

var issuer = "Test"
var qrFilename = filepath.FromSlash("tmp/qr.png")
var totpConfig dgoogauth.OTPConfig

type QRCode struct {
	Base64Image string `json:"base64_image"`
}

func InitGoogleAuthenticator() {
	// Generate random secret
	secret := make([]byte, 10)
	_, err := rand.Read(secret)
	CheckError(err)

	secretBase32 := base32.StdEncoding.EncodeToString(secret)

	// The OTPConfig gets modified by otpc.Authenticate() to prevent passcode replay, etc.,
	// so allocate it once and reuse it for multiple calls.
	totpConfig = dgoogauth.OTPConfig{
		Secret:      secretBase32,
		WindowSize:  5,
		HotpCounter: 0,
		UTC:         true,
	}

	// Create directory to store temp files
	if !DirectoryExists("tmp") {
		err = os.Mkdir("tmp", 0777)
		CheckError(err)
	}
}

func HandleGenerateQRCode(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		var user User
		_ = json.NewDecoder(r.Body).Decode(&user)
		ctx := r.Context()

		if err := db.PingContext(ctx); err != nil {
			CheckError(err)
		}

		err := db.QueryRowContext(ctx,
			`SELECT email
				FROM auth.public.users WHERE email = $1`, user.Email,
		).Scan(
			&user.Email,
		)

		if err != nil {
			CheckError(err)
			// Return error here if user wasn't found
			http.Error(w, "User not found", http.StatusNotFound)
		}
		qrPath := GenerateQRCode(user.Email)
		qrFile, err := os.Open(qrPath)

		if err != nil {
			CheckError(err)
			http.Error(w, "Error generating QR code", http.StatusInternalServerError)
			return
		}

		defer qrFile.Close()

		// Create a new buffer based on file size
		fileInfo, _ := qrFile.Stat()
		var size int64 = fileInfo.Size()
		buffer := make([]byte, size)

		// Read file into buffer
		fileReader := bufio.NewReader(qrFile)
		fileReader.Read(buffer)

		// Convert buffer bytes to base64 string
		qrBase64 := base64.StdEncoding.EncodeToString(buffer)
		var qrCode QRCode
		qrCode.Base64Image = qrBase64

		json.NewEncoder(w).Encode(qrCode)
	}
}

func GenerateQRCode(email string) string {
	URL, err := url.Parse("otpauth://totp")
	CheckError(err)

	URL.Path += "/" + url.PathEscape(issuer) + ":" + url.PathEscape(email)

	params := url.Values{}
	params.Add("secret", totpConfig.Secret)
	params.Add("issuer", issuer)

	URL.RawQuery = params.Encode()
	code, err := qr.Encode(URL.String(), qr.Q)
	CheckError(err)

	b := code.PNG()
	err = ioutil.WriteFile(qrFilename, b, 0777)
	CheckError(err)

	return qrFilename
}

func HandleVerifyGoogleAuthenticator(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		var code VerificationCode
		_ = json.NewDecoder(r.Body).Decode(&code)
		validCode := verify2FACode(code.Code)

		if validCode {
			// Find user in database to pass to CreateJWT
			ctx := r.Context()
			var user User

			if err := db.PingContext(ctx); err != nil {
				CheckError(err)
			}

			err := db.QueryRowContext(ctx,
				`SELECT id,
				username,
				email,
				email_verification,
				google_verification
				FROM auth.public.users WHERE email = $1`, code.Email,
			).Scan(
				&user.ID,
				&user.Username,
				&user.Email,
				&user.EmailVerification,
				&user.GoogleVerification,
			)

			if err != nil {
				CheckError(err)
				// Return error here if user wasn't found
				http.Error(w, "User not found", http.StatusNotFound)
			}

			user.Password = ""
			userToken := CustomToken{}
			userToken.Token = CreateJWT(user)
			userToken.User = user
			json.NewEncoder(w).Encode(userToken)
		} else {
			http.Error(w, "Google authenticator code is invalid", http.StatusBadRequest)
		}
	}
}

func verify2FACode(code string) bool {
	val, err := totpConfig.Authenticate(code)
	CheckError(err)

	if !val {
		return false
	}

	return true
}
