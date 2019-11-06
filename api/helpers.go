package main

import (
	"fmt"
	"os"
	"strconv"
)

// Require missing env variable
func MustEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("%s required on environment variables", key))
	}
	return v
}

// Require env variable if present or fallback on existing value provided
func env(key, fallbackValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallbackValue
	}
	return v
}

// Require env variable if present or string representation of an integer data type or fallback on existing value provided
func intEnv(key string, fallbackValue int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallbackValue
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallbackValue
	}
	return i
}

// FileExists checks if a file exists at a given path
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

// DirectoryExists checks if a directory exists at a given path
func DirectoryExists(directory string) bool {
	info, err := os.Stat(directory)
	if os.IsNotExist(err) {
		return false
	}

	return info.IsDir()
}

func CheckError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
