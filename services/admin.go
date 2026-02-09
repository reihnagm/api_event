package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"superapps/entities"
	helper "superapps/helpers"
	"time"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

func AdminUpdateProject(r *http.Request, aup *entities.AdminUpdateProject) (entities.AdminUpdateProjectResponse, error) {
	tx := dbDefault.Begin()
	if tx.Error != nil {
		helper.Logger("error", "Failed to start transaction: "+tx.Error.Error())
		return entities.AdminUpdateProjectResponse{}, tx.Error
	}
	defer func() {
		if rr := recover(); rr != nil {
			tx.Rollback()
			panic(rr)
		}
	}()

	sets := make([]string, 0, 12)
	args := make([]any, 0, 12)

	// NOTE: update hanya jika "terisi"
	if strings.TrimSpace(aup.KodeEfek) != "" {
		sets = append(sets, "code_effect = ?")
		args = append(args, strings.TrimSpace(aup.KodeEfek))
	}
	if aup.Tenor != "" {
		sets = append(sets, "loan_term = ?")
		args = append(args, aup.Tenor)
	}
	if aup.HargaUnit != "" {
		sets = append(sets, "unit_price = ?")
		args = append(args, aup.HargaUnit)
	}
	if aup.TotalUnit != "" {
		sets = append(sets, "capital = ?")
		args = append(args, aup.TotalUnit)
	}
	if aup.JumlahUnit != "" {
		sets = append(sets, "number_of_unit = ?")
		args = append(args, aup.JumlahUnit)
	}
	if aup.ProfitPercentage != "" {
		sets = append(sets, "profit_percentage = ?")
		args = append(args, aup.ProfitPercentage)
	}
	if strings.TrimSpace(aup.Prospoectus) != "" {
		sets = append(sets, "doc_prospect = ?")
		args = append(args, strings.TrimSpace(aup.Prospoectus))
	}
	if aup.MinInvest != "" {
		sets = append(sets, "min_invest = ?")
		args = append(args, aup.MinInvest)
	}
	if aup.AmountSharesPerLot != "" {
		sets = append(sets, "amount_shares_per_lot = ?")
		args = append(args, aup.AmountSharesPerLot)
	}

	// kalau tidak ada field yang mau diupdate
	if len(sets) == 0 {
		tx.Rollback()
		return entities.AdminUpdateProjectResponse{}, errors.New("no fields to update")
	}

	// WHERE uid = ?
	query := fmt.Sprintf("UPDATE projects SET %s WHERE uid = ?", strings.Join(sets, ", "))
	args = append(args, aup.ProjectId)

	res := tx.Exec(query, args...)
	if res.Error != nil {
		tx.Rollback()
		helper.Logger("error", "Failed to update project: "+res.Error.Error())
		return entities.AdminUpdateProjectResponse{}, res.Error
	}
	if res.RowsAffected == 0 {
		tx.Rollback()
		err := errors.New("no project found with the given uid")
		helper.Logger("warn", err.Error())
		return entities.AdminUpdateProjectResponse{}, err
	}

	if err := tx.Commit().Error; err != nil {
		helper.Logger("error", "Failed to commit transaction: "+err.Error())
		return entities.AdminUpdateProjectResponse{}, err
	}

	// logging tetap
	email, _ := helper.GetEmailByProjectId(dbDefault, aup.ProjectId)
	role, _ := helper.GetRoleByProjectId(dbDefault, aup.ProjectId)
	title, _ := helper.GetTitleByProjectId(dbDefault, aup.ProjectId)

	ip := helper.GetClientIP(r)
	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s memperbarui data project %s pada waktu %s",
			ip,
			email,
			role,
			title,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	// response: balikin payload yang dikirim (sesuai pola kamu sekarang)
	return entities.AdminUpdateProjectResponse{
		KodeEfek:           aup.KodeEfek,
		JumlahUnit:         aup.JumlahUnit,
		HargaUnit:          aup.HargaUnit,
		TotalUnit:          aup.TotalUnit,
		Tenor:              aup.Tenor,
		Prospoectus:        aup.Prospoectus,
		ProjectId:          aup.ProjectId,
		MinInvest:          aup.MinInvest,
		ProfitPercentage:   aup.ProfitPercentage,
		AmountSharesPerLot: aup.AmountSharesPerLot,
	}, nil
}

func AdminGetProfile(userID string) (entities.AdminGetResponseProfile, error) {
	var profile entities.AdminGetProfile

	const query = `
		SELECT
			p.user_id AS id,
			p.fullname,
			u.phone, 
			u.email,
			p.avatar
		FROM users u
		LEFT JOIN profiles p ON p.user_id = u.uid
		WHERE u.uid = ?
		LIMIT 1;
	`

	tx := dbDefault.Raw(query, userID).Scan(&profile)
	if tx.Error != nil {
		helper.Logger("error", "AdminGetProfile: "+tx.Error.Error())
		return entities.AdminGetResponseProfile{}, tx.Error
	}
	if tx.RowsAffected == 0 {
		return entities.AdminGetResponseProfile{}, gorm.ErrRecordNotFound
	}

	return entities.AdminGetResponseProfile{
		ID:       profile.ID,
		Fullname: profile.Fullname,
		Email:    profile.Email,
		Phone:    profile.Phone,
		Avatar:   helper.DefaultIfEmpty(profile.Avatar, "-"),
	}, nil
}

func AdminUpdateProfile(r *http.Request, ctx context.Context, aup *entities.AdminUpdateProfile) (entities.AdminUpdateProfile, error) {
	if aup == nil {
		return entities.AdminUpdateProfile{}, errors.New("nil payload")
	}

	err := dbDefault.WithContext(ctx).Debug().Transaction(func(tx *gorm.DB) error {
		// --- Update profiles (selalu diizinkan) ---
		profileQuery := `
			UPDATE profiles
			SET fullname = ?, avatar = ?
			WHERE user_id = ?
		`
		resProfiles := tx.Exec(profileQuery, aup.Fullname, aup.Avatar, aup.UserId)
		if resProfiles.Error != nil {
			helper.Logger("error", "AdminUpdateProfile profiles: "+resProfiles.Error.Error())
			return resProfiles.Error
		}

		// --- Update users (opsional: hanya email/phone yang tidak kosong) ---
		var (
			setClauses []string
			args       []any
		)

		if strings.TrimSpace(aup.Email) != "" {
			setClauses = append(setClauses, "email = ?")
			args = append(args, aup.Email)
		}
		if strings.TrimSpace(aup.Phone) != "" {
			setClauses = append(setClauses, "phone = ?")
			args = append(args, aup.Phone)
		}

		var usersAffected int64 = 0
		if len(setClauses) > 0 {
			userQuery := fmt.Sprintf(`
				UPDATE users
				SET %s
				WHERE uid = ?
			`, strings.Join(setClauses, ", "))
			args = append(args, aup.UserId)

			resUsers := tx.Exec(userQuery, args...)
			if resUsers.Error != nil {
				helper.Logger("error", "AdminUpdateProfile users: "+resUsers.Error.Error())
				return resUsers.Error
			}
			usersAffected = resUsers.RowsAffected
		}

		// Jika tidak ada baris yang berubah di KEDUA tabel, anggap data tidak ditemukan
		if resProfiles.RowsAffected+usersAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
	if err != nil {
		return entities.AdminUpdateProfile{}, err
	}

	email, _ := helper.GetEmailByUID(dbDefault, aup.UserId)
	role, _ := helper.GetRoleByUID(dbDefault, aup.UserId)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s memperbarui profil akun pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return entities.AdminUpdateProfile{
		Fullname: aup.Fullname,
		Avatar:   aup.Avatar,
		Email:    aup.Email, // mungkin kosong jika tidak diupdate
		Phone:    aup.Phone, // mungkin kosong jika tidak diupdate
		UserId:   aup.UserId,
	}, nil
}

func AdminCreateUser(r *http.Request, acu *entities.AdminCreateUser) (entities.AdminCreateUserResponse, error) {
	var roles []entities.CheckRole

	// Generate UUID for the new user
	acu.Id = uuid.NewV4().String()

	// Hash password
	hashedPassword, err := helper.Hash(acu.Password)
	if err != nil {
		helper.Logger("error", "Failed to hash password: "+err.Error())
		return entities.AdminCreateUserResponse{}, err
	}

	// Start transaction
	tx := dbDefault.Begin()
	if tx.Error != nil {
		helper.Logger("error", "Failed to start transaction: "+tx.Error.Error())
		return entities.AdminCreateUserResponse{}, tx.Error
	}

	// Insert into users
	if err := tx.Exec(
		`INSERT INTO users (uid, email, phone, password, role) VALUES (?, ?, ?, ?, ?)`,
		acu.Id, acu.Email, acu.Phone, hashedPassword, acu.Role,
	).Error; err != nil {
		tx.Rollback()
		helper.Logger("error", "Failed to insert user: "+err.Error())
		return entities.AdminCreateUserResponse{}, err
	}

	// Insert into profiles
	if err := tx.Exec(
		`INSERT INTO profiles (user_id, fullname) VALUES (?, ?)`,
		acu.Id, acu.Fullname,
	).Error; err != nil {
		tx.Rollback()
		helper.Logger("error", "Failed to insert profile: "+err.Error())
		return entities.AdminCreateUserResponse{}, err
	}

	// Validate role exists
	if err := tx.Raw(
		`SELECT id, name FROM roles WHERE id = ?`, acu.Role,
	).Scan(&roles).Error; err != nil {
		tx.Rollback()
		helper.Logger("error", "Failed to check role: "+err.Error())
		return entities.AdminCreateUserResponse{}, err
	}

	if len(roles) == 0 {
		tx.Rollback()
		helper.Logger("error", "ROLE_NOT_FOUND")
		return entities.AdminCreateUserResponse{}, errors.New("ROLE_NOT_FOUND")
	}

	if err := tx.Commit().Error; err != nil {
		helper.Logger("error", "Failed to commit transaction: "+err.Error())
		return entities.AdminCreateUserResponse{}, err
	}

	email, _ := helper.GetEmailByUID(dbDefault, acu.Id)
	role, _ := helper.GetRoleByUID(dbDefault, acu.Id)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s membuat profil akun pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	// Success response
	return entities.AdminCreateUserResponse{
		Id:    acu.Id,
		Email: acu.Email,
		Phone: acu.Phone,
		Role: entities.AdminRoleUser{
			Id:   roles[0].Id,
			Name: roles[0].Name,
		},
	}, nil
}

func AdminUpdateUser(r *http.Request, acu *entities.AdminUpdateUser) (entities.AdminUpdateUserResponse, error) {
	if acu.Id == "" {
		return entities.AdminUpdateUserResponse{}, errors.New("user id is required for update")
	}

	roles := []entities.CheckRole{}

	// --- check role exists ---
	queryCheckRole := `SELECT id, name FROM roles WHERE id = ?`
	errCheckRole := dbDefault.Raw(queryCheckRole, acu.Role).Scan(&roles).Error
	if errCheckRole != nil {
		helper.Logger("error", "In Server: "+errCheckRole.Error())
		return entities.AdminUpdateUserResponse{}, errors.New(errCheckRole.Error())
	}
	if len(roles) == 0 {
		helper.Logger("error", "In Server: ROLE_NOT_FOUND")
		return entities.AdminUpdateUserResponse{}, errors.New("ROLE_NOT_FOUND")
	}

	// --- start transaction ---
	tx := dbDefault.Begin()
	if tx.Error != nil {
		return entities.AdminUpdateUserResponse{}, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// --- update user ---
	if strings.TrimSpace(acu.Password) != "" {
		// update password only if provided
		var errHasshed error
		hashedPasswordBytes, errHasshed := helper.Hash(acu.Password)
		if errHasshed != nil {
			tx.Rollback()
			helper.Logger("error", "In Server: "+errHasshed.Error())
			return entities.AdminUpdateUserResponse{}, errors.New(errHasshed.Error())
		}
		hashedPassword := string(hashedPasswordBytes)

		queryUpdateUser := `
			UPDATE users 
			SET email = ?, phone = ?, password = ?, role = ?
			WHERE uid = ?
		`
		if err := tx.Debug().Exec(queryUpdateUser, acu.Email, acu.Phone, hashedPassword, acu.Role, acu.Id).Error; err != nil {
			tx.Rollback()
			helper.Logger("error", "In Server: "+err.Error())
			return entities.AdminUpdateUserResponse{}, errors.New(err.Error())
		}
	} else {
		// no password update
		queryUpdateUser := `
			UPDATE users 
			SET email = ?, phone = ?, role = ?
			WHERE uid = ?
		`
		if err := tx.Debug().Exec(queryUpdateUser, acu.Email, acu.Phone, acu.Role, acu.Id).Error; err != nil {
			tx.Rollback()
			helper.Logger("error", "In Server: "+err.Error())
			return entities.AdminUpdateUserResponse{}, errors.New(err.Error())
		}
	}

	// --- update profile ---
	queryUpdateProfile := `UPDATE profiles SET fullname = ? WHERE user_id = ?`
	if err := tx.Debug().Exec(queryUpdateProfile, acu.Fullname, acu.Id).Error; err != nil {
		tx.Rollback()
		helper.Logger("error", "In Server: "+err.Error())
		return entities.AdminUpdateUserResponse{}, errors.New(err.Error())
	}

	// --- commit ---
	if err := tx.Commit().Error; err != nil {
		helper.Logger("error", "In Server commit: "+err.Error())
		return entities.AdminUpdateUserResponse{}, fmt.Errorf("transaction commit failed: %w", err)
	}

	email, _ := helper.GetEmailByUID(dbDefault, acu.Id)
	role, _ := helper.GetRoleByUID(dbDefault, acu.Id)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s memperbarui profil akun pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	// --- response ---
	return entities.AdminUpdateUserResponse{
		Id:    acu.Id,
		Email: acu.Email,
		Phone: acu.Phone,
		Role: entities.AdminRoleUser{
			Id:   roles[0].Id,
			Name: roles[0].Name,
		},
	}, nil
}

func AdminAssignRole(r *http.Request, asr *entities.AdminAssignRole) (map[string]any, error) {

	RoleId := strings.TrimSpace(asr.RoleId)
	UserId := strings.TrimSpace(asr.UserId)

	if RoleId == "" {
		helper.Logger("error", "In Server : role_id is required")
		return nil, errors.New("role_id is required")
	}

	if UserId == "" {
		helper.Logger("error", "In Server : user_id is required")
		return nil, errors.New("user_id is required")
	}

	query := `UPDATE users SET role = ? 
	WHERE uid = ?`

	err := dbDefault.Exec(query, asr.RoleId, asr.UserId).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	role, _ := helper.GetRoleByUID(dbDefault, asr.UserId)
	email, _ := helper.GetEmailByUID(dbDefault, asr.UserId)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] assign role [%s] pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{
		"data": "Successfully Admin Assign Role",
	}, nil
}

func AdminRevokeRole(r *http.Request, arr *entities.AdminRevokeRole) (map[string]any, error) {

	UserId := strings.TrimSpace(arr.UserId)

	if UserId == "" {
		helper.Logger("error", "In Server : user_id is required")
		return nil, errors.New("user_id is required")
	}

	query := `UPDATE users SET role = ? 
	WHERE uid = ?`

	err := dbDefault.Exec(query, 4, arr.UserId).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	role, err := helper.GetRoleByUID(dbDefault, arr.UserId)
	if err != nil || role == "" {
		if err != nil {
			helper.Logger("error", "get role by uid: "+err.Error())
		}
	}

	email, err := helper.GetEmailByUID(dbDefault, arr.UserId)
	if err != nil || email == "" {
		if err != nil {
			helper.Logger("error", "get email by uid: "+err.Error())
		}
	}

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] revoke role [%s] pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{
		"data": "Successfully Admin Revoke Role",
	}, nil
}

func AdminListRole(page, limit string) (map[string]any, error) {
	url := os.Getenv("API_URL_PROD")

	var dataAdminListRole []entities.AdminListRole
	var totalAdminListRole []entities.AdminListRole
	var adminListRole entities.AdminListRole

	pageinteger, _ := strconv.Atoi(page)
	limitinteger, _ := strconv.Atoi(limit)

	var offset = strconv.Itoa((pageinteger - 1) * limitinteger)

	errAllRole := dbDefault.Raw(`SELECT id FROM roles`).Scan(&totalAdminListRole).Error

	if errAllRole != nil {
		helper.Logger("error", "In Server: "+errAllRole.Error())
		return nil, errors.New(errAllRole.Error())
	}

	var resultTotal = len(totalAdminListRole)

	var perPage = math.Ceil(float64(resultTotal) / float64(limitinteger))

	var prevPage int
	var nextPage int

	if pageinteger == 1 {
		prevPage = 1
	} else {
		prevPage = pageinteger - 1
	}

	nextPage = pageinteger + 1

	query := `SELECT id, name FROM roles
	ORDER BY created_at DESC
	LIMIT ?, ?`

	rows, err := dbDefault.Raw(query, offset, limit).Rows()

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	var nextUrl = strconv.Itoa(nextPage)
	var prevUrl = strconv.Itoa(prevPage)

	for rows.Next() {

		errAdminListRoleRows := dbDefault.ScanRows(rows, &adminListRole)
		if errAdminListRoleRows != nil {
			helper.Logger("error", "In Server: "+errAdminListRoleRows.Error())
			return nil, errors.New(errAdminListRoleRows.Error())
		}

		dataAdminListRole = append(dataAdminListRole, entities.AdminListRole{
			Id:   adminListRole.Id,
			Name: adminListRole.Name,
		})

	}

	return map[string]any{
		"total":        resultTotal,
		"current_page": pageinteger,
		"per_page":     int(perPage),
		"prev_page":    prevPage,
		"next_page":    nextPage,
		"next_url":     url + "/api/v1/admin/list/role?page=" + nextUrl,
		"prev_url":     url + "/api/v1/admin/list/role?page=" + prevUrl,
		"data":         dataAdminListRole,
	}, nil
}

func AdminListUser(page, limit string) (map[string]any, error) {
	url := os.Getenv("API_URL_PROD")

	var dataAdminListUser []entities.AdminListUserResponse
	var totalAdminListUser []entities.AdminListUser
	var adminListUser entities.AdminListUser

	pageinteger, _ := strconv.Atoi(page)
	limitinteger, _ := strconv.Atoi(limit)

	var offset = strconv.Itoa((pageinteger - 1) * limitinteger)

	errAllUser := dbDefault.Raw(`SELECT id FROM users`).Scan(&totalAdminListUser).Error

	if errAllUser != nil {
		helper.Logger("error", "In Server: "+errAllUser.Error())
	}

	var resultTotal = len(totalAdminListUser)

	var perPage = math.Ceil(float64(resultTotal) / float64(limitinteger))

	var prevPage int
	var nextPage int

	if pageinteger == 1 {
		prevPage = 1
	} else {
		prevPage = pageinteger - 1
	}

	nextPage = pageinteger + 1

	query := `SELECT u.uid AS id, p.fullname, p.avatar, u.email, u.phone,
	u.sku,
	p.selfie, p.photo_ktp, p.no_ktp, p.no_npwp, p.position, p.gender, 
	p.province_name, p.city_name, p.district_name, p.subdistrict_name, 
	p.status_marital, p.occupation, p.last_education, p.address_detail,
	p.beneficiary_name, p.beneficiary_phone, r.name AS role, 
	u.created_at, u.updated_at, u.verify, u.verify_emiten, u.verify_investor
	FROM users u 
	INNER JOIN roles r ON r.id = u.role
	INNER JOIN profiles p ON u.uid = p.user_id
	ORDER BY u.created_at DESC
	LIMIT ?, ?`

	rows, err := dbDefault.Raw(query, offset, limit).Rows()

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	for rows.Next() {
		errAdminListUserRows := dbDefault.ScanRows(rows, &adminListUser)
		if errAdminListUserRows != nil {
			helper.Logger("error", "In Server: "+errAdminListUserRows.Error())
			return nil, errors.New(errAdminListUserRows.Error())
		}

		var adminUserKtp entities.AdminUserKtp
		var adminUserJob entities.AdminUserJob
		var adminUserBank entities.AdminUserBank
		var adminAdditionalDoc entities.AdditionalDoc
		var adminUserSignature entities.AdminUserSignature
		var adminUserSecurityAccount entities.AdminUserSecurityAccount
		var adminUserRisk entities.AdminUserRisk

		var adminUserSlipPay entities.AdminUserSlipPay

		var adminEmitenCompany entities.AdminEmitenCompany
		var adminEmitenPositions []entities.AdminEmitenPositionCompany
		var adminEmitenProjects []entities.AdminEmitenProjectCompany
		var adminEmitenAddress []entities.AdminEmitenAddressCompany

		dataAdminEmitenProjects := make([]entities.AdminEmitenProjectCompanyResponse, 0)

		// Ktp

		queryKtp := `SELECT name, nik, place_datebirth, path FROM ktps WHERE user_id = ?`

		errKtp := dbDefault.Raw(queryKtp, adminListUser.Id).Scan(&adminUserKtp).Error
		if errKtp != nil {
			helper.Logger("error", "In Server: "+errKtp.Error())
			return nil, errors.New(errKtp.Error())
		}

		// Job

		queryJob := `SELECT company_name, province_name, city_name, district_name, subdistrict_name, company_address, monthly_income, annual_income_num AS annual_income, 
		position, npwp, npwp_path 
		FROM jobs WHERE user_id = ?`

		errJob := dbDefault.Raw(queryJob, adminListUser.Id).Scan(&adminUserJob).Error
		if errJob != nil {
			helper.Logger("error", "In Server: "+errJob.Error())
			return nil, errors.New(errJob.Error())
		}

		// Bank

		queryBank := `SELECT no, bank_name AS name, bank_owner AS owner, bank_branch AS branch, rek_koran_path FROM accounts WHERE user_id = ?`

		errBank := dbDefault.Raw(queryBank, adminListUser.Id).Scan(&adminUserBank).Error
		if errBank != nil {
			helper.Logger("error", "In Server: "+errBank.Error())
			return nil, errors.New(errBank.Error())
		}

		// Additional Doc

		queryAdditionalDoc := `SELECT type, path, user_id FROM additional_docs WHERE user_id = ?`

		errAdditionalDoc := dbDefault.Raw(queryAdditionalDoc, adminListUser.Id).Scan(&adminAdditionalDoc).Error
		if errAdditionalDoc != nil {
			helper.Logger("error", "In Server: "+errAdditionalDoc.Error())
			return nil, errors.New(errAdditionalDoc.Error())
		}

		// Company

		queryEmitenCompany := `SELECT c.uid AS id, c.company_name AS name, c.company_nib AS nib, c.site, c.email, c.phone, c.est, 
		toc.name AS jenis_perusahaan,
		tob.name AS jenis_usaha,
		tocp.name AS status_kantor,
		c.bank_name, c.bank_account, c.bank_owner_company, c.siup, c.tdp, c.company_nib_path AS nib_path, c.deed_of_incorporation AS akta_pendirian, 
		c.latest_amendment_deed AS akta_perubahan_terahkir, 
		c.latest_amendment_deed_path AS akta_perubahan_terahkir_path,
		c.certificate_of_company_est AS sk_pendirian_perusahaan,
		c.sk_kumham, c.sk_kumham_last, c.sk_kumham_path, c.npwp, c.npwp_path, c.total_employees, 
		c.financial_statement AS laporan_keuangan, c.bank_statement AS rekening_koran
 		FROM companies c 
		INNER JOIN type_of_companies toc ON toc.id = c.type_of_company
		INNER JOIN type_of_businesses tob ON tob.id = c.type_of_business
		INNER JOIN type_of_company_places tocp ON tocp.id = c.type_of_company_place
		WHERE user_id = ?`

		errEmitenCompany := dbDefault.Raw(queryEmitenCompany, adminListUser.Id).Scan(&adminEmitenCompany).Error
		if errEmitenCompany != nil {
			helper.Logger("error", "In Server: "+errEmitenCompany.Error())
			return nil, errors.New(errEmitenCompany.Error())
		}

		// Company Position

		queryEmitenPositionCompany := `SELECT title, name, position, ktp, ktp_path, npwp, npwp_path FROM positions WHERE company_id = ?`

		errEmitenPositionCompany := dbDefault.Raw(queryEmitenPositionCompany, adminEmitenCompany.Id).Scan(&adminEmitenPositions).Error

		if errEmitenPositionCompany != nil {
			helper.Logger("error", "In Server: "+errEmitenPositionCompany.Error())
			return nil, errors.New(errEmitenPositionCompany.Error())
		}

		// Company Project

		queryEmitenProjectCompany := `
			SELECT 
				p.uid AS id, 
				p.title, 
				top.name AS jenis_project, 
				p.is_apbn, 
				p.is_approved, 
				p.nominal_value AS jumlah_minimal,
				p.time_periode AS jangka_waktu, 
				p.interest_rate AS tingkat_bunga, 
				p.interest_payment_schedule AS jadwal_pembayaran_bunga, 
				p.principal_payment_schedule AS jadwal_pembayaran_pokok, 
				p.desc_job AS deskripsi_pekerjaan,
				p.company_profile, 
				p.collateral_guarantee AS jaminan_kolateral, 
				ps.name AS status,
				p.sku,
				p.start_project, p.end_project,
				p.provider_address, p.provider_province_name, p.provider_city_name, p.provider_district_name,
				p.provider_subdistrict_name, p.provider_postal_code
			FROM 
				projects p
			INNER JOIN 
				project_statuses ps ON p.status = ps.id
			INNER JOIN
				type_of_projects top ON top.id = p.type_of_project
			WHERE 
				p.company_id = ?
		`

		errEmitenProjectCompany := dbDefault.Raw(queryEmitenProjectCompany, adminEmitenCompany.Id).Scan(&adminEmitenProjects).Error

		if errEmitenProjectCompany != nil {
			helper.Logger("error", "In Server: "+errEmitenProjectCompany.Error())
			return nil, errors.New(errEmitenProjectCompany.Error())
		}

		// Company Address

		queryEmitenAddressCompany := `SELECT ca.name, ca.address, ca.postal_code,
		ca.province_name AS province, ca.city_name AS city, ca.district_name AS district, 
		ca.subdistrict_name AS subdistrict
		FROM company_addresses ca 
		WHERE ca.company_id = ?`

		errEmitenAddressCompany := dbDefault.Raw(queryEmitenAddressCompany, adminEmitenCompany.Id).Scan(&adminEmitenAddress).Error

		if errEmitenAddressCompany != nil {
			helper.Logger("error", "In Server: "+errEmitenAddressCompany.Error())
			return nil, errors.New(errEmitenAddressCompany.Error())
		}

		// User Signature

		queryUserSignature := `SELECT path FROM user_signatures 
		WHERE user_id = ?`

		errQueryUserSignature := dbDefault.Raw(queryUserSignature, adminListUser.Id).Scan(&adminUserSignature).Error

		if errQueryUserSignature != nil {
			helper.Logger("error", "In Server: "+errQueryUserSignature.Error())
			return nil, errors.New(errQueryUserSignature.Error())
		}

		// User Slip Pay

		querySlipPay := `SELECT path FROM pay_slips 
		WHERE user_id = ?`

		errSlipPay := dbDefault.Raw(querySlipPay, adminListUser.Id).Scan(&adminUserSlipPay).Error

		if errSlipPay != nil {
			helper.Logger("error", "In Server: "+errSlipPay.Error())
			return nil, errors.New(errSlipPay.Error())
		}

		// User Security Account

		querySecurityAccount := `SELECT account_name, account_no, account_sub_no, account_bank
		FROM security_accounts 
		WHERE user_id = ?`

		errSecurityAccount := dbDefault.Raw(querySecurityAccount, adminListUser.Id).Scan(&adminUserSecurityAccount).Error

		if errSecurityAccount != nil {
			helper.Logger("error", "In Server: "+errSecurityAccount.Error())
			return nil, errors.New(errSecurityAccount.Error())
		}

		// User Risk

		queryUserRisk := `SELECT goal, tolerance, experience, capital_market_knowledge AS pengetahuan_pasar_modal FROM user_risks WHERE user_id = ?`

		errQueryUserRisk := dbDefault.Raw(queryUserRisk, adminListUser.Id).Scan(&adminUserRisk).Error

		if errQueryUserRisk != nil {
			helper.Logger("error", "In Server: "+errQueryUserRisk.Error())
			return nil, errors.New(errQueryUserRisk.Error())
		}

		for i := range adminEmitenProjects {

			var projectContract entities.AdminProjectContract

			var dataMedias []entities.ProjectMediaPath
			var dataProjectUseOfFunds []entities.AdminProjectUseOfFunds
			var dataProjectCollateralGuarantee []entities.AdminProjectCollateralGuarantee

			// Get Project Media

			errMedia := dbDefault.Raw(`SELECT path FROM project_medias WHERE project_id = ?`, adminEmitenProjects[i].Id).Scan(&dataMedias).Error
			if errMedia != nil {
				helper.Logger("error", "Load project medias error: "+errMedia.Error())
				return nil, errMedia
			}

			// Get Project Use Of Funds

			queryProjectUseOfFunds := `SELECT id, name FROM use_of_funds WHERE project_id = ?`

			errProjectUseOfFunds := dbDefault.
				Raw(queryProjectUseOfFunds, adminEmitenProjects[i].Id).
				Scan(&dataProjectUseOfFunds).Error

			if errProjectUseOfFunds != nil {
				helper.Logger("error", "In Server: "+errProjectUseOfFunds.Error())
				return nil, errProjectUseOfFunds
			}

			// Get Project Document Verify

			var result entities.ProjectDocumentVerify

			queryProjectDocumentVerify := `
		SELECT 
			id, skd, cv, rab, other_license_document,
			video_profile_company, project_summary,
			revenue_projection, work_of_timeline, annual_tax_report,
			list_of_employment, list_of_supplier_data,
			latest_receivable_account
		FROM document_verify_projects
		WHERE project_id = ?`

			if err := dbDefault.
				Raw(queryProjectDocumentVerify, adminEmitenProjects[i].Id).
				Scan(&result).Error; err != nil {
				log.Printf("FetchProjectDocumentVerify: project docs query error: %v", err)
				return nil, err
			}

			var mediaDocs []entities.MediaDocumentVerifyProject
			queryMedia := `
		SELECT path, type
		FROM media_document_verify_projects
		WHERE id = ?`

			if err := dbDefault.
				Raw(queryMedia, result.Id).
				Scan(&mediaDocs).Error; err != nil {
				log.Printf("FetchProjectDocumentVerify: media query error: %v", err)
				return nil, err
			}

			result.Media = mediaDocs

			// Get Project Contract

			queryProjectContract := `SELECT value, path FROM project_contracts WHERE project_id = ?`

			errProjectContract := dbDefault.
				Raw(queryProjectContract, adminEmitenProjects[i].Id).
				Scan(&projectContract).Error

			if errProjectContract != nil {
				helper.Logger("error", "In Server: "+errProjectContract.Error())
				return nil, errProjectContract
			}

			// Get Project Collateral Guarantees

			queryProjectCollateralGurantee := `SELECT id, name FROM collateral_guarantees WHERE project_id = ?`

			errProjectCollateralGurantee := dbDefault.
				Raw(queryProjectCollateralGurantee, adminEmitenProjects[i].Id).
				Scan(&dataProjectCollateralGuarantee).Error

			if errProjectCollateralGurantee != nil {
				helper.Logger("error", "In Server: "+errProjectCollateralGurantee.Error())
				return nil, errProjectCollateralGurantee
			}

			if dataProjectUseOfFunds == nil {
				dataProjectUseOfFunds = []entities.AdminProjectUseOfFunds{}
			}

			if dataProjectCollateralGuarantee == nil {
				dataProjectCollateralGuarantee = []entities.AdminProjectCollateralGuarantee{}
			}

			if dataMedias == nil {
				dataMedias = []entities.ProjectMediaPath{}
			}

			dataAdminEmitenProjects = append(dataAdminEmitenProjects, entities.AdminEmitenProjectCompanyResponse{
				Id:                     adminEmitenProjects[i].Id,
				Title:                  adminEmitenProjects[i].Title,
				JenisProject:           adminEmitenProjects[i].JenisProject,
				JumlahMinimal:          adminEmitenProjects[i].JumlahMinimal,
				JangkaWaktu:            adminEmitenProjects[i].JangkaWaktu,
				TingkatBunga:           adminEmitenProjects[i].TingkatBunga,
				JadwalPembayaranBunga:  adminEmitenProjects[i].JadwalPembayaranBunga,
				JadwalPembayaranPokok:  adminEmitenProjects[i].JadwalPembayaranPokok,
				DeskripsiPekerjaan:     adminEmitenProjects[i].DeskripsiPekerjaan,
				CompanyProfile:         adminEmitenProjects[i].CompanyProfile,
				PenggunaanData:         dataProjectUseOfFunds,
				JaminanKolateral:       dataProjectCollateralGuarantee,
				TenorPinjaman:          adminEmitenProjects[i].TenorPinjaman,
				DocumentVerify:         result,
				Kontrak:                projectContract,
				Website:                adminEmitenProjects[i].Website,
				IsApbn:                 adminEmitenProjects[i].IsApbn,
				IsApproved:             adminEmitenProjects[i].IsApproved,
				Status:                 adminEmitenProjects[i].Status,
				Medias:                 dataMedias,
				MulaiProject:           adminEmitenProjects[i].StartProject,
				SelesaiProject:         adminEmitenProjects[i].EndProject,
				AlamatPenyediaProject:  adminEmitenProjects[i].ProviderAddress,
				AlamatPenyediaProvinsi: adminEmitenProjects[i].ProviderProvinceName,
				AlamatPenyediaKota:     adminEmitenProjects[i].ProviderCityName,
				AlamatPenyediaDaerah:   adminEmitenProjects[i].ProviderDistrictName,
				AlamatPenyediaWilayah:  adminEmitenProjects[i].ProviderSubdistrictName,
				AlamatPenyediaKodePos:  adminEmitenProjects[i].ProviderPostalCode,
			})
		}

		if adminEmitenPositions == nil {
			adminEmitenPositions = []entities.AdminEmitenPositionCompany{}
		}

		if adminEmitenProjects == nil {
			adminEmitenProjects = []entities.AdminEmitenProjectCompany{}
		}

		if adminEmitenAddress == nil {
			adminEmitenAddress = []entities.AdminEmitenAddressCompany{}
		}

		dataAdminListUser = append(dataAdminListUser, entities.AdminListUserResponse{
			Id:               adminListUser.Id,
			Fullname:         helper.DefaultIfEmpty(adminListUser.Fullname, "-"),
			Avatar:           helper.DefaultIfEmpty(adminListUser.Avatar, "-"),
			Selfie:           helper.DefaultIfEmpty(adminListUser.Selfie, "-"),
			PhotoKtp:         helper.DefaultIfEmpty(adminListUser.PhotoKtp, "-"),
			NoKtp:            helper.DefaultIfEmpty(adminListUser.NoKtp, "-"),
			NoNpwp:           helper.DefaultIfEmpty(adminListUser.NoNpwp, "-"),
			Jabatan:          helper.DefaultIfEmpty(adminListUser.Position, "-"),
			Gender:           helper.DefaultIfEmpty(adminListUser.Gender, "-"),
			StatusMarital:    helper.DefaultIfEmpty(adminListUser.StatusMarital, "-"),
			LastEducation:    helper.DefaultIfEmpty(adminListUser.LastEducation, "-"),
			AddressDetail:    helper.DefaultIfEmpty(adminListUser.AddressDetail, "-"),
			Occupation:       helper.DefaultIfEmpty(adminListUser.Occupation, "-"),
			ProvinceName:     helper.DefaultIfEmpty(adminListUser.ProvinceName, "-"),
			CityName:         helper.DefaultIfEmpty(adminListUser.CityName, "-"),
			DistrictName:     helper.DefaultIfEmpty(adminListUser.DistrictName, "-"),
			SubdistrictName:  helper.DefaultIfEmpty(adminListUser.SubdistrictName, "-"),
			Email:            helper.DefaultIfEmpty(adminListUser.Email, "-"),
			Phone:            helper.DefaultIfEmpty(adminListUser.Phone, "-"),
			Sku:              helper.DefaultIfEmpty(adminListUser.Sku, "-"),
			Role:             helper.DefaultIfEmpty(adminListUser.Role, "-"),
			Verified:         adminListUser.Verify,
			VerifiedEmiten:   adminListUser.VerifyEmiten,
			VerifiedInvestor: adminListUser.VerifyInvestor,
			NamaAhliWaris:    helper.DefaultIfEmpty(adminListUser.BeneficiaryName, "-"),
			PhoneAhliWaris:   helper.DefaultIfEmpty(adminListUser.BeneficiaryPhone, "-"),
			RekeningEfek: entities.RekeningEfek{
				AccountName:  helper.DefaultIfEmpty(adminUserSecurityAccount.AccountName, "-"),
				AccountNo:    helper.DefaultIfEmpty(adminUserSecurityAccount.AccountNo, "-"),
				AccountSubNo: helper.DefaultIfEmpty(adminUserSecurityAccount.AccountSubNo, "-"),
				AccountBank:  helper.DefaultIfEmpty(adminUserSecurityAccount.AccountBank, "-"),
			},
			SuratKuasa: entities.SuratKuasa{
				Type: helper.DefaultIfEmpty(adminAdditionalDoc.Type, "-"),
				Path: helper.DefaultIfEmpty(adminAdditionalDoc.Path, "-"),
			},
			Signature: entities.AdminUserSignature{
				Path: helper.DefaultIfEmpty(adminUserSignature.Path, "-"),
			},
			Risk: entities.AdminUserRisk{
				Goal:                  helper.DefaultIfEmpty(adminUserRisk.Goal, "-"),
				Tolerance:             helper.DefaultIfEmpty(adminUserRisk.Tolerance, "-"),
				Experience:            helper.DefaultIfEmpty(adminUserRisk.Experience, "-"),
				PengetahuanPasarModal: helper.DefaultIfEmpty(adminUserRisk.PengetahuanPasarModal, "-"),
			},
			Emiten: entities.AdminEmiten{
				Company: entities.AdminEmitenCompanyResponse{
					Id:                        helper.DefaultIfEmpty(adminEmitenCompany.Id, "-"),
					Name:                      helper.DefaultIfEmpty(adminEmitenCompany.Name, "-"),
					Nib:                       helper.DefaultIfEmpty(adminEmitenCompany.Nib, "-"),
					NibPath:                   helper.DefaultIfEmpty(adminEmitenCompany.NibPath, "-"),
					AktaPendirian:             helper.DefaultIfEmpty(adminEmitenCompany.AktaPendirian, "-"),
					AktaPerubahanTerahkir:     helper.DefaultIfEmpty(adminEmitenCompany.AktaPerubahanTerahkir, "-"),
					AktaPerubahanTerahkirPath: helper.DefaultIfEmpty(adminEmitenCompany.AktaPerubahanTerahkirPath, "-"),
					SkPendirianPerusahaan:     helper.DefaultIfEmpty(adminEmitenCompany.SkPendirianPerusahaan, "-"),
					SkKumham:                  helper.DefaultIfEmpty(adminEmitenCompany.SkKumham, "-"),
					SkKumhamTerahkir:          helper.DefaultIfEmpty(adminEmitenCompany.SkKumhamLast, "-"),
					SkKumhamPath:              helper.DefaultIfEmpty(adminEmitenCompany.SkKumhamPath, "-"),
					Npwp:                      helper.DefaultIfEmpty(adminEmitenCompany.Npwp, "-"),
					NpwpPath:                  helper.DefaultIfEmpty(adminEmitenCompany.NpwpPath, "-"),
					TotalEmployees:            helper.DefaultIfEmpty(adminEmitenCompany.TotalEmployees, "-"),
					LaporanKeuangan:           helper.DefaultIfEmpty(adminEmitenCompany.LaporanKeuangan, "-"),
					RekeningKoran:             helper.DefaultIfEmpty(adminEmitenCompany.RekeningKoran, "-"),
					Site:                      helper.DefaultIfEmpty(adminEmitenCompany.Site, "-"),
					Email:                     helper.DefaultIfEmpty(adminEmitenCompany.Email, "-"),
					Phone:                     helper.DefaultIfEmpty(adminEmitenCompany.Phone, "-"),
					Est:                       helper.DefaultIfEmpty(adminEmitenCompany.Est, "-"),
					BankName:                  helper.DefaultIfEmpty(adminEmitenCompany.BankName, "-"),
					BankAccount:               helper.DefaultIfEmpty(adminEmitenCompany.BankAccount, "-"),
					BankOwnerCompany:          helper.DefaultIfEmpty(adminEmitenCompany.BankOwnerCompany, "-"),
					Siup:                      helper.DefaultIfEmpty(adminEmitenCompany.Siup, "-"),
					Tdp:                       helper.DefaultIfEmpty(adminEmitenCompany.Tdp, "-"),
					JenisPerusahaan:           helper.DefaultIfEmpty(adminEmitenCompany.JenisPerusahaan, "-"),
					JenisUsaha:                helper.DefaultIfEmpty(adminEmitenCompany.JenisUsaha, "-"),
					StatusKantor:              helper.DefaultIfEmpty(adminEmitenCompany.StatusKantor, "-"),
					Address:                   adminEmitenAddress,
					Positions:                 adminEmitenPositions,
					Projects:                  dataAdminEmitenProjects,
				},
			},
			Investor: entities.AdminInvestor{
				Ktp:      adminUserKtp,
				Bank:     adminUserBank,
				Job:      adminUserJob,
				SlipGaji: adminUserSlipPay,
			},
			CreatedAt: adminListUser.CreatedAt,
			UpdatedAt: adminListUser.UpdatedAt,
		})
	}

	var nextUrl = strconv.Itoa(nextPage)
	var prevUrl = strconv.Itoa(prevPage)

	return map[string]any{
		"total":        resultTotal,
		"current_page": pageinteger,
		"per_page":     int(perPage),
		"prev_page":    prevPage,
		"next_page":    nextPage,
		"next_url":     url + "/api/v1/admin/list/user?page=" + nextUrl,
		"prev_url":     url + "/api/v1/admin/list/user?page=" + prevUrl,
		"data":         dataAdminListUser,
	}, nil
}

func AdminDetailUser(userId string) (map[string]any, error) {
	var dataAdminListUser []entities.AdminListUserResponse
	var adminListUser entities.AdminListUser

	query := `SELECT u.uid AS id, p.avatar, p.fullname, u.email, u.phone, u.verify_emiten, p.selfie, 
	p.photo_ktp, p.no_ktp, p.no_npwp, u.sku, p.position, p.gender, p.province_name, p.city_name, p.district_name, p.subdistrict_name, p.status_marital, p.occupation, p.last_education, p.address_detail,
	r.name AS role, p.beneficiary_name, p.beneficiary_phone,
	u.created_at, u.verify, u.verify_emiten, u.verify_investor, 
	u.updated_at
	FROM users u 
	INNER JOIN roles r ON r.id = u.role
	INNER JOIN profiles p ON u.uid = p.user_id
	WHERE u.uid = ?
	ORDER BY u.created_at DESC`

	rows, err := dbDefault.Raw(query, userId).Rows()

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	for rows.Next() {
		errAdminListUserRows := dbDefault.ScanRows(rows, &adminListUser)
		if errAdminListUserRows != nil {
			helper.Logger("error", "In Server: "+errAdminListUserRows.Error())
			return nil, errors.New(errAdminListUserRows.Error())
		}

		var adminUserKtp entities.AdminUserKtp
		var adminUserJob entities.AdminUserJob
		var adminUserBank entities.AdminUserBank
		var adminUserSignature entities.AdminUserSignature
		var adminUserRisk entities.AdminUserRisk
		var adminUserSecurityAccount entities.AdminUserSecurityAccount
		var adminUserSlipPay entities.AdminUserSlipPay

		var adminAdditionalDoc entities.AdditionalDoc

		var adminEmitenCompany entities.AdminEmitenCompany
		var adminEmitenPositions []entities.AdminEmitenPositionCompany
		var adminEmitenProjects []entities.AdminEmitenProjectCompany
		var adminEmitenAddress []entities.AdminEmitenAddressCompany

		dataAdminEmitenProjects := make([]entities.AdminEmitenProjectCompanyResponse, 0)

		queryKtp := `SELECT name, nik, place_datebirth, path FROM ktps WHERE user_id = ?`

		errKtp := dbDefault.Raw(queryKtp, adminListUser.Id).Scan(&adminUserKtp).Error
		if errKtp != nil {
			helper.Logger("error", "In Server: "+errKtp.Error())
			return nil, errors.New(errKtp.Error())
		}

		queryJob := `SELECT company_name, province_name, city_name, district_name, 
		subdistrict_name, company_address, monthly_income, annual_income, 
		position, npwp, npwp_path 
		FROM jobs 
		WHERE user_id = ?`

		errJob := dbDefault.Raw(queryJob, adminListUser.Id).Scan(&adminUserJob).Error
		if errJob != nil {
			helper.Logger("error", "In Server: "+errJob.Error())
			return nil, errors.New(errJob.Error())
		}

		queryBank := `SELECT no, bank_name AS name, bank_owner AS owner, bank_branch AS branch, rek_koran_path FROM accounts WHERE user_id = ?`

		errBank := dbDefault.Raw(queryBank, adminListUser.Id).Scan(&adminUserBank).Error
		if errBank != nil {
			helper.Logger("error", "In Server: "+errBank.Error())
			return nil, errors.New(errBank.Error())
		}

		queryAdditionalDoc := `SELECT type, path, user_id FROM additional_docs WHERE user_id = ?`

		errAdditionalDoc := dbDefault.Raw(queryAdditionalDoc, adminListUser.Id).Scan(&adminAdditionalDoc).Error
		if errBank != nil {
			helper.Logger("error", "In Server: "+errAdditionalDoc.Error())
			return nil, errors.New(errAdditionalDoc.Error())
		}

		queryEmitenCompany := `SELECT c.uid AS id, c.company_name AS name, c.company_nib AS nib, 
		c.site, c.email, c.phone, c.est, c.bank_name, c.bank_account, c.npwp, c.siup, c.tdp,
		toc.name AS jenis_perusahaan,
		tob.name AS jenis_usaha,
		tocp.name AS status_kantor,
		c.bank_owner_company, c.company_nib_path AS nib_path, c.deed_of_incorporation AS akta_pendirian, 
		c.latest_amendment_deed AS akta_perubahan_terahkir, 
		c.latest_amendment_deed_path AS akta_perubahan_terahkir_path,
		c.certificate_of_company_est AS sk_pendirian_perusahaan,
		c.sk_kumham, 
		c.sk_kumham_last, c.sk_kumham_path, 
		c.npwp, c.npwp_path, c.total_employees, 
		c.financial_statement AS laporan_keuangan, c.bank_statement AS rekening_koran
 		FROM companies c 
		INNER JOIN type_of_companies toc ON toc.id = c.type_of_company
		INNER JOIN type_of_businesses tob ON tob.id = c.type_of_business
		INNER JOIN type_of_company_places tocp ON tocp.id = c.type_of_company_place
		WHERE user_id = ?`

		errEmitenCompany := dbDefault.Raw(queryEmitenCompany, adminListUser.Id).Scan(&adminEmitenCompany).Error
		if errEmitenCompany != nil {
			helper.Logger("error", "In Server: "+errEmitenCompany.Error())
			return nil, errors.New(errEmitenCompany.Error())
		}

		// Company Position

		queryEmitenPositionCompany := `SELECT id, title, name, position, ktp, ktp_path, npwp, npwp_path FROM positions WHERE company_id = ?`

		errEmitenPositionCompany := dbDefault.Raw(queryEmitenPositionCompany, adminEmitenCompany.Id).Scan(&adminEmitenPositions).Error

		if errEmitenPositionCompany != nil {
			helper.Logger("error", "In Server: "+errEmitenPositionCompany.Error())
			return nil, errors.New(errEmitenPositionCompany.Error())
		}

		// Company Project

		queryEmitenProjectCompany := `
			SELECT 
				p.uid AS id, 
				p.title, 
				top.name AS jenis_project, 
				p.is_apbn, 
				p.is_approved, 
				p.site AS website,
				p.loan_term AS tenor_pinjaman,
				p.spk,
				p.loa,
				p.sku,
				p.nominal_value AS jumlah_minimal,
				p.time_periode AS jangka_waktu, 
				p.interest_rate AS tingkat_bunga, 
				p.interest_payment_schedule AS jadwal_pembayaran_bunga, 
				p.principal_payment_schedule AS jadwal_pembayaran_pokok, 
				p.desc_job AS deskripsi_pekerjaan,
				p.company_profile, 
				p.collateral_guarantee AS jaminan_kolateral, 
				ps.name AS status,
				p.start_project, p.end_project,
				p.provider_address, p.provider_province_name, p.provider_city_name, p.provider_district_name,
				p.provider_subdistrict_name, p.provider_postal_code
			FROM 
				projects p
			INNER JOIN 
				project_statuses ps ON p.status = ps.id
			INNER JOIN
				type_of_projects top ON top.id = p.type_of_project
			WHERE 
				p.company_id = ?
		`

		errEmitenProjectCompany := dbDefault.Raw(queryEmitenProjectCompany, adminEmitenCompany.Id).Scan(&adminEmitenProjects).Error

		if errEmitenProjectCompany != nil {
			helper.Logger("error", "In Server: "+errEmitenProjectCompany.Error())
			return nil, errors.New(errEmitenProjectCompany.Error())
		}

		// Company Address

		queryEmitenAddressCompany := `SELECT ca.name, ca.address, ca.postal_code,
		ca.province_name AS province, ca.city_name AS city, ca.district_name AS district, ca.subdistrict_name AS subdistrict FROM company_addresses ca 
		WHERE ca.company_id = ?`

		errEmitenAddressCompany := dbDefault.Raw(queryEmitenAddressCompany, adminEmitenCompany.Id).Scan(&adminEmitenAddress).Error

		if errEmitenAddressCompany != nil {
			helper.Logger("error", "In Server: "+errEmitenAddressCompany.Error())
			return nil, errors.New(errEmitenAddressCompany.Error())
		}

		// User Signature

		queryUserSignature := `SELECT path FROM user_signatures WHERE user_id = ?`

		errQueryUserSignature := dbDefault.Raw(queryUserSignature, adminListUser.Id).Scan(&adminUserSignature).Error

		if errQueryUserSignature != nil {
			helper.Logger("error", "In Server: "+errQueryUserSignature.Error())
			return nil, errors.New(errQueryUserSignature.Error())
		}

		// User Security Account

		querySecurityAccount := `SELECT account_name, account_no, account_sub_no, account_bank
		FROM security_accounts 
		WHERE user_id = ?`

		errSecurityAccount := dbDefault.Raw(querySecurityAccount, adminListUser.Id).Scan(&adminUserSecurityAccount).Error

		if errSecurityAccount != nil {
			helper.Logger("error", "In Server: "+errSecurityAccount.Error())
			return nil, errors.New(errSecurityAccount.Error())
		}

		// User Slip Pay

		querySlipPay := `SELECT path FROM pay_slips 
		WHERE user_id = ?`

		errSlipPay := dbDefault.Raw(querySlipPay, adminListUser.Id).Scan(&adminUserSlipPay).Error

		if errSlipPay != nil {
			helper.Logger("error", "In Server: "+errSlipPay.Error())
			return nil, errors.New(errSlipPay.Error())
		}

		// User Risk

		queryUserRisk := `SELECT goal, tolerance, experience, capital_market_knowledge AS pengetahuan_pasar_modal FROM user_risks WHERE user_id = ?`

		errQueryUserRisk := dbDefault.Raw(queryUserRisk, adminListUser.Id).Scan(&adminUserRisk).Error

		if errQueryUserRisk != nil {
			helper.Logger("error", "In Server: "+errQueryUserRisk.Error())
			return nil, errors.New(errQueryUserRisk.Error())
		}

		for i := range adminEmitenProjects {

			var dataMedias []entities.ProjectMediaPath

			var projectContract entities.AdminProjectContract

			var dataProjectUseOfFunds []entities.AdminProjectUseOfFunds
			var dataProjectCollateralGuarantee []entities.AdminProjectCollateralGuarantee

			// Get Project Media

			errMedia := dbDefault.Raw(`SELECT path FROM project_medias WHERE project_id = ?`, adminEmitenProjects[i].Id).Scan(&dataMedias).Error
			if errMedia != nil {
				helper.Logger("error", "Load project medias error: "+errMedia.Error())
				return nil, errMedia
			}

			// Get Project Use Of Funds

			queryProjectUseOfFunds := `SELECT id, name FROM use_of_funds WHERE project_id = ?`

			errProjectUseOfFunds := dbDefault.
				Raw(queryProjectUseOfFunds, adminEmitenProjects[i].Id).
				Scan(&dataProjectUseOfFunds).Error

			if errProjectUseOfFunds != nil {
				helper.Logger("error", "In Server: "+errProjectUseOfFunds.Error())
				return nil, errProjectUseOfFunds
			}

			// Get Project Contract

			queryProjectContract := `SELECT value, path FROM project_contracts WHERE project_id = ?`

			errProjectContract := dbDefault.
				Raw(queryProjectContract, adminEmitenProjects[i].Id).
				Scan(&projectContract).Error

			if errProjectContract != nil {
				helper.Logger("error", "In Server: "+errProjectContract.Error())
				return nil, errProjectContract
			}

			// Get Project Document Verify

			var result entities.ProjectDocumentVerify

			queryProjectDocumentVerify := `
		SELECT 
			id, skd, cv, rab, other_license_document,
			video_profile_company, project_summary,
			revenue_projection, work_of_timeline, annual_tax_report,
			list_of_employment, list_of_supplier_data,
			latest_receivable_account
		FROM document_verify_projects
		WHERE project_id = ?`

			if err := dbDefault.
				Raw(queryProjectDocumentVerify, adminEmitenProjects[i].Id).
				Scan(&result).Error; err != nil {
				log.Printf("FetchProjectDocumentVerify: project docs query error: %v", err)
				return nil, err
			}

			var mediaDocs []entities.MediaDocumentVerifyProject
			queryMedia := `
		SELECT path, type
		FROM media_document_verify_projects
		WHERE id = ?`

			if err := dbDefault.
				Raw(queryMedia, result.Id).
				Scan(&mediaDocs).Error; err != nil {
				log.Printf("FetchProjectDocumentVerify: media query error: %v", err)
				return nil, err
			}

			result.Media = mediaDocs

			// Get Project Collateral Guarantees

			queryProjectCollateralGurantee := `SELECT id, name FROM collateral_guarantees WHERE project_id = ?`

			errProjectCollateralGurantee := dbDefault.
				Raw(queryProjectCollateralGurantee, adminEmitenProjects[i].Id).
				Scan(&dataProjectCollateralGuarantee).Error

			if errProjectCollateralGurantee != nil {
				helper.Logger("error", "In Server: "+errProjectCollateralGurantee.Error())
				return nil, errProjectCollateralGurantee
			}

			if dataProjectUseOfFunds == nil {
				dataProjectUseOfFunds = []entities.AdminProjectUseOfFunds{}
			}

			if dataProjectCollateralGuarantee == nil {
				dataProjectCollateralGuarantee = []entities.AdminProjectCollateralGuarantee{}
			}

			if dataMedias == nil {
				dataMedias = []entities.ProjectMediaPath{}
			}

			dataAdminEmitenProjects = append(dataAdminEmitenProjects, entities.AdminEmitenProjectCompanyResponse{
				Id:                     adminEmitenProjects[i].Id,
				Title:                  adminEmitenProjects[i].Title,
				JenisProject:           adminEmitenProjects[i].JenisProject,
				JumlahMinimal:          adminEmitenProjects[i].JumlahMinimal,
				JangkaWaktu:            adminEmitenProjects[i].JangkaWaktu,
				TingkatBunga:           adminEmitenProjects[i].TingkatBunga,
				JadwalPembayaranBunga:  adminEmitenProjects[i].JadwalPembayaranBunga,
				JadwalPembayaranPokok:  adminEmitenProjects[i].JadwalPembayaranPokok,
				DeskripsiPekerjaan:     adminEmitenProjects[i].DeskripsiPekerjaan,
				CompanyProfile:         adminEmitenProjects[i].CompanyProfile,
				PenggunaanData:         dataProjectUseOfFunds,
				TenorPinjaman:          adminEmitenProjects[i].TenorPinjaman,
				DocumentVerify:         result,
				Website:                adminEmitenProjects[i].Website,
				JaminanKolateral:       dataProjectCollateralGuarantee,
				Kontrak:                projectContract,
				IsApbn:                 adminEmitenProjects[i].IsApbn,
				IsApproved:             adminEmitenProjects[i].IsApproved,
				Status:                 adminEmitenProjects[i].Status,
				Medias:                 dataMedias,
				MulaiProject:           adminEmitenProjects[i].StartProject,
				SelesaiProject:         adminEmitenProjects[i].EndProject,
				AlamatPenyediaProject:  adminEmitenProjects[i].ProviderAddress,
				AlamatPenyediaProvinsi: adminEmitenProjects[i].ProviderProvinceName,
				AlamatPenyediaKota:     adminEmitenProjects[i].ProviderCityName,
				AlamatPenyediaDaerah:   adminEmitenProjects[i].ProviderDistrictName,
				AlamatPenyediaWilayah:  adminEmitenProjects[i].ProviderSubdistrictName,
				AlamatPenyediaKodePos:  adminEmitenProjects[i].ProviderPostalCode,
			})
		}

		if adminEmitenPositions == nil {
			adminEmitenPositions = []entities.AdminEmitenPositionCompany{}
		}

		if adminEmitenProjects == nil {
			adminEmitenProjects = []entities.AdminEmitenProjectCompany{}
		}

		if adminEmitenAddress == nil {
			adminEmitenAddress = []entities.AdminEmitenAddressCompany{}
		}

		dataAdminListUser = append(dataAdminListUser, entities.AdminListUserResponse{
			Id:               adminListUser.Id,
			Fullname:         helper.DefaultIfEmpty(adminListUser.Fullname, "-"),
			Avatar:           helper.DefaultIfEmpty(adminListUser.Avatar, "-"),
			Selfie:           helper.DefaultIfEmpty(adminListUser.Selfie, "-"),
			PhotoKtp:         helper.DefaultIfEmpty(adminListUser.PhotoKtp, "-"),
			NoKtp:            helper.DefaultIfEmpty(adminListUser.NoKtp, "-"),
			NoNpwp:           helper.DefaultIfEmpty(adminListUser.NoNpwp, "-"),
			Jabatan:          helper.DefaultIfEmpty(adminListUser.Position, "-"),
			Gender:           helper.DefaultIfEmpty(adminListUser.Gender, "-"),
			StatusMarital:    helper.DefaultIfEmpty(adminListUser.StatusMarital, "-"),
			LastEducation:    helper.DefaultIfEmpty(adminListUser.LastEducation, "-"),
			AddressDetail:    helper.DefaultIfEmpty(adminListUser.AddressDetail, "-"),
			Occupation:       helper.DefaultIfEmpty(adminListUser.Occupation, "-"),
			ProvinceName:     helper.DefaultIfEmpty(adminListUser.ProvinceName, "-"),
			CityName:         helper.DefaultIfEmpty(adminListUser.CityName, "-"),
			DistrictName:     helper.DefaultIfEmpty(adminListUser.DistrictName, "-"),
			SubdistrictName:  helper.DefaultIfEmpty(adminListUser.SubdistrictName, "-"),
			Email:            helper.DefaultIfEmpty(adminListUser.Email, "-"),
			Phone:            helper.DefaultIfEmpty(adminListUser.Phone, "-"),
			Sku:              helper.DefaultIfEmpty(adminListUser.Sku, "-"),
			Role:             helper.DefaultIfEmpty(adminListUser.Role, "-"),
			Verified:         adminListUser.Verify,
			VerifiedEmiten:   adminListUser.VerifyEmiten,
			VerifiedInvestor: adminListUser.VerifyInvestor,
			NamaAhliWaris:    helper.DefaultIfEmpty(adminListUser.BeneficiaryName, "-"),
			PhoneAhliWaris:   helper.DefaultIfEmpty(adminListUser.BeneficiaryPhone, "-"),
			RekeningEfek: entities.RekeningEfek{
				AccountName:  helper.DefaultIfEmpty(adminUserSecurityAccount.AccountName, "-"),
				AccountNo:    helper.DefaultIfEmpty(adminUserSecurityAccount.AccountNo, "-"),
				AccountSubNo: helper.DefaultIfEmpty(adminUserSecurityAccount.AccountSubNo, "-"),
				AccountBank:  helper.DefaultIfEmpty(adminUserSecurityAccount.AccountBank, "-"),
			},
			SuratKuasa: entities.SuratKuasa{
				Type: helper.DefaultIfEmpty(adminAdditionalDoc.Type, "-"),
				Path: helper.DefaultIfEmpty(adminAdditionalDoc.Path, "-"),
			},
			Signature: entities.AdminUserSignature{
				Path: helper.DefaultIfEmpty(adminUserSignature.Path, "-"),
			},
			Risk: entities.AdminUserRisk{
				Goal:                  helper.DefaultIfEmpty(adminUserRisk.Goal, "-"),
				Tolerance:             helper.DefaultIfEmpty(adminUserRisk.Tolerance, "-"),
				Experience:            helper.DefaultIfEmpty(adminUserRisk.Experience, "-"),
				PengetahuanPasarModal: helper.DefaultIfEmpty(adminUserRisk.PengetahuanPasarModal, "-"),
			},
			Emiten: entities.AdminEmiten{
				Company: entities.AdminEmitenCompanyResponse{
					Id:                        helper.DefaultIfEmpty(adminEmitenCompany.Id, "-"),
					Name:                      helper.DefaultIfEmpty(adminEmitenCompany.Name, "-"),
					Nib:                       helper.DefaultIfEmpty(adminEmitenCompany.Nib, "-"),
					NibPath:                   helper.DefaultIfEmpty(adminEmitenCompany.NibPath, "-"),
					AktaPendirian:             helper.DefaultIfEmpty(adminEmitenCompany.AktaPendirian, "-"),
					AktaPerubahanTerahkir:     helper.DefaultIfEmpty(adminEmitenCompany.AktaPerubahanTerahkir, "-"),
					AktaPerubahanTerahkirPath: helper.DefaultIfEmpty(adminEmitenCompany.AktaPerubahanTerahkirPath, "-"),
					SkPendirianPerusahaan:     helper.DefaultIfEmpty(adminEmitenCompany.SkPendirianPerusahaan, "-"),
					SkKumham:                  helper.DefaultIfEmpty(adminEmitenCompany.SkKumham, "-"),
					SkKumhamTerahkir:          helper.DefaultIfEmpty(adminEmitenCompany.SkKumhamLast, "-"),
					SkKumhamPath:              helper.DefaultIfEmpty(adminEmitenCompany.SkKumhamPath, "-"),
					Npwp:                      helper.DefaultIfEmpty(adminEmitenCompany.Npwp, "-"),
					NpwpPath:                  helper.DefaultIfEmpty(adminEmitenCompany.NpwpPath, "-"),
					TotalEmployees:            helper.DefaultIfEmpty(adminEmitenCompany.TotalEmployees, "-"),
					LaporanKeuangan:           helper.DefaultIfEmpty(adminEmitenCompany.LaporanKeuangan, "-"),
					RekeningKoran:             helper.DefaultIfEmpty(adminEmitenCompany.RekeningKoran, "-"),
					Site:                      helper.DefaultIfEmpty(adminEmitenCompany.Site, "-"),
					Email:                     helper.DefaultIfEmpty(adminEmitenCompany.Email, "-"),
					Phone:                     helper.DefaultIfEmpty(adminEmitenCompany.Phone, "-"),
					Est:                       helper.DefaultIfEmpty(adminEmitenCompany.Est, "-"),
					BankName:                  helper.DefaultIfEmpty(adminEmitenCompany.BankName, "-"),
					BankAccount:               helper.DefaultIfEmpty(adminEmitenCompany.BankAccount, "-"),
					BankOwnerCompany:          helper.DefaultIfEmpty(adminEmitenCompany.BankOwnerCompany, "-"),
					Siup:                      helper.DefaultIfEmpty(adminEmitenCompany.Siup, "-"),
					Tdp:                       helper.DefaultIfEmpty(adminEmitenCompany.Tdp, "-"),
					JenisPerusahaan:           helper.DefaultIfEmpty(adminEmitenCompany.JenisPerusahaan, "-"),
					JenisUsaha:                helper.DefaultIfEmpty(adminEmitenCompany.JenisUsaha, "-"),
					StatusKantor:              helper.DefaultIfEmpty(adminEmitenCompany.StatusKantor, "-"),
					Address:                   adminEmitenAddress,
					Positions:                 adminEmitenPositions,
					Projects:                  dataAdminEmitenProjects,
				},
			},
			Investor: entities.AdminInvestor{
				Ktp:      adminUserKtp,
				Bank:     adminUserBank,
				Job:      adminUserJob,
				SlipGaji: adminUserSlipPay,
			},
			CreatedAt: adminListUser.CreatedAt,
			UpdatedAt: adminListUser.UpdatedAt,
		})
	}

	return map[string]any{
		"data": dataAdminListUser[0],
	}, nil
}

func AdminListProject(page, limit string) (map[string]any, error) {
	url := os.Getenv("API_URL_DEV")

	var dataAdminListProject []entities.AdminListProjectResponse
	var totalAdminListProject []entities.AdminListProject

	var adminListProject entities.AdminListProject

	pageinteger, _ := strconv.Atoi(page)
	limitinteger, _ := strconv.Atoi(limit)

	var offset = strconv.Itoa((pageinteger - 1) * limitinteger)

	errAllUser := dbDefault.Raw(`SELECT id FROM projects`).Scan(&totalAdminListProject).Error

	if errAllUser != nil {
		helper.Logger("error", "In Server: "+errAllUser.Error())
	}

	var resultTotal = len(totalAdminListProject)

	var perPage = math.Ceil(float64(resultTotal) / float64(limitinteger))

	var prevPage int
	var nextPage int

	if pageinteger == 1 {
		prevPage = 1
	} else {
		prevPage = pageinteger - 1
	}

	nextPage = pageinteger + 1

	query := `SELECT p.uid AS id, p.title, p.goal, p.capital, p.min_invest, p.min_invest,
	ps.name AS status,
	p.unit_price, p.unit_total, p.number_of_unit, p.periode, top.name AS type_of_project,
	p.required_fund,
	p.nominal_value, p.time_periode, p.interest_rate, p.interest_payment_schedule, p.principal_payment_schedule,
	p.use_of_funds, p.collateral_guarantee, p.desc_job, p.is_apbn, p.is_approved,
	p.profit_percentage,
	u.uid AS user_id, u.email AS user_email, u.phone AS user_phone, pro.fullname AS user_name,
	p.loan_term, p.spk, p.loa,
	p.unit_price, p.unit_total, p.min_invest, p.number_of_unit,
	p.code_effect,
	p.created_at,
	p.sku,
	p.amount_shares_per_lot,
	toc.name AS type_of_contracting_authority,
	p.site,
	p.start_project, p.end_project,
	p.doc_bank_statement, 
	p.doc_financial_statement, 
	p.doc_contract,
	p.doc_prospect,
	p.contracting_authority,
	p.provider_address, 
	p.provider_province_name, p.provider_city_name, 
	p.provider_district_name, p.provider_subdistrict_name, p.provider_postal_code
	FROM projects p
	INNER JOIN companies c ON c.uid = p.company_id
	INNER JOIN users u ON u.uid = c.user_id
	INNER JOIN profiles pro ON pro.user_id = u.uid
	INNER JOIN type_of_projects top ON top.id = p.type_of_project
	INNER JOIN type_of_contracting_authorities toc ON toc.id = p.type_of_contracting_authority  
	INNER JOIN project_statuses ps ON ps.id = p.status
	ORDER BY p.created_at DESC
	LIMIT ?, ?`

	rows, err := dbDefault.Raw(query, offset, limit).Rows()

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	for rows.Next() {
		errAdminListUserRows := dbDefault.ScanRows(rows, &adminListProject)

		if errAdminListUserRows != nil {
			helper.Logger("error", "In Server: "+errAdminListUserRows.Error())
			return nil, errors.New(errAdminListUserRows.Error())
		}

		//  Project Media

		dataProjectMedia := make([]entities.AdminListMedia, 0)

		queryProjectMedia := `SELECT id, path FROM project_medias WHERE project_id = ?`

		errProjectMedia := dbDefault.
			Raw(queryProjectMedia, adminListProject.Id).
			Scan(&dataProjectMedia).Error

		if errProjectMedia != nil {
			helper.Logger("error", "In Server: "+errProjectMedia.Error())
			return nil, errProjectMedia
		}

		// Project Location

		var dataProjectLoc entities.ProjectLocation

		queryProjectLoc := `SELECT id, url, name, lat, lng FROM project_locations WHERE project_id = ?`

		errProjectLoc := dbDefault.
			Raw(queryProjectLoc, adminListProject.Id).
			Scan(&dataProjectLoc).Error

		if errProjectLoc != nil {
			helper.Logger("error", "In Server: "+errProjectLoc.Error())
			return nil, errProjectLoc
		}

		// Project Company

		var dataProjectCompany entities.ProjectCompany
		queryProjectCompany := `SELECT c.uid AS id, c.company_name, 
		ca.address, ca.postal_code, ca.province_name, ca.city_name, 
		ca.district_name, ca.subdistrict_name 
		FROM companies c 
		INNER JOIN company_addresses ca ON c.uid = ca.company_id
		WHERE c.user_id = ?`
		errProjectCompany := dbDefault.Raw(queryProjectCompany, adminListProject.UserId).Scan(&dataProjectCompany).Error
		if errProjectCompany != nil {
			helper.Logger("error", "In Server: "+errProjectCompany.Error())
			return nil, errProjectCompany
		}

		// Project Contract

		var projectContract entities.AdminProjectContract

		queryProjectContract := `SELECT value, path FROM project_contracts WHERE project_id = ?`

		errProjectContract := dbDefault.
			Raw(queryProjectContract, adminListProject.Id).
			Scan(&projectContract).Error

		if errProjectContract != nil {
			helper.Logger("error", "In Server: "+errProjectContract.Error())
			return nil, errProjectContract
		}

		// Project Document Verify

		var result entities.ProjectDocumentVerify

		queryProjectDocumentVerify := `
		SELECT id,
			skd, cv, rab, other_license_document,
			cashflow_project,
			video_profile_company, project_summary,
			revenue_projection, work_of_timeline, annual_tax_report,
			list_of_employment, list_of_supplier_data,
			latest_receivable_account
		FROM document_verify_projects
		WHERE project_id = ?`

		if err := dbDefault.
			Raw(queryProjectDocumentVerify, adminListProject.Id).
			Scan(&result).Error; err != nil {
			log.Printf("FetchProjectDocumentVerify: project docs query error: %v", err)
			return nil, err
		}

		var mediaDocs []entities.MediaDocumentVerifyProject
		queryMedia := `
		SELECT path, type
		FROM media_document_verify_projects
		WHERE document_verify_project_id = ?`

		if err := dbDefault.
			Raw(queryMedia, result.Id).
			Scan(&mediaDocs).Error; err != nil {
			log.Printf("FetchProjectDocumentVerify: media query error: %v", err)
			return nil, err
		}

		result.Media = mediaDocs

		// Project Payment

		var projectPayment entities.ProjectPayment

		queryProjectPayment := `SELECT path, is_approve 
		FROM project_payments WHERE project_id = ?`

		errProjectPayment := dbDefault.
			Raw(queryProjectPayment, adminListProject.Id).
			Scan(&projectPayment).Error

		if errProjectPayment != nil {
			helper.Logger("error", "In Server: "+errProjectPayment.Error())
			return nil, errProjectPayment
		}

		// Project Use Of Funds

		var dataProjectUseOfFunds []entities.AdminProjectUseOfFunds

		queryProjectUseOfFunds := `SELECT id, name FROM use_of_funds WHERE project_id = ?`

		errProjectUseOfFunds := dbDefault.
			Raw(queryProjectUseOfFunds, adminListProject.Id).
			Scan(&dataProjectUseOfFunds).Error

		if errProjectUseOfFunds != nil {
			helper.Logger("error", "In Server: "+errProjectUseOfFunds.Error())
			return nil, errProjectUseOfFunds
		}

		if dataProjectUseOfFunds == nil {
			dataProjectUseOfFunds = []entities.AdminProjectUseOfFunds{}
		}

		// Project Collateral Guarantee

		var dataProjectCollateralGuarantee []entities.AdminProjectCollateralGuarantee

		queryProjectCollateralGurantee := `SELECT id, name FROM collateral_guarantees WHERE project_id = ?`

		errProjectCollateralGurantee := dbDefault.
			Raw(queryProjectCollateralGurantee, adminListProject.Id).
			Scan(&dataProjectCollateralGuarantee).Error

		if errProjectCollateralGurantee != nil {
			helper.Logger("error", "In Server: "+errProjectCollateralGurantee.Error())
			return nil, errProjectCollateralGurantee
		}

		if dataProjectCollateralGuarantee == nil {
			dataProjectCollateralGuarantee = []entities.AdminProjectCollateralGuarantee{}
		}

		// Project User

		var dataAdminListUser []entities.AdminListUserResponse
		var adminListUser entities.AdminListUser

		userQuery := `SELECT u.uid AS id, p.fullname, p.avatar, u.email, u.phone,
		p.selfie, p.photo_ktp, p.no_ktp, p.no_npwp, p.position, p.gender, 
		p.province_name, p.city_name, p.district_name, p.subdistrict_name, 
		p.status_marital, p.occupation, p.last_education, p.address_detail,
		p.beneficiary_name, p.beneficiary_phone, r.name AS role, 
		u.created_at, u.updated_at, u.verify, u.verify_emiten, u.verify_investor
		FROM users u 
		INNER JOIN roles r ON r.id = u.role
		INNER JOIN profiles p ON u.uid = p.user_id
		WHERE u.uid = ?
		ORDER BY u.created_at DESC`

		rows, err := dbDefault.Raw(userQuery, adminListProject.UserId).Rows()

		if err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return nil, errors.New(err.Error())
		}

		for rows.Next() {
			errAdminListUserRows := dbDefault.ScanRows(rows, &adminListUser)
			if errAdminListUserRows != nil {
				helper.Logger("error", "In Server: "+errAdminListUserRows.Error())
				return nil, errors.New(errAdminListUserRows.Error())
			}

			var adminUserKtp entities.AdminUserKtp
			var adminUserJob entities.AdminUserJob
			var adminUserBank entities.AdminUserBank
			var adminAdditionalDoc entities.AdditionalDoc
			var adminUserSignature entities.AdminUserSignature
			var adminUserSecurityAccount entities.AdminUserSecurityAccount
			var adminUserRisk entities.AdminUserRisk

			var adminUserSlipPay entities.AdminUserSlipPay

			var adminEmitenCompany entities.AdminEmitenCompany
			var adminEmitenPositions []entities.AdminEmitenPositionCompany

			var adminEmitenProjects []entities.AdminEmitenProjectCompany

			var adminEmitenAddress []entities.AdminEmitenAddressCompany

			dataAdminEmitenProjects := make([]entities.AdminEmitenProjectCompanyResponse, 0)

			// Ktp

			queryKtp := `SELECT name, nik, place_datebirth, path FROM ktps WHERE user_id = ?`

			errKtp := dbDefault.Raw(queryKtp, adminListUser.Id).Scan(&adminUserKtp).Error
			if errKtp != nil {
				helper.Logger("error", "In Server: "+errKtp.Error())
				return nil, errors.New(errKtp.Error())
			}

			// Job

			queryJob := `SELECT company_name, province_name, city_name, district_name, subdistrict_name, company_address, monthly_income, annual_income_num AS annual_income, 
		position, npwp, npwp_path 
		FROM jobs WHERE user_id = ?`

			errJob := dbDefault.Raw(queryJob, adminListUser.Id).Scan(&adminUserJob).Error
			if errJob != nil {
				helper.Logger("error", "In Server: "+errJob.Error())
				return nil, errors.New(errJob.Error())
			}

			// Bank

			queryBank := `SELECT no, bank_name AS name, bank_owner AS owner, bank_branch AS branch, rek_koran_path FROM accounts WHERE user_id = ?`

			errBank := dbDefault.Raw(queryBank, adminListUser.Id).Scan(&adminUserBank).Error
			if errBank != nil {
				helper.Logger("error", "In Server: "+errBank.Error())
				return nil, errors.New(errBank.Error())
			}

			// Additional Doc

			queryAdditionalDoc := `SELECT type, path, user_id FROM additional_docs WHERE user_id = ?`

			errAdditionalDoc := dbDefault.Raw(queryAdditionalDoc, adminListUser.Id).Scan(&adminAdditionalDoc).Error
			if errAdditionalDoc != nil {
				helper.Logger("error", "In Server: "+errAdditionalDoc.Error())
				return nil, errors.New(errAdditionalDoc.Error())
			}

			// Company

			queryEmitenCompany := `SELECT c.uid AS id, c.company_name AS name, c.company_nib AS nib, c.site, c.email, c.phone, c.est, 
			toc.name AS jenis_perusahaan,
			tob.name AS jenis_usaha,
			tocp.name AS status_kantor,
			c.bank_name, c.bank_account, c.bank_owner_company, c.siup, c.tdp, c.company_nib_path AS nib_path, c.deed_of_incorporation AS akta_pendirian, 
			c.latest_amendment_deed AS akta_perubahan_terahkir, 
			c.certificate_of_company_est AS sk_pendirian_perusahaan,
			c.sk_kumham, c.sk_kumham_last, c.sk_kumham_path, c.npwp, c.npwp_path, c.total_employees, 
			c.financial_statement AS laporan_keuangan, c.bank_statement AS rekening_koran
			FROM companies c 
			INNER JOIN type_of_companies toc ON toc.id = c.type_of_company
			INNER JOIN type_of_businesses tob ON tob.id = c.type_of_business
			INNER JOIN type_of_company_places tocp ON tocp.id = c.type_of_company_place
			WHERE user_id = ?`

			errEmitenCompany := dbDefault.Raw(queryEmitenCompany, adminListUser.Id).Scan(&adminEmitenCompany).Error
			if errEmitenCompany != nil {
				helper.Logger("error", "In Server: "+errEmitenCompany.Error())
				return nil, errors.New(errEmitenCompany.Error())
			}

			// Company Position

			queryEmitenPositionCompany := `SELECT title, name, position, ktp, ktp_path, npwp, npwp_path FROM positions WHERE company_id = ?`

			errEmitenPositionCompany := dbDefault.Raw(queryEmitenPositionCompany, adminEmitenCompany.Id).Scan(&adminEmitenPositions).Error

			if errEmitenPositionCompany != nil {
				helper.Logger("error", "In Server: "+errEmitenPositionCompany.Error())
				return nil, errors.New(errEmitenPositionCompany.Error())
			}

			// Company Project

			queryEmitenProjectCompany := `
			SELECT 
				p.uid AS id, 
				p.title, 
				top.name AS jenis_project, 
				p.is_apbn, 
				p.is_approved, 
				p.nominal_value AS jumlah_minimal,
				p.time_periode AS jangka_waktu, 
				p.interest_rate AS tingkat_bunga, 
				p.interest_payment_schedule AS jadwal_pembayaran_bunga, 
				p.principal_payment_schedule AS jadwal_pembayaran_pokok, 
				p.desc_job AS deskripsi_pekerjaan,
				p.company_profile, 
				p.collateral_guarantee AS jaminan_kolateral, 
				ps.name AS status,
				p.start_project, p.end_project,
				p.provider_address, p.provider_province_name, p.provider_city_name, p.provider_district_name,
				p.provider_subdistrict_name, p.provider_postal_code
			FROM 
				projects p
			INNER JOIN 
				project_statuses ps ON p.status = ps.id
			INNER JOIN
				type_of_projects top ON top.id = p.type_of_project
			WHERE 
				p.company_id = ?
		`

			errEmitenProjectCompany := dbDefault.Raw(queryEmitenProjectCompany, adminEmitenCompany.Id).Scan(&adminEmitenProjects).Error

			if errEmitenProjectCompany != nil {
				helper.Logger("error", "In Server: "+errEmitenProjectCompany.Error())
				return nil, errors.New(errEmitenProjectCompany.Error())
			}

			// Company Address

			queryEmitenAddressCompany := `SELECT ca.name, ca.address, ca.postal_code,
			ca.province_name AS province, ca.city_name AS city, ca.district_name AS district, 
			ca.subdistrict_name AS subdistrict
			FROM company_addresses ca 
			WHERE ca.company_id = ?`

			errEmitenAddressCompany := dbDefault.Raw(queryEmitenAddressCompany, adminEmitenCompany.Id).Scan(&adminEmitenAddress).Error

			if errEmitenAddressCompany != nil {
				helper.Logger("error", "In Server: "+errEmitenAddressCompany.Error())
				return nil, errors.New(errEmitenAddressCompany.Error())
			}

			// User Signature

			queryUserSignature := `SELECT path FROM user_signatures 
			WHERE user_id = ?`

			errQueryUserSignature := dbDefault.Raw(queryUserSignature, adminListUser.Id).Scan(&adminUserSignature).Error

			if errQueryUserSignature != nil {
				helper.Logger("error", "In Server: "+errQueryUserSignature.Error())
				return nil, errors.New(errQueryUserSignature.Error())
			}

			// User Slip Pay

			querySlipPay := `SELECT path FROM pay_slips 
			WHERE user_id = ?`

			errSlipPay := dbDefault.Raw(querySlipPay, adminListUser.Id).Scan(&adminUserSlipPay).Error

			if errSlipPay != nil {
				helper.Logger("error", "In Server: "+errSlipPay.Error())
				return nil, errors.New(errSlipPay.Error())
			}

			// User Security Account

			querySecurityAccount := `SELECT account_name, account_no, account_sub_no, account_bank
			FROM security_accounts 
			WHERE user_id = ?`

			errSecurityAccount := dbDefault.Raw(querySecurityAccount, adminListUser.Id).Scan(&adminUserSecurityAccount).Error

			if errSecurityAccount != nil {
				helper.Logger("error", "In Server: "+errSecurityAccount.Error())
				return nil, errors.New(errSecurityAccount.Error())
			}

			// User Risk

			queryUserRisk := `SELECT goal, tolerance, experience, capital_market_knowledge AS pengetahuan_pasar_modal FROM user_risks WHERE user_id = ?`

			errQueryUserRisk := dbDefault.Raw(queryUserRisk, adminListUser.Id).Scan(&adminUserRisk).Error

			if errQueryUserRisk != nil {
				helper.Logger("error", "In Server: "+errQueryUserRisk.Error())
				return nil, errors.New(errQueryUserRisk.Error())
			}

			for i := range adminEmitenProjects {

				var projectContract entities.AdminProjectContract

				var dataMedias []entities.ProjectMediaPath
				var dataProjectUseOfFunds []entities.AdminProjectUseOfFunds
				var dataProjectCollateralGuarantee []entities.AdminProjectCollateralGuarantee

				// Get Project Media

				errMedia := dbDefault.Raw(`SELECT path FROM project_medias WHERE project_id = ?`, adminEmitenProjects[i].Id).Scan(&dataMedias).Error
				if errMedia != nil {
					helper.Logger("error", "Load project medias error: "+errMedia.Error())
					return nil, errMedia
				}

				// Get Project Use Of Funds

				queryProjectUseOfFunds := `SELECT id, name FROM use_of_funds WHERE project_id = ?`

				errProjectUseOfFunds := dbDefault.
					Raw(queryProjectUseOfFunds, adminEmitenProjects[i].Id).
					Scan(&dataProjectUseOfFunds).Error

				if errProjectUseOfFunds != nil {
					helper.Logger("error", "In Server: "+errProjectUseOfFunds.Error())
					return nil, errProjectUseOfFunds
				}

				// Get Project Document Verify

				var result entities.ProjectDocumentVerify

				queryProjectDocumentVerify := `
				SELECT 
					id, skd, cv, rab, other_license_document,
					video_profile_company, project_summary,
					revenue_projection, work_of_timeline, annual_tax_report,
					list_of_employment, list_of_supplier_data,
					latest_receivable_account
				FROM document_verify_projects
				WHERE project_id = ?`

				if err := dbDefault.
					Raw(queryProjectDocumentVerify, adminEmitenProjects[i].Id).
					Scan(&result).Error; err != nil {
					log.Printf("FetchProjectDocumentVerify: project docs query error: %v", err)
					return nil, err
				}

				var mediaDocs []entities.MediaDocumentVerifyProject
				queryMedia := `
				SELECT path, type
				FROM media_document_verify_projects
				WHERE id = ?`

				if err := dbDefault.
					Raw(queryMedia, result.Id).
					Scan(&mediaDocs).Error; err != nil {
					log.Printf("FetchProjectDocumentVerify: media query error: %v", err)
					return nil, err
				}

				result.Media = mediaDocs

				// Get Project Contract

				queryProjectContract := `SELECT value, path FROM project_contracts WHERE project_id = ?`

				errProjectContract := dbDefault.
					Raw(queryProjectContract, adminEmitenProjects[i].Id).
					Scan(&projectContract).Error

				if errProjectContract != nil {
					helper.Logger("error", "In Server: "+errProjectContract.Error())
					return nil, errProjectContract
				}

				// Get Project Collateral Guarantees

				queryProjectCollateralGurantee := `SELECT id, name FROM collateral_guarantees WHERE project_id = ?`

				errProjectCollateralGurantee := dbDefault.
					Raw(queryProjectCollateralGurantee, adminEmitenProjects[i].Id).
					Scan(&dataProjectCollateralGuarantee).Error

				if errProjectCollateralGurantee != nil {
					helper.Logger("error", "In Server: "+errProjectCollateralGurantee.Error())
					return nil, errProjectCollateralGurantee
				}

				if dataProjectUseOfFunds == nil {
					dataProjectUseOfFunds = []entities.AdminProjectUseOfFunds{}
				}

				if dataProjectCollateralGuarantee == nil {
					dataProjectCollateralGuarantee = []entities.AdminProjectCollateralGuarantee{}
				}

				if dataMedias == nil {
					dataMedias = []entities.ProjectMediaPath{}
				}

				dataAdminEmitenProjects = append(dataAdminEmitenProjects, entities.AdminEmitenProjectCompanyResponse{
					Id:                     adminEmitenProjects[i].Id,
					Title:                  adminEmitenProjects[i].Title,
					JenisProject:           adminEmitenProjects[i].JenisProject,
					JumlahMinimal:          adminEmitenProjects[i].JumlahMinimal,
					JangkaWaktu:            adminEmitenProjects[i].JangkaWaktu,
					TingkatBunga:           adminEmitenProjects[i].TingkatBunga,
					JadwalPembayaranBunga:  adminEmitenProjects[i].JadwalPembayaranBunga,
					JadwalPembayaranPokok:  adminEmitenProjects[i].JadwalPembayaranPokok,
					DeskripsiPekerjaan:     adminEmitenProjects[i].DeskripsiPekerjaan,
					CompanyProfile:         adminEmitenProjects[i].CompanyProfile,
					PenggunaanData:         dataProjectUseOfFunds,
					JaminanKolateral:       dataProjectCollateralGuarantee,
					TenorPinjaman:          adminEmitenProjects[i].TenorPinjaman,
					DocumentVerify:         result,
					Kontrak:                projectContract,
					Website:                adminEmitenProjects[i].Website,
					IsApbn:                 adminEmitenProjects[i].IsApbn,
					IsApproved:             adminEmitenProjects[i].IsApproved,
					Status:                 adminEmitenProjects[i].Status,
					Medias:                 dataMedias,
					MulaiProject:           adminEmitenProjects[i].StartProject,
					SelesaiProject:         adminEmitenProjects[i].EndProject,
					AlamatPenyediaProject:  adminEmitenProjects[i].ProviderAddress,
					AlamatPenyediaProvinsi: adminEmitenProjects[i].ProviderProvinceName,
					AlamatPenyediaKota:     adminEmitenProjects[i].ProviderCityName,
					AlamatPenyediaDaerah:   adminEmitenProjects[i].ProviderDistrictName,
					AlamatPenyediaWilayah:  adminEmitenProjects[i].ProviderSubdistrictName,
					AlamatPenyediaKodePos:  adminEmitenProjects[i].ProviderPostalCode,
				})
			}

			if adminEmitenPositions == nil {
				adminEmitenPositions = []entities.AdminEmitenPositionCompany{}
			}

			if adminEmitenProjects == nil {
				adminEmitenProjects = []entities.AdminEmitenProjectCompany{}
			}

			if adminEmitenAddress == nil {
				adminEmitenAddress = []entities.AdminEmitenAddressCompany{}
			}

			dataAdminListUser = append(dataAdminListUser, entities.AdminListUserResponse{
				Id:               adminListUser.Id,
				Fullname:         helper.DefaultIfEmpty(adminListUser.Fullname, "-"),
				Avatar:           helper.DefaultIfEmpty(adminListUser.Avatar, "-"),
				Selfie:           helper.DefaultIfEmpty(adminListUser.Selfie, "-"),
				PhotoKtp:         helper.DefaultIfEmpty(adminListUser.PhotoKtp, "-"),
				NoKtp:            helper.DefaultIfEmpty(adminListUser.NoKtp, "-"),
				NoNpwp:           helper.DefaultIfEmpty(adminListUser.NoNpwp, "-"),
				Jabatan:          helper.DefaultIfEmpty(adminListUser.Position, "-"),
				Gender:           helper.DefaultIfEmpty(adminListUser.Gender, "-"),
				StatusMarital:    helper.DefaultIfEmpty(adminListUser.StatusMarital, "-"),
				LastEducation:    helper.DefaultIfEmpty(adminListUser.LastEducation, "-"),
				AddressDetail:    helper.DefaultIfEmpty(adminListUser.AddressDetail, "-"),
				Occupation:       helper.DefaultIfEmpty(adminListUser.Occupation, "-"),
				ProvinceName:     helper.DefaultIfEmpty(adminListUser.ProvinceName, "-"),
				CityName:         helper.DefaultIfEmpty(adminListUser.CityName, "-"),
				DistrictName:     helper.DefaultIfEmpty(adminListUser.DistrictName, "-"),
				SubdistrictName:  helper.DefaultIfEmpty(adminListUser.SubdistrictName, "-"),
				Email:            helper.DefaultIfEmpty(adminListUser.Email, "-"),
				Phone:            helper.DefaultIfEmpty(adminListUser.Phone, "-"),
				Role:             helper.DefaultIfEmpty(adminListUser.Role, "-"),
				Verified:         adminListUser.Verify,
				VerifiedEmiten:   adminListUser.VerifyEmiten,
				VerifiedInvestor: adminListUser.VerifyInvestor,
				NamaAhliWaris:    helper.DefaultIfEmpty(adminListUser.BeneficiaryName, "-"),
				PhoneAhliWaris:   helper.DefaultIfEmpty(adminListUser.BeneficiaryPhone, "-"),
				RekeningEfek: entities.RekeningEfek{
					AccountName:  helper.DefaultIfEmpty(adminUserSecurityAccount.AccountName, "-"),
					AccountNo:    helper.DefaultIfEmpty(adminUserSecurityAccount.AccountNo, "-"),
					AccountSubNo: helper.DefaultIfEmpty(adminUserSecurityAccount.AccountSubNo, "-"),
					AccountBank:  helper.DefaultIfEmpty(adminUserSecurityAccount.AccountBank, "-"),
				},
				SuratKuasa: entities.SuratKuasa{
					Type: helper.DefaultIfEmpty(adminAdditionalDoc.Type, "-"),
					Path: helper.DefaultIfEmpty(adminAdditionalDoc.Path, "-"),
				},
				Signature: entities.AdminUserSignature{
					Path: helper.DefaultIfEmpty(adminUserSignature.Path, "-"),
				},
				Risk: entities.AdminUserRisk{
					Goal:                  helper.DefaultIfEmpty(adminUserRisk.Goal, "-"),
					Tolerance:             helper.DefaultIfEmpty(adminUserRisk.Tolerance, "-"),
					Experience:            helper.DefaultIfEmpty(adminUserRisk.Experience, "-"),
					PengetahuanPasarModal: helper.DefaultIfEmpty(adminUserRisk.PengetahuanPasarModal, "-"),
				},
				Emiten: entities.AdminEmiten{
					Company: entities.AdminEmitenCompanyResponse{
						Id:                    helper.DefaultIfEmpty(adminEmitenCompany.Id, "-"),
						Name:                  helper.DefaultIfEmpty(adminEmitenCompany.Name, "-"),
						Nib:                   helper.DefaultIfEmpty(adminEmitenCompany.Nib, "-"),
						NibPath:               helper.DefaultIfEmpty(adminEmitenCompany.NibPath, "-"),
						AktaPendirian:         helper.DefaultIfEmpty(adminEmitenCompany.AktaPendirian, "-"),
						AktaPerubahanTerahkir: helper.DefaultIfEmpty(adminEmitenCompany.AktaPerubahanTerahkir, "-"),
						SkPendirianPerusahaan: helper.DefaultIfEmpty(adminEmitenCompany.SkPendirianPerusahaan, "-"),
						SkKumham:              helper.DefaultIfEmpty(adminEmitenCompany.SkKumham, "-"),
						SkKumhamTerahkir:      helper.DefaultIfEmpty(adminEmitenCompany.SkKumhamLast, "-"),
						SkKumhamPath:          helper.DefaultIfEmpty(adminEmitenCompany.SkKumhamPath, "-"),
						Npwp:                  helper.DefaultIfEmpty(adminEmitenCompany.Npwp, "-"),
						NpwpPath:              helper.DefaultIfEmpty(adminEmitenCompany.NpwpPath, "-"),
						TotalEmployees:        helper.DefaultIfEmpty(adminEmitenCompany.TotalEmployees, "-"),
						LaporanKeuangan:       helper.DefaultIfEmpty(adminEmitenCompany.LaporanKeuangan, "-"),
						RekeningKoran:         helper.DefaultIfEmpty(adminEmitenCompany.RekeningKoran, "-"),
						Site:                  helper.DefaultIfEmpty(adminEmitenCompany.Site, "-"),
						Email:                 helper.DefaultIfEmpty(adminEmitenCompany.Email, "-"),
						Phone:                 helper.DefaultIfEmpty(adminEmitenCompany.Phone, "-"),
						Est:                   helper.DefaultIfEmpty(adminEmitenCompany.Est, "-"),
						BankName:              helper.DefaultIfEmpty(adminEmitenCompany.BankName, "-"),
						BankAccount:           helper.DefaultIfEmpty(adminEmitenCompany.BankAccount, "-"),
						BankOwnerCompany:      helper.DefaultIfEmpty(adminEmitenCompany.BankOwnerCompany, "-"),
						Siup:                  helper.DefaultIfEmpty(adminEmitenCompany.Siup, "-"),
						Tdp:                   helper.DefaultIfEmpty(adminEmitenCompany.Tdp, "-"),
						JenisPerusahaan:       helper.DefaultIfEmpty(adminEmitenCompany.JenisPerusahaan, "-"),
						JenisUsaha:            helper.DefaultIfEmpty(adminEmitenCompany.JenisUsaha, "-"),
						StatusKantor:          helper.DefaultIfEmpty(adminEmitenCompany.StatusKantor, "-"),
						Address:               adminEmitenAddress,
						Positions:             adminEmitenPositions,
						Projects:              dataAdminEmitenProjects,
					},
				},
				Investor: entities.AdminInvestor{
					Ktp:      adminUserKtp,
					Bank:     adminUserBank,
					Job:      adminUserJob,
					SlipGaji: adminUserSlipPay,
				},
				CreatedAt: adminListUser.CreatedAt,
				UpdatedAt: adminListUser.UpdatedAt,
			})
		}

		dataAdminListProject = append(dataAdminListProject, entities.AdminListProjectResponse{
			Id:                    adminListProject.Id,
			Title:                 helper.DefaultIfEmpty(adminListProject.Title, "-"),
			Deskripsi:             helper.DefaultIfEmpty(adminListProject.DescJob, "-"),
			Modal:                 adminListProject.Capital,
			JumlahMinimal:         adminListProject.NominalValue,
			JadwalPembayaranBunga: helper.DefaultIfEmpty(adminListProject.InterestPaymentSchedule, "-"),
			JadwalPembayaranPokok: helper.DefaultIfEmpty(adminListProject.PrincipalPaymentSchedule, "-"),
			PersentaseKeuntungan:  helper.DefaultIfEmpty(adminListProject.ProfitPercentage, "-"),
			TingkatBunga:          helper.DefaultIfEmpty(adminListProject.InterestRate, "-"),
			JangkaWaktu:           helper.DefaultIfEmpty(adminListProject.TimePeriode, "-"),
			Spk:                   helper.DefaultIfEmpty(adminListProject.Spk, "-"),
			Loa:                   helper.DefaultIfEmpty(adminListProject.Loa, "-"),
			JenisProject:          helper.DefaultIfEmpty(adminListProject.TypeOfProject, "-"),
			BatasAkhirPengerjaan:  adminListProject.ExpireDate,
			PenggunaanDana:        dataProjectUseOfFunds,
			DanaYangDibutuhkan:    adminListProject.RequiredFund,
			JaminanKolateral:      dataProjectCollateralGuarantee,
			Website:               helper.DefaultIfEmpty(adminListProject.Site, "-"),
			TenorPinjaman:         helper.DefaultIfEmpty(adminListProject.LoanTerm, "-"),
			Kontrak:               projectContract,
			HargaPerlot:           adminListProject.MinInvest,
			JumlahLot:             adminListProject.AmountSharesPerLot,
			JumlahUnit:            adminListProject.NumberOfUnit,
			HargaPerlembar:        adminListProject.UnitPrice,
			UnitTotal:             adminListProject.UnitTotal,
			KodeEfek:              adminListProject.CodeEffect,
			Sku:                   adminListProject.Sku,
			IsApbn:                adminListProject.IsApbn,
			DocumentVerify:        result,
			BuktiPembayaran: entities.BuktiPembayaran{
				Path:      helper.DefaultIfEmpty(projectPayment.Path, "-"),
				IsApprove: projectPayment.IsApprove,
			},
			IsApproved:                  adminListProject.IsApproved,
			Status:                      adminListProject.Status,
			MulaiProject:                adminListProject.StartProject,
			SelesaiProject:              adminListProject.EndProject,
			AlamatPenyediaProject:       adminListProject.ProviderAddress,
			AlamatPenyediaProvinsi:      adminListProject.ProviderProvinceName,
			AlamatPenyediaKota:          adminListProject.ProviderCityName,
			AlamatPenyediaDaerah:        adminListProject.ProviderDistrictName,
			AlamatPenyediaWilayah:       adminListProject.ProviderSubdistrictName,
			AlamatPenyediaKodePos:       adminListProject.ProviderPostalCode,
			DocRekeningKoran:            adminListProject.DocBankStatement,
			DocLaporanKeuangan:          adminListProject.DocFinancialStatement,
			DocContract:                 adminListProject.DocContract,
			DocProspect:                 adminListProject.DocProspect,
			JenisInstansiPemberiProject: adminListProject.TypeOfContractingAuthority,
			InstansiPemberiProject:      adminListProject.ContractingAuthority,
			Company: entities.AdminListCompany{
				Id:          dataProjectCompany.Id,
				Name:        dataProjectCompany.CompanyName,
				Address:     dataProjectCompany.Address,
				Province:    dataProjectCompany.ProvinceName,
				City:        dataProjectCompany.CityName,
				District:    dataProjectCompany.DistrictName,
				Subdistrict: dataProjectCompany.SubdistrictName,
			},
			Location: entities.AdminProjectLocation{
				Id:   dataProjectLoc.Id,
				Name: dataProjectLoc.Name,
				Url:  dataProjectLoc.Url,
				Lat:  dataProjectLoc.Lat,
				Lng:  dataProjectLoc.Lng,
			},
			Media:     dataProjectMedia,
			User:      dataAdminListUser[0],
			CreatedAt: adminListProject.CreatedAt,
		})

	}

	var nextUrl = strconv.Itoa(nextPage)
	var prevUrl = strconv.Itoa(prevPage)

	if dataAdminListProject == nil {
		dataAdminListProject = []entities.AdminListProjectResponse{}
	}

	return map[string]any{
		"total":        resultTotal,
		"current_page": pageinteger,
		"per_page":     int(perPage),
		"prev_page":    prevPage,
		"next_page":    nextPage,
		"next_url":     url + "/api/v1/admin/list/project?page=" + nextUrl,
		"prev_url":     url + "/api/v1/admin/list/project?page=" + prevUrl,
		"data":         dataAdminListProject,
	}, nil
}

func AdminDetailProject(id string) (map[string]any, error) {
	var adminProject entities.AdminListProject
	var projectContract entities.AdminProjectContract

	query := `
		SELECT p.uid AS id, p.title, p.goal, p.capital, p.min_invest,
		p.unit_price, p.unit_total, p.number_of_unit, p.periode, top.name AS type_of_project,
		p.nominal_value, p.time_periode, p.interest_rate, p.interest_payment_schedule, p.principal_payment_schedule,
		p.use_of_funds, p.required_fund, p.collateral_guarantee, p.profit_percentage, p.desc_job, ps.name AS status, p.is_apbn, p.is_approved,
		u.uid AS user_id, u.email AS user_email, u.phone AS user_phone, 
		p.spk, p.loa,
		p.unit_price, p.unit_total, p.min_invest, p.number_of_unit,
		p.code_effect,
		p.created_at,
		p.sku,
		toc.name AS type_of_contracting_authority,
		p.site, p.loan_term,
		p.start_project, p.end_project,
		p.doc_bank_statement, 
		p.doc_financial_statement, 
		p.doc_contract,
		p.doc_prospect,
		p.amount_shares_per_lot,
		p.contracting_authority,
		p.provider_address, 
		p.provider_province_name, p.provider_city_name, 
		p.provider_district_name, p.provider_subdistrict_name, p.provider_postal_code
		FROM projects p
		INNER JOIN companies c ON c.uid = p.company_id
		INNER JOIN users u ON u.uid = c.user_id
		INNER JOIN project_statuses ps ON ps.id = p.status
		INNER JOIN type_of_projects top ON top.id = p.type_of_project
		INNER JOIN type_of_contracting_authorities toc ON toc.id = p.type_of_contracting_authority  
		WHERE p.uid = ?`

	err := dbDefault.Raw(query, id).Scan(&adminProject).Error
	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	if adminProject.Id == "" {
		return nil, errors.New("PROJECT_NOT_FOUND")
	}

	// Project Media

	dataProjectMedia := make([]entities.AdminListMedia, 0)

	queryProjectMedia := `SELECT id, path FROM project_medias WHERE project_id = ?`

	errProjectMedia := dbDefault.
		Raw(queryProjectMedia, adminProject.Id).
		Scan(&dataProjectMedia).Error

	if errProjectMedia != nil {
		helper.Logger("error", "In Server: "+errProjectMedia.Error())
		return nil, errProjectMedia
	}

	// Project Location

	var projectLoc entities.ProjectLocation
	if err := dbDefault.
		Raw(`SELECT id, url, name, lat, lng FROM project_locations WHERE project_id = ?`, adminProject.Id).
		Scan(&projectLoc).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	// Project Company

	var projectCompany entities.ProjectCompany
	if err := dbDefault.
		Raw(`SELECT c.uid AS id, c.company_name,
		ca.address, ca.postal_code, ca.province_name, ca.city_name, 
		ca.district_name, ca.subdistrict_name
		FROM companies c 
		INNER JOIN company_addresses ca ON c.uid = ca.company_id
		WHERE c.user_id = ?`, adminProject.UserId).
		Scan(&projectCompany).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	// Project Use Of Funds

	var dataProjectUseOfFunds []entities.AdminProjectUseOfFunds

	queryProjectUseOfFunds := `SELECT id, name FROM use_of_funds WHERE project_id = ?`

	errProjectUseOfFunds := dbDefault.
		Raw(queryProjectUseOfFunds, adminProject.Id).
		Scan(&dataProjectUseOfFunds).Error

	if errProjectUseOfFunds != nil {
		helper.Logger("error", "In Server: "+errProjectUseOfFunds.Error())
		return nil, errProjectUseOfFunds
	}

	// Project Location

	var dataProjectLoc entities.ProjectLocation

	queryProjectLoc := `SELECT id, url, name, lat, lng FROM project_locations WHERE project_id = ?`

	errProjectLoc := dbDefault.
		Raw(queryProjectLoc, adminProject.Id).
		Scan(&dataProjectLoc).Error

	if errProjectLoc != nil {
		helper.Logger("error", "In Server: "+errProjectLoc.Error())
		return nil, errProjectLoc
	}

	// Project Contract

	queryProjectContract := `SELECT value, path 
	FROM project_contracts WHERE project_id = ?`

	errProjectContract := dbDefault.
		Raw(queryProjectContract, adminProject.Id).
		Scan(&projectContract).Error

	if errProjectContract != nil {
		helper.Logger("error", "In Server: "+errProjectContract.Error())
		return nil, errProjectContract
	}

	if dataProjectUseOfFunds == nil {
		dataProjectUseOfFunds = []entities.AdminProjectUseOfFunds{}
	}

	// Project Collateral Guarantee

	var dataProjectCollateralGuarantee []entities.AdminProjectCollateralGuarantee

	queryProjectCollateralGurantee := `SELECT id, name FROM collateral_guarantees WHERE project_id = ?`

	errProjectCollateralGurantee := dbDefault.
		Raw(queryProjectCollateralGurantee, adminProject.Id).
		Scan(&dataProjectCollateralGuarantee).Error

	if errProjectCollateralGurantee != nil {
		helper.Logger("error", "In Server: "+errProjectCollateralGurantee.Error())
		return nil, errProjectCollateralGurantee
	}

	if dataProjectCollateralGuarantee == nil {
		dataProjectCollateralGuarantee = []entities.AdminProjectCollateralGuarantee{}
	}

	// Project Document Verify

	var result entities.ProjectDocumentVerify

	queryProjectDocumentVerify := `
		SELECT 
			id, skd, cv, rab, other_license_document,
			cashflow_project,
			video_profile_company, project_summary,
			revenue_projection, work_of_timeline, annual_tax_report,
			list_of_employment, list_of_supplier_data,
			latest_receivable_account
		FROM document_verify_projects
		WHERE project_id = ?`

	if err := dbDefault.
		Raw(queryProjectDocumentVerify, adminProject.Id).
		Scan(&result).Error; err != nil {
		log.Printf("FetchProjectDocumentVerify: project docs query error: %v", err)
		return nil, err
	}

	var mediaDocs []entities.MediaDocumentVerifyProject
	queryMedia := `
		SELECT path, type
		FROM media_document_verify_projects
		WHERE document_verify_project_id = ?`

	if err := dbDefault.
		Raw(queryMedia, result.Id).
		Scan(&mediaDocs).Error; err != nil {
		log.Printf("FetchProjectDocumentVerify: media query error: %v", err)
		return nil, err
	}

	result.Media = mediaDocs

	// Project Payment

	var projectPayment entities.ProjectPayment

	queryProjectPayment := `SELECT path, is_approve 
		FROM project_payments WHERE project_id = ?`

	errProjectPayment := dbDefault.
		Raw(queryProjectPayment, adminProject.Id).
		Scan(&projectPayment).Error

	if errProjectPayment != nil {
		helper.Logger("error", "In Server: "+errProjectPayment.Error())
		return nil, errProjectPayment
	}

	// Project User

	var dataAdminListUser []entities.AdminListUserResponse
	var adminListUser entities.AdminListUser

	userQuery := `SELECT u.uid AS id, p.fullname, p.avatar, u.email, u.phone,
	p.selfie, p.photo_ktp, p.no_ktp, p.no_npwp, p.position, p.gender, 
	p.province_name, p.city_name, p.district_name, p.subdistrict_name, 
	p.status_marital, p.occupation, p.last_education, p.address_detail,
	p.beneficiary_name, p.beneficiary_phone, r.name AS role, 
	u.created_at, u.updated_at, u.verify, u.verify_emiten, u.verify_investor
	FROM users u 
	INNER JOIN roles r ON r.id = u.role
	INNER JOIN profiles p ON u.uid = p.user_id
	WHERE u.uid = ?
	ORDER BY u.created_at DESC`

	rows, err := dbDefault.Raw(userQuery, adminProject.UserId).Rows()

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	for rows.Next() {
		errAdminListUserRows := dbDefault.ScanRows(rows, &adminListUser)
		if errAdminListUserRows != nil {
			helper.Logger("error", "In Server: "+errAdminListUserRows.Error())
			return nil, errors.New(errAdminListUserRows.Error())
		}

		var adminUserKtp entities.AdminUserKtp
		var adminUserJob entities.AdminUserJob
		var adminUserBank entities.AdminUserBank
		var adminAdditionalDoc entities.AdditionalDoc
		var adminUserSignature entities.AdminUserSignature
		var adminUserSecurityAccount entities.AdminUserSecurityAccount
		var adminUserRisk entities.AdminUserRisk

		var adminUserSlipPay entities.AdminUserSlipPay

		var adminEmitenCompany entities.AdminEmitenCompany
		var adminEmitenPositions []entities.AdminEmitenPositionCompany

		var adminEmitenProjects []entities.AdminEmitenProjectCompany

		var adminEmitenAddress []entities.AdminEmitenAddressCompany

		dataAdminEmitenProjects := make([]entities.AdminEmitenProjectCompanyResponse, 0)

		// Ktp

		queryKtp := `SELECT name, nik, place_datebirth, path FROM ktps WHERE user_id = ?`

		errKtp := dbDefault.Raw(queryKtp, adminListUser.Id).Scan(&adminUserKtp).Error
		if errKtp != nil {
			helper.Logger("error", "In Server: "+errKtp.Error())
			return nil, errors.New(errKtp.Error())
		}

		// Job

		queryJob := `SELECT company_name, province_name, city_name, district_name, subdistrict_name, company_address, monthly_income, annual_income_num AS annual_income, 
		position, npwp, npwp_path 
		FROM jobs WHERE user_id = ?`

		errJob := dbDefault.Raw(queryJob, adminListUser.Id).Scan(&adminUserJob).Error
		if errJob != nil {
			helper.Logger("error", "In Server: "+errJob.Error())
			return nil, errors.New(errJob.Error())
		}

		// Bank

		queryBank := `SELECT no, bank_name AS name, bank_owner AS owner, bank_branch AS branch, rek_koran_path FROM accounts WHERE user_id = ?`

		errBank := dbDefault.Raw(queryBank, adminListUser.Id).Scan(&adminUserBank).Error
		if errBank != nil {
			helper.Logger("error", "In Server: "+errBank.Error())
			return nil, errors.New(errBank.Error())
		}

		// Additional Doc

		queryAdditionalDoc := `SELECT type, path, user_id FROM additional_docs WHERE user_id = ?`

		errAdditionalDoc := dbDefault.Raw(queryAdditionalDoc, adminListUser.Id).Scan(&adminAdditionalDoc).Error
		if errAdditionalDoc != nil {
			helper.Logger("error", "In Server: "+errAdditionalDoc.Error())
			return nil, errors.New(errAdditionalDoc.Error())
		}

		// Company

		queryEmitenCompany := `SELECT c.uid AS id, c.company_name AS name, c.company_nib AS nib, c.site, c.email, c.phone, c.est, 
		toc.name AS jenis_perusahaan,
		tob.name AS jenis_usaha,
		tocp.name AS status_kantor,
		c.bank_name, c.bank_account, c.bank_owner_company, c.siup, c.tdp, c.company_nib_path AS nib_path, c.deed_of_incorporation AS akta_pendirian, 
		c.latest_amendment_deed AS akta_perubahan_terahkir, 
		c.certificate_of_company_est AS sk_pendirian_perusahaan,
		c.sk_kumham, c.sk_kumham_last, c.sk_kumham_path, c.npwp, c.npwp_path, c.total_employees, 
		c.financial_statement AS laporan_keuangan, c.bank_statement AS rekening_koran
 		FROM companies c 
		INNER JOIN type_of_companies toc ON toc.id = c.type_of_company
		INNER JOIN type_of_businesses tob ON tob.id = c.type_of_business
		INNER JOIN type_of_company_places tocp ON tocp.id = c.type_of_company_place
		WHERE user_id = ?`

		errEmitenCompany := dbDefault.Raw(queryEmitenCompany, adminListUser.Id).Scan(&adminEmitenCompany).Error
		if errEmitenCompany != nil {
			helper.Logger("error", "In Server: "+errEmitenCompany.Error())
			return nil, errors.New(errEmitenCompany.Error())
		}

		// Company Position

		queryEmitenPositionCompany := `SELECT title, name, position, ktp, ktp_path, npwp, npwp_path FROM positions WHERE company_id = ?`

		errEmitenPositionCompany := dbDefault.Raw(queryEmitenPositionCompany, adminEmitenCompany.Id).Scan(&adminEmitenPositions).Error

		if errEmitenPositionCompany != nil {
			helper.Logger("error", "In Server: "+errEmitenPositionCompany.Error())
			return nil, errors.New(errEmitenPositionCompany.Error())
		}

		// Company Project

		queryEmitenProjectCompany := `
			SELECT 
				p.uid AS id, 
				p.title, 
				top.name AS jenis_project, 
				p.is_apbn, 
				p.is_approved, 
				p.nominal_value AS jumlah_minimal,
				p.time_periode AS jangka_waktu, 
				p.interest_rate AS tingkat_bunga, 
				p.interest_payment_schedule AS jadwal_pembayaran_bunga, 
				p.principal_payment_schedule AS jadwal_pembayaran_pokok, 
				p.desc_job AS deskripsi_pekerjaan,
				p.company_profile, 
				p.collateral_guarantee AS jaminan_kolateral, 
				ps.name AS status,
				p.start_project, p.end_project,
				p.provider_address, p.provider_province_name, p.provider_city_name, p.provider_district_name,
				p.provider_subdistrict_name, p.provider_postal_code
			FROM 
				projects p
			INNER JOIN 
				project_statuses ps ON p.status = ps.id
			INNER JOIN
				type_of_projects top ON top.id = p.type_of_project
			WHERE 
				p.company_id = ?
		`

		errEmitenProjectCompany := dbDefault.Raw(queryEmitenProjectCompany, adminEmitenCompany.Id).Scan(&adminEmitenProjects).Error

		if errEmitenProjectCompany != nil {
			helper.Logger("error", "In Server: "+errEmitenProjectCompany.Error())
			return nil, errors.New(errEmitenProjectCompany.Error())
		}

		// Company Address

		queryEmitenAddressCompany := `SELECT ca.name, ca.address, ca.postal_code,
		ca.province_name AS province, ca.city_name AS city, ca.district_name AS district, 
		ca.subdistrict_name AS subdistrict
		FROM company_addresses ca 
		WHERE ca.company_id = ?`

		errEmitenAddressCompany := dbDefault.Raw(queryEmitenAddressCompany, adminEmitenCompany.Id).Scan(&adminEmitenAddress).Error

		if errEmitenAddressCompany != nil {
			helper.Logger("error", "In Server: "+errEmitenAddressCompany.Error())
			return nil, errors.New(errEmitenAddressCompany.Error())
		}

		// User Signature

		queryUserSignature := `SELECT path FROM user_signatures 
		WHERE user_id = ?`

		errQueryUserSignature := dbDefault.Raw(queryUserSignature, adminListUser.Id).Scan(&adminUserSignature).Error

		if errQueryUserSignature != nil {
			helper.Logger("error", "In Server: "+errQueryUserSignature.Error())
			return nil, errors.New(errQueryUserSignature.Error())
		}

		// User Slip Pay

		querySlipPay := `SELECT path FROM pay_slips 
		WHERE user_id = ?`

		errSlipPay := dbDefault.Raw(querySlipPay, adminListUser.Id).Scan(&adminUserSlipPay).Error

		if errSlipPay != nil {
			helper.Logger("error", "In Server: "+errSlipPay.Error())
			return nil, errors.New(errSlipPay.Error())
		}

		// User Security Account

		querySecurityAccount := `SELECT account_name, account_no, account_sub_no, account_bank
		FROM security_accounts 
		WHERE user_id = ?`

		errSecurityAccount := dbDefault.Raw(querySecurityAccount, adminListUser.Id).Scan(&adminUserSecurityAccount).Error

		if errSecurityAccount != nil {
			helper.Logger("error", "In Server: "+errSecurityAccount.Error())
			return nil, errors.New(errSecurityAccount.Error())
		}

		// User Risk

		queryUserRisk := `SELECT goal, tolerance, experience, capital_market_knowledge AS pengetahuan_pasar_modal FROM user_risks WHERE user_id = ?`

		errQueryUserRisk := dbDefault.Raw(queryUserRisk, adminListUser.Id).Scan(&adminUserRisk).Error

		if errQueryUserRisk != nil {
			helper.Logger("error", "In Server: "+errQueryUserRisk.Error())
			return nil, errors.New(errQueryUserRisk.Error())
		}

		for i := range adminEmitenProjects {

			var projectContract entities.AdminProjectContract

			var dataMedias []entities.ProjectMediaPath
			var dataProjectUseOfFunds []entities.AdminProjectUseOfFunds
			var dataProjectCollateralGuarantee []entities.AdminProjectCollateralGuarantee

			// Get Project Media

			errMedia := dbDefault.Raw(`SELECT path FROM project_medias WHERE project_id = ?`, adminEmitenProjects[i].Id).Scan(&dataMedias).Error
			if errMedia != nil {
				helper.Logger("error", "Load project medias error: "+errMedia.Error())
				return nil, errMedia
			}

			// Get Project Use Of Funds

			queryProjectUseOfFunds := `SELECT id, name FROM use_of_funds WHERE project_id = ?`

			errProjectUseOfFunds := dbDefault.
				Raw(queryProjectUseOfFunds, adminEmitenProjects[i].Id).
				Scan(&dataProjectUseOfFunds).Error

			if errProjectUseOfFunds != nil {
				helper.Logger("error", "In Server: "+errProjectUseOfFunds.Error())
				return nil, errProjectUseOfFunds
			}

			// Get Project Document Verify

			var result entities.ProjectDocumentVerify

			queryProjectDocumentVerify := `SELECT 
			id, skd, cv, rab, other_license_document,
			video_profile_company, project_summary,
			revenue_projection, work_of_timeline, annual_tax_report,
			list_of_employment, list_of_supplier_data,
			latest_receivable_account
			FROM document_verify_projects
			WHERE project_id = ?`

			if err := dbDefault.
				Raw(queryProjectDocumentVerify, adminEmitenProjects[i].Id).
				Scan(&result).Error; err != nil {
				log.Printf("FetchProjectDocumentVerify: project docs query error: %v", err)
				return nil, err
			}

			var mediaDocs []entities.MediaDocumentVerifyProject
			queryMedia := `SELECT path, type
			FROM media_document_verify_projects
			WHERE id = ?`

			if err := dbDefault.
				Raw(queryMedia, result.Id).
				Scan(&mediaDocs).Error; err != nil {
				log.Printf("FetchProjectDocumentVerify: media query error: %v", err)
				return nil, err
			}

			result.Media = mediaDocs

			// Get Project Contract

			queryProjectContract := `SELECT value, path FROM project_contracts WHERE project_id = ?`

			errProjectContract := dbDefault.
				Raw(queryProjectContract, adminEmitenProjects[i].Id).
				Scan(&projectContract).Error

			if errProjectContract != nil {
				helper.Logger("error", "In Server: "+errProjectContract.Error())
				return nil, errProjectContract
			}

			// Get Project Collateral Guarantees

			queryProjectCollateralGurantee := `SELECT id, name FROM collateral_guarantees WHERE project_id = ?`

			errProjectCollateralGurantee := dbDefault.
				Raw(queryProjectCollateralGurantee, adminEmitenProjects[i].Id).
				Scan(&dataProjectCollateralGuarantee).Error

			if errProjectCollateralGurantee != nil {
				helper.Logger("error", "In Server: "+errProjectCollateralGurantee.Error())
				return nil, errProjectCollateralGurantee
			}

			if dataProjectUseOfFunds == nil {
				dataProjectUseOfFunds = []entities.AdminProjectUseOfFunds{}
			}

			if dataProjectCollateralGuarantee == nil {
				dataProjectCollateralGuarantee = []entities.AdminProjectCollateralGuarantee{}
			}

			if dataMedias == nil {
				dataMedias = []entities.ProjectMediaPath{}
			}

			dataAdminEmitenProjects = append(dataAdminEmitenProjects, entities.AdminEmitenProjectCompanyResponse{
				Id:                     adminEmitenProjects[i].Id,
				Title:                  adminEmitenProjects[i].Title,
				JenisProject:           adminEmitenProjects[i].JenisProject,
				JumlahMinimal:          adminEmitenProjects[i].JumlahMinimal,
				JangkaWaktu:            adminEmitenProjects[i].JangkaWaktu,
				TingkatBunga:           adminEmitenProjects[i].TingkatBunga,
				JadwalPembayaranBunga:  adminEmitenProjects[i].JadwalPembayaranBunga,
				JadwalPembayaranPokok:  adminEmitenProjects[i].JadwalPembayaranPokok,
				DeskripsiPekerjaan:     adminEmitenProjects[i].DeskripsiPekerjaan,
				CompanyProfile:         adminEmitenProjects[i].CompanyProfile,
				PenggunaanData:         dataProjectUseOfFunds,
				JaminanKolateral:       dataProjectCollateralGuarantee,
				TenorPinjaman:          adminEmitenProjects[i].TenorPinjaman,
				DocumentVerify:         result,
				Kontrak:                projectContract,
				Website:                adminEmitenProjects[i].Website,
				IsApbn:                 adminEmitenProjects[i].IsApbn,
				IsApproved:             adminEmitenProjects[i].IsApproved,
				Status:                 adminEmitenProjects[i].Status,
				Medias:                 dataMedias,
				MulaiProject:           adminEmitenProjects[i].StartProject,
				SelesaiProject:         adminEmitenProjects[i].EndProject,
				AlamatPenyediaProject:  adminEmitenProjects[i].ProviderAddress,
				AlamatPenyediaProvinsi: adminEmitenProjects[i].ProviderProvinceName,
				AlamatPenyediaKota:     adminEmitenProjects[i].ProviderCityName,
				AlamatPenyediaDaerah:   adminEmitenProjects[i].ProviderDistrictName,
				AlamatPenyediaWilayah:  adminEmitenProjects[i].ProviderSubdistrictName,
				AlamatPenyediaKodePos:  adminEmitenProjects[i].ProviderPostalCode,
			})
		}

		if adminEmitenPositions == nil {
			adminEmitenPositions = []entities.AdminEmitenPositionCompany{}
		}

		if adminEmitenProjects == nil {
			adminEmitenProjects = []entities.AdminEmitenProjectCompany{}
		}

		if adminEmitenAddress == nil {
			adminEmitenAddress = []entities.AdminEmitenAddressCompany{}
		}

		dataAdminListUser = append(dataAdminListUser, entities.AdminListUserResponse{
			Id:               adminListUser.Id,
			Fullname:         helper.DefaultIfEmpty(adminListUser.Fullname, "-"),
			Avatar:           helper.DefaultIfEmpty(adminListUser.Avatar, "-"),
			Selfie:           helper.DefaultIfEmpty(adminListUser.Selfie, "-"),
			PhotoKtp:         helper.DefaultIfEmpty(adminListUser.PhotoKtp, "-"),
			NoKtp:            helper.DefaultIfEmpty(adminListUser.NoKtp, "-"),
			NoNpwp:           helper.DefaultIfEmpty(adminListUser.NoNpwp, "-"),
			Jabatan:          helper.DefaultIfEmpty(adminListUser.Position, "-"),
			Gender:           helper.DefaultIfEmpty(adminListUser.Gender, "-"),
			StatusMarital:    helper.DefaultIfEmpty(adminListUser.StatusMarital, "-"),
			LastEducation:    helper.DefaultIfEmpty(adminListUser.LastEducation, "-"),
			AddressDetail:    helper.DefaultIfEmpty(adminListUser.AddressDetail, "-"),
			Occupation:       helper.DefaultIfEmpty(adminListUser.Occupation, "-"),
			ProvinceName:     helper.DefaultIfEmpty(adminListUser.ProvinceName, "-"),
			CityName:         helper.DefaultIfEmpty(adminListUser.CityName, "-"),
			DistrictName:     helper.DefaultIfEmpty(adminListUser.DistrictName, "-"),
			SubdistrictName:  helper.DefaultIfEmpty(adminListUser.SubdistrictName, "-"),
			Email:            helper.DefaultIfEmpty(adminListUser.Email, "-"),
			Phone:            helper.DefaultIfEmpty(adminListUser.Phone, "-"),
			Role:             helper.DefaultIfEmpty(adminListUser.Role, "-"),
			Verified:         adminListUser.Verify,
			VerifiedEmiten:   adminListUser.VerifyEmiten,
			VerifiedInvestor: adminListUser.VerifyInvestor,
			NamaAhliWaris:    helper.DefaultIfEmpty(adminListUser.BeneficiaryName, "-"),
			PhoneAhliWaris:   helper.DefaultIfEmpty(adminListUser.BeneficiaryPhone, "-"),
			RekeningEfek: entities.RekeningEfek{
				AccountName:  helper.DefaultIfEmpty(adminUserSecurityAccount.AccountName, "-"),
				AccountNo:    helper.DefaultIfEmpty(adminUserSecurityAccount.AccountNo, "-"),
				AccountSubNo: helper.DefaultIfEmpty(adminUserSecurityAccount.AccountSubNo, "-"),
				AccountBank:  helper.DefaultIfEmpty(adminUserSecurityAccount.AccountBank, "-"),
			},
			SuratKuasa: entities.SuratKuasa{
				Type: helper.DefaultIfEmpty(adminAdditionalDoc.Type, "-"),
				Path: helper.DefaultIfEmpty(adminAdditionalDoc.Path, "-"),
			},
			Signature: entities.AdminUserSignature{
				Path: helper.DefaultIfEmpty(adminUserSignature.Path, "-"),
			},
			Risk: entities.AdminUserRisk{
				Goal:                  helper.DefaultIfEmpty(adminUserRisk.Goal, "-"),
				Tolerance:             helper.DefaultIfEmpty(adminUserRisk.Tolerance, "-"),
				Experience:            helper.DefaultIfEmpty(adminUserRisk.Experience, "-"),
				PengetahuanPasarModal: helper.DefaultIfEmpty(adminUserRisk.PengetahuanPasarModal, "-"),
			},
			Emiten: entities.AdminEmiten{
				Company: entities.AdminEmitenCompanyResponse{
					Id:                    helper.DefaultIfEmpty(adminEmitenCompany.Id, "-"),
					Name:                  helper.DefaultIfEmpty(adminEmitenCompany.Name, "-"),
					Nib:                   helper.DefaultIfEmpty(adminEmitenCompany.Nib, "-"),
					NibPath:               helper.DefaultIfEmpty(adminEmitenCompany.NibPath, "-"),
					AktaPendirian:         helper.DefaultIfEmpty(adminEmitenCompany.AktaPendirian, "-"),
					AktaPerubahanTerahkir: helper.DefaultIfEmpty(adminEmitenCompany.AktaPerubahanTerahkir, "-"),
					SkPendirianPerusahaan: helper.DefaultIfEmpty(adminEmitenCompany.SkPendirianPerusahaan, "-"),
					SkKumham:              helper.DefaultIfEmpty(adminEmitenCompany.SkKumham, "-"),
					SkKumhamTerahkir:      helper.DefaultIfEmpty(adminEmitenCompany.SkKumhamLast, "-"),
					SkKumhamPath:          helper.DefaultIfEmpty(adminEmitenCompany.SkKumhamPath, "-"),
					Npwp:                  helper.DefaultIfEmpty(adminEmitenCompany.Npwp, "-"),
					NpwpPath:              helper.DefaultIfEmpty(adminEmitenCompany.NpwpPath, "-"),
					TotalEmployees:        helper.DefaultIfEmpty(adminEmitenCompany.TotalEmployees, "-"),
					LaporanKeuangan:       helper.DefaultIfEmpty(adminEmitenCompany.LaporanKeuangan, "-"),
					RekeningKoran:         helper.DefaultIfEmpty(adminEmitenCompany.RekeningKoran, "-"),
					Site:                  helper.DefaultIfEmpty(adminEmitenCompany.Site, "-"),
					Email:                 helper.DefaultIfEmpty(adminEmitenCompany.Email, "-"),
					Phone:                 helper.DefaultIfEmpty(adminEmitenCompany.Phone, "-"),
					Est:                   helper.DefaultIfEmpty(adminEmitenCompany.Est, "-"),
					BankName:              helper.DefaultIfEmpty(adminEmitenCompany.BankName, "-"),
					BankAccount:           helper.DefaultIfEmpty(adminEmitenCompany.BankAccount, "-"),
					BankOwnerCompany:      helper.DefaultIfEmpty(adminEmitenCompany.BankOwnerCompany, "-"),
					Siup:                  helper.DefaultIfEmpty(adminEmitenCompany.Siup, "-"),
					Tdp:                   helper.DefaultIfEmpty(adminEmitenCompany.Tdp, "-"),
					JenisPerusahaan:       helper.DefaultIfEmpty(adminEmitenCompany.JenisPerusahaan, "-"),
					JenisUsaha:            helper.DefaultIfEmpty(adminEmitenCompany.JenisUsaha, "-"),
					StatusKantor:          helper.DefaultIfEmpty(adminEmitenCompany.StatusKantor, "-"),
					Address:               adminEmitenAddress,
					Positions:             adminEmitenPositions,
					Projects:              dataAdminEmitenProjects,
				},
			},
			Investor: entities.AdminInvestor{
				Ktp:      adminUserKtp,
				Bank:     adminUserBank,
				Job:      adminUserJob,
				SlipGaji: adminUserSlipPay,
			},
			CreatedAt: adminListUser.CreatedAt,
			UpdatedAt: adminListUser.UpdatedAt,
		})
	}

	response := entities.AdminListProjectResponse{
		Id:                    adminProject.Id,
		Title:                 helper.DefaultIfEmpty(adminProject.Title, "-"),
		Deskripsi:             helper.DefaultIfEmpty(adminProject.DescJob, "-"),
		Modal:                 adminProject.Capital,
		PersentaseKeuntungan:  adminProject.ProfitPercentage,
		Spk:                   helper.DefaultIfEmpty(adminProject.Spk, "-"),
		Loa:                   helper.DefaultIfEmpty(adminProject.Loa, "-"),
		BatasAkhirPengerjaan:  adminProject.ExpireDate,
		JenisProject:          helper.DefaultIfEmpty(adminProject.TypeOfProject, "-"),
		JumlahMinimal:         adminProject.NominalValue,
		JangkaWaktu:           helper.DefaultIfEmpty(adminProject.TimePeriode, "-"),
		TingkatBunga:          helper.DefaultIfEmpty(adminProject.InterestRate, "-"),
		JadwalPembayaranBunga: helper.DefaultIfEmpty(adminProject.InterestPaymentSchedule, "-"),
		JadwalPembayaranPokok: helper.DefaultIfEmpty(adminProject.PrincipalPaymentSchedule, "-"),
		PenggunaanDana:        dataProjectUseOfFunds,
		DanaYangDibutuhkan:    adminProject.RequiredFund,
		JaminanKolateral:      dataProjectCollateralGuarantee,
		Kontrak:               projectContract,
		Website:               helper.DefaultIfEmpty(adminProject.Site, "-"),
		TenorPinjaman:         helper.DefaultIfEmpty(adminProject.LoanTerm, "-"),
		HargaPerlot:           adminProject.MinInvest,
		JumlahUnit:            adminProject.NumberOfUnit,
		HargaPerlembar:        adminProject.UnitPrice,
		JumlahLot:             adminProject.AmountSharesPerLot,
		UnitTotal:             adminProject.UnitTotal,
		KodeEfek:              adminProject.CodeEffect,
		Sku:                   adminProject.Sku,
		IsApbn:                adminProject.IsApbn,
		DocumentVerify:        result,
		BuktiPembayaran: entities.BuktiPembayaran{
			Path:      helper.DefaultIfEmpty(projectPayment.Path, "-"),
			IsApprove: projectPayment.IsApprove,
		},
		IsApproved:                  adminProject.IsApproved,
		Status:                      adminProject.Status,
		MulaiProject:                adminProject.StartProject,
		SelesaiProject:              adminProject.EndProject,
		AlamatPenyediaProject:       adminProject.ProviderAddress,
		AlamatPenyediaProvinsi:      adminProject.ProviderProvinceName,
		AlamatPenyediaKota:          adminProject.ProviderCityName,
		AlamatPenyediaDaerah:        adminProject.ProviderDistrictName,
		AlamatPenyediaWilayah:       adminProject.ProviderSubdistrictName,
		AlamatPenyediaKodePos:       adminProject.ProviderPostalCode,
		DocRekeningKoran:            adminProject.DocBankStatement,
		DocLaporanKeuangan:          adminProject.DocFinancialStatement,
		DocContract:                 adminProject.DocContract,
		DocProspect:                 adminProject.DocProspect,
		JenisInstansiPemberiProject: adminProject.TypeOfContractingAuthority,
		InstansiPemberiProject:      adminProject.ContractingAuthority,
		Company: entities.AdminListCompany{
			Id:          projectCompany.Id,
			Name:        projectCompany.CompanyName,
			Address:     projectCompany.Address,
			Province:    projectCompany.ProvinceName,
			City:        projectCompany.CityName,
			District:    projectCompany.DistrictName,
			Subdistrict: projectCompany.SubdistrictName,
		},
		Location: entities.AdminProjectLocation{
			Id:   dataProjectLoc.Id,
			Name: dataProjectLoc.Name,
			Url:  dataProjectLoc.Url,
			Lat:  dataProjectLoc.Lat,
			Lng:  dataProjectLoc.Lng,
		},
		Media:     dataProjectMedia,
		User:      dataAdminListUser[0],
		CreatedAt: adminProject.CreatedAt,
	}

	return map[string]any{
		"data": response,
	}, nil
}

func VerifyUser(r *http.Request, avu *entities.AdminVerifyUser) (map[string]any, error) {

	query := `UPDATE users SET verify = ? WHERE uid = ?`

	err := dbDefault.Exec(query, 1, avu.UserId).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	email, _ := helper.GetEmailByUID(dbDefault, avu.UserId)
	role, _ := helper.GetRoleByUID(dbDefault, avu.UserId)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s berhasil memverifikasi akun pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{}, nil
}

func VerifyUserEmiten(r *http.Request, avu *entities.AdminVerifyUser) (map[string]any, error) {
	email, _ := helper.GetEmailByUID(dbDefault, avu.UserId)
	role, _ := helper.GetRoleByUID(dbDefault, avu.UserId)

	sku, _ := helper.GenNumeric8()

	query := `UPDATE users SET verify_emiten = ?, sku = ? WHERE uid = ?`

	err := dbDefault.Exec(query, 1, sku, avu.UserId).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	subject := "Verifikasi Penerbit Berhasil"
	body := fmt.Sprintf(
		"Halo,\n\nAkun Anda telah diverifikasi sebagai penerbit di Fulusme.\nID: %s\nTanggal: %s\n\nTerima kasih.",
		sku, helper.FormatDate(time.Now()),
	)
	if err := helper.SendEmail(email, "Fulusme", subject, body, "another-otp"); err != nil {
		helper.Logger("error", "Failed to send email: "+err.Error())
	}

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s berhasil memverifikasi akun penerbit pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{}, nil
}

func VerifyUserInvestor(r *http.Request, avu *entities.AdminVerifyUser) (map[string]any, error) {

	email, _ := helper.GetEmailByUID(dbDefault, avu.UserId)
	role, _ := helper.GetRoleByUID(dbDefault, avu.UserId)

	sku, _ := helper.GenPemodalID()

	tx := dbDefault.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if err := tx.Debug().
		Exec(`UPDATE users SET verify_investor = ?, sku = ? WHERE uid = ?`, 1, sku, avu.UserId).Error; err != nil {
		tx.Rollback()
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		helper.Logger("error", "commit tx: "+err.Error())
		return nil, err
	}

	subject := "Verifikasi Investor Berhasil"
	body := fmt.Sprintf(
		"Halo,\n\nAkun Anda telah diverifikasi sebagai investor di Fulusme.\nID: %s\nTanggal: %s\n\nTerima kasih.",
		sku, helper.FormatDate(time.Now()),
	)
	if err := helper.SendEmail(email, "Fulusme", subject, body, "another-otp"); err != nil {
		helper.Logger("error", "Failed to send email: "+err.Error())
	}

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s berhasil memverifikasi akun investor pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{
		"uid":             avu.UserId,
		"email":           email,
		"sku":             sku,
		"verify_investor": 1,
	}, nil
}

func VerifyProject(r *http.Request, avp *entities.AdminVerifyProject) (map[string]any, error) {

	title, _ := helper.GetTitleByProjectId(dbDefault, avp.Id)
	email, _ := helper.GetEmailByProjectId(dbDefault, avp.Id)
	role, _ := helper.GetRoleByProjectId(dbDefault, avp.Id)
	sku, _ := helper.GenNumeric8()

	queryProject := `UPDATE projects SET status = ?, sku = ?, updated_at = TIMESTAMP(CURDATE()) WHERE uid = ?`

	errProject := dbDefault.Exec(queryProject, avp.Status, sku, avp.Id).Error

	if errProject != nil {
		helper.Logger("error", "In Server: "+errProject.Error())
		return nil, errors.New(errProject.Error())
	}

	queryInbox := `UPDATE inboxes SET status = ? WHERE field_2 = ?`

	errInbox := dbDefault.Exec(queryInbox, avp.Status, avp.Id).Error

	if errInbox != nil {
		helper.Logger("error", "In Server: "+errInbox.Error())
		return nil, errors.New(errInbox.Error())
	}

	emailEmiten, _ := helper.GetEmailByProjectId(dbDefault, avp.Id)

	if avp.Status == "2" {

		// Send Email Emiten

		subject := fmt.Sprintf(`Pembayaran Administrasi: Proyek "%s"`, safe(title, avp.Id))

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
										<p style="margin:0 0 12px 0;color:#111827;font-size:14px;">Proyek %[2]s Berhasil Diverifikasi</p>

										<p style="margin:16px 0 6px 0;color:#374151;font-size:14px;">
											Proyek Anda telah berhasil diverifikasi. Selanjutnya, silakan melakukan pembayaran administrasi sesuai ketentuan yang berlaku. Detail informasi lengkap beserta instruksi pembayaran telah kami kirimkan melalui inbox Anda di website Fulusme. Mohon segera lakukan pembayaran agar proses dapat berjalan tanpa hambatan.
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
			safe(title, "-"),
			time.Now().Year(),
		)

		if err := helper.SendEmail(emailEmiten, "Fulusme", subject, htmlBody, "another-otp"); err != nil {
			helper.Logger("error", fmt.Sprintf("gagal kirim email ke %s: %v", emailEmiten, err))
		} else {
			helper.Logger("info", fmt.Sprintf("email HTML konfirmasi terkirim ke %s", emailEmiten))
		}

		// Send Email Emiten

		email, _ := helper.GetEmailByProjectId(dbDefault, avp.Id)

		if err := helper.SendEmail(email, "Fulusme", subject, htmlBody, "another-otp"); err != nil {
			helper.Logger("error", fmt.Sprintf("gagal kirim email ke %s: %v", email, err))
		}

		helper.Logger("info", fmt.Sprintf("email konfirmasi pembayaran terkirim ke %s", email))
	}

	if avp.Status == "4" {

		// Send Email Emiten

		subject := fmt.Sprintf(`Pembayaran Dikonfirmasi: Proyek "%s"`, safe(title, "-"))

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
								<p style="margin:0 0 16px 0;color:#374151;font-size:14px;line-height:1.6;">
									Kami informasikan bahwa bukti pembayaran Anda telah berhasil dikonfirmasi oleh Admin Tim Publish. 
									Selanjutnya, kami mohon Anda segera mengisi Form Dokumen Pelengkap melalui akses yang telah kami kirimkan via inbox di website Fulusme. 
									Pastikan untuk menyiapkan seluruh dokumen yang dibutuhkan agar proses verifikasi dapat berjalan dengan lancar.
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
			time.Now().Year(),
		)

		if err := helper.SendEmail(emailEmiten, "Fulusme", subject, htmlBody, "another-otp"); err != nil {
			helper.Logger("error", fmt.Sprintf("gagal kirim email ke %s: %v", emailEmiten, err))
		} else {
			helper.Logger("info", fmt.Sprintf("email HTML konfirmasi terkirim ke %s", emailEmiten))
		}

	}

	if avp.Status == "5" {

		// Send Email Emiten

		subject := fmt.Sprintf(`Proyek "%s" telah tayang`, safe(title, "-"))

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
								<p style="margin:0 0 16px 0;color:#374151;font-size:14px;line-height:1.6;">
									Dengan ini kami informasikan bahwa proyek %[2]s dengan Proyek ID %[3]s Anda telah berhasil ditayangkan di platform Fulusme. 
									Untuk mengetahui informasi lebih lanjut mengenai status dan detail proyek, 
									silakan kunjungi website Fulusme melalui akun Anda.
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
			safe(title, "-"),
			safe(sku, "-"),
			time.Now().Year(),
		)

		if err := helper.SendEmail(emailEmiten, "Fulusme", subject, htmlBody, "another-otp"); err != nil {
			helper.Logger("error", fmt.Sprintf("gagal kirim email ke %s: %v", emailEmiten, err))
		} else {
			helper.Logger("info", fmt.Sprintf("email HTML konfirmasi terkirim ke %s", emailEmiten))
		}

	}

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s berhasil memverifikasi project %s pada waktu %s",
			ip,
			email,
			role,
			title,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{}, nil
}
