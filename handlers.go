package main

import (
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"os"
	"time"
)

func (s *APIServer) HandleHome(w http.ResponseWriter, r *http.Request) error {
	payload := struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}{
		Name:    "Go auth",
		Version: "1.0",
	}

	return writeJSON(w, http.StatusOK, payload)
}

func (s *APIServer) HandleLogin(w http.ResponseWriter, r *http.Request) error {
	var loginRequest LoginRequest

	err := readJSON(w, r, &loginRequest)

	if err != nil {
		return apiError{Err: "bad login request", Status: http.StatusBadRequest}
	}

	user, err := s.store.SelectUserByUserName(loginRequest.UserName)

	if err != nil {
		return apiError{Err: "user not found", Status: http.StatusBadRequest}
	}

	err = checkPassword(loginRequest.Password, user.Password)

	if err != nil {
		return apiError{Err: "bad user password", Status: http.StatusBadRequest}
	}

	JWTTokenClaims := JWTClaims{
		UserName: loginRequest.UserName,
		Roles:    []int{2, 1, 3, 4},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: createTokenExpirationTimeForJWTToken(),
			Issuer:    os.Getenv("SERVER_IP"),
			IssuedAt:  &jwt.NumericDate{Time: time.Now().UTC()},
			Audience:  []string{r.Header.Get("Referer")},
		},
	}

	refreshTokenClaims := JWTClaims{
		UserName: loginRequest.UserName,
		Roles:    []int{2, 1, 3, 4},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: createTokenExpirationTimeForJWTRefreshToken(),
			Issuer:    os.Getenv("SERVER_IP"),
			IssuedAt:  &jwt.NumericDate{Time: time.Now().UTC()},
			Audience:  []string{r.Header.Get("Referer")},
		},
	}

	encodedJWT, _ := generateJWTToken(JWTTokenClaims)
	encodedRefreshToken, _ := generateJWTToken(refreshTokenClaims)

	_, err = s.store.UpdateUserRefreshToken(user.ID, encodedRefreshToken)
	log.Printf("err: %+v\n", err)
	if err != nil {
		return apiError{Err: "Internal server error", Status: http.StatusInternalServerError}
	}

	cookie := http.Cookie{
		Name:     "jwt",
		Value:    encodedRefreshToken,
		HttpOnly: true,
		MaxAge:   24 * 60 * 60,
		Secure:   !isEnvDev(),
		Path:     "/",
		SameSite: 4,
	}

	http.SetCookie(w, &cookie)

	return writeJSON(w, http.StatusOK, LoginResponse{encodedJWT})
}

func (s *APIServer) HandleLogout(w http.ResponseWriter, r *http.Request) error {
	jwtCookie, err := r.Cookie("jwt")

	if err != nil {
		return writeJSON(w, http.StatusNoContent, "")
	}

	user, err := s.store.SelectUserByRefreshToken(jwtCookie.Value)

	if err != nil {
		cookie := http.Cookie{
			Name:     "jwt",
			Value:    "",
			HttpOnly: true,
			MaxAge:   -1,
			Secure:   !isEnvDev(),
			Path:     "/",
			SameSite: 4,
		}
		http.SetCookie(w, &cookie)
		return apiError{Err: "user not found", Status: http.StatusForbidden}
	}

	_, err = s.store.UpdateUserRefreshToken(user.ID, "")

	if err != nil {
		return apiError{Err: "internal server error", Status: http.StatusInternalServerError}
	}

	cookie := http.Cookie{
		Name:     "jwt",
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
		Secure:   !isEnvDev(),
		Path:     "/",
		SameSite: 4,
	}
	http.SetCookie(w, &cookie)

	return writeJSON(w, http.StatusNoContent, "Logout successful")
}

func (s *APIServer) HandlePostRefreshToken(w http.ResponseWriter, r *http.Request) error {
	var postRefreshTokenRequest PostRefreshTokenRequest

	err := readJSON(w, r, &postRefreshTokenRequest)

	if err != nil {
		return apiError{Err: "refresh token not found", Status: http.StatusUnauthorized}
	}

	user, err := s.store.SelectUserByRefreshToken(postRefreshTokenRequest.RefreshToken)

	if err != nil {
		return apiError{Err: "user not found", Status: http.StatusUnauthorized}
	}

	token, err := parseToken(postRefreshTokenRequest.RefreshToken)

	if err != nil {
		return apiError{Err: err.Error(), Status: http.StatusUnauthorized}
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		userName := claims.UserName

		if user.UserName != userName {
			log.Println("user name from refresh token not match with user name from db")
			return apiError{Err: "unauthorized", Status: http.StatusUnauthorized}
		}

		JWTTokenClaims := JWTClaims{
			UserName: user.UserName,
			Roles:    []int{2, 1, 3, 4},
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: createTokenExpirationTimeForJWTToken(),
				Issuer:    os.Getenv("SERVER_IP"),
				IssuedAt:  &jwt.NumericDate{Time: time.Now().UTC()},
				Audience:  []string{r.Header.Get("Referer")},
			},
		}

		encodedJWT, _ := generateJWTToken(JWTTokenClaims)

		log.Println("encodedJWT: ", encodedJWT)

		return writeJSON(w, http.StatusOK, LoginResponse{encodedJWT})

	} else {
		return apiError{Err: "JWT Claims are not correct", Status: http.StatusUnauthorized}
	}

}

func (s *APIServer) HandleRefreshToken(w http.ResponseWriter, r *http.Request) error {
	jwtCookie, err := r.Cookie("jwt")

	if err != nil {
		return apiError{Err: "refresh token not found", Status: http.StatusUnauthorized}
	}

	user, err := s.store.SelectUserByRefreshToken(jwtCookie.Value)

	if err != nil {
		return apiError{Err: "user not found", Status: http.StatusUnauthorized}
	}

	token, err := parseToken(jwtCookie.Value)

	if err != nil {
		return apiError{Err: err.Error(), Status: http.StatusUnauthorized}
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		userName := claims.UserName

		if user.UserName != userName {
			log.Println("user name from refresh token not match with user name from db")
			return apiError{Err: "unauthorized", Status: http.StatusUnauthorized}
		}

		JWTTokenClaims := JWTClaims{
			UserName: user.UserName,
			Roles:    []int{2, 1, 3, 4},
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: createTokenExpirationTimeForJWTToken(),
				Issuer:    os.Getenv("SERVER_IP"),
				IssuedAt:  &jwt.NumericDate{Time: time.Now().UTC()},
				Audience:  []string{r.Header.Get("Referer")},
			},
		}

		encodedJWT, _ := generateJWTToken(JWTTokenClaims)

		return writeJSON(w, http.StatusOK, LoginResponse{encodedJWT})

	} else {
		return apiError{Err: "JWT Claims are not correct", Status: http.StatusUnauthorized}
	}

}

func (s *APIServer) HandleCreateUser(w http.ResponseWriter, r *http.Request) error {
	var createUserRequest CreateUserRequest

	err := readJSON(w, r, &createUserRequest)

	if err != nil {
		return apiError{Err: "bad create user request", Status: http.StatusBadRequest}
	}

	password, err := hashPassword(createUserRequest.Password)

	if err != nil {
		return apiError{Err: err.Error(), Status: http.StatusBadRequest}
	}

	userId, err := s.store.InsertUser(User{
		UserName: createUserRequest.UserName,
		Password: password,
	})

	if err != nil {
		return apiError{Err: err.Error(), Status: http.StatusBadRequest}
	}

	return writeJSON(w, http.StatusCreated, userId)
}
