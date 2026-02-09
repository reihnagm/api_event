package jobs

import (
	"context"
	"fmt"
	"time"

	helper "superapps/helpers"
	"superapps/services"

	"gorm.io/gorm"
)

const batchSize = 1000

// RunExpireInvoices: ubah ISSUED -> EXPIRED jika expires_at <= NOW()
// return rowsAffected untuk keperluan logging/observability
func RunExpireInvoices(ctx context.Context) (int64, error) {
	db := services.GetDefaultDB()
	if db == nil {
		return 0, fmt.Errorf("dbPayment is nil")
	}

	start := time.Now()
	var rows int64

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// NOTE: kalau MySQL kamu belum support UPDATE ... LIMIT, hapus "LIMIT ?"
		res := tx.Exec(`
		UPDATE invoices
		SET invoice_status = 'EXPIRED',
			updated_at     = NOW()
		WHERE invoice_status = 'ISSUED'
			AND expires_at IS NOT NULL
			AND CONVERT_TZ(expires_at, '+00:00', '+07:00') <= NOW()
		ORDER BY id
		LIMIT ?`, batchSize)
		if res.Error != nil {
			return res.Error
		}
		rows = res.RowsAffected
		return nil
	})

	if err != nil {
		helper.Logger("error", fmt.Sprintf("[expire-job] ERROR: %v", err))
	} else {
		helper.Logger("info", fmt.Sprintf("[expire-job] expired=%d took=%s", rows, time.Since(start)))
	}

	return rows, err
}
