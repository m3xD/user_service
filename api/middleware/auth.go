package middleware

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
	"user_service/internal/util"
)

type AuthMiddleware struct {
	JwtService util.Jwt
}

func NewAuthMiddleware(jwtService util.Jwt) *AuthMiddleware {
	return &AuthMiddleware{JwtService: jwtService}
}

func (auth *AuthMiddleware) AuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				util.ResponseErr(w, util.ResponseError{
					Status:    "UNAUTHORIZED",
					TimeStamp: time.Now().String(),
					Message:   "authorization header is required",
					Errors:    nil,
				}, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(token, " ")

			if len(parts) != 2 || parts[0] != "Bearer" {
				util.ResponseErr(w, util.ResponseError{
					Status:    "UNAUTHORIZED",
					TimeStamp: time.Now().String(),
					Message:   "authorization header is required",
					Errors:    nil,
				}, http.StatusUnauthorized)
				return
			}
			token = parts[1]

			claims, err := auth.JwtService.ValidateAccessToken(token)
			if err != nil {
				util.ResponseErr(w, util.ResponseError{
					Status:    "UNAUTHORIZED",
					TimeStamp: time.Now().String(),
					Message:   "invalid token",
					Errors:    nil,
				}, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (auth *AuthMiddleware) ACLMiddleware(allowRoles ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value("user").(jwt.MapClaims)
			role := claims["role"].(string)
			for _, allowRole := range allowRoles {
				if role == allowRole {
					next.ServeHTTP(w, r)
					return
				}
			}
			util.ResponseErr(w, util.ResponseError{
				Status:    "FORBIDDEN",
				TimeStamp: time.Now().String(),
				Message:   "forbidden",
				Errors:    nil,
			}, http.StatusForbidden)
		})
	}
}

func (auth *AuthMiddleware) OwnerMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value("user").(jwt.MapClaims)
			role := claims["role"].(string)
			user := claims["userID"].(string)
			if role == "admin" ||
				(role == "user" && (r.URL.Query().Get("id") == "" || user == r.URL.Query().Get("id"))) {
				next.ServeHTTP(w, r)
				return
			}
			util.ResponseErr(w, util.ResponseError{
				Status:    "FORBIDDEN",
				TimeStamp: time.Now().String(),
				Message:   "forbidden",
				Errors:    nil,
			}, http.StatusForbidden)
		})
	}
}
