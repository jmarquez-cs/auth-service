package main

import (
	"database/sql"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type VerificationCode struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

var sendgridApiKey string

func init() {
	godotenv.Load()
	sendgridApiKey = MustEnv("SENDGRID_API_KEY")
}

func HandleEmailCode(db *sql.DB) http.HandlerFunc {
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
			username,
			email
			FROM auth.public.users WHERE email = $1`, user.Email,
		).Scan(
			&user.Username,
			&user.Email,
		)

		if err != nil {
			CheckError(err)
			// Return error here if user wasn't found
			http.Error(w, "User not found", http.StatusBadRequest)
		}

		codeSent, verificationCode := EmailCode(user)

		if codeSent {
			// Insert the code into the database to verify a user against
			_, err = db.ExecContext(ctx, `
				INSERT INTO auth.public.login_codes (
					email,
					code
				) VALUES ($1, $2);`,
				user.Email,
				verificationCode,
			)

			if err != nil {
				CheckError(err)
				http.Error(w, "Error associating verification code to user", http.StatusInternalServerError)
			}

			w.Write([]byte("Email verification code sent"))
		} else {
			http.Error(w, "Error sending email", http.StatusInternalServerError)
		}
	}
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func generateVerificationCode(length int) string {
	rand.Seed(time.Now().UnixNano())
	return strconv.Itoa(randomInt(100000, 999999))
}

func EmailCode(user User) (bool, string) {
	// Generate random six digit number to save in database for user
	emailVerificationCode := generateVerificationCode(6)

	from := mail.NewEmail("Test", "Test@no-reply.com")
	subject := "One time verification code for Test"
	to := mail.NewEmail(user.Username, user.Email)
	plainTextConent := "Test Verification Code"
	htmlContent := "Your Test verification code is: " + emailVerificationCode
	message := mail.NewSingleEmail(from, subject, to, plainTextConent, htmlContent)
	client := sendgrid.NewSendClient(sendgridApiKey)
	response, err := client.Send(message)
	CheckError(err)

	// Client.Send returns 202 on success
	return response.StatusCode == 202, emailVerificationCode
}

func HandleVerifyEmail(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		var code VerificationCode
		_ = json.NewDecoder(r.Body).Decode(&code)
		verified := verifyEmailCode(code.Email, code.Code, db, r)

		if verified {
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
			http.Error(w, "Email code is invalid", http.StatusBadRequest)
		}
	}
}

func verifyEmailCode(email, code string, db *sql.DB, r *http.Request) bool {
	ctx := r.Context()

	if err := db.PingContext(ctx); err != nil {
		CheckError(err)
	}

	result, err := db.ExecContext(ctx, `
		DELETE FROM auth.public.login_codes
		WHERE email = $1
		AND code = $2`,
		email,
		code,
	)

	CheckError(err)

	rows, err := result.RowsAffected()
	CheckError(err)

	return rows > 0
}