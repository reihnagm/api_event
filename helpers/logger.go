package helper

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

// Logger menulis log ke stdout + file harian di logs/go-YYYY-MM-DD.log.
// Akan otomatis membuat folder "logs" jika belum ada.
func Logger(logType string, message string) {
	// Pastikan folder logs ada
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		// fallback: kalau gagal bikin folder, tulis ke stdout saja
		std := logrus.New()
		std.SetOutput(os.Stdout)
		std.SetFormatter(&easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%",
		})
		std.Errorf("failed to create log dir: %v; msg=%s", err, message)
		return
	}

	dateStr := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, "go-"+dateStr+".log")

	// Buka file (buat kalau belum ada)
	file, err := os.OpenFile(filepath.Clean(logPath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		// fallback ke stdout
		std := logrus.New()
		std.SetOutput(os.Stdout)
		std.SetFormatter(&easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%",
		})
		std.Errorf("failed to open log file: %v; msg=%s", err, message)
		return
	}
	defer file.Close()

	logger := &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.DebugLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%",
		},
	}

	// Tulis ke file + stdout (biar kelihatan di `air`)
	logger.SetOutput(io.MultiWriter(file, os.Stdout))

	switch logType {
	case "info":
		logger.Info(message)
	case "error":
		logger.Error(message)
	default:
		logger.Print(message)
	}
}
