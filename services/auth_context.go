package services

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func getBearerToken(r *http.Request) string {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// Ambil user UID dari JWT claims ("uid")
func getAuthUIDFromRequest(r *http.Request) string {
	tokenStr := getBearerToken(r)
	if tokenStr == "" {
		return ""
	}

	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return ""
	}

	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("INVALID_SIGNING_METHOD")
		}
		return []byte(secret), nil
	})
	if err != nil || tok == nil || !tok.Valid {
		return ""
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}

	uid, _ := claims["uid"].(string)
	return strings.TrimSpace(uid)
}
