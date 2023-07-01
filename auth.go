package main

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"strings"
	"time"
)

type JWTClaims struct {
	UserName string `json:"user_name"`
	Roles    []int  `json:"roles"`
	jwt.RegisteredClaims
}

func (claims JWTClaims) Validate() error {
	if claims.UserName == "" {
		return errors.New("user name claims are missing")
	}
	return nil
}

func createTokenExpirationTimeForJWTToken() *jwt.NumericDate {
	ttl := 60 * time.Second
	expirationTime := time.Now().UTC().Add(ttl)
	return &jwt.NumericDate{Time: expirationTime}
}

func createTokenExpirationTimeForJWTRefreshToken() *jwt.NumericDate {
	ttl := 24 * time.Hour
	expirationTime := time.Now().UTC().Add(ttl)
	return &jwt.NumericDate{Time: expirationTime}
}

func generateJWTToken(claims JWTClaims) (string, error) {
	var (
		key []byte
		t   *jwt.Token
		s   string
	)

	key = []byte(os.Getenv("JWT_SECRET"))

	t = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	s, err := t.SignedString(key)

	if err != nil {
		return "", err
	}

	return s, nil
}

func parseToken(jwtString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(jwtString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})
}

func validateUserRoles(userRoles []int, validRoles []int) error {
	if len(validRoles) > 0 {

		isUserHasValidRoles := every(validRoles, func(value int, index int) bool {
			for _, role := range userRoles {

				if role == value {
					return true
				}
			}
			return false
		})

		if !isUserHasValidRoles {
			return errors.New("you do not have enough permissions")
		}
	}

	return nil
}

func isJWTTokenValid(tokenString string, validRoles ...int) error {

	var err error

	token, err := parseToken(tokenString)

	if err != nil {
		return err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		userRoles := claims.Roles
		return validateUserRoles(userRoles, validRoles)
	} else {
		return errors.New("JWT Claims are not correct")
	}
}

func bearerToken(r *http.Request, header string) (string, error) {
	rawToken := r.Header.Get(header)
	pieces := strings.SplitN(rawToken, " ", 2)

	if len(pieces) < 2 {
		return "", errors.New("token with incorrect bearer format")
	}

	token := strings.TrimSpace(pieces[1])

	return token, nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func checkPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
