package main

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"net/http"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

type CreateUserRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type LoginRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

func NewApiServer(listenAddr string, storage Storage) *APIServer {
	return &APIServer{
		listenAddr,
		storage,
	}
}

func (s *APIServer) Run() {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	//mux.Use(app.enableCORS)

	mux.Get("/api/v1/refresh", makeHTTPHandler(s.HandleRefreshToken))
	mux.Get("/api/v1/logout", makeHTTPHandler(s.HandleLogout))

	mux.Post("/api/v1", s.requireAuth(2, 3)(makeHTTPHandler(s.HandleHome)))
	mux.Post("/api/v1/login", makeHTTPHandler(s.HandleLogin))

	mux.Post("/api/v1/users", makeHTTPHandler(s.HandleCreateUser))

	log.Println("JSON API server running on port: ", s.listenAddr)

	//err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), mux)
	err := http.ListenAndServe(s.listenAddr, mux)
	if err != nil {
		log.Fatal(err)
	}

}

func makeHTTPHandler(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			if e, ok := err.(apiError); ok {
				_ = writeJSON(w, e.Status, e)
				return
			}
			_ = writeJSON(w, http.StatusInternalServerError, apiError{Err: "Internal sever", Status: http.StatusInternalServerError})
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func readJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1024 * 1024 //one megabyte

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	dec.DisallowUnknownFields()

	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
