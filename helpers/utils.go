package helper

import (
	"bytes"
	"crypto/rand"
	crand "crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	entities "superapps/entities"
	"time"

	// "time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func GetEmailByUID(db *gorm.DB, uid string) (string, error) {
	if db == nil {
		return "", errors.New("nil *gorm.DB")
	}
	var email string
	row := db.Raw(`SELECT email FROM users WHERE uid = ? LIMIT 1`, uid).Row()
	if err := row.Scan(&email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("uid %q not found: %w", uid, err)
		}
		return "", err
	}
	if email == "" {
		return "", fmt.Errorf("uid %q has no email", uid)
	}
	return email, nil
}

func GetEmailAndTitleAndRoleByMediaProject(db *gorm.DB, id string) (string, string, string, error) {
	if db == nil {
		return "", "", "", errors.New("nil *gorm.DB")
	}

	var email, title, role string

	row := db.Raw(`
		SELECT u.email, p.title, r.name AS role
		FROM media_document_verify_projects mdvp
		INNER JOIN document_verify_projects dvp 
			ON dvp.id = mdvp.document_verify_project_id
		INNER JOIN projects p 
			ON p.uid = dvp.project_id
		INNER JOIN companies c 
			ON c.uid = p.company_id
		INNER JOIN users u 
			ON u.uid = c.user_id
		INNER JOIN roles r
			ON u.uid = r.id
		WHERE dvp.id = ?
		LIMIT 1
	`, id).Row()

	if err := row.Scan(&email, &title, &role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", fmt.Errorf("id %q not found: %w", id, err)
		}
		return "", "", "", err
	}

	if email == "" {
		return "", "", "", fmt.Errorf("id %q has no email", id)
	}
	if title == "" {
		return "", "", "", fmt.Errorf("id %q has no title", id)
	}
	if role == "" {
		return "", "", "", fmt.Errorf("id %q has no role", id)
	}

	return email, title, role, nil
}

func GetEmailByEmail(db *gorm.DB, uid string) (string, error) {
	if db == nil {
		return "", errors.New("nil *gorm.DB")
	}
	var email string
	row := db.Raw(`SELECT email FROM users WHERE email = ? LIMIT 1`, uid).Row()
	if err := row.Scan(&email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("uid %q not found: %w", uid, err)
		}
		return "", err
	}
	if email == "" {
		return "", fmt.Errorf("uid %q has no email", uid)
	}
	return email, nil
}

func GetRoleByEmail(db *gorm.DB, uid string) (string, error) {
	if db == nil {
		return "", errors.New("nil *gorm.DB")
	}

	var role string
	row := db.Raw(`SELECT r.name
	FROM roles r
	INNER JOIN users u
	ON u.role = r.id 
	WHERE u.email = ? LIMIT 1`, uid).Row()
	if err := row.Scan(&role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("uid %q not found: %w", uid, err)
		}
		return "", err
	}
	if role == "" {
		return "", fmt.Errorf("uid %q has no email", uid)
	}
	return role, nil
}

func GetRoleByUID(db *gorm.DB, uid string) (string, error) {
	if db == nil {
		return "", errors.New("nil *gorm.DB")
	}

	var role string
	row := db.Raw(`SELECT r.name
	FROM roles r 
	INNER JOIN users u
	ON u.role = r.id 
	WHERE u.uid = ? LIMIT 1`, uid).Row()
	if err := row.Scan(&role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("uid %q not found: %w", uid, err)
		}
		return "", err
	}
	if role == "" {
		return "", fmt.Errorf("uid %q has no email", uid)
	}
	return role, nil
}

func GetEmailInboxByUID(db *gorm.DB, uid string) (string, error) {
	if db == nil {
		return "", errors.New("nil *gorm.DB")
	}
	var email string
	row := db.Raw(`SELECT u.email FROM inboxes i
	INNER JOIN users u ON u.uid = i.user_id
	WHERE i.receiver_id = ? ORDER BY i.created_at DESC LIMIT 1`, uid).Row()
	if err := row.Scan(&email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("uid %q not found: %w", uid, err)
		}
		return "", err
	}
	if email == "" {
		return "", fmt.Errorf("uid %q has no email", uid)
	}
	return email, nil
}

func GetFullnameByUID(db *gorm.DB, uid string) (string, error) {
	if db == nil {
		return "", errors.New("nil *gorm.DB")
	}
	var fullname string
	row := db.Raw(`SELECT fullname FROM profiles WHERE user_id = ? LIMIT 1`, uid).Row()
	if err := row.Scan(&fullname); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("uid %q not found: %w", uid, err)
		}
		return "", err
	}
	if fullname == "" {
		return "", fmt.Errorf("uid %q has no fullname", uid)
	}
	return fullname, nil
}

func GetEmailsByRoleProjectAnalyst(db *gorm.DB) ([]string, error) {
	if db == nil {
		return nil, errors.New("nil *gorm.DB")
	}

	rows, err := db.Raw(`SELECT email FROM users WHERE role = ?`, "6").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err != nil {
			return nil, err
		}
		emails = append(emails, e)
	}
	return emails, rows.Err()
}

func GetEmailsByRoleProjectPublish(db *gorm.DB) ([]string, error) {
	if db == nil {
		return nil, errors.New("nil *gorm.DB")
	}

	rows, err := db.Raw(`SELECT email FROM users WHERE role = ?`, "8").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err != nil {
			return nil, err
		}
		emails = append(emails, e)
	}
	return emails, rows.Err()
}

func GetTitleByProjectId(db *gorm.DB, uid string) (string, error) {
	if db == nil {
		return "", errors.New("nil *gorm.DB")
	}
	var val string

	row := db.Raw(`SELECT title 
	FROM projects
	WHERE uid = ? 
	LIMIT 1`, uid).Row()

	if err := row.Scan(&val); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("uid %q not found: %w", uid, err)
		}
		return "", err
	}

	if val == "" {
		return "", fmt.Errorf("uid %q has no uid", uid)
	}

	return val, nil
}

func GetSkuByProjectId(db *gorm.DB, uid string) (string, error) {
	if db == nil {
		return "", errors.New("nil *gorm.DB")
	}
	var val string

	row := db.Raw(`SELECT sku 
	FROM projects
	WHERE uid = ? 
	LIMIT 1`, uid).Row()

	if err := row.Scan(&val); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("uid %q not found: %w", uid, err)
		}
		return "", err
	}

	if val == "" {
		return "", fmt.Errorf("uid %q has no uid", uid)
	}

	return val, nil
}

func GetEmailAndTitleAndRoleByPayment(db *gorm.DB, paymentID string) (string, string, string, error) {
	if db == nil {
		return "", "", "", errors.New("nil *gorm.DB")
	}

	var (
		email string
		title string
		role  string
	)

	row := db.Raw(`
		SELECT u.email, pr.title, r.name AS role
		FROM payments p 
		INNER JOIN projects pr ON pr.uid = p.project_uid
		INNER JOIN companies c ON c.uid = pr.company_id
		INNER JOIN users u ON u.uid = c.user_id
		INNER JOIN roles r ON r.id = u.role
		WHERE p.uid = ?
		LIMIT 1
	`, paymentID).Row()

	if err := row.Scan(&email, &title, &role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", fmt.Errorf("payment_id %q not found: %w", paymentID, err)
		}
		return "", "", "", err
	}

	if email == "" {
		return "", "", "", fmt.Errorf("payment_id %q has no email", paymentID)
	}
	if title == "" {
		return "", "", "", fmt.Errorf("payment_id %q has no title", paymentID)
	}
	if role == "" {
		return "", "", "", fmt.Errorf("payment_id %q has no role", paymentID)
	}

	return email, title, role, nil
}

func GetEmailAndTitleAndRoleByContractLetterPayment(db *gorm.DB, paymentID string) (string, string, string, error) {
	if db == nil {
		return "", "", "", errors.New("nil *gorm.DB")
	}

	var (
		email string
		title string
		role  string
	)

	row := db.Raw(`
		SELECT u.email, pj.title, r.name AS role
		FROM contract_letter_payments clp
		INNER JOIN payments p ON clp.payment_id = p.id
		INNER JOIN projects pj ON pj.uid = p.project_uid
		INNER JOIN companies c ON c.uid = pj.company_id
		INNER JOIN users u ON u.uid = c.user_id
		INNER JOIN roles r ON r.id = u.role
		WHERE clp.payment_id = ? LIMIT 1
	`, paymentID).Row()

	if err := row.Scan(&email, &title, &role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", fmt.Errorf("payment_id %q not found: %w", paymentID, err)
		}
		return "", "", "", err
	}

	if email == "" {
		return "", "", "", fmt.Errorf("payment_id %q has no email", paymentID)
	}
	if title == "" {
		return "", "", "", fmt.Errorf("payment_id %q has no title", paymentID)
	}
	if role == "" {
		return "", "", "", fmt.Errorf("payment_id %q has no role", paymentID)
	}

	return email, title, role, nil
}

func GetEmailByProjectId(db *gorm.DB, uid string) (string, error) {
	if db == nil {
		return "", errors.New("nil *gorm.DB")
	}
	var email string

	row := db.Raw(`SELECT u.email 
	FROM projects p 
	INNER JOIN companies c ON c.uid = p.company_id 
	INNER JOIN users u ON u.uid = c.user_id
	WHERE p.uid = ? 
	LIMIT 1`, uid).Row()

	if err := row.Scan(&email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("uid %q not found: %w", uid, err)
		}
		return "", err
	}

	if email == "" {
		return "", fmt.Errorf("uid %q has no email", uid)
	}

	return email, nil
}

func GetRoleByProjectId(db *gorm.DB, uid string) (string, error) {
	if db == nil {
		return "", errors.New("nil *gorm.DB")
	}
	var role string

	row := db.Raw(`SELECT r.name AS role 
	FROM projects p 
	INNER JOIN companies c ON c.uid = p.company_id 
	INNER JOIN users u ON u.uid = c.user_id
	INNER JOIN roles r ON r.id = u.role
	WHERE p.uid = ? 
	LIMIT 1`, uid).Row()

	if err := row.Scan(&role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("uid %q not found: %w", uid, err)
		}
		return "", err
	}

	if role == "" {
		return "", fmt.Errorf("uid %q has no role", uid)
	}

	return role, nil
}

func GetPriceByProjectId(db *gorm.DB, uid string) (float64, error) {
	if db == nil {
		return 0, errors.New("nil *gorm.DB")
	}
	var amountIdr float64

	row := db.Raw(`SELECT amount_idr 
	FROM payments 
	WHERE project_uid = ? 
	LIMIT 1`, uid).Row()

	if err := row.Scan(&amountIdr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("uid %q not found: %w", uid, err)
		}
		return 0, err
	}

	if amountIdr == 0 {
		return 0, fmt.Errorf("uid %q has no amount_idr", uid)
	}

	return amountIdr, nil
}

func SafeSub(quota, used uint64) uint64 {
	if used >= quota {
		return 0
	}
	return quota - used
}

func GetClientIP(r *http.Request) string {
	// prioritas: X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// fallback: X-Real-IP
	if rip := r.Header.Get("X-Real-IP"); rip != "" {
		return rip
	}

	// terakhir: RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

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

func SendEmail(to, app, subject, data, Type string) error {

	body := data

	emailData := &entities.SendEmailRequest{
		To:      to,
		App:     app,
		Subject: subject,
		Body:    body,
		Type:    Type,
	}

	jsonData, err := json.Marshal(emailData)
	if err != nil {
		return err
	}

	resp, err := http.Post(os.Getenv("EMAIL_URL"), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to send email, status code: " + resp.Status)
	} else {
		Logger("info", "Send Email Success")
	}

	return nil
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
