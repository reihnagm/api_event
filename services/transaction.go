package services

import (
	"context"
	"encoding/json"
	"strings"

	"superapps/entities"
	helper "superapps/helpers"

	"gorm.io/gorm"
)

func normalizeStatuses(csv string) []string {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	raw := strings.Split(csv, ",")
	out := make([]string, 0, len(raw))
	for _, s := range raw {
		s = strings.ToUpper(strings.TrimSpace(s))
		switch s {
		case "PENDING", "PAID", "FAILED", "REFUNDED", "CANCELLED":
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func ProjectListTransactions(
	ctx context.Context,
	userID string,
	statusCSV string,
	page, perPage int,
) (entities.TransactionListResp, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	statuses := normalizeStatuses(statusCSV)

	where := "1=1"
	args := make([]any, 0, 3)

	if userID != "" {
		where += " AND j.user_id = ?"
		args = append(args, userID)
	}
	if len(statuses) > 0 {
		where += " AND p.payment_status IN ?"
		args = append(args, statuses)
	}

	var total int64
	countSQL := `
		SELECT COUNT(1)
		FROM payments p
		JOIN jobs j ON j.id = p.investor_job_id
		WHERE ` + where

	if err := dbDefault.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return entities.TransactionListResp{}, err
	}

	type row struct {
		InvestorJobId string  `json:"-"`
		PaymentId     uint64  `json:"-"`
		ProjectUid    string  `json:"-"`
		ProjectTitle  string  `json:"-"`
		ProjectSku    string  `json:"-"`
		Amount        int64   `json:"-"`
		PaymentStatus string  `json:"-"`
		CreatedAt     string  `json:"-"`
		PaidAt        *string `json:"-"`
		OrderID       *string `json:"-"`
		Provider      *string `json:"-"`
		InvoiceStatus *string `json:"-"`
		ChannelCode   *string `json:"-"`
		ChannelRef    *string `json:"-"`
		ExpiresAt     *string `json:"-"`
	}
	var rows []row

	listSQL := `
		SELECT 
			p.id                           AS payment_id,
			p.project_uid                  AS project_uid,
			pr.title                       AS project_title,
			p.amount_idr                   AS amount,
			p.payment_status               AS payment_status,
			p.investor_job_id,
			pr.sku AS project_sku,
			DATE_FORMAT(p.created_at,'%Y-%m-%d %H:%i:%s') AS created_at,
			IFNULL(DATE_FORMAT(p.paid_at,'%Y-%m-%d %H:%i:%s'), NULL) AS paid_at,
			i.order_id, i.provider, i.invoice_status,
			i.channel_code, i.channel_ref,
			IFNULL(
				DATE_FORMAT(CONVERT_TZ(i.expires_at, '+00:00', '+07:00'), '%Y-%m-%d %H:%i:%s'),
				NULL
			) AS expires_at,

			-- can_refund: true jika paid_at ada dan selisih WAKTU < 2 jam
			CASE 
				WHEN p.paid_at IS NOT NULL 
				 AND TIMESTAMPDIFF(SECOND, p.paid_at, NOW()) >= 0
				 AND TIMESTAMPDIFF(SECOND, p.paid_at, NOW()) < 7200
				THEN TRUE ELSE FALSE 
			END AS can_refund
		FROM payments p
		JOIN jobs j      ON j.id = p.investor_job_id
		LEFT JOIN projects pr ON pr.uid = p.project_uid
		LEFT JOIN (
			SELECT i1.*
			FROM invoices i1
			JOIN (
				SELECT payment_id, MAX(id) AS max_id
				FROM invoices
				GROUP BY payment_id
			) t ON t.payment_id = i1.payment_id AND t.max_id = i1.id
		) i ON i.payment_id = p.id
		WHERE ` + where + `
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?`

	args2 := append(append([]any{}, args...), perPage, offset)

	if err := dbDefault.WithContext(ctx).Raw(listSQL, args2...).Scan(&rows).Error; err != nil {
		return entities.TransactionListResp{}, err
	}

	items := make([]entities.TransactionListItem, 0, len(rows))
	for _, r := range rows {
		// Get Contract Letter Payment
		paymentContractLetter := entities.PaymentContractLetter{}
		queryPaymentMethod := `SELECT path
		FROM contract_letter_payments 
		WHERE payment_id = ?`
		if err := dbDefault.Raw(queryPaymentMethod, r.PaymentId).Scan(&paymentContractLetter).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return entities.TransactionListResp{}, err
		}

		// User
		getUser := entities.PaymentInvestor{}
		queryGetUser := `SELECT u.uid AS id, p.fullname, p.selfie, u.email, u.sku
		FROM jobs j
		INNER JOIN users u ON j.user_id = u.uid
		INNER JOIN profiles p ON j.user_id = p.user_id
		WHERE j.id = ?`
		if err := dbDefault.Raw(queryGetUser, r.InvestorJobId).Scan(&getUser).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return entities.TransactionListResp{}, err
		}

		// Company
		getCompany := entities.PaymentCompany{}
		queryCompany := `SELECT c.uid AS id, c.company_name AS name 
		FROM projects p 
		INNER JOIN companies c ON c.uid = p.company_id
		WHERE p.uid = ?`
		if err := dbDefault.Raw(queryCompany, r.ProjectUid).Scan(&getCompany).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return entities.TransactionListResp{}, err
		}

		// Refund Document
		getPaymentProjectRefundTransDoc := entities.PaymentProjectRefundTransDoc{}
		queryPaymentProjectRefundTransDoc := `SELECT path, transaction_id, created_at 
		FROM refund_transfer_documents
		WHERE transaction_id = ?`
		errGetPaymentProjectRefundTransDoc := dbDefault.Raw(queryPaymentProjectRefundTransDoc, r.PaymentId).Scan(&getPaymentProjectRefundTransDoc).Error
		if errGetPaymentProjectRefundTransDoc != nil {
			helper.Logger("error", "In Server: "+errGetPaymentProjectRefundTransDoc.Error())
			return entities.TransactionListResp{}, errGetPaymentProjectRefundTransDoc
		}

		var isRefund bool

		if getPaymentProjectRefundTransDoc.Path == "" {
			isRefund = false
		} else {
			isRefund = true
		}

		items = append(items, entities.TransactionListItem{
			PaymentId:      r.PaymentId,
			ProjectId:      r.ProjectUid,
			ProjectTitle:   r.ProjectTitle,
			ProjectSku:     r.ProjectSku,
			Amount:         r.Amount,
			PaymentStatus:  r.PaymentStatus,
			CreatedAt:      r.CreatedAt,
			PaidAt:         r.PaidAt,
			ContractLetter: paymentContractLetter,
			Investor:       getUser,
			Company:        getCompany,
			Refund: entities.PaymentProjectRefundTransDoc{
				Path:      helper.DefaultIfEmpty(getPaymentProjectRefundTransDoc.Path, "-"),
				CreatedAt: getPaymentProjectRefundTransDoc.CreatedAt,
			},
			IsRefund:      isRefund,
			OrderId:       r.OrderID,
			Provider:      r.Provider,
			InvoiceStatus: r.InvoiceStatus,
			ChannelCode:   r.ChannelCode,
			ChannelRef:    r.ChannelRef,
			ExpiresAt:     r.ExpiresAt,
		})
	}

	return entities.TransactionListResp{
		Items:      items,
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
	}, nil
}

func ProjectTransactionDetail(ctx context.Context, userID string, paymentID string) (entities.TransactionDetail, error) {
	type mainRow struct {
		InvestorJobId  int
		PaymentId      uint64
		ProjectUid     string
		ProjectTitle   string
		ProjectSku     string
		Amount         int64
		PaymentStatus  string
		CreatedAt      string
		PaidAt         *string
		FundingStatus  string
		TargetAmount   uint64
		PaidAmount     uint64
		ReservedAmount uint64

		CanRefund bool
	}
	var m mainRow

	sqlMain := `
	  SELECT
	    p.id AS payment_id,
	    p.project_uid,
	    pr.title AS project_title,
		pr.sku AS project_sku,
	    p.amount_idr AS amount,
	   	p.payment_status               AS payment_status,
		p.investor_job_id,
	    DATE_FORMAT(p.created_at,'%Y-%m-%d %H:%i:%s') AS created_at,
	    IFNULL(DATE_FORMAT(p.paid_at,'%Y-%m-%d %H:%i:%s'), NULL) AS paid_at,
	    pr.funding_status,
	    pr.target_amount_idr     AS target_amount,
	    pr.paid_amount_idr       AS paid_amount,
	    pr.reserved_amount_idr   AS reserved_amount,

	    -- can_refund: true jika paid_at ada dan selisih waktu < 2 jam
	    CASE
	      WHEN p.paid_at IS NOT NULL
	       AND TIMESTAMPDIFF(SECOND, p.paid_at, NOW()) >= 0
	       AND TIMESTAMPDIFF(SECOND, p.paid_at, NOW()) < 7200
	      THEN TRUE ELSE FALSE
	    END AS can_refund
	  FROM payments p
	  JOIN jobs j       ON j.id  = p.investor_job_id
	  LEFT JOIN projects pr ON pr.uid = p.project_uid
	  LEFT JOIN (
			SELECT i1.*
			FROM invoices i1
			JOIN (
				SELECT payment_id, MAX(id) AS max_id
				FROM invoices
				GROUP BY payment_id
			) t ON t.payment_id = i1.payment_id AND t.max_id = i1.id
		) i ON i.payment_id = p.id
	  WHERE j.user_id = ? AND p.id = ?
	  
	  LIMIT 1`

	if err := dbDefault.WithContext(ctx).Raw(sqlMain, userID, paymentID).Scan(&m).Error; err != nil {
		return entities.TransactionDetail{}, err
	}
	if m.PaymentId == 0 {
		return entities.TransactionDetail{}, gorm.ErrRecordNotFound
	}

	type invRow struct {
		InvoiceId     uint64
		Provider      string
		OrderId       string
		Amount        int64
		InvoiceStatus string
		ChannelCode   *string
		RawResponse   *json.RawMessage
		ChannelRef    *string
		ExpiresAt     *string
		PaidAt        *string
		CreatedAt     string
	}
	var invs []invRow

	sqlInv := `
	  SELECT 
	    id AS invoice_id, provider, order_id, amount_idr AS amount,
	    invoice_status, channel_code, channel_ref, raw_response,
	    IFNULL(
			DATE_FORMAT(CONVERT_TZ(expires_at, '+00:00', '+07:00'), '%Y-%m-%d %H:%i:%s'),
			NULL
		) AS expires_at,
	    IFNULL(DATE_FORMAT(paid_at,'%Y-%m-%d %H:%i:%s'), NULL) AS paid_at,
	    DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s') AS created_at
	  FROM invoices
	  WHERE payment_id = ?
	  ORDER BY id DESC`

	if err := dbDefault.WithContext(ctx).Raw(sqlInv, paymentID).Scan(&invs).Error; err != nil {
		return entities.TransactionDetail{}, err
	}

	// Contract Letter
	paymentContractLetter := entities.PaymentContractLetter{}
	queryPaymentMethod := `SELECT path
		FROM contract_letter_payments 
		WHERE payment_id = ?`
	if err := dbDefault.Raw(queryPaymentMethod, m.PaymentId).Scan(&paymentContractLetter).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return entities.TransactionDetail{}, err
	}

	// User
	getUser := entities.PaymentInvestor{}
	queryGetUser := `SELECT u.uid AS id, p.fullname, p.selfie, u.email, u.sku
		FROM jobs j 
		INNER JOIN users u ON j.user_id = u.uid
		INNER JOIN profiles p ON j.user_id = p.user_id
		WHERE j.id = ?`
	if err := dbDefault.Raw(queryGetUser, m.InvestorJobId).Scan(&getUser).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return entities.TransactionDetail{}, err
	}

	// Company
	getCompany := entities.PaymentCompany{}
	queryCompany := `SELECT c.uid AS id, c.company_name AS name 
		FROM projects p 
		INNER JOIN companies c ON c.uid = p.company_id
		WHERE p.uid = ?`
	if err := dbDefault.Raw(queryCompany, m.ProjectUid).Scan(&getCompany).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return entities.TransactionDetail{}, err
	}

	out := entities.TransactionDetail{
		PaymentId:            m.PaymentId,
		ProjectId:            m.ProjectUid,
		ProjectTitle:         m.ProjectTitle,
		ProjectSku:           m.ProjectSku,
		Amount:               m.Amount,
		PaymentStatus:        m.PaymentStatus,
		CreatedAt:            m.CreatedAt,
		PaidAt:               m.PaidAt,
		ContractLetter:       paymentContractLetter,
		Investor:             getUser,
		Company:              getCompany,
		ProjectFundingStatus: m.FundingStatus,
		ProjectTarget:        m.TargetAmount,
		ProjectPaid:          m.PaidAmount,
		ProjectReserved:      m.ReservedAmount,
		Invoices:             make([]entities.InvoiceItem, 0, len(invs)),
	}

	for _, iv := range invs {
		// Payment Method
		paymentMethods := entities.PaymentMethod{}
		queryPaymentMethod := `SELECT id, name, nameCode as name_code, logo, platform, fee 
		FROM Channels 
		WHERE id = ?`
		if err := dbPayment.Debug().Raw(queryPaymentMethod, iv.ChannelCode).Scan(&paymentMethods).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return entities.TransactionDetail{}, err
		}

		out.Invoices = append(out.Invoices, entities.InvoiceItem{
			InvoiceId:     iv.InvoiceId,
			Provider:      iv.Provider,
			OrderId:       iv.OrderId,
			Amount:        iv.Amount,
			PaymentMethod: paymentMethods,
			InvoiceStatus: iv.InvoiceStatus,
			ChannelCode:   iv.ChannelCode,
			RawResponse:   iv.RawResponse,
			ChannelRef:    iv.ChannelRef,
			ExpiresAt:     iv.ExpiresAt,
			PaidAt:        iv.PaidAt,
			CreatedAt:     iv.CreatedAt,
		})
	}

	return out, nil
}

func AdminProjectListTransactions(
	ctx context.Context,
	statusCSV string,
	page, perPage int,
) (entities.TransactionListResp, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	statuses := normalizeStatuses(statusCSV)

	where := "1=1"
	args := make([]any, 0, 3)

	if len(statuses) > 0 {
		where += " AND p.payment_status IN ?"
		args = append(args, statuses)
	}

	var total int64
	countSQL := `
		SELECT COUNT(1)
		FROM payments p
		JOIN jobs j ON j.id = p.investor_job_id
		WHERE ` + where

	if err := dbDefault.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return entities.TransactionListResp{}, err
	}

	type row struct {
		InvestorJobId int
		PaymentId     uint64
		ProjectUid    string
		ProjectTitle  string
		ProjectSku    string
		Amount        int64
		PaymentStatus string
		CreatedAt     string
		PaidAt        *string
		OrderID       *string
		Provider      *string
		InvoiceStatus *string
		ChannelCode   *string
		ChannelRef    *string
		ExpiresAt     *string
	}
	var rows []row

	listSQL := `
		SELECT 
			p.id                           AS payment_id,
			p.project_uid                  AS project_uid,
			pr.title                       AS project_title,
			pr.sku 						   AS project_sku,
			p.amount_idr                   AS amount,
			p.payment_status               AS payment_status,
			p.investor_job_id,
			DATE_FORMAT(p.created_at,'%Y-%m-%d %H:%i:%s') AS created_at,
			IFNULL(DATE_FORMAT(p.paid_at,'%Y-%m-%d %H:%i:%s'), NULL) AS paid_at,
			i.order_id, i.provider, i.invoice_status,
			i.channel_code, i.channel_ref,
			IFNULL(
				DATE_FORMAT(CONVERT_TZ(i.expires_at, '+00:00', '+07:00'), '%Y-%m-%d %H:%i:%s'),
				NULL
			) AS expires_at
		FROM payments p
		JOIN jobs j      ON j.id = p.investor_job_id
		LEFT JOIN projects pr ON pr.uid = p.project_uid
		LEFT JOIN (
			SELECT i1.*
			FROM invoices i1
			JOIN (
				SELECT payment_id, MAX(id) AS max_id
				FROM invoices
				GROUP BY payment_id
			) t ON t.payment_id = i1.payment_id AND t.max_id = i1.id
		) i ON i.payment_id = p.id
		WHERE ` + where + `
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?`

	args2 := append(append([]any{}, args...), perPage, offset)

	if err := dbDefault.WithContext(ctx).Raw(listSQL, args2...).Scan(&rows).Error; err != nil {
		return entities.TransactionListResp{}, err
	}

	items := make([]entities.TransactionListItem, 0, len(rows))
	for _, r := range rows {

		// Contract Letter Payment
		paymentContractLetter := entities.PaymentContractLetter{}
		queryPaymentMethod := `SELECT path
		FROM contract_letter_payments 
		WHERE payment_id = ?`
		errPaymentContractLetter := dbDefault.Raw(queryPaymentMethod, r.PaymentId).Scan(&paymentContractLetter).Error
		if errPaymentContractLetter != nil {
			helper.Logger("error", "In Server: "+errPaymentContractLetter.Error())
			return entities.TransactionListResp{}, errPaymentContractLetter
		}

		// User
		getUser := entities.PaymentInvestor{}
		queryGetUser := `SELECT u.uid AS id, p.fullname, p.selfie, u.sku, u.email,
		a.bank_name, a.no AS bank_no
		FROM jobs j 
		LEFT JOIN users u ON j.user_id = u.uid
		LEFT JOIN profiles p ON j.user_id = p.user_id
		LEFT JOIN accounts a ON a.user_id = p.user_id
		WHERE j.id = ?`
		errGetUser := dbDefault.Raw(queryGetUser, r.InvestorJobId).Scan(&getUser).Error
		if errGetUser != nil {
			helper.Logger("error", "In Server: "+errGetUser.Error())
			return entities.TransactionListResp{}, errGetUser
		}

		// Refund Document
		getPaymentProjectRefundTransDoc := entities.PaymentProjectRefundTransDoc{}
		queryPaymentProjectRefundTransDoc := `SELECT path, transaction_id, created_at 
		FROM refund_transfer_documents
		WHERE transaction_id = ?`
		errGetPaymentProjectRefundTransDoc := dbDefault.Raw(queryPaymentProjectRefundTransDoc, r.PaymentId).Scan(&getPaymentProjectRefundTransDoc).Error
		if errGetPaymentProjectRefundTransDoc != nil {
			helper.Logger("error", "In Server: "+errGetPaymentProjectRefundTransDoc.Error())
			return entities.TransactionListResp{}, errGetPaymentProjectRefundTransDoc
		}

		// Company
		getCompany := entities.PaymentCompany{}
		queryCompany := `SELECT c.uid AS id, c.company_name AS name, c.bank_name, 
		c.latest_amendment_deed AS akta_perubahan_terakhir, 
		c.latest_amendment_deed_path AS akta_perubahan_terahkir_path,
		c.bank_account AS bank_no 
		FROM projects p 
		LEFT JOIN companies c ON c.uid = p.company_id
		WHERE p.uid = ?`
		errGetCompany := dbDefault.Raw(queryCompany, r.ProjectUid).Scan(&getCompany).Error
		if errGetCompany != nil {
			helper.Logger("error", "In Server: "+errGetCompany.Error())
			return entities.TransactionListResp{}, errGetCompany
		}

		var isRefund bool

		if getPaymentProjectRefundTransDoc.Path == "" {
			isRefund = false
		} else {
			isRefund = true
		}

		items = append(items, entities.TransactionListItem{
			PaymentId:     r.PaymentId,
			ProjectId:     r.ProjectUid,
			ProjectTitle:  r.ProjectTitle,
			ProjectSku:    r.ProjectSku,
			Amount:        r.Amount,
			PaymentStatus: r.PaymentStatus,
			CreatedAt:     r.CreatedAt,
			PaidAt:        r.PaidAt,
			ContractLetter: entities.PaymentContractLetter{
				Path: helper.DefaultIfEmpty(paymentContractLetter.Path, "-"),
			},
			Investor: entities.PaymentInvestor{
				Id:       getUser.Id,
				Fullname: getUser.Fullname,
				Selfie:   getUser.Selfie,
				Sku:      getUser.Sku,
				Email:    getUser.Email,
				BankName: helper.DefaultIfEmpty(getUser.BankName, getCompany.BankName),
				BankNo:   helper.DefaultIfEmpty(getUser.BankNo, getCompany.BankNo),
			},
			Company: getCompany,
			Refund: entities.PaymentProjectRefundTransDoc{
				Path:      helper.DefaultIfEmpty(getPaymentProjectRefundTransDoc.Path, "-"),
				CreatedAt: getPaymentProjectRefundTransDoc.CreatedAt,
			},
			IsRefund:      isRefund,
			OrderId:       r.OrderID,
			Provider:      r.Provider,
			InvoiceStatus: r.InvoiceStatus,
			ChannelCode:   r.ChannelCode,
			ChannelRef:    r.ChannelRef,
			ExpiresAt:     r.ExpiresAt,
		})
	}

	return entities.TransactionListResp{
		Items:      items,
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
	}, nil
}

func AdminProjectGetTransactionDetail(ctx context.Context, paymentID string) (entities.TransactionDetail, error) {
	type mainRow struct {
		InvestorJobId int
		PaymentId     uint64
		ProjectUid    string
		ProjectTitle  string
		ProjectSku    string
		Amount        int64
		PaymentStatus string
		CreatedAt     string
		PaidAt        *string

		FundingStatus  string
		TargetAmount   uint64
		PaidAmount     uint64
		ReservedAmount uint64
	}
	var m mainRow

	sqlMain := `
	  SELECT
	  	p.investor_job_id,
	    p.id AS payment_id,
	    p.project_uid,
	    pr.title AS project_title,
		pr.sku AS project_sku,
	    p.amount_idr AS amount,
	    p.payment_status,
	    DATE_FORMAT(p.created_at,'%Y-%m-%d %H:%i:%s') AS created_at,
	    IFNULL(DATE_FORMAT(p.paid_at,'%Y-%m-%d %H:%i:%s'), NULL) AS paid_at,
	    pr.funding_status,
	    pr.target_amount_idr AS target_amount,
	    pr.paid_amount_idr   AS paid_amount,
	    pr.reserved_amount_idr AS reserved_amount
	  FROM payments p
	  JOIN jobs j ON j.id = p.investor_job_id
	  LEFT JOIN projects pr ON pr.uid = p.project_uid
	  WHERE p.id = ?
	  LIMIT 1`
	if err := dbDefault.WithContext(ctx).Raw(sqlMain, paymentID).Scan(&m).Error; err != nil {
		return entities.TransactionDetail{}, err
	}
	if m.PaymentId == 0 {
		return entities.TransactionDetail{}, gorm.ErrRecordNotFound
	}

	// User
	getUser := entities.PaymentInvestor{}
	queryGetUser := `SELECT u.uid AS id, p.fullname, p.selfie, u.sku, u.email,
		a.bank_name, a.no AS bank_no
		FROM jobs j 
		LEFT JOIN users u ON j.user_id = u.uid
		LEFT JOIN profiles p ON j.user_id = p.user_id
		LEFT JOIN accounts a ON a.user_id = p.user_id
		WHERE j.id = ?`
	errGetUser := dbDefault.Raw(queryGetUser, m.InvestorJobId).Scan(&getUser).Error
	if errGetUser != nil {
		helper.Logger("error", "In Server: "+errGetUser.Error())
		return entities.TransactionDetail{}, errGetUser
	}

	// Refund Document
	getPaymentProjectRefundTransDoc := entities.PaymentProjectRefundTransDoc{}
	queryPaymentProjectRefundTransDoc := `SELECT path, transaction_id, created_at 
		FROM refund_transfer_documents
		WHERE transaction_id = ?`
	errGetPaymentProjectRefundTransDoc := dbDefault.Raw(queryPaymentProjectRefundTransDoc, m.PaymentId).Scan(&getPaymentProjectRefundTransDoc).Error
	if errGetPaymentProjectRefundTransDoc != nil {
		helper.Logger("error", "In Server: "+errGetPaymentProjectRefundTransDoc.Error())
		return entities.TransactionDetail{}, errGetPaymentProjectRefundTransDoc
	}

	// Company
	getCompany := entities.PaymentCompany{}
	queryCompany := `SELECT c.uid AS id, c.company_name AS name,
		c.bank_name, c.bank_account AS bank_no,
		c.latest_amendment_deed AS akta_perubahan_terakhir, 
		c.latest_amendment_deed_path AS akta_perubahan_terahkir_path
		FROM projects p 
		INNER JOIN companies c ON c.uid = p.company_id
		WHERE p.uid = ?`
	errGetCompany := dbDefault.Raw(queryCompany, m.ProjectUid).Scan(&getCompany).Error
	if errGetCompany != nil {
		helper.Logger("error", "In Server: "+errGetCompany.Error())
		return entities.TransactionDetail{}, errGetCompany
	}

	type invRow struct {
		InvoiceID     uint64
		Provider      string
		OrderID       string
		Amount        int64
		InvoiceStatus string
		ChannelCode   *string
		RawResponse   *json.RawMessage
		ChannelRef    *string
		ExpiresAt     *string
		PaidAt        *string
		CreatedAt     string
	}
	var invs []invRow
	sqlInv := `
	  SELECT 
	    id AS invoice_id, provider, order_id, amount_idr AS amount,
	    invoice_status, channel_code, channel_ref, raw_response,
    	IFNULL(
			DATE_FORMAT(CONVERT_TZ(expires_at, '+00:00', '+07:00'), '%Y-%m-%d %H:%i:%s'),
			NULL
		) AS expires_at,
	    IFNULL(DATE_FORMAT(paid_at,'%Y-%m-%d %H:%i:%s'), NULL) AS paid_at,
	    DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s') AS created_at
	  FROM invoices
	  WHERE payment_id = ?
	  ORDER BY id DESC`
	if err := dbDefault.WithContext(ctx).Raw(sqlInv, paymentID).Scan(&invs).Error; err != nil {
		return entities.TransactionDetail{}, err
	}

	// Get Contract Letter Payment
	paymentContractLetter := entities.PaymentContractLetter{}
	queryPaymentMethod := `SELECT path
		FROM contract_letter_payments 
		WHERE payment_id = ?`
	errUser := dbDefault.Raw(queryPaymentMethod, m.PaymentId).Scan(&paymentContractLetter).Error
	if errUser != nil {
		helper.Logger("error", "In Server: "+errUser.Error())
		return entities.TransactionDetail{}, errUser
	}

	var isRefund bool

	if getPaymentProjectRefundTransDoc.Path == "" {
		isRefund = false
	} else {
		isRefund = true
	}

	out := entities.TransactionDetail{
		PaymentId:      m.PaymentId,
		ProjectId:      m.ProjectUid,
		ProjectTitle:   m.ProjectTitle,
		ProjectSku:     m.ProjectSku,
		Amount:         m.Amount,
		PaymentStatus:  m.PaymentStatus,
		CreatedAt:      m.CreatedAt,
		PaidAt:         m.PaidAt,
		ContractLetter: paymentContractLetter,
		Investor: entities.PaymentInvestor{
			Id:       getUser.Id,
			Fullname: getUser.Fullname,
			Selfie:   getUser.Selfie,
			Sku:      getUser.Sku,
			Email:    getUser.Email,
			BankName: helper.DefaultIfEmpty(getUser.BankName, getCompany.BankName),
			BankNo:   helper.DefaultIfEmpty(getUser.BankNo, getCompany.BankNo),
		},
		Company: getCompany,
		Refund: entities.PaymentProjectRefundTransDoc{
			Path:      helper.DefaultIfEmpty(getPaymentProjectRefundTransDoc.Path, "-"),
			CreatedAt: getPaymentProjectRefundTransDoc.CreatedAt,
		},
		IsRefund:             isRefund,
		ProjectFundingStatus: m.FundingStatus,
		ProjectTarget:        m.TargetAmount,
		ProjectPaid:          m.PaidAmount,
		ProjectReserved:      m.ReservedAmount,
		Invoices:             make([]entities.InvoiceItem, 0, len(invs)),
	}
	for _, iv := range invs {

		// Get Payment Method
		paymentMethods := entities.PaymentMethod{}
		queryPaymentMethod := `SELECT id, name, nameCode as name_code, logo, platform, fee 
		FROM Channels 
		WHERE id = ?`
		errUser := dbPayment.Debug().Raw(queryPaymentMethod, iv.ChannelCode).Scan(&paymentMethods).Error
		if errUser != nil {
			helper.Logger("error", "In Server: "+errUser.Error())
			return entities.TransactionDetail{}, errUser
		}

		out.Invoices = append(out.Invoices, entities.InvoiceItem{
			InvoiceId:     iv.InvoiceID,
			Provider:      iv.Provider,
			OrderId:       iv.OrderID,
			Amount:        iv.Amount,
			PaymentMethod: paymentMethods,
			InvoiceStatus: iv.InvoiceStatus,
			ChannelCode:   iv.ChannelCode,
			RawResponse:   iv.RawResponse,
			ChannelRef:    iv.ChannelRef,
			ExpiresAt:     iv.ExpiresAt,
			PaidAt:        iv.PaidAt,
			CreatedAt:     iv.CreatedAt,
		})
	}
	return out, nil
}
