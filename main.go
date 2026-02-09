package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"superapps/controllers"
	helper "superapps/helpers"
	"superapps/jobs"

	middleware "superapps/middlewares"
	"superapps/services"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	_ = godotenv.Load()

	services.InitDBs()

	router := mux.NewRouter()

	// ✅ CORS
	router.Use(middleware.CorsMiddleware)

	// ✅ JWT auth
	router.Use(middleware.JwtAuthentication)

	// ✅ Rate Limiter
	rl := middleware.NewRateLimiterFromEnv()
	router.Use(rl.Middleware)

	// --- STATIC ---
	if err := os.MkdirAll("public", os.ModePerm); err != nil {
		log.Fatalf("Failed to create or access directory: %v", err)
	}
	dir, err := os.Open("public")
	if err != nil {
		log.Fatalf("Failed to open public directory: %v", err)
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		log.Fatalf("Failed to read directory contents: %v", err)
	}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			staticPath := "/" + fileInfo.Name() + "/"
			publicPath := "./public/" + fileInfo.Name() + "/"
			log.Printf("Serving static files from %s at %s", publicPath, staticPath)
			router.PathPrefix(staticPath).
				Handler(http.StripPrefix(staticPath, http.FileServer(http.Dir(publicPath))))
		}
	}

	// Callback
	router.HandleFunc("/api/v1/callback", controllers.Callback).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/project-payment-callback", controllers.ProjectPaymentCallback).Methods("POST", "OPTIONS")

	// Order
	router.HandleFunc("/api/v1/payment/order", controllers.Order).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/payment/method", controllers.PaymentMethod).Methods("GET", "OPTIONS")

	// Inbox
	router.HandleFunc("/api/v1/inbox/list", controllers.InboxList).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/inbox/store", controllers.InboxStore).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/inbox/detail/{id}", controllers.InboxDetail).Methods("GET", "OPTIONS")

	// Dashboard
	router.HandleFunc("/api/v1/dashboard/investor", controllers.DashboardInvestor).Methods("GET", "OPTIONS")

	// Portfolio
	router.HandleFunc("/api/v1/portfolio/info", controllers.PortfolioInfo).Methods("GET", "OPTIONS")

	// Admin
	router.HandleFunc("/api/v1/admin/profile", controllers.AdminGetProfile).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/admin/update/profile", controllers.AdminUpdateProfile).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/v1/admin/list/role", controllers.AdminListRole).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/admin/assign/role", controllers.AdminAssignRole).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/admin/revoke/role", controllers.AdminRevokeRole).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/v1/admin/create/user", controllers.AdminCreateUser).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/admin/update/user", controllers.AdminUpdateUser).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/v1/admin/list/user", controllers.AdminListUser).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/admin/detail/user/{id}", controllers.AdminDetailUser).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/admin/verify/user", controllers.UpdateAdminVerifyUser).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/v1/admin/verify/user-investor", controllers.UpdateAdminVerifyUserInvestor).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/v1/admin/verify/user-emiten", controllers.UpdateAdminVerifyUserEmiten).Methods("PUT", "OPTIONS")

	// Contract
	router.HandleFunc("/api/v1/contract-letter-project-payment/upload", controllers.ContractLetterPaymentUpload).Methods("POST", "OPTIONS")

	// Admin Project Transaction
	router.HandleFunc("/api/v1/admin/transaction/project/list", controllers.AdminProjectTransactionList).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/admin/transaction/project/detail/{id}", controllers.AdminProjectTransactionDetail).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/admin/transaction/project/refund-transfer-document", controllers.ProjectRefundTransferDocument).Methods("POST", "OPTIONS")

	// Admin Project
	router.HandleFunc("/api/v1/admin/list/project", controllers.AdminListProject).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/admin/detail/project/{id}", controllers.AdminDetailProject).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/admin/update/project/{id}", controllers.AdminUpdateProject).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/v1/admin/verify/project", controllers.UpdateAdminVerifyProject).Methods("PUT", "OPTIONS")

	// Administration
	router.HandleFunc("/api/v1/administration/province", controllers.GetProvince).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/administration/city/{province_id}", controllers.GetCity).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/administration/district/{city_id}", controllers.GetDistrict).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/administration/subdistrict/{district_id}", controllers.GetSubdistrict).Methods("GET", "OPTIONS")

	// Auth
	router.HandleFunc("/api/v1/auth/admin/login", controllers.AdminLogin).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/login", controllers.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/logout", controllers.Logout).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/register", controllers.Register).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/register-as-emiten", controllers.RegisterAsEmiten).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/register-as-investor-institute", controllers.RegisterAsInvestorInstitute).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/assign/role", controllers.AssignRole).Methods("POST", "OPTIONS")

	// Broadcast
	router.HandleFunc("/api/v1/broadcast/list", controllers.BroadcastList).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/broadcast/detail/{id}", controllers.BroadcastDetail).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/broadcast/send", controllers.BroadcastSend).Methods("POST", "OPTIONS")

	// Document
	router.HandleFunc("/api/v1/document/verify/project", controllers.DocumentVerifyProject).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/document/update/user/{type}", controllers.UpdateValUser).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/v1/document/update/{type}", controllers.DocumentUpdate).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/v1/document/transaction/payment", controllers.DocumentTransactionPayment).Methods("POST", "OPTIONS")

	// Profile
	router.HandleFunc("/api/v1/profile", controllers.GetProfile).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/profile/update", controllers.UpdateProfile).Methods("PUT", "OPTIONS")

	// Account
	router.HandleFunc("/api/v1/account/update", controllers.UpdateAccount).Methods("PUT", "OPTIONS")

	// Project
	router.HandleFunc("/api/v1/project/list", controllers.ProjectList).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/project/detail/{id}", controllers.ProjectDetail).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/project/type/list", controllers.ProjectTypeList).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/project/authority/type/list", controllers.ProjectAuthorityTypeList).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/project-by-emiten/list", controllers.ProjectEmitenList).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/project/store", controllers.ProjectStore).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/project/payment", controllers.ProjectPayment).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/project/refund", controllers.ProjectRefund).Methods("POST", "OPTIONS")

	// Transaction
	router.HandleFunc("/api/v1/transaction/project/list", controllers.ProjectTransactionList).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/transaction/project/detail/{id}", controllers.ProjectTransactionDetail).Methods("GET", "OPTIONS")

	// Business
	router.HandleFunc("/api/v1/business/type/list", controllers.BusinessTypeList).Methods("GET", "OPTIONS")

	// Company
	router.HandleFunc("/api/v1/company/type/list", controllers.CompanyTypeList).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/company/type/place/list", controllers.CompanyTypePlaceList).Methods("GET", "OPTIONS")

	// Project Inquiry
	router.HandleFunc("/api/v1/project/inquiry", controllers.ProjectInquiry).Methods("POST", "OPTIONS")

	// Project Cost Of Fund Template
	router.HandleFunc("/api/v1/project/cost-of-fund/template/list", controllers.ProjectCostOfFundTemplateList).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/project/cost-of-fund/template/detail/{id}", controllers.ProjectCostOfFundTemplateDetail).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/project/cost-of-fund/template/create", controllers.ProjectCostOfFundTemplateCreate).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/project/cost-of-fund/template/delete/{id}", controllers.ProjectCostOfFundTemplateDelete).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/api/v1/project/cost-of-fund/template/update/{id}", controllers.ProjectCostOfFundTemplateUpdate).Methods("PUT", "OPTIONS")

	// Project Cost Of Fund Template Without
	router.HandleFunc("/api/v1/project/cost-of-fund/template-without/list", controllers.ProjectCostOfFundTemplateWithoutList).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/project/cost-of-fund/template-without/detail/{id}", controllers.ProjectCostOfFundTemplateWithoutDetail).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/project/cost-of-fund/template-without/create", controllers.ProjectCostOfFundTemplateWithoutCreate).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/project/cost-of-fund/template-without/delete/{id}", controllers.ProjectCostOfFundTemplateWithoutDelete).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/api/v1/project/cost-of-fund/template-without/update/{id}", controllers.ProjectCostOfFundTemplateWithoutUpdate).Methods("PUT", "OPTIONS")

	// Otp
	router.HandleFunc("/api/v1/resend-otp", controllers.ResendOtp).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/verify-otp", controllers.VerifyOtp).Methods("POST", "OPTIONS")

	// Forgot Password
	router.HandleFunc("/api/v1/forgot-password", controllers.ForgotPassword).Methods("PUT", "OPTIONS")

	// Logs
	router.HandleFunc("/api/v1/logs", controllers.LogList).Methods("GET", "OPTIONS")

	// FCM
	router.HandleFunc("/api/v1/fcm/initialize", controllers.InitializeFcm).Methods("POST", "OPTIONS")

	// --- SERVER ---
	port := ":" + firstNonEmpty(os.Getenv("PORT"))
	helper.Logger("info", "Starting server at "+port)

	server := &http.Server{
		Addr:              port,
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
	}

	// --- CRON (logging supaya kebaca di Air) ---
	loc := time.FixedZone("WIB", 7*60*60)
	cronLogger := cron.VerbosePrintfLogger(log.New(os.Stdout, "[cron] ", log.LstdFlags))

	c := cron.New(
		cron.WithLocation(loc),
		cron.WithLogger(cronLogger),
		cron.WithChain(
			cron.Recover(cronLogger),            // kalau panic, jangan matiin process
			cron.SkipIfStillRunning(cronLogger), // hindari overlap job
		),
	)

	// Tiap menit (ubah sesuai kebutuhan)
	_, _ = c.AddFunc("*/10 * * * *", func() {
		_, _ = jobs.RunExpireInvoices(context.Background())
	})
	c.Start()
	defer c.Stop()

	// Jalankan server di goroutine agar bisa shutdown rapi
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			helper.Logger("error", fmt.Sprintf("HTTP server error: %v", err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	helper.Logger("info", "Server stopped gracefully")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
