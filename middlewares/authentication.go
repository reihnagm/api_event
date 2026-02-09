package middlewares

import (
	"context"
	"net/http"
	"os"
	"strings"
	helper "superapps/helpers"

	"github.com/dgrijalva/jwt-go"
)

type contextKey string

const userKey contextKey = "user"

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight (OPTIONS) requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

func JwtAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var url string = r.URL.Path

		// Allow images to be accessed without authentication
		if strings.Contains(url, "jpg") {
			next.ServeHTTP(w, r)
			return
		}

		// Define public paths that do not require authentication
		publicPaths := []string{
			"/api/v1/auth/login",
			"/api/v1/auth/logout",
			"/api/v1/auth/admin/login",
			"/api/v1/callback",
			"/api/v1/project-payment-callback",
			"/api/v1/auth/register",
			"/api/v1/project/list",
			"/api/v1/administration/province",
			"/api/v1/verify-otp",
			"/api/v1/broadcast/list",
			"/api/v1/logs",
			"/api/v1/resend-otp",
		}

		// Allow access to job-detail/:id
		if strings.HasPrefix(r.URL.Path, "/api/v1/project/detail/") {
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/v1/administration/city/") {
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/v1/administration/district/") {
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/v1/administration/subdistrict/") {
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/v1/broadcast/detail/") {
			next.ServeHTTP(w, r)
			return
		}

		// Allow access to job-detail/:id
		// if strings.HasPrefix(r.URL.Path, "/api/v1/job-detail/") {
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		for _, path := range publicPaths {
			if r.URL.Path == path {
				next.ServeHTTP(w, r)
				return
			}
		}

		tokenHeader := r.Header.Get("Authorization")

		if tokenHeader == "" {
			helper.Logger("error", "In Server: Missing auth token")
			helper.Response(w, http.StatusUnauthorized, true, "Missing auth token", map[string]any{})
			return
		}

		splitted := strings.Split(tokenHeader, " ")

		if len(splitted) != 2 {
			helper.Logger("error", "In Server: Invalid token format")
			helper.Response(w, http.StatusUnauthorized, true, "Invalid token format", map[string]any{})
			return
		}

		tokenPart := splitted[1]
		claims := jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(tokenPart, claims, func(token *jwt.Token) (any, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			helper.Logger("error", "In Server: Token is invalid")
			helper.Response(w, http.StatusUnauthorized, true, "Token is invalid", map[string]any{})
			return
		}

		ctx := context.WithValue(r.Context(), userKey, claims)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func CreateToken(userId string) (map[string]string, error) {

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["id"] = userId
	// claims["exp"] = time.Now().Add(time.Hour * 168).Unix()

	access, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]string{"token": access}, nil
}
