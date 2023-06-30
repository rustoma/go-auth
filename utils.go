package main

import "os"

func every[T comparable](slice []T, callback func(value T, index int) bool) bool {
	length := len(slice)

	for i := 0; i < length; i++ {
		value := slice[i]

		if !callback(value, i) {
			return false
		}
	}

	return true
}

func isEnvDev() bool {
	if os.Getenv("APP_ENV") == "dev" {
		return true
	}
	return false
}
