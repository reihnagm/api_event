package services

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"superapps/entities"
	helper "superapps/helpers"
	"time"

	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
)

func Order(r *http.Request, o *entities.Order) (*entities.OrderResult, error) {
	var lastOrder entities.OrderScan
	var project entities.ProjectOrder

	// Get Invoice Last
	queryInvoiceLast := `SELECT no FROM orders ORDER BY id DESC LIMIT 1`
	errInvoiceLast := dbDefault.Raw(queryInvoiceLast).Scan(&lastOrder).Error
	if errInvoiceLast != nil && !errors.Is(errInvoiceLast, gorm.ErrRecordNotFound) {
		helper.Logger("error", "In Server: "+errInvoiceLast.Error())
		return nil, errInvoiceLast
	}

	// Get Project
	queryProject := `SELECT p.title, c.user_id FROM projects p 
	INNER JOIN companies c ON c.uid = p.company_id 
	WHERE p.uid = ?`
	errProject := dbDefault.Raw(queryProject, o.ProjectId).Scan(&project).Error
	if errProject != nil {
		helper.Logger("error", "In Server: "+errProject.Error())
		return nil, errProject
	}

	counterNumber := 1
	if lastOrder.No != 0 {
		counterNumber = lastOrder.No + 1
	}

	var paymentLogo string
	var paymentName string
	var paymentFee int
	var inboxId int
	var paymentAccess string
	var paymentType string
	var paymentExpire string

	var realPrice = o.Price

	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(100000)
	invoice := fmt.Sprintf("FULUSME-INV%d-%05d", counterNumber, randomNumber)

	if o.PaymentMethod != "billing" {
		tx := dbDefault.Begin()
		if tx.Error != nil {
			return nil, tx.Error
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Get Payment Method
		paymentMethods := []entities.PaymentMethod{}
		queryPaymentMethod := `SELECT id, name, nameCode as name_code, logo, platform, fee 
		FROM Channels 
		WHERE id = ?`
		errUser := dbPayment.Debug().Raw(queryPaymentMethod, o.PaymentMethod).Scan(&paymentMethods).Error
		if errUser != nil {
			tx.Rollback()
			helper.Logger("error", "In Server: "+errUser.Error())
			return nil, errUser
		}
		if len(paymentMethods) == 0 {
			tx.Rollback()
			return nil, errors.New("payment method not found")
		}

		paymentLogo = paymentMethods[0].Logo
		paymentName = paymentMethods[0].Name
		paymentFee = paymentMethods[0].Fee

		// Check Project Existing
		var projectTitle string
		queryCheckProject := `SELECT title FROM projects WHERE uid = ? LIMIT 1`

		errCheckProject := tx.Raw(queryCheckProject, o.ProjectId).Scan(&projectTitle).Error
		if errCheckProject != nil {
			if errors.Is(errCheckProject, gorm.ErrRecordNotFound) || projectTitle == "" {
				return nil, fmt.Errorf("PROJECT_NOT_FOUND")
			}
			return nil, fmt.Errorf("error checking project: %v", errCheckProject)
		}

		// Store Order
		queryOrder := `INSERT INTO orders (invoice, no, project_id, cut_price, real_price) VALUES (?, ?, ?, ?, ?)`
		errOrder := tx.Debug().Exec(queryOrder, invoice, counterNumber, o.ProjectId, realPrice, realPrice).Error
		if errOrder != nil {
			tx.Rollback()
			helper.Logger("error", "In Server: "+errOrder.Error())
			return nil, errOrder
		}

		// Payload Midtrans
		payload := map[string]any{
			"channel_id":  o.PaymentMethod,
			"orderId":     invoice,
			"amount":      realPrice,
			"app":         "FULUSME",
			"callbackUrl": os.Getenv("CALLBACK_URL"),
		}

		// Send Request
		client := resty.New()
		midtransUrl := os.Getenv("PAY_MIDTRANS")

		var midtransRes entities.MidtransResponse
		var midtransErr entities.MidtransErrorResponse

		resp, err := client.R().
			SetBody(payload).
			SetResult(&midtransRes).
			Post(midtransUrl)

		if err != nil || resp.StatusCode() != 200 {
			tx.Rollback()
			errMsg := fmt.Sprintf("gagal request: %s", midtransErr.Message)
			if midtransErr.Message == "" {
				errMsg = fmt.Sprintf("gagal request: %s", string(resp.Body()))
			}
			helper.Logger("error", errMsg)
			return nil, errors.New(errMsg)
		}

		// Parse hasil Midtrans
		if o.PaymentMethod == "4" {
			paymentAccess = midtransRes.Data.Data.Actions[0].Url
			paymentType = "emoney"
			loc, _ := time.LoadLocation("Asia/Jakarta")
			paymentExpire = time.Now().In(loc).Add(30 * time.Minute).Format("2006-01-02 15:04:05")
		} else {
			paymentAccess = midtransRes.Data.Data.VANumber
			paymentType = "va"
			paymentExpire = midtransRes.Data.Expire
		}

		inbox := entities.Inboxes{
			Title:      "Proyek [" + project.Title + "]",
			Content:    "Silahkan melakukan pembayaran lebih lanjut sebesar " + helper.FormatIDR(float64(realPrice)),
			Field1:     paymentName,
			Field2:     paymentFee,
			Field3:     paymentAccess,
			Field4:     paymentExpire,
			Field5:     paymentType,
			Field6:     paymentLogo,
			Field7:     realPrice,
			Field8:     o.ProjectId,
			UserId:     o.UserId,
			ReceiverId: project.UserId,
			Type:       "2",
		}

		if errInbox := tx.Create(&inbox).Error; errInbox != nil {
			return nil, errInbox
		}

		inboxId = int(inbox.ID)

		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
	}

	result := &entities.OrderResult{
		Price:   realPrice,
		Invoice: invoice,
		Payment: entities.OrderPaymentResult{
			Logo:   paymentLogo,
			Name:   paymentName,
			Fee:    paymentFee,
			Access: paymentAccess,
			Expire: paymentExpire,
			Type:   paymentType,
		},
		Project: entities.OrderProjectResult{
			Id:    o.ProjectId,
			Title: project.Title,
		},
		Inbox: entities.OrderInbox{
			Id: inboxId,
		},
	}

	email, _ := helper.GetEmailByProjectId(dbDefault, o.ProjectId)
	title, _ := helper.GetTitleByProjectId(dbDefault, o.ProjectId)
	role, _ := helper.GetRoleByProjectId(dbDefault, o.ProjectId)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s membuat order untuk project [%s] pada waktu %s",
			ip,
			email,
			role,
			title,
			time.Now().Format(time.Now().Format("2006-01-02 15:04:05")),
		),
		role,
	)

	return result, nil
}
