module superapps

go 1.24.0

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-resty/resty/v2 v2.16.5
	github.com/gorilla/mux v1.8.1
	github.com/joho/godotenv v1.5.1
	github.com/sirupsen/logrus v1.9.3
	github.com/t-tomalak/logrus-easy-formatter v0.0.0-20190827215021-c074f06c5816
	golang.org/x/crypto v0.31.0
	golang.org/x/time v0.14.0
	gorm.io/driver/mysql v1.5.7
	gorm.io/gorm v1.30.1
)

require golang.org/x/net v0.33.0 // indirect

require (
	github.com/go-sql-driver/mysql v1.7.0
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/robfig/cron/v3 v3.0.1
	github.com/satori/go.uuid v1.2.0
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)
