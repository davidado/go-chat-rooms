// Package auth provides authentication utilities for the chat service.
package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/davidado/go-chat-rooms/config"

	"github.com/Pallinder/go-randomdata"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

// UserKey is the key used to store the username in the context.
const UserKey contextKey = "username"

// CreateJWT creates a new JWT token with the given secret and username.
func CreateJWT(secret string, username string) (string, error) {
	expiration := time.Second * time.Duration(config.Envs.JWTExpirationInSeconds)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username":  username,
		"expiredAt": time.Now().Add(expiration).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// WithJWTAuth is a middleware that checks for a valid JWT token in the request.
func WithJWTAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := getTokenFromCookie(r)
		if tokenString == "" {
			tokenString = getTokenFromRequest(r)
		}

		shouldReturn := handleEmptyTokenString(tokenString, handlerFunc, w, r)
		if shouldReturn {
			return
		}

		token, err := validateToken(tokenString)
		if err != nil {
			log.Println("unable to validate token:", err)
			permissionDenied(w)
			return
		}
		if !token.Valid {
			log.Println("invalid token")
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		username := claims["username"].(string)

		r = setUsernameInContext(r, username)
		handlerFunc(w, r)
	}
}

func getTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(config.Envs.CookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func getTokenFromRequest(r *http.Request) string {
	tokenAuth := r.Header.Get("Authorization")
	if tokenAuth == "" {
		return tokenAuth
	}
	return ""
}

func handleEmptyTokenString(tokenString string, handlerFunc http.HandlerFunc, w http.ResponseWriter, r *http.Request) bool {
	if tokenString == "" {
		if config.Envs.AllowAnon {
			username := randomdata.SillyName()
			r = setUsernameInContext(r, username)
			handlerFunc(w, r)
			return true
		}
		permissionDenied(w)
		return true
	}
	return false
}

func validateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method alg for security.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Envs.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func permissionDenied(w http.ResponseWriter) {
	http.Error(w, "Permission Denied", http.StatusForbidden)
}

func setUsernameInContext(r *http.Request, username string) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, UserKey, username)
	r = r.WithContext(ctx)
	return r
}

// GetUsernameFromContext returns the username stored in the context.
func GetUsernameFromContext(ctx context.Context) string {
	username, ok := ctx.Value(UserKey).(string)
	if !ok {
		return ""
	}
	return username
}
