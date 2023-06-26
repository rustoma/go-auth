package main

import "net/http"

type apiError struct {
	Err    string
	Status int
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func (e apiError) Error() string {
	return e.Err
}
