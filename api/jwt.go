package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

// CustomToken is a struct to hold the generated JWT token
type CustomToken struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// CreateJWT will create a JWT signed by a private RSA key
func CreateJWT(user User) string {
	var data = make(map[string]string)
	data["id"] = user.ID
	data["username"] = user.Username

	jsonData, err := json.Marshal(data)
	CheckError(err)

	authURL := MustEnv("AUTH_URL")
	createURL := authURL + "/create_jwt"

	request, err := http.NewRequest("POST", createURL, bytes.NewReader(jsonData))
	CheckError(err)

	client := &http.Client{
		Timeout: time.Second * 45,
	}

	resp, err := client.Do(request)
	CheckError(err)

	var response struct {
		Token string `json:"token"`
	}

	if err == nil {
		defer resp.Body.Close()
	}

	_ = json.NewDecoder(resp.Body).Decode(&response)

	return response.Token
}

func VerifyJWT(token string) bool {
	var data = make(map[string]string)
	data["token"] = token

	jsonData, err := json.Marshal(data)
	CheckError(err)

	authURL := MustEnv("AUTH_URL")
	verifyURL := authURL + "/verify_jwt"

	request, err := http.NewRequest("POST", verifyURL, bytes.NewReader(jsonData))
	CheckError(err)

	client := &http.Client{
		Timeout: time.Second * 15,
	}

	resp, err := client.Do(request)
	CheckError(err)

	var response struct {
		Verified bool `json:"verified"`
	}

	if err == nil {
		defer resp.Body.Close()
	}

	_ = json.NewDecoder(resp.Body).Decode(&response)

	return response.Verified
}
