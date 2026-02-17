package helper

import (
	"crypto/rand"
	crand "crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SendLogs(db *gorm.DB, msg string, field1 string) error {
	const insertSQL = `INSERT INTO logs (msg, field1) VALUES (?, ?)`

	if err := db.Exec(insertSQL, msg, field1).Error; err != nil {
		return err
	}
	return nil
}

func DefaultIfEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// Ex : Kamis, 20 Maret 2025 13:24 WIB
func FormatDate(t time.Time) string {
	days := map[string]string{
		"Sunday":    "Minggu",
		"Monday":    "Senin",
		"Tuesday":   "Selasa",
		"Wednesday": "Rabu",
		"Thursday":  "Kamis",
		"Friday":    "Jumat",
		"Saturday":  "Sabtu",
	}

	months := map[string]string{
		"January":   "Januari",
		"February":  "Februari",
		"March":     "Maret",
		"April":     "April",
		"May":       "Mei",
		"June":      "Juni",
		"July":      "Juli",
		"August":    "Agustus",
		"September": "September",
		"October":   "Oktober",
		"November":  "November",
		"December":  "Desember",
	}

	day := days[t.Weekday().String()]
	month := months[t.Month().String()]
	return fmt.Sprintf("%s, %02d %s %d %02d:%02d WIB", day, t.Day(), month, t.Year(), t.Hour(), t.Minute())
}

func DecodeJwt(tokenP string) *jwt.Token {
	splitted := strings.Split(tokenP, " ")

	tokenPart := splitted[1]

	token, _ := jwt.Parse(tokenPart, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	return token
}

func RandHex16() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return strings.ToUpper(hex.EncodeToString(b[:]))
}

func NullableStr(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

// coba parse beberapa format umum → kembalikan "YYYY-MM-DD HH:MM:SS" (local time) atau ""
func ParseToMySQLDatetime(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	layouts := []string{
		time.RFC3339Nano, time.RFC3339,
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
	}
	var t time.Time
	var err error
	for _, l := range layouts {
		t, err = time.Parse(l, v)
		if err == nil {
			loc, _ := time.LoadLocation("Asia/Jakarta")
			return t.In(loc).Format("2006-01-02 15:04:05")
		}
	}
	return ""
}

func CodeOtpSecure() string {
	// Calculate the number of random bytes needed for the specified length
	numBytes := (4 * 5) / 8

	// Generate random bytes
	randomBytes := make([]byte, numBytes)
	_, err := crand.Read(randomBytes)
	if err != nil {
		return ""
	}

	// Encode the random bytes to base32
	otp := base32.StdEncoding.EncodeToString(randomBytes)

	// Truncate to the desired length
	otp = otp[:4]

	return string(otp)
}

func FormatIDR(amount float64) string {
	// Convert amount to a string with no decimal places
	amountStr := fmt.Sprintf("%.0f", amount)
	n := len(amountStr)

	if n <= 3 {
		return "Rp." + amountStr
	}

	var result []string
	for i, c := range amountStr {
		if (n-i)%3 == 0 && i != 0 {
			result = append(result, ".")
		}
		result = append(result, string(c))
	}

	return "Rp." + strings.Join(result, "")
}

// FormatIDRInt formats an int64 into Indonesian Rupiah with thousand separators.
// Examples:
//
//	0        -> "Rp 0"
//	1200     -> "Rp 1.200"
//	-9876543 -> "Rp -9.876.543"
func FormatIDRInt(v int64) string {
	neg := v < 0
	if neg {
		v = -v
	}

	s := strconv.FormatInt(v, 10)
	var b strings.Builder
	// "Rp " + optional "-" + digits + separators
	b.Grow(4 + 1 + len(s) + len(s)/3)

	b.WriteString("Rp ")
	if neg {
		b.WriteByte('-')
	}

	n := len(s)
	first := n % 3
	if first == 0 {
		first = 3
	}
	b.WriteString(s[:first])
	for i := first; i < n; i += 3 {
		b.WriteByte('.')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

func GenPemodalID() (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	out := make([]byte, 8)

	// 2 huruf
	for i := 0; i < 2; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		out[i] = letters[idx.Int64()]
	}

	// 6 digit
	for i := 2; i < 8; i++ {
		d, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		out[i] = byte('0' + d.Int64())
	}

	return string(out), nil
}

func GenNumeric8() (string, error) {
	val, err := rand.Int(rand.Reader, big.NewInt(100_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08d", val.Int64()), nil
}
func FormatIDRUint(v uint64) string {
	s := strconv.FormatUint(v, 10)
	var b strings.Builder
	b.Grow(len(s) + len(s)/3 + 4)

	n := len(s)
	first := n % 3
	if first == 0 {
		first = 3
	}
	b.WriteString(s[:first])
	for i := first; i < n; i += 3 {
		b.WriteByte('.')
		b.WriteString(s[i : i+3])
	}
	return "Rp " + b.String()
}

func IsValidEmail(email string) bool {
	// Optional 1
	// _, err := mail.ParseAddress(email)
	// return err == nil

	emailRegex := regexp.MustCompile(`^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|.(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`)
	return emailRegex.MatchString(email)
}

func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
