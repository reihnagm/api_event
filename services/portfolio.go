package services

import (
	"context"
	"errors"
	"strings"

	"superapps/entities"
)

func PortfolioInfo(ctx context.Context, userId string) (entities.PortfolioOnlyResponse, error) {
	userId = strings.TrimSpace(userId)
	empty := entities.PortfolioOnlyResponse{Portfolio: make([]entities.MinimalPortfolioItem, 0)}
	if userId == "" {
		return empty, errors.New("user_id is required")
	}

	// Ambil latest job
	var jobID int64
	const qJob = `
	  SELECT j.id
	  FROM jobs j
	  WHERE j.user_id = ?
	  ORDER BY j.updated_at DESC
	  LIMIT 1`
	if err := dbDefault.WithContext(ctx).Raw(qJob, userId).Scan(&jobID).Error; err != nil {
		return empty, err
	}
	if jobID == 0 {
		return empty, nil
	}

	// VALIDASI: hanya tampilkan portfolio jika ada contract_letter_payments untuk job ini
	// (hanya hitung yang path-nya ada/isi)
	var clpCount int64
	const qHasCLP = `
	  SELECT COUNT(1)
	  FROM contract_letter_payments clp
	  JOIN payments p ON p.id = clp.payment_id
	  WHERE p.investor_job_id = ?
	    AND clp.path IS NOT NULL
	    AND clp.path <> ''`
	if err := dbDefault.WithContext(ctx).Raw(qHasCLP, jobID).Scan(&clpCount).Error; err != nil {
		return empty, err
	}
	if clpCount == 0 {
		// Tidak ada contract letter => portfolio disembunyikan ([])
		return empty, nil
	}

	// Query portfolio minimal — hanya proyek yang PEMBAYARANNYA punya contract letter
	const qPortfolio = `
	  SELECT
	    p.project_uid,
	    pr.title AS project_title,
	    pr.funding_status,
	    pr.target_amount_idr AS target_amount_idr,
	    SUM(IF(p.payment_status='PAID',    p.amount_idr, 0)) AS user_paid_idr,
	    SUM(IF(p.payment_status='PENDING', p.amount_idr, 0)) AS user_pending_idr,
	    pr.paid_amount_idr     AS project_paid_amount_idr,
	    pr.reserved_amount_idr AS project_reserved_amount_idr
	  FROM payments p
	  LEFT JOIN projects pr ON pr.uid = p.project_uid
	  WHERE p.investor_job_id = ?
	    AND p.payment_status IN ('PAID','PENDING')
	    AND EXISTS (
	      SELECT 1
	      FROM contract_letter_payments clp
	      WHERE clp.payment_id = p.id
	        AND clp.path IS NOT NULL
	        AND clp.path <> ''
	    )
	  GROUP BY p.project_uid, pr.title, pr.funding_status, pr.target_amount_idr,
	           pr.paid_amount_idr, pr.reserved_amount_idr
	  ORDER BY pr.title ASC`

	items := make([]entities.MinimalPortfolioItem, 0)
	if err := dbDefault.WithContext(ctx).Raw(qPortfolio, jobID).Scan(&items).Error; err != nil {
		return empty, err
	}

	return entities.PortfolioOnlyResponse{Portfolio: items}, nil
}
