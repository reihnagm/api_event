package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"superapps/entities"
	helper "superapps/helpers"

	"github.com/go-resty/resty/v2"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func PaymentMethod() ([]entities.PaymentMethod, error) {
	paymentMethods := []entities.PaymentMethod{}

	queryUserExist := `SELECT id, name, nameCode as name_code, logo, platform, fee FROM Channels WHERE nameCode != ?`

	errUser := dbPayment.Debug().Raw(queryUserExist, "bca").Scan(&paymentMethods).Error
	if errUser != nil {
		helper.Logger("error", "In Server: "+errUser.Error())
		return nil, errors.New(errUser.Error())
	}

	return paymentMethods, nil
}

func ProjectRefundTransferDocument(req *entities.PaymentProjectReqRefundTransDoc) (entities.PaymentProjectReqRefundTransDoc, error) {
	if req == nil {
		return entities.PaymentProjectReqRefundTransDoc{}, errors.New("nil payload")
	}

	updateQuery := `UPDATE refund_transfer_documents SET path = ? WHERE transaction_id = ?`

	tx := dbDefault.Exec(updateQuery, req.Path, req.TransactionId)
	if tx.Error != nil {
		helper.Logger("error", "In Server: "+tx.Error.Error())
		return entities.PaymentProjectReqRefundTransDoc{}, tx.Error
	}

	if tx.RowsAffected == 0 {
		insertQuery := `INSERT INTO refund_transfer_documents (path, transaction_id) VALUES (?, ?)`
		if err := dbDefault.Exec(insertQuery, req.Path, req.TransactionId).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return entities.PaymentProjectReqRefundTransDoc{}, err
		}
	}

	return entities.PaymentProjectReqRefundTransDoc{
		Path:          req.Path,
		TransactionId: req.TransactionId,
	}, nil
}

func ProjectPayment(r *http.Request, req *entities.PaymentProjectReq) (entities.PaymentProjectReq, error) {
	if req == nil {
		return entities.PaymentProjectReq{}, errors.New("nil payload")
	}
	req.ProjectId = strings.TrimSpace(req.ProjectId)
	if req.ProjectId == "" {
		return entities.PaymentProjectReq{}, errors.New("project_id is required")
	}
	if req.Amount <= 0 {
		return entities.PaymentProjectReq{}, errors.New("amount must be > 0")
	}

	// --- map payment method (1..6) ---
	type chInfo struct {
		ChannelID   string // untuk payload gateway
		ChannelCode string // disimpan di invoices.channel_code
		Method      string // disimpan di invoices.payment_method (opsional info)
		TTL         time.Duration
	}
	pmMap := map[string]chInfo{
		"1": {ChannelID: "1", ChannelCode: "bca", Method: "VA", TTL: 24 * time.Hour},
		"2": {ChannelID: "2", ChannelCode: "mandiri", Method: "VA", TTL: 24 * time.Hour}, // echannel juga OK jika gateway butuh
		"3": {ChannelID: "3", ChannelCode: "bri", Method: "VA", TTL: 24 * time.Hour},
		"4": {ChannelID: "4", ChannelCode: "gopay", Method: "EWALLET", TTL: 1 * time.Hour},
		"5": {ChannelID: "5", ChannelCode: "bni", Method: "VA", TTL: 24 * time.Hour},
		"6": {ChannelID: "6", ChannelCode: "cimb", Method: "VA", TTL: 24 * time.Hour},
	}
	pmKey := strings.TrimSpace(req.PaymentMethod) // biar backward-compatible kalau masih string angka
	ch, ok := pmMap[pmKey]
	if !ok {
		return entities.PaymentProjectReq{}, fmt.Errorf("payment_method invalid. gunakan: 1=bca, 2=mandiri, 3=bri, 4=gopay, 5=bni, 6=cimb")
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	nowWIB := time.Now().In(loc)

	expiresAt := nowWIB.Add(30 * time.Minute)

	// --- role institusi? (users.role = 9) ---
	const institusiRoleID = 9
	var isInstitusi bool
	if err := dbDefault.
		Raw(`SELECT EXISTS(SELECT 1 FROM users WHERE uid=? AND role=?)`, req.UserId, institusiRoleID).
		Scan(&isInstitusi).Error; err != nil {
		return entities.PaymentProjectReq{}, err
	}

	// --- Cek masih ada PENDING yg belum expired di project yg sama (by user) ---
	// type pendingRow struct {
	// 	PaymentID  uint64
	// 	OrderID    *string
	// 	ExpiresAt  *string
	// 	PaymentURL *string
	// }
	// var pend pendingRow
	// const qPending = `
	//   SELECT
	//     p.id AS payment_id,
	//     i.order_id,
	//     IFNULL(DATE_FORMAT(i.expires_at,'%Y-%m-%d %H:%i:%s'), NULL) AS expires_at,
	//     JSON_UNQUOTE(JSON_EXTRACT(i.raw_response, '$.data.data.actions[0].url')) AS payment_url
	//   FROM payments p
	//   JOIN jobs j ON j.id = p.investor_job_id
	//   LEFT JOIN (
	//     SELECT i1.*
	//     FROM invoices i1
	//     JOIN (
	//       SELECT payment_id, MAX(id) AS max_id
	//       FROM invoices
	//       GROUP BY payment_id
	//     ) t ON t.payment_id = i1.payment_id AND t.max_id = i1.id
	//   ) i ON i.payment_id = p.id
	//   WHERE j.user_id = ?
	//     AND p.project_uid = ?
	//     AND p.payment_status = 'PENDING'
	//     AND i.invoice_status != 'EXPIRED'
	//   LIMIT 1`
	// if err := dbDefault.Raw(qPending, req.UserId, req.ProjectId).Scan(&pend).Error; err != nil {
	// 	return entities.PaymentProjectReq{}, err
	// }
	// if pend.PaymentID != 0 {
	// 	if isInstitusi {
	// 		return entities.PaymentProjectReq{}, fmt.Errorf("Perusahaan Anda masih punya pembelian efek yang belum dibayar. Selesaikan atau batalkan sebelum membuat pembayaran baru.")
	// 	}
	// 	return entities.PaymentProjectReq{}, fmt.Errorf("Anda masih punya pembelian efek yang belum dibayar. Selesaikan atau batalkan sebelum membuat pembayaran baru.")
	// }

	// --- Pastikan ada job; jika belum ada, create minimal (tanpa kolom generated) ---
	var jobID int64
	if err := dbDefault.Raw(`SELECT id FROM jobs WHERE user_id=? ORDER BY updated_at DESC LIMIT 1`, req.UserId).
		Scan(&jobID).Error; err != nil {
		return entities.PaymentProjectReq{}, err
	}
	if jobID == 0 {
		if errInsertJobs := dbDefault.Exec(`INSERT INTO jobs (user_id, created_at, updated_at) VALUES (?, NOW(), NOW())`, req.UserId).Error; errInsertJobs != nil {
			var me *mysql.MySQLError
			if errors.As(errInsertJobs, &me) && me.Number == 3105 {
				return entities.PaymentProjectReq{}, fmt.Errorf("table jobs punya kolom generated; jangan set kolom generated saat INSERT: %v", me.Message)
			}
			if errors.As(errInsertJobs, &me) && me.Number == 1364 {
				return entities.PaymentProjectReq{}, fmt.Errorf("jobs memiliki kolom NOT NULL tanpa default. Tambahkan default di schema atau isi kolom tersebut saat INSERT. Detail: %v", me.Message)
			}
			return entities.PaymentProjectReq{}, fmt.Errorf("failed to create minimal job: %w", errInsertJobs)
		}

		if errSelectJobs := dbDefault.Raw(`SELECT id FROM jobs WHERE user_id=? ORDER BY updated_at DESC, id DESC LIMIT 1`, req.UserId).
			Scan(&jobID).Error; errSelectJobs != nil || jobID == 0 {
			return entities.PaymentProjectReq{}, errors.New("failed to fetch created job id")
		}

	}

	// --- Provider & order_id (institusi beda prefix) ---
	provider := "midtrans"
	prefix := "FULUSME-INV-"
	if isInstitusi {
		prefix = "FULUSME-INV9-"
	}
	orderID := prefix + helper.RandHex16()

	var paymentID uint64

	// --- TX: create payment + issue invoice (pakai expires_at hasil TTL) ---
	if err := dbDefault.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("CALL sp_create_payment_job_uid(?,?,?)",
			jobID, req.ProjectId, req.Amount,
		).Error; err != nil {
			return err
		}
		if err := tx.Raw("SELECT LAST_INSERT_ID() AS id").Scan(&paymentID).Error; err != nil {
			return err
		}
		// kirim method, channel_code, dan expires_at spesifik
		if err := tx.Exec(
			"CALL sp_issue_invoice(?,?,?,?,?,?,?,?)",
			paymentID, provider, orderID,
			ch.Method, ch.ChannelID, nil, req.Lot, expiresAt,
		).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1644 {
			return entities.PaymentProjectReq{}, fmt.Errorf("%s", me.Message)
		}
		return entities.PaymentProjectReq{}, err
	}

	// --- Call gateway ---
	midtransURL := os.Getenv("PAY_MIDTRANS")
	callbackURL := os.Getenv("CALLBACK_PROJECT_PAYMENT")
	if midtransURL == "" {
		_ = cancelInvoiceSafely(provider, orderID, "MISSING_MIDTRANS_ENDPOINT")
		return entities.PaymentProjectReq{}, errors.New("payment gateway endpoint not configured")
	}

	appName := "FULUSME-INV"
	if isInstitusi {
		appName = "FULUSME-INVCORP"
	}
	payload := map[string]any{
		"channel_id":  ch.ChannelID, // <— pakai kode mapped, bukan angka mentah
		"orderId":     orderID,
		"amount":      req.Amount,
		"app":         appName,
		"callbackUrl": callbackURL,
	}

	client := resty.New()
	var mtRes entities.MidtransResponse
	var mtErr entities.MidtransErrorResponse

	resp, err := client.R().
		SetBody(payload).
		SetResult(&mtRes).
		SetError(&mtErr).
		Post(midtransURL)

	if err != nil || resp.IsError() {
		msg := "payment gateway error"
		if err != nil {
			msg = err.Error()
		} else if mtErr.Message != "" {
			msg = mtErr.Message
		}
		_ = cancelInvoiceSafely(provider, orderID, "GATEWAY_ERROR")
		return entities.PaymentProjectReq{}, fmt.Errorf("%s", msg)
	}

	// --- Update invoice dgn VA/expiry/raw_response ---
	va := strings.TrimSpace(mtRes.Data.Data.VANumber)
	expMySQL := helper.ParseToMySQLDatetime(mtRes.Data.Expire) // kalau gateway balikin expire sendiri; jika kosong, pakai yang kita set
	if expMySQL == "" {
		expMySQL = expiresAt.Format("2006-01-02 15:04:05")
	}
	rawJSON, _ := json.Marshal(mtRes)
	if err := updateInvoiceGatewayResponse(provider, orderID, rawJSON); err != nil {
		helper.Logger("error", "update invoice gateway response: "+err.Error())
	}

	var payURL string
	if len(mtRes.Data.Data.Actions) > 0 {
		payURL = mtRes.Data.Data.Actions[0].Url
	}

	out := *req
	out.Id = paymentID
	out.Provider = provider
	out.OrderId = orderID
	out.VANumber = va
	out.ExpireAt = expiresAt.Format("2006-01-02 15:04:05")
	out.PaymentURL = payURL
	out.PaymentMethod = pmKey

	title, _ := helper.GetTitleByProjectId(dbDefault, req.ProjectId)
	email, _ := helper.GetEmailByProjectId(dbDefault, req.ProjectId)
	role, _ := helper.GetRoleByProjectId(dbDefault, req.ProjectId)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s membuat pembayaran untuk proyek %s dengan metode pembayaran %s sebesar %s pada %s",
			ip,
			email,
			role,
			title,
			pmKey,
			helper.FormatIDRInt(int64(req.Amount)),
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return out, nil
}

func ProjectRefund(r *http.Request, pprr *entities.PaymentProjectReqRefund) (entities.PaymentProjectResRefund, error) {
	if pprr == nil {
		return entities.PaymentProjectResRefund{}, errors.New("nil payload")
	}

	if err := dbDefault.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("CALL sp_refund_payment(?)",
			pprr.PaymentId,
		).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1644 {
			return entities.PaymentProjectResRefund{}, fmt.Errorf("%s", me.Message)
		}
		return entities.PaymentProjectResRefund{}, err
	}

	email, title, role, _ := helper.GetEmailAndTitleAndRoleByPayment(dbDefault, pprr.PaymentId)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s mengajukan refund untuk proyek [%s] pada %s",
			ip,
			email,
			role,
			title,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return entities.PaymentProjectResRefund{
		PaymentId: pprr.PaymentId,
	}, nil
}

func cancelInvoiceSafely(provider, orderID, reason string) error {
	return dbDefault.Exec("CALL sp_cancel_invoice(?,?,?)", provider, orderID, reason).Error
}

func updateInvoiceGatewayResponse(provider, orderID string, rawResp []byte) error {
	// return dbDefault.Exec(`
	// 	UPDATE invoices
	// 	SET channel_ref = COALESCE(?, channel_ref),
	// 	    expires_at  = COALESCE(?, expires_at),
	// 	    raw_response= ?
	// 	WHERE provider=? AND order_id=?`,
	// 	channelRef, expiresAt, string(rawResp), provider, orderID,
	// ).Error

	return dbDefault.Exec(`
		UPDATE invoices
		SET raw_response= ?
		WHERE provider=? AND order_id=?`,
		string(rawResp), provider, orderID,
	).Error
}
