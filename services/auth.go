package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	helper "superapps/helpers"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type registerPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Fullname string `json:"fullname"`
}

type loginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userRow struct {
	ID        int64      `json:"id"`
	UID       string     `json:"uid"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	Phone     *string    `json:"phone"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type profileRow struct {
	ID        int64      `json:"id"`
	Fullname  *string    `json:"fullname"`
	UserID    string     `json:"user_id"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func signJWT(userUID, email string) (string, error) {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		// jangan bocorin detail secret kosong ke client
		return "", errors.New("INTERNAL_SERVER_ERROR")
	}

	ttlMin := 60
	if v := strings.TrimSpace(os.Getenv("JWT_TTL_MINUTES")); v != "" {
		// optional: kalau mau parse, tapi biar aman default aja
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"uid":   userUID,
		"email": email,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Duration(ttlMin) * time.Minute).Unix(),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

// =====================================================
// REGISTER
// Body: {email, password, phone, fullname}
// =====================================================
func Register(r *http.Request) (map[string]any, error) {
	var payload registerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return map[string]any{"message": "INVALID_JSON"}, errors.New("INVALID_JSON")
	}

	email := normalizeEmail(payload.Email)
	pass := strings.TrimSpace(payload.Password)
	phone := strings.TrimSpace(payload.Phone)
	fullname := strings.TrimSpace(payload.Fullname)

	if email == "" || pass == "" {
		return map[string]any{"message": "BAD_REQUEST"}, errors.New("BAD_REQUEST")
	}
	if len(pass) < 6 {
		return map[string]any{"message": "BAD_REQUEST", "reason": "PASSWORD_MIN_6"}, errors.New("BAD_REQUEST")
	}

	// check email exists
	var exist []struct {
		Total int64 `json:"total"`
	}
	if err := dbDefault.Raw(`SELECT COUNT(*) AS total FROM users WHERE LOWER(email) = LOWER(?)`, email).
		Scan(&exist).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}
	if len(exist) > 0 && exist[0].Total > 0 {
		return map[string]any{"message": "EMAIL_EXISTS"}, errors.New("EMAIL_EXISTS")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		helper.Logger("error", "bcrypt: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	uid := uuid.NewV4().String()

	tx := dbDefault.Begin()
	if tx.Error != nil {
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	// insert users
	if err := tx.Exec(`
		INSERT INTO users (uid, email, password, phone, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`, uid, email, string(hashed), phone).Error; err != nil {
		tx.Rollback()
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	// insert profiles (optional fullname)
	if fullname != "" {
		if err := tx.Exec(`
			INSERT INTO profiles (fullname, user_id, created_at, updated_at)
			VALUES (?, ?, NOW(), NOW())
		`, fullname, uid).Error; err != nil {
			tx.Rollback()
			helper.Logger("error", "In Server: "+err.Error())
			return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
		}
	} else {
		// kalau mau selalu ada row profile walau kosong:
		if err := tx.Exec(`
			INSERT INTO profiles (fullname, user_id, created_at, updated_at)
			VALUES (NULL, ?, NOW(), NOW())
		`, uid).Error; err != nil {
			tx.Rollback()
			helper.Logger("error", "In Server: "+err.Error())
			return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
		}
	}

	if err := tx.Commit().Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	token, err := signJWT(uid, email)
	if err != nil {
		return map[string]any{}, err
	}

	return map[string]any{
		"message": "REGISTER_SUCCESS",
		"token":   token,
		"user": map[string]any{
			"uid":      uid,
			"email":    email,
			"phone":    phone,
			"fullname": fullname,
		},
	}, nil
}

// =====================================================
// LOGIN
// Body: {email, password}
// =====================================================
func Login(r *http.Request) (map[string]any, error) {
	var payload loginPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return map[string]any{"message": "INVALID_JSON"}, errors.New("INVALID_JSON")
	}

	email := normalizeEmail(payload.Email)
	pass := strings.TrimSpace(payload.Password)

	if email == "" || pass == "" {
		return map[string]any{"message": "BAD_REQUEST"}, errors.New("BAD_REQUEST")
	}

	// get user
	var u userRow
	if err := dbDefault.Raw(`
		SELECT id, uid, email, password, phone, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER(?)
		LIMIT 1
	`, email).Scan(&u).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	if strings.TrimSpace(u.UID) == "" {
		return map[string]any{"message": "USER_NOT_FOUND"}, errors.New("USER_NOT_FOUND")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pass)); err != nil {
		return map[string]any{"message": "INVALID_CREDENTIALS"}, errors.New("INVALID_CREDENTIALS")
	}

	// profile
	var pr profileRow
	_ = dbDefault.Raw(`
		SELECT id, fullname, user_id, created_at, updated_at
		FROM profiles
		WHERE user_id = ?
		LIMIT 1
	`, u.UID).Scan(&pr).Error

	fullname := ""
	if pr.Fullname != nil {
		fullname = strings.TrimSpace(*pr.Fullname)
	}

	phone := ""
	if u.Phone != nil {
		phone = strings.TrimSpace(*u.Phone)
	}

	token, err := signJWT(u.UID, u.Email)
	if err != nil {
		return map[string]any{}, err
	}

	return map[string]any{
		"message": "LOGIN_SUCCESS",
		"token":   token,
		"user": map[string]any{
			"uid":      u.UID,
			"email":    u.Email,
			"phone":    phone,
			"fullname": fullname,
		},
	}, nil
}

// =====================================================
// LOGOUT
// (stateless JWT) -> cuma response sukses
// =====================================================
func Logout(r *http.Request) (map[string]any, error) {
	// Kalau kamu pakai blacklist token, taruh logic di sini.
	return map[string]any{
		"message": "LOGOUT_SUCCESS",
	}, nil
}
