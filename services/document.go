package services

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"superapps/entities"
	helper "superapps/helpers"
	"time"

	"gorm.io/gorm"
)

func DocumentTransactionPayment(r *http.Request, dtp *entities.DocumentTransactionPayment) (map[string]any, error) {
	if dtp == nil {
		return nil, fmt.Errorf("nil DocumentTransactionPayment")
	}

	dtp.Path = strings.TrimSpace(dtp.Path)

	fields := map[string]string{
		"Path":      dtp.Path,
		"ProjectId": dtp.ProjectId,
	}

	var missing []string

	for name, value := range fields {
		if value == "" {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("validation failed: missing/empty fields: %s", strings.Join(missing, ", "))
	}

	var projectPaymentId uint64

	// ----- write DB (atomic) -----
	if err := dbDefault.Transaction(func(tx *gorm.DB) error {
		const queryInsert = `
			INSERT INTO project_payments (path, project_id)
			VALUES (?, ?)
		`
		if err := tx.Exec(queryInsert, dtp.Path, dtp.ProjectId).Error; err != nil {
			helper.Logger("error", "insert project_payments: "+err.Error())
			return fmt.Errorf("insert project_payments failed: %w", err)
		}

		if err := tx.Raw("SELECT LAST_INSERT_ID()").Scan(&projectPaymentId).Error; err != nil {
			helper.Logger("error", "fetching LAST_INSERT_ID: "+err.Error())
			return fmt.Errorf("could not fetch inserted id: %w", err)
		}

		const queryUpdate = `UPDATE inboxes SET type = '2' WHERE id = ?`
		if err := tx.Exec(queryUpdate, dtp.InboxId).Error; err != nil {
			helper.Logger("error", "update inboxes: "+err.Error())
			return fmt.Errorf("update inboxes failed: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var ps struct {
		Id    string
		Title string
		Sku   string
	}
	if err := dbDefault.
		Raw(`SELECT uid AS id, title, sku FROM projects WHERE uid = ? LIMIT 1`, dtp.ProjectId).
		Scan(&ps).Error; err != nil {
		helper.Logger("error", "get project: "+err.Error())
	}

	emailProjectAnalys, err := helper.GetEmailsByRoleProjectAnalyst(dbDefault)
	if err != nil {
		helper.Logger("error", fmt.Sprintf("cannot get analyst emails: %v", err))
	} else {
		subject := fmt.Sprintf(`Bukti Pembayaran Proyek "%s" siap ditinjau`, safe(ps.Title, ps.Id))

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
							<p style="margin:0 0 12px 0;color:#111827;font-size:14px;">Halo Tim Project Analyst dan Tim Publish</p>
							<p style="margin:0 0 16px 0;color:#374151;font-size:14px;line-height:1.6;">
								Dengan ini kami informasikan bahwa bukti pembayaran untuk proyek:
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
								<td style="padding:12px 16px;color:#6b7280;font-size:13px;">Bukti Pembayaran</td>
								<td style="padding:12px 16px;color:#111827;font-size:13px;"><a href=%[4]s>Link</a></td>
							</tr>
							</table>

							<p style="color:#374151;font-size:14px;">Telah dikirim oleh penerbit. Mohon kepada Admin Tim Publish untuk segera melakukan verifikasi atas bukti pembayaran tersebut agar proses dapat dilanjutkan sesuai SOP yang berlaku. </p>
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
			ps.Sku,
			ps.Title,
			dtp.Path,
			time.Now().Year(),
		)

		for _, email := range emailProjectAnalys {
			if err := helper.SendEmail(email, "Fulusme", subject, htmlBody, "another-otp"); err != nil {
				helper.Logger("error", fmt.Sprintf("gagal kirim email ke %s: %v", email, err))
				continue
			}
			helper.Logger("info", fmt.Sprintf("email konfirmasi pembayaran terkirim ke %s", email))
		}
	}

	emailProjectPublish, err := helper.GetEmailsByRoleProjectPublish(dbDefault)
	if err != nil {
		helper.Logger("error", fmt.Sprintf("cannot get publish emails: %v", err))
	} else {
		subject := fmt.Sprintf(`Bukti Pembayaran Proyek "%s" siap ditinjau`, safe(ps.Title, ps.Id))

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
							<p style="margin:0 0 12px 0;color:#111827;font-size:16px;">Halo Tim Project Analyst dan Tim Publish</p>
							<p style="margin:0 0 16px 0;color:#374151;font-size:14px;line-height:1.6;">
								Dengan ini kami informasikan bahwa bukti pembayaran untuk proyek:
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
									<td style="padding:12px 16px;color:#6b7280;font-size:13px;">Butki Pembayaran</td>
									<td style="padding:12px 16px;color:#111827;font-size:13px;"><a href=%[4]s>Link</a></td>
								</tr>
							</table>

							<p style="color:#374151;font-size:14px;">Telah dikirim oleh penerbit. Mohon kepada Admin Tim Publish untuk segera melakukan verifikasi atas bukti pembayaran tersebut agar proses dapat dilanjutkan sesuai SOP yang berlaku. </p>
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
			ps.Sku,
			ps.Title,
			dtp.Path,
			time.Now().Year(),
		)

		for _, email := range emailProjectPublish {
			if err := helper.SendEmail(email, "Fulusme", subject, htmlBody, "another-otp"); err != nil {
				helper.Logger("error", fmt.Sprintf("gagal kirim email ke %s: %v", email, err))
				continue
			}
			helper.Logger("info", fmt.Sprintf("email konfirmasi pembayaran terkirim ke %s", email))
		}
	}

	title, _ := helper.GetTitleByProjectId(dbDefault, ps.Id)
	email, _ := helper.GetEmailByProjectId(dbDefault, ps.Id)
	role, _ := helper.GetRoleByProjectId(dbDefault, ps.Id)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s konfirmasi dokumen transaksi payment project %s untuk pada waktu %s",
			ip,
			email,
			role,
			title,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{
		"message":          "project_payments success",
		"project_payments": projectPaymentId,
	}, nil

}

func DocumentVerifyProject(r *http.Request, dvp *entities.DocumentVerifyProject) (map[string]any, error) {
	if dvp == nil {
		return nil, fmt.Errorf("nil DocumentVerifyProject payload")
	}

	dvp.Skd = strings.TrimSpace(dvp.Skd)
	dvp.Cv = strings.TrimSpace(dvp.Cv)
	dvp.Rab = strings.TrimSpace(dvp.Rab)
	dvp.DokumenPerizinanLainnya = strings.TrimSpace(dvp.DokumenPerizinanLainnya)
	dvp.VideoProfilPerusahaan = strings.TrimSpace(dvp.VideoProfilPerusahaan)
	dvp.ProjectSummary = strings.TrimSpace(dvp.ProjectSummary)
	dvp.ProjectPendapatan = strings.TrimSpace(dvp.ProjectPendapatan)
	dvp.TimelinePekerjaan = strings.TrimSpace(dvp.TimelinePekerjaan)
	dvp.LaporanPajakTahunan = strings.TrimSpace(dvp.LaporanPajakTahunan)
	dvp.DaftarPekerjaan = strings.TrimSpace(dvp.DaftarPekerjaan)
	dvp.DaftarSupplier = strings.TrimSpace(dvp.DaftarSupplier)
	dvp.DaftarPiutang = strings.TrimSpace(dvp.DaftarPiutang)
	dvp.CashflowProject = strings.TrimSpace(dvp.CashflowProject)
	dvp.ProjectId = strings.TrimSpace(dvp.ProjectId)

	query := `DELETE FROM inboxes WHERE id = ?`

	errInbox := dbDefault.Raw(query, dvp.InboxId).Exec(query).Error

	if errInbox != nil {
		helper.Logger("error", "In Server: "+errInbox.Error())
		return nil, errInbox
	}

	fields := map[string]string{
		"Skd":                     dvp.Skd,
		"Cv":                      dvp.Cv,
		"Rab":                     dvp.Rab,
		"DokumenPerizinanLainnya": dvp.DokumenPerizinanLainnya,
		"VideoProfileCompany":     dvp.VideoProfilPerusahaan,
		"ProjectSummary":          dvp.ProjectSummary,
		"ProjectPendapatan":       dvp.ProjectPendapatan,
		"TimelinePekerjaan":       dvp.TimelinePekerjaan,
		"LaporanPajakTahunan":     dvp.LaporanPajakTahunan,
		"DaftarPekerjaan":         dvp.DaftarPekerjaan,
		"DaftarSupplier":          dvp.DaftarSupplier,
		"DaftarPiutang":           dvp.DaftarPiutang,
		"CashflowProject":         dvp.CashflowProject,
		"ProjectId":               dvp.ProjectId,
	}
	var missing []string
	for name, value := range fields {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("validation failed: missing/empty fields: %s", strings.Join(missing, ", "))
	}

	var docVerifyID uint64

	if err := dbDefault.Transaction(func(tx *gorm.DB) error {
		const queryInsert = `
			INSERT INTO document_verify_projects 
				(skd, cv, rab, other_license_document, video_profile_company, project_summary, 
				 revenue_projection, work_of_timeline, annual_tax_report, list_of_employment, 
				 list_of_supplier_data, latest_receivable_account, cashflow_project, project_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		if err := tx.Exec(
			queryInsert,
			dvp.Skd,
			dvp.Cv,
			dvp.Rab,
			dvp.DokumenPerizinanLainnya,
			dvp.VideoProfilPerusahaan,
			dvp.ProjectSummary,
			dvp.ProjectPendapatan,
			dvp.TimelinePekerjaan,
			dvp.LaporanPajakTahunan,
			dvp.DaftarPekerjaan,
			dvp.DaftarSupplier,
			dvp.DaftarPiutang,
			dvp.CashflowProject,
			dvp.ProjectId,
		).Error; err != nil {
			helper.Logger("error", "insert document_verify_projects: "+err.Error())
			return fmt.Errorf("insert document_verify_projects failed: %w", err)
		}

		if err := tx.Raw("SELECT LAST_INSERT_ID()").Scan(&docVerifyID).Error; err != nil {
			helper.Logger("error", "LAST_INSERT_ID: "+err.Error())
			return fmt.Errorf("could not fetch inserted id: %w", err)
		}

		const queryInsertDocMediaProject = `
			INSERT INTO media_document_verify_projects (path, document_verify_project_id, type) 
			VALUES (?, ?, ?)
		`

		for _, raw := range dvp.FotoKaryawanKantor {
			p := strings.TrimSpace(raw.Path)
			if p == "" {
				continue
			}
			if err := tx.Exec(queryInsertDocMediaProject, p, docVerifyID, "foto_karyawan_kantor").Error; err != nil {
				helper.Logger("error", "insert media (karyawan_kantor): "+err.Error())
				return fmt.Errorf("failed to insert media_document_verify_projects (path=%q): %w", p, err)
			}
		}
		for _, raw := range dvp.FotoKegiatanUsaha {
			p := strings.TrimSpace(raw.Path)
			if p == "" {
				continue
			}
			if err := tx.Exec(queryInsertDocMediaProject, p, docVerifyID, "foto_kegiatan_usaha").Error; err != nil {
				helper.Logger("error", "insert media (kegiatan_usaha): "+err.Error())
				return fmt.Errorf("failed to insert media_document_verify_projects (path=%q): %w", p, err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	title, _ := helper.GetTitleByProjectId(dbDefault, dvp.ProjectId)
	email, _ := helper.GetEmailByProjectId(dbDefault, dvp.ProjectId)
	role, _ := helper.GetRoleByProjectId(dbDefault, dvp.ProjectId)
	projectId, _ := helper.GetSkuByProjectId(dbDefault, dvp.ProjectId)

	subject := fmt.Sprintf(`Dokumen Pelengkap "%s" Siap Ditinjau`, title)

	emailProjectPublish, _ := helper.GetEmailsByRoleProjectPublish(dbDefault)

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
								<td style="padding:24px;">
									<p style="margin:0 0 12px 0;color:#111827;font-size:14px;">ID Proyek : %[2]s</p>
									<p style="margin:0 0 12px 0;color:#111827;font-size:14px;">Judul Proyek : %[3]s</p>

									<p style="margin:16px 0 6px 0;color:#374151;font-size:14px;">
										Mohon kepada Admin Tim Publish untuk memastikan bahwa seluruh dokumen yang telah diterima lengkap, sesuai, dan valid. 
										Setelah proses pengecekan selesai, silakan segera mengisi Form Pra-Publish Proyek agar tahapan publikasi dapat dilanjutkan sesuai prosedur.
									</p>
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
		projectId,
		title,
		time.Now().Year(),
	)

	for _, email := range emailProjectPublish {
		if err := helper.SendEmail(email, "Fulusme", subject, htmlBody, "another-otp"); err != nil {
			helper.Logger("error", fmt.Sprintf("gagal kirim email ke %s: %v", email, err))
			continue
		}
		helper.Logger("info", fmt.Sprintf("email konfirmasi pembayaran terkirim ke %s", email))
	}

	docs := []string{}

	if dvp.Cv != "" {
		docs = append(docs, dvp.Cv)
	}
	if dvp.Rab != "" {
		docs = append(docs, dvp.Rab)
	}
	if dvp.DokumenPerizinanLainnya != "" {
		docs = append(docs, dvp.DokumenPerizinanLainnya)
	}
	if dvp.VideoProfilPerusahaan != "" {
		docs = append(docs, dvp.VideoProfilPerusahaan)
	}
	if dvp.ProjectSummary != "" {
		docs = append(docs, dvp.ProjectSummary)
	}
	if dvp.ProjectPendapatan != "" {
		docs = append(docs, dvp.ProjectPendapatan)
	}
	if dvp.TimelinePekerjaan != "" {
		docs = append(docs, dvp.TimelinePekerjaan)
	}
	if dvp.LaporanPajakTahunan != "" {
		docs = append(docs, dvp.LaporanPajakTahunan)
	}
	if dvp.DaftarPekerjaan != "" {
		docs = append(docs, dvp.DaftarPekerjaan)
	}
	if dvp.DaftarSupplier != "" {
		docs = append(docs, dvp.DaftarSupplier)
	}
	if dvp.DaftarPiutang != "" {
		docs = append(docs, dvp.DaftarPiutang)
	}
	if dvp.CashflowProject != "" {
		docs = append(docs, dvp.CashflowProject)
	}

	if dvp.Skd != "" {
		docs = append(docs, dvp.Skd)
	}

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s konfirmasi dokumen transaksi pembayaran project %s pada %s",
			ip,
			email,
			role,
			title,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{
		"message":                    "document verify project success",
		"document_verify_project_id": docVerifyID,
	}, nil
}

func DocumentMediaVerifyProject(r *http.Request, dvpm *entities.DocumentMediaVerifyProject) (map[string]any, error) {
	if dvpm == nil {
		return nil, fmt.Errorf("nil document media verify project payload")
	}

	dvpm.DocumentVerifyProjectId = strings.TrimSpace(dvpm.DocumentVerifyProjectId)
	dvpm.Path = strings.TrimSpace(dvpm.Path)
	dvpm.Type = strings.TrimSpace(dvpm.Type)

	fields := map[string]string{
		"DocumentVerifyProjectId": dvpm.DocumentVerifyProjectId,
		"Path":                    dvpm.Path,
		"Type":                    dvpm.Type,
	}

	var missing []string
	for name, value := range fields {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("validation failed: missing/empty fields: %s", strings.Join(missing, ", "))
	}

	queryInsert := `
		INSERT INTO media_document_verify_projects
			(document_verify_project_id, path, type)
		VALUES
			(?, ?, ?)`

	errInsert := dbDefault.Exec(
		queryInsert,
		dvpm.DocumentVerifyProjectId,
		dvpm.Path,
		dvpm.Type,
	).Error
	if errInsert != nil {
		helper.Logger("error", "In Server: "+errInsert.Error())
		return nil, errors.New(errInsert.Error())
	}

	email, title, role, _ := helper.GetEmailAndTitleAndRoleByMediaProject(dbDefault, dvpm.DocumentVerifyProjectId)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s memperbarui project %s untuk media %s pada waktu %s",
			ip,
			email,
			role,
			title,
			dvpm.Type,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{"message": "document media verify project success"}, nil
}

func DocumentUpdate(r *http.Request, d *entities.DocumentUpdate) (map[string]any, error) {
	docType := strings.ToLower(strings.TrimSpace(d.Type))

	aliases := map[string]string{
		"nama-perusaahaan":             "nama-perusahaan",
		"npwp-perusahaan":              "npwp-perusahaan",
		"upload-ktp-pic":               "photo_ktp",
		"sk-kumham-path":               "sk-kumham-path",
		"sk-pendirian-perusahaan":      "sk-pendirian-perusahaan",
		"surat_kuasa":                  "surat-kuasa",
		"akta-perubahan-terakhir":      "akta-perubahan-terakhir",
		"akta_pendirian":               "akta-pendirian-perusahaan",
		"akta_perubahan_terahkir_path": "akta-perubahan-terakhir-path",
	}
	if v, ok := aliases[docType]; ok {
		docType = v
	}

	exec := func(tx *gorm.DB, query string, args ...any) error {
		if err := tx.Exec(query, args...).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return err
		}
		return nil
	}

	var (
		notifyEmail string
		notifyName  string
		notifyRole  string
		needNotify  bool
	)

	email, _ := helper.GetEmailInboxByUID(dbDefault, d.UserId)
	fullname, _ := helper.GetFullnameByUID(dbDefault, d.UserId)
	role, _ := helper.GetRoleByUID(dbDefault, d.UserId)

	notifyEmail = email
	notifyName = fullname
	notifyRole = role
	needNotify = true

	type action struct {
		label string
		query string
		args  func(*entities.DocumentUpdate) ([]any, error)
	}

	actions := map[string]action{
		"nama-perusahaan": {
			label: "Nama Perusahaan",
			query: `UPDATE companies SET company_name = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"npwp-perusahaan": {
			label: "Npwp Perusahaan",
			query: `UPDATE companies SET npwp_path = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"surat-kuasa": {
			label: "Surat Kuasa",
			query: `UPDATE additional_docs SET path = ? WHERE user_id = ? AND type = 'surat-kuasa'`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.UserId == "" {
					return nil, errors.New("userId is required")
				}
				return []any{d.Val, d.UserId}, nil
			},
		},
		"photo_ktp": {
			label: "Photo KTP",
			query: `UPDATE profiles SET photo_ktp = ? WHERE user_id = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.UserId == "" {
					return nil, errors.New("userId is required")
				}
				return []any{d.Val, d.UserId}, nil
			},
		},
		"slip-gaji": {
			label: "Slip Gaji",
			query: `UPDATE pay_slips SET path = ? WHERE user_id = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.UserId == "" {
					return nil, errors.New("userId is required")
				}
				return []any{d.Val, d.UserId}, nil
			},
		},
		"sk-kumham-pendirian": {
			label: "SK Kumham Pendirian",
			query: `UPDATE companies SET sk_kumham_path = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"sk-kumham-path": {
			label: "SK Kumham",
			query: `UPDATE companies SET sk_kumham_path = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"npwp": {
			label: "NPWP",
			query: `UPDATE companies SET npwp = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"npwp_path": {
			label: "NPWP",
			query: `UPDATE companies SET npwp_path = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"nib": {
			label: "NIB",
			query: `UPDATE companies SET company_nib_path = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"akta-pendirian-perusahaan": {
			label: "Akta Pendirian Perusahaan",
			query: `UPDATE companies SET deed_of_incorporation = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"alamat-perusahaan": {
			label: "Alamat Perusahaan",
			query: `UPDATE companies SET address = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"siup": {
			label: "SIUP",
			query: `UPDATE companies SET siup = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"tdp": {
			label: "TDP",
			query: `UPDATE companies SET tdp = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"alamat-korespondensi": {
			label: "Alamat Korespondensi",
			query: `UPDATE companies SET address = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"jumlah-karyawan": {
			label: "Jumlah Karyawan",
			query: `UPDATE companies SET total_employees = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"laporan-keuangan": {
			label: "Laporan Keuangan",
			query: `UPDATE companies SET financial_statement = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"rekening-koran": {
			label: "Rekening Koran",
			query: `UPDATE companies SET bank_statement = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"title-proyek": {
			label: "Judul Proyek",
			query: `UPDATE projects SET title = ? WHERE company_id = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"foto-proyek": {
			label: "Foto Proyek",
			query: `UPDATE projects SET project_image_path = ? WHERE company_id = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"nilai-nominal": {
			label: "Nilai Nominal",
			query: `UPDATE projects SET nominal_value = ? WHERE company_id = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"sk-kumham-terakhir": {
			label: "SK Kumham Terahkir",
			query: `UPDATE companies SET sk_kumham_last = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"sk_kumham_path": {
			label: "SK Kumham",
			query: `UPDATE companies SET sk_kumham_path = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"akta-perubahan-terakhir": {
			label: "Akta Perubahan Terakhir",
			query: `UPDATE companies SET latest_amendment_deed = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"akta-perubahan-terakhir-path": {
			label: "Akta Perubahan Terakhir",
			query: `UPDATE companies SET latest_amendment_deed_path = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"company-profile": {
			label: "Company Profile",
			query: `UPDATE projects SET company_profile = ? WHERE company_id = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
		"doc-kontrak": {
			label: "Dokumen Kontrak",
			query: `UPDATE project_contracts SET path = ? WHERE project_id = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.ProjectId == "" {
					return nil, errors.New("projectId is required")
				}
				return []any{d.Val, d.ProjectId}, nil
			},
		},
		"sk-pendirian-perusahaan": {
			label: "SK Pendirian Perusahaan",
			query: `UPDATE companies SET certificate_of_company_est = ? WHERE uid = ?`,
			args: func(d *entities.DocumentUpdate) ([]any, error) {
				if d.CompanyId == "" {
					return nil, errors.New("companyId is required")
				}
				return []any{d.Val, d.CompanyId}, nil
			},
		},
	}

	tx := dbDefault.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if d.InboxId != "" {
		if err := exec(tx, `DELETE FROM inboxes WHERE id = ?`, d.InboxId); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	roleLabel := notifyRole
	if strings.EqualFold(notifyRole, "emiten") {
		roleLabel = "Penerbit"
	} else {
		roleLabel = strings.ToTitle(notifyRole)
	}

	if docType == "update-direktur-komisaris" {
		if len(d.ValArray) > 0 {
			for _, v := range d.ValArray {
				switch strings.ToLower(strings.TrimSpace(v.Type)) {
				case "ktp":
					if err := exec(tx, `UPDATE positions SET ktp_path = ? WHERE id = ?`, v.Val, v.Id); err != nil {
						tx.Rollback()
						return nil, err
					}
					if needNotify {
						body := fmt.Sprintf(
							"Kami informasikan bahwa dokumen %s telah diperbaharui oleh %s ( %s ).\n\nSilahkan meninjau dokumen terbaru melalui tautan berikut %s\n\nTerima kasih.",
							"KTP", notifyName, roleLabel, v.Val,
						)
						if err := helper.SendEmail(notifyEmail, "Fulusme", "Pembaharuan Dokumen", body, "another-otp"); err != nil {
							helper.Logger("error", "Failed to send email: "+err.Error())
						}
					}
				case "npwp":
					if err := exec(tx, `UPDATE positions SET npwp_path = ? WHERE id = ?`, v.Val, v.Id); err != nil {
						tx.Rollback()
						return nil, err
					}

					if needNotify {
						body := fmt.Sprintf(
							"Kami informasikan bahwa dokumen %s telah diperbaharui oleh %s ( %s ).\n\nSilahkan meninjau dokumen terbaru melalui tautan berikut %s\n\nTerima kasih.",
							"NPWP", notifyName, roleLabel, v.Val,
						)
						if err := helper.SendEmail(notifyEmail, "Fulusme", "Pembaharuan Dokumen", body, "another-otp"); err != nil {
							helper.Logger("error", "Failed to send email: "+err.Error())
						}
					}
				}
			}
		}

		if err := tx.Commit().Error; err != nil {
			return nil, err
		}

		return map[string]any{"message": "document update success"}, nil
	}

	act, ok := actions[docType]
	if !ok {
		tx.Rollback()
		return nil, errors.New("tipe dokumen tidak dikenali")
	}

	args, err := act.args(d)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := exec(tx, act.query, args...); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if needNotify {
		body := fmt.Sprintf(
			"Kami informasikan bahwa dokumen %s telah diperbaharui oleh %s ( %s ).\n\nSilahkan meninjau dokumen terbaru melalui tautan berikut %s\n\nTerima kasih.",
			act.label, notifyName, roleLabel, d.Val,
		)
		if err := helper.SendEmail(notifyEmail, "Fulusme", "Pembaharuan Dokumen", body, "another-otp"); err != nil {
			helper.Logger("error", "Failed to send email: "+err.Error())
		}
	}

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s memperbarui data dokumen %s pada waktu %s",
			ip,
			email,
			role,
			act.label,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{"message": "document update success"}, nil
}

func UpdateValUser(r *http.Request, d *entities.UpdateValUser) (map[string]any, error) {
	switch d.Type {
	case "npwp_path":
		queryUpdate := `UPDATE jobs SET npwp_path = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	case "ktp":
		queryUpdate := `UPDATE ktps SET nik = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	case "ktp_path":
		queryUpdate := `UPDATE ktps SET path = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	case "slip_gaji":
		queryUpdate := `UPDATE pay_slips SET path = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	case "tujuan":
		queryUpdate := `UPDATE user_risks SET goal = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	case "toleransi":
		queryUpdate := `UPDATE user_risks SET tolerance = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	case "sk-kumham-pendirian":
		queryUpdate := `UPDATE companies SET sk_kumham = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	case "npwp":
		queryUpdate := `UPDATE companies SET npwp = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	case "pengalaman":
		queryUpdate := `UPDATE user_risks SET experience = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	case "pengetahuan_pasar_modal":
		queryUpdate := `UPDATE user_risks SET capital_market_knowledge = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	case "akta_pendirian":
		queryUpdate := `UPDATE companies SET deed_of_incorporation = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	case "akta_perubahan_terahkir_path":
		queryUpdate := `UPDATE companies SET latest_amendment_deed_path = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	case "sk_kumham_path":
		queryUpdate := `UPDATE companies SET sk_kumham_path = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	case "sk_pendirian_perusahaan":
		queryUpdate := `UPDATE companies SET certificate_of_company_est = ? WHERE user_id = ?`
		errUpdate := dbDefault.Exec(queryUpdate, d.Val, d.UserId).Error
		if errUpdate != nil {
			helper.Logger("error", "In Server: "+errUpdate.Error())
			return nil, errors.New(errUpdate.Error())
		}
	default:
		return nil, errors.New("tipe dokumen tidak dikenali")

	}

	query := `DELETE FROM inboxes 
	WHERE receiver_id = ?`

	errInbox := dbDefault.Raw(query, d.UserId).Exec(query).Error

	if errInbox != nil {
		helper.Logger("error", "In Server: "+errInbox.Error())
		return nil, errInbox
	}

	email, _ := helper.GetEmailByUID(dbDefault, d.UserId)
	role, _ := helper.GetRoleByUID(dbDefault, d.UserId)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s memperbarui data dokumen %s pada waktu %s",
			ip,
			email,
			role,
			d.Type,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{"message": "update value user success"}, nil
}

func safe(val, fallback string) string {
	v := strings.TrimSpace(val)
	if v == "" {
		return fallback
	}
	return v
}
