package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"superapps/entities"
	helper "superapps/helpers"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func HandleProjectPaymentCallback(in *entities.Callback) error {
	if in == nil {
		return errors.New("nil payload")
	}

	orderID := strings.TrimSpace(in.OrderId)
	if orderID == "" {
		return errors.New("order_id is required")
	}

	status := strings.ToUpper(strings.TrimSpace(in.Status))

	var inv struct {
		ProjectId     string `json:"project_id"`
		InvoiceStatus string `json:"invoice_status"`
	}
	if err := dbDefault.Raw(
		`SELECT project_uid AS project_id, invoice_status
		   FROM invoices
		  WHERE provider = 'midtrans' AND order_id = ?`,
		orderID,
	).Scan(&inv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || inv.InvoiceStatus == "" {
			return fmt.Errorf("invoice not found for order_id=%s", orderID)
		}
		return err
	}
	if inv.InvoiceStatus == "" {
		return fmt.Errorf("invoice not found for order_id=%s", orderID)
	}

	if inv.InvoiceStatus == "PAID" && status == "PAID" {
		_ = saveCallbackSnapshot(orderID, in)
		return nil
	}
	if inv.InvoiceStatus != "ISSUED" && status != "PAID" {
		_ = saveCallbackSnapshot(orderID, in)
		return nil
	}

	switch status {
	case "PAID":
		if err := dbDefault.Exec(
			"CALL sp_mark_invoice_paid(?,?,?)",
			"midtrans", orderID, nil,
		).Error; err != nil {
			var me *mysql.MySQLError
			if errors.As(err, &me) && me.Number == 1644 {
				helper.Logger("warn", "sp_mark_invoice_paid: "+me.Message)
			} else {
				_ = saveCallbackSnapshot(orderID, in)
				return err
			}
		}

		var ps struct {
			Id    string `json:"id"`
			Title string `json:"title"`
		}
		if err := dbDefault.Raw(
			`SELECT 
				sku AS id,
				title
			   FROM projects
			  WHERE uid = ?
			  LIMIT 1`,
			inv.ProjectId,
		).Scan(&ps).Error; err != nil {
			helper.Logger("error", fmt.Sprintf("load project summary failed (project_uid=%s): %v", inv.ProjectId, err))
		}

		if ps.Title != "" {
			emails, err := helper.GetEmailsByRoleProjectAnalyst(dbDefault)
			if err != nil {
				helper.Logger("error", fmt.Sprintf("cannot get analyst emails: %v", err))
			} else {

				subject := fmt.Sprintf(`Pembayaran Berhasil: Proyek "%s" siap ditinjau`, safe(ps.Title, ps.Id))

				htmlBody := fmt.Sprintf(`<!doctype html>
					<html lang="id">
					<head>
						<meta charset="utf-8">
						<title>%[1]s</title>
						<meta name="viewport" content="width=device-width, initial-scale=1">
					</head>
					<body style="margin:0;padding:0;background:#f6f7f9;font-family:Arial,Helvetica,sans-serif;">
						<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background:#f6f7f9;padding:24px 0;">
						<tr>
							<td align="center">
							<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="600" style="background:#ffffff;border-radius:12px;overflow:hidden;border:1px solid #e5e7eb;">
								<tr>
								<td style="background:#111827;color:#ffffff;padding:20px 24px;font-size:18px;font-weight:700;">
									Fulusme &mdash; Konfirmasi Pembayaran Proyek
								</td>
								</tr>
								<tr>
								<td style="padding:24px;">
									<p style="margin:0 0 12px 0;color:#111827;font-size:16px;">Halo Tim Project Analyst,</p>
									<p style="margin:0 0 16px 0;color:#374151;font-size:14px;line-height:1.6;">
									Transaksi pembayaran untuk proyek berikut telah <strong>BERHASIL</strong> diverifikasi.
									Silakan lanjutkan proses review sesuai SOP.
									</p>

									<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background:#f9fafb;border:1px solid #e5e7eb;border-radius:8px;margin:12px 0;">
									<tr>
										<td style="padding:12px 16px;width:35%%;color:#6b7280;font-size:13px;">ID Proyek</td>
										<td style="padding:12px 16px;color:#111827;font-size:13px;">%[2]s</td>
									</tr>
									<tr>
										<td style="padding:12px 16px;color:#6b7280;font-size:13px;">Judul Proyek</td>
										<td style="padding:12px 16px;color:#111827;font-size:13px;">%[3]s</td>
									</tr>
									<tr>
										<td style="padding:12px 16px;color:#6b7280;font-size:13px;">Order ID</td>
										<td style="padding:12px 16px;color:#111827;font-size:13px;">%[4]s</td>
									</tr>
									<tr>
										<td style="padding:12px 16px;color:#6b7280;font-size:13px;">Status</td>
										<td style="padding:12px 16px;color:#111827;font-size:13px;">%[5]s</td>
									</tr>
									</table>
								</td>
								</tr>
								<tr>
								<td style="background:#f3f4f6;color:#6b7280;padding:14px 24px;text-align:center;font-size:12px;">
									&copy; %d Fulusme. All rights reserved.
								</td>
								</tr>
							</table>
							</td>
						</tr>
						</table>
					</body>
					</html>`,
					subject,
					ps.Id,
					ps.Title,
					orderID,
					status,
					time.Now().Year(),
				)

				for _, email := range emails {
					if err := helper.SendEmail(email, "Fulusme", subject, htmlBody, "another-otp"); err != nil {
						helper.Logger("error", fmt.Sprintf("gagal kirim email ke %s: %v", email, err))
						continue
					}
					helper.Logger("info", fmt.Sprintf("email konfirmasi pembayaran terkirim ke %s", email))
				}
			}
		}

	default:
		// PENDING/CANCELLED/EXPIRED → tidak ada aksi khusus di sini
	}

	// 4) Snapshot raw callback (opsional)
	_ = saveCallbackSnapshot(orderID, in)

	return nil
}

func saveCallbackSnapshot(orderID string, in *entities.Callback) error {
	b, _ := json.Marshal(in)
	return dbDefault.Exec(`
		UPDATE invoices
		   SET raw_response = JSON_SET(
				COALESCE(raw_response, JSON_OBJECT()),
				'$.callback',
				CAST(? AS JSON)
		   )
		 WHERE provider='midtrans'
		   AND order_id = ?`,
		string(b), orderID,
	).Error
}
