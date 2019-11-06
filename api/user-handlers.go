package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"github.com/lib/pq"
)

// Init users as slice User struct
var users []User
var results []string

func (s *server) handleRegistration() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ctx := r.Context()
		var user User
		_ = json.NewDecoder(r.Body).Decode(&user)

		if err := s.db.PingContext(ctx); err != nil {
			fmt.Println(err)
		}

		hash := HandleCrypto(user.Password)

		result, err := s.db.ExecContext(ctx, `
			INSERT INTO auth.public.users (
				username,
				email,
				pass
			) VALUES ($1, $2, $3);`,
			user.Username,
			user.Email,
			hash,
		)

		if err != nil {
			switch err.(type) {
			case *pq.Error:
				registrationError := errors.New(err.(*pq.Error).Message)
				fmt.Println(err)
				http.Error(w, err.(*pq.Error).Message, http.StatusInternalServerError)
				return
			default:
				fmt.Println(err)
				http.Error(w, "Error registering user", http.StatusInternalServerError)
				return
			}
		}

		if result != nil {
			// Get id of user just inserted
			err := s.db.QueryRowContext(ctx,
				`SELECT
				id
				FROM auth.public.users WHERE email = $1`, user.Email,
			).Scan(
				&user.ID,
			)

			if err != nil {
				fmt.Println(err)
				// Return error here if user wasn't found
				http.Error(w, "User not found", http.StatusNotFound)
			}

			user.Password = ""
			userToken := CustomToken{}
			userToken.Token = CreateJWT(user)
			userToken.User = user
			json.NewEncoder(w).Encode(userToken)
		} else {
			http.Error(w, "Error registering user", http.StatusInternalServerError)
			return
		}
	}
}

func (s *server) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ctx := r.Context()
		var user User
		_ = json.NewDecoder(r.Body).Decode(&user)
		providedPass := user.Password

		if err := s.db.PingContext(ctx); err != nil {
			fmt.Println(err)
		}

		err := s.db.QueryRowContext(ctx,
			`SELECT id,
			email,
			username,
			pass,
			phone_number,
			email_verification,
			sms_verification,
			google_verification
			FROM auth.public.users WHERE email = $1`, user.Email,
		).Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.Password,
			&user.PhoneNumber,
			&user.EmailVerification,
			&user.SMSVerification,
			&user.GoogleVerification,
		)

		if err != nil {
			// Email address sent by user doesn't exist in our database
			fmt.Println(err)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		match, err := ComparePasswordAndHash(providedPass, user.Password)

		if match {
			// Check if user has 2FA enabled
			if user.EmailVerification || user.SMSVerification || user.GoogleVerification {
				if user.EmailVerification {

					// Send user the verification code by email
					emailSent, verificationCode := EmailCode(user)
					if emailSent {
						// Insert the code into the database to verify a user against
						_, err = s.db.ExecContext(ctx, `
							INSERT INTO auth.public.login_codes (
								email,
								code
							) VALUES ($1, $2);`,
							user.Email,
							verificationCode,
						)

						if err != nil {
							fmt.Println(err)
							http.Error(w, "Error associating verification code to user", http.StatusInternalServerError)
						}

						verificationData := UserVerification{
							User: user,
						}
						json.NewEncoder(w).Encode(verificationData)
						// Redirect to form to input code
					} else {
						http.Error(w, "Error sending email", http.StatusInternalServerError)
					}
				}

				if user.SMSVerification {
					// Send user the verification code by sms
					messageSent := SendSMSCode(user)
					if messageSent {
						verificationData := UserVerification{
							EmailVerification: user.EmailVerification,
							GoogleVerification: user.GoogleVerification,
						}
						json.NewEncoder(w).Encode(verificationData)
						// Redirect to form to input code
					} else {
						http.Error(w, "Error sending SMS", http.StatusInternalServerError)
					}
				}

				if user.GoogleVerification {
					// Redirect to form to type in google authenticator code
					verificationData := UserVerification{
						User: user,
					}
					json.NewEncoder(w).Encode(verificationData)
				}
			} else {
				// If no 2FA enabled, return authenticated jwt
				user.Password = ""
				userToken := CustomToken{}
				userToken.Token = CreateJWT(user)

				userToken.User = user

				json.NewEncoder(w).Encode(userToken)
			}
		} else {
			loginError := errors.New("Invalid Credentials")
			CheckError(loginError)
			http.Error(w, invalidCredentials, http.StatusForbidden)
			return
		}
	}
}