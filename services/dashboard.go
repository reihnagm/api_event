package services

import (
	"context"
	"time"

	"superapps/entities"
	helper "superapps/helpers"
)

const InstitusiRoleID = 9

func u64ptr(v uint64) *uint64 { return &v }

func DashboardInvestor(ctx context.Context, userId string, recentLimit int) (entities.InvestorDashboard, error) {
	if recentLimit <= 0 || recentLimit > 50 {
		recentLimit = 10
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	curYear := now.Year()

	// 1) verifikasi user
	var verifyInvestor bool
	if err := dbDefault.WithContext(ctx).
		Raw(`SELECT COALESCE(verify_investor,0) FROM users WHERE uid=? LIMIT 1`, userId).
		Scan(&verifyInvestor).Error; err != nil {
		return entities.InvestorDashboard{}, err
	}

	// 2) job terbaru (boleh kosong)
	var job struct {
		Id              int64
		AnnualIncomeIdr uint64
		AnnualQuotaIdr  uint64
	}
	if err := dbDefault.WithContext(ctx).Raw(`
	  SELECT id, annual_income_idr, annual_quota_idr
	  FROM jobs
	  WHERE user_id = ?
	  ORDER BY updated_at DESC
	  LIMIT 1`, userId).Scan(&job).Error; err != nil {
		return entities.InvestorDashboard{}, err
	}

	// 3) punya rekening efek?
	var hasSecAccount bool
	if err := dbDefault.WithContext(ctx).
		Raw(`
        SELECT EXISTS(
            SELECT 1
            FROM security_accounts
            WHERE user_id = ?
              AND account_no IS NOT NULL
              AND account_no <> ''
            LIMIT 1
        )
    `, userId).
		Scan(&hasSecAccount).Error; err != nil {
		return entities.InvestorDashboard{}, err
	}

	// 4) role institusi? (users.role = 9)
	var isInstitusi bool
	if err := dbDefault.WithContext(ctx).
		Raw(`SELECT EXISTS(SELECT 1 FROM users WHERE uid=? AND role=?)`, userId, InstitusiRoleID).
		Scan(&isInstitusi).Error; err != nil {
		return entities.InvestorDashboard{}, err
	}

	// 5) SUMMARY — HANYA PAID
	var sumRow struct {
		PaidAllTime  uint64
		PaidThisYear uint64
		ProjectsCnt  int64
		ActiveInvCnt int64
	}
	if err := dbDefault.WithContext(ctx).Raw(`
	  SELECT
	    /* total PAID all time */
	    COALESCE(SUM(IF(p.payment_status='PAID', p.amount_idr, 0)),0) AS paid_all_time,

	    /* total PAID this year */
	    COALESCE(SUM(
	      IF(p.payment_year = ? AND p.payment_status='PAID', p.amount_idr, 0)
	    ),0) AS paid_this_year,

	    /* jumlah proyek yang sudah PAID (tanpa pending) */
	    COUNT(DISTINCT IF(p.payment_status='PAID', p.project_uid, NULL)) AS projects_cnt,

	    /* active invoices (ISSUED & belum expired) tetap ditampilkan */
	    (
	      SELECT COUNT(1)
	      FROM invoices i
	      WHERE i.invoice_status='ISSUED'
	        AND (i.expires_at IS NULL OR i.expires_at > UTC_TIMESTAMP())
	        AND EXISTS(SELECT 1 FROM payments px WHERE px.id=i.payment_id AND px.investor_job_id=?)
	    ) AS active_inv_cnt
	  FROM payments p
	  WHERE p.investor_job_id = ?`,
		curYear, job.Id, job.Id,
	).Scan(&sumRow).Error; err != nil {
		return entities.InvestorDashboard{}, err
	}

	// used_this_year_idr = paid_this_year_idr (tanpa pending)
	summary := entities.InvestorSummary{
		JobID:           job.Id,
		AnnualIncomeIDR: job.AnnualIncomeIdr,
		UsedThisYearIDR: sumRow.PaidThisYear, // <-- hanya PAID
		PaidAllTimeIDR:  sumRow.PaidAllTime,
		PaidThisYearIDR: sumRow.PaidThisYear,
		ProjectsCount:   sumRow.ProjectsCnt,
		ActiveInvoices:  sumRow.ActiveInvCnt,
		QuotaEnforced:   !(hasSecAccount || isInstitusi),
	}
	if hasSecAccount || isInstitusi || job.Id == 0 {
		summary.AnnualQuotaIDR = nil
		summary.RemainingQuotaIDR = nil
	} else {
		remain := helper.SafeSub(job.AnnualQuotaIdr, sumRow.PaidThisYear) // <-- pakai PAID this year
		summary.AnnualQuotaIDR = u64ptr(job.AnnualQuotaIdr)
		summary.RemainingQuotaIDR = &remain
	}

	// 6) recent transactions (tetap)
	type txRow struct {
		PaymentId     uint64
		ProjectUid    string
		ProjectTitle  string
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
		PaymentURL    *string
	}
	recentByProject := make(map[string][]entities.InvestorTransactionListItem, 32)
	if job.Id != 0 && recentLimit > 0 {
		var txs []txRow
		const qRecent = `
		  WITH i_latest AS (
		    SELECT
		      id, payment_id, order_id, provider, invoice_status,
		      channel_code, channel_ref, expires_at, paid_at, created_at,
		      JSON_UNQUOTE(JSON_EXTRACT(raw_response, '$.data.data.actions[0].url')) AS payment_url,
		      ROW_NUMBER() OVER (PARTITION BY payment_id ORDER BY id DESC) AS rn
		    FROM invoices
		  )
		  SELECT 
		    p.id AS payment_id,
		    p.project_uid,
		    pr.title AS project_title,
		    p.amount_idr AS amount,
		    p.payment_status,
		    DATE_FORMAT(p.created_at,'%Y-%m-%d %H:%i:%s') AS created_at,
		    IFNULL(DATE_FORMAT(p.paid_at,'%Y-%m-%d %H:%i:%s'), NULL) AS paid_at,
		    i.order_id, i.provider, i.invoice_status,
		    i.channel_code, i.channel_ref,
		    IFNULL(DATE_FORMAT(i.expires_at,'%Y-%m-%d %H:%i:%s'), NULL) AS expires_at,
		    i.payment_url
		  FROM payments p
		  LEFT JOIN projects pr ON pr.uid = p.project_uid
		  LEFT JOIN i_latest i ON i.payment_id = p.id AND i.rn = 1
		  WHERE p.investor_job_id = ?
		  ORDER BY p.created_at DESC
		  LIMIT ?`
		if err := dbDefault.WithContext(ctx).Raw(qRecent, job.Id, recentLimit).Scan(&txs).Error; err != nil {
			return entities.InvestorDashboard{}, err
		}
		for _, r := range txs {
			item := entities.InvestorTransactionListItem{
				PaymentID:     r.PaymentId,
				ProjectUID:    r.ProjectUid,
				ProjectTitle:  r.ProjectTitle,
				Amount:        r.Amount,
				PaymentStatus: r.PaymentStatus,
				CreatedAt:     r.CreatedAt,
				PaidAt:        r.PaidAt,
				OrderID:       r.OrderID,
				Provider:      r.Provider,
				InvoiceStatus: r.InvoiceStatus,
				ChannelCode:   r.ChannelCode,
				ChannelRef:    r.ChannelRef,
				ExpiresAt:     r.ExpiresAt,
				PaymentURL:    r.PaymentURL,
			}
			recentByProject[r.ProjectUid] = append(recentByProject[r.ProjectUid], item)
		}
	}

	// 7) PORTFOLIO — pending aktif tetap dihitung (tidak memengaruhi kuota)
	portfolio := make([]entities.ProjectPortfolioItem, 0)
	if job.Id != 0 {
		type pfRow struct {
			ProjectUID               string
			ProjectTitle             string
			FundingStatus            string
			TargetAmount             uint64
			UserPaidIdr              int64
			UserPendingIdr           int64
			ProjectPaidAmountIdr     uint64
			ProjectReservedAmountIdr uint64
		}
		var pf []pfRow
		const qPortfolio = `
		  SELECT
		    p.project_uid,
		    pr.title AS project_title,
		    pr.funding_status,
		    pr.target_amount_idr AS target_amount,

		    /* paid & pending user ini */
		    SUM(IF(p.payment_status='PAID', p.amount_idr, 0)) AS user_paid_idr,
		    SUM(
		      IF(
		        p.payment_status='PENDING'
		        AND EXISTS(
		          SELECT 1 FROM invoices ix
		          WHERE ix.payment_id = p.id
		            AND ix.invoice_status = 'ISSUED'
		            AND (ix.expires_at IS NULL OR ix.expires_at > UTC_TIMESTAMP())
		        ),
		        p.amount_idr, 0
		      )
		    ) AS user_pending_idr,

		    /* PAID proyek (agregat di projects) */
		    pr.paid_amount_idr AS project_paid_amount_idr,

		    /* RESERVED proyek = semua PENDING aktif semua investor (bukan expired) */
		    (
		      SELECT COALESCE(SUM(pp.amount_idr),0)
		      FROM payments pp
		      WHERE pp.project_uid = p.project_uid
		        AND pp.payment_status = 'PENDING'
		        AND EXISTS(
		          SELECT 1 FROM invoices ii
		          WHERE ii.payment_id = pp.id
		            AND ii.invoice_status = 'ISSUED'
		            AND (ii.expires_at IS NULL OR ii.expires_at > UTC_TIMESTAMP())
		        )
		    ) AS project_reserved_amount_idr

		  FROM payments p
		  LEFT JOIN projects pr ON pr.uid = p.project_uid
		  WHERE p.investor_job_id = ?
		    AND (p.payment_status IN ('PAID','PENDING'))
		  GROUP BY p.project_uid, pr.title, pr.funding_status, pr.target_amount_idr, pr.paid_amount_idr
		  ORDER BY pr.title ASC`
		if err := dbDefault.WithContext(ctx).Raw(qPortfolio, job.Id).Scan(&pf).Error; err != nil {
			return entities.InvestorDashboard{}, err
		}

		portfolio = make([]entities.ProjectPortfolioItem, 0, len(pf))
		for _, r := range pf {
			portfolio = append(portfolio, entities.ProjectPortfolioItem{
				ProjectUID:               r.ProjectUID,
				ProjectTitle:             r.ProjectTitle,
				FundingStatus:            r.FundingStatus,
				TargetAmount:             r.TargetAmount,
				UserPaidIdr:              r.UserPaidIdr,
				UserPending:              r.UserPendingIdr,
				ProjectPaidAmountIdr:     r.ProjectPaidAmountIdr,
				ProjectReservedAmountIdr: r.ProjectReservedAmountIdr,
				RecentTransactions:       recentByProject[r.ProjectUID],
			})
		}
	}

	return entities.InvestorDashboard{
		VerifiedInvestor: verifyInvestor,
		RekEfek:          hasSecAccount,
		IsInstitusi:      isInstitusi,
		Summary:          summary,
		Portfolio:        portfolio,
	}, nil
}
