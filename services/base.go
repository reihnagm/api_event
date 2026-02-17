package services

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	helper "superapps/helpers"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	dbDefault *gorm.DB
	dbPayment *gorm.DB
)

// InitDBs initializes all database connections with env-aware database names.
func InitDBs() {
	// Load .env if present (tidak error kalau tidak ada, karena PM2 inject ENV)
	_ = godotenv.Load()

	// --- Resolve envs ---
	env := strings.ToLower(strings.TrimSpace(os.Getenv("GO_ENV")))
	if env == "" {
		env = "development"
	}

	// Timezone (optional override), default Asia/Jakarta
	appTZ := strings.TrimSpace(os.Getenv("APP_TZ"))
	if appTZ == "" {
		appTZ = "Asia/Jakarta"
	}

	// Koneksi Default (app utama)
	defaultDBName := resolveDBName(env,
		os.Getenv("DB_NAME"),
		os.Getenv("DB_STAGING_NAME"),
	)

	dbDefault = connectMySQL(MySQLConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		DBName:   defaultDBName,
		TZ:       appTZ,
	}, "default")

	// Koneksi Payment (schema berbeda)
	paymentDBName := resolveDBName(env,
		os.Getenv("DB_PG"),              // production
		os.Getenv("DB_PG_STAGING_NAME"), // staging (opsional)
	)
	if paymentDBName == "" {
		// fallback: kalau tidak ada DB_PG_STAGING_NAME, pakai DB_PG
		paymentDBName = os.Getenv("DB_PG")
	}

	// Bisa gunakan host/cred berbeda untuk payment dengan *_PG_*, fallback ke yang utama
	dbPayment = connectMySQL(MySQLConfig{
		User:     firstNonEmpty(os.Getenv("DB_USER_PG"), os.Getenv("DB_USER")),
		Password: firstNonEmpty(os.Getenv("DB_PASSWORD_PG"), os.Getenv("DB_PASSWORD")),
		Host:     firstNonEmpty(os.Getenv("DB_HOST_PG"), os.Getenv("DB_HOST")),
		Port:     firstNonEmpty(os.Getenv("DB_PORT_PG"), os.Getenv("DB_PORT")),
		DBName:   paymentDBName,
		TZ:       appTZ,
	}, "payment")

	if dbDefault == nil {
		panic("❌ dbDefault is nil, failed to connect to default DB")
	}
}

// resolveDBName memilih nama DB berdasarkan env (staging vs selainnya)
func resolveDBName(env, prodName, stagingName string) string {
	if env == "staging" && strings.TrimSpace(stagingName) != "" {
		return stagingName
	}
	return prodName
}

// MySQLConfig konfigurasi minimal untuk DSN
type MySQLConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
	TZ       string // e.g. "Asia/Jakarta" atau "+07:00"
	Params   string // opsional tambahan, e.g. "tls=true"
}

// connectMySQL membuka koneksi GORM v2 ke MySQL dengan TZ & retry.
func connectMySQL(cfg MySQLConfig, tag string) *gorm.DB {
	if cfg.Port == "" {
		cfg.Port = "3306"
	}
	// if cfg.TZ == "" {
	// 	cfg.TZ = "Asia/Jakarta"
	// }

	// Untuk DSN MySQL, gunakan loc= (URL-escaped)
	// locEscaped := url.QueryEscape(cfg.TZ)

	// Param default
	// baseParams := "charset=utf8mb4&parseTime=True&loc=" + locEscaped

	baseParams := "charset=utf8mb4&parseTime=True"
	if strings.TrimSpace(cfg.Params) != "" {
		baseParams += "&" + strings.TrimPrefix(cfg.Params, "&")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, baseParams,
	)

	var conn *gorm.DB
	var err error

	// Retry ringan (3x) untuk kasus DB baru bangun atau jaringan delay
	for attempt := 1; attempt <= 3; attempt++ {
		conn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			NowFunc: func() time.Time { return nowInTZ(cfg.TZ) },
		})
		if err == nil {
			break
		}
		helper.Logger("error", fmt.Sprintf("(%s) gorm.Open attempt %d failed: %v", tag, attempt, err))
		time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
	}
	if err != nil {
		helper.Logger("error", fmt.Sprintf("Failed to connect to MySQL (%s:%s/%s) tag=%s: %v",
			cfg.Host, cfg.Port, cfg.DBName, tag, err))
		return nil
	}

	// Set session time_zone agar NOW() di MySQL seragam dengan app
	// sessionTZ := toMySQLSessionTZ(cfg.TZ) // "+07:00" atau "Asia/Jakarta"
	// if execErr := conn.Exec("SET time_zone = ?", sessionTZ).Error; execErr != nil {
	// 	helper.Logger("error", fmt.Sprintf("(%s) set session time_zone failed: %v", tag, execErr))
	// }

	sqlDB, err := conn.DB()
	if err != nil {
		helper.Logger("error", fmt.Sprintf("(%s) get *sql.DB failed: %v", tag, err))
		return nil
	}

	// Pooling
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

	// Ping untuk memastikan koneksi OK
	if pingErr := pingWithTimeout(sqlDB, 3*time.Second); pingErr != nil {
		helper.Logger("error", fmt.Sprintf("(%s) ping failed: %v", tag, pingErr))
		// tetap kembalikan conn; caller bisa memilih untuk lanjut/stop
	}

	helper.Logger("info", fmt.Sprintf("Connected to MySQL [%s] %s:%s/%s (tz=%s)",
		tag, cfg.Host, cfg.Port, cfg.DBName, cfg.TZ))
	return conn
}

func pingWithTimeout(db *sql.DB, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		done <- db.Ping()
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("ping timeout after %s", timeout.String())
	}
}

func nowInTZ(tz string) time.Time {
	if strings.HasPrefix(tz, "+") || strings.HasPrefix(tz, "-") {
		// offset format, e.g. +07:00
		secs := parseOffsetToSeconds(tz)
		return time.Now().In(time.FixedZone("APP_TZ", secs))
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		// fallback ke WIB
		return time.Now().In(time.FixedZone("WIB", 7*60*60))
	}
	return time.Now().In(loc)
}

func toMySQLSessionTZ(tz string) string {
	// MySQL menerima "Asia/Jakarta" atau "+07:00"
	if strings.HasPrefix(tz, "+") || strings.HasPrefix(tz, "-") || strings.Contains(tz, "/") {
		return tz
	}
	// fallback
	return "Asia/Jakarta"
}

func parseOffsetToSeconds(off string) int {
	// expects +HH:MM / -HH:MM
	sign := 1
	s := off
	if strings.HasPrefix(off, "-") {
		sign = -1
		s = off[1:]
	} else if strings.HasPrefix(off, "+") {
		s = off[1:]
	}
	hh, mm := 0, 0
	fmt.Sscanf(s, "%d:%d", &hh, &mm)
	return sign * (hh*3600 + mm*60)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func GetDefaultDB() *gorm.DB { return dbDefault }
func GetPaymentDB() *gorm.DB { return dbPayment }
