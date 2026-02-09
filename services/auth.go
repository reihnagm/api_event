package services

import (
	"errors"
	"fmt"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	middleware "superapps/middlewares"
	"time"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

func AdminLogin(r *http.Request, l *entities.AdminLogin) (entities.LoginResponse, error) {

	users := []entities.LoginScan{}
	roles := []entities.CheckRole{}

	queryUserExist := `SELECT uid AS id, enabled, password, verify, role, email FROM users WHERE phone = ?`

	errUser := dbDefault.Raw(queryUserExist, l.Val).Scan(&users).Error

	if errUser != nil {
		helper.Logger("error", "In Server: "+errUser.Error())
		return entities.LoginResponse{}, errors.New(errUser.Error())
	}

	isUserExist := len(users)

	if isUserExist == 0 {
		helper.Logger("error", "In Server: USER_NOT_FOUND")
		return entities.LoginResponse{}, errors.New("USER_NOT_FOUND")
	}

	queryCheckRole := `SELECT id, name FROM roles WHERE id = ?`

	errCheckRole := dbDefault.Raw(queryCheckRole, users[0].Role).Scan(&roles).Error

	if errCheckRole != nil {
		helper.Logger("error", "In Server: "+errCheckRole.Error())
		return entities.LoginResponse{}, errors.New(errCheckRole.Error())
	}

	isCheckRoleExist := len(roles)

	if isCheckRoleExist == 0 {
		helper.Logger("error", "In Server: ROLE_NOT_FOUND")
		return entities.LoginResponse{}, errors.New("ROLE_NOT_FOUND")
	}

	passHashed := users[0].Password

	errVerify := helper.VerifyPassword(passHashed, l.Password)

	if errVerify != nil {
		helper.Logger("error", "In Server: CREDENTIALS_IS_INCORRECT")
		return entities.LoginResponse{}, errors.New("CREDENTIALS_IS_INCORRECT")
	}

	token, errToken := middleware.CreateToken(users[0].Id)
	if errToken != nil {
		helper.Logger("error", "In Server: "+errToken.Error())
		return entities.LoginResponse{}, errToken
	}

	access := token["token"]

	email, _ := helper.GetEmailByUID(dbDefault, users[0].Id)
	role, _ := helper.GetRoleByUID(dbDefault, users[0].Id)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s melakukan login pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return entities.LoginResponse{
		Id:      users[0].Id,
		Email:   users[0].Email,
		Enabled: users[0].Enabled,
		Verify:  users[0].Verify,
		Role:    roles[0].Name,
		Token:   access,
	}, nil
}

func Login(r *http.Request, l *entities.Login) (entities.LoginResponse, error) {

	users := []entities.LoginScan{}
	roles := []entities.CheckRole{}

	queryUserExist := `SELECT uid AS id, enabled, password, verify, role, email FROM users WHERE email = ?`

	errUser := dbDefault.Raw(queryUserExist, l.Email).Scan(&users).Error

	if errUser != nil {
		helper.Logger("error", "In Server: "+errUser.Error())
		return entities.LoginResponse{}, errors.New(errUser.Error())
	}

	isUserExist := len(users)

	if isUserExist == 0 {
		helper.Logger("error", "In Server: USER_NOT_FOUND")
		return entities.LoginResponse{}, errors.New("USER_NOT_FOUND")
	}

	queryCheckRole := `SELECT id, name FROM roles WHERE id = ?`

	errCheckRole := dbDefault.Raw(queryCheckRole, users[0].Role).Scan(&roles).Error

	if errCheckRole != nil {
		helper.Logger("error", "In Server: "+errCheckRole.Error())
		return entities.LoginResponse{}, errors.New(errCheckRole.Error())
	}

	isCheckRoleExist := len(roles)

	if isCheckRoleExist == 0 {
		helper.Logger("error", "In Server: ROLE_NOT_FOUND")
		return entities.LoginResponse{}, errors.New("ROLE_NOT_FOUND")
	}

	passHashed := users[0].Password

	errVerify := helper.VerifyPassword(passHashed, l.Password)

	if errVerify != nil {
		helper.Logger("error", "In Server: CREDENTIALS_IS_INCORRECT")
		return entities.LoginResponse{}, errors.New("CREDENTIALS_IS_INCORRECT")
	}

	token, errToken := middleware.CreateToken(users[0].Id)
	if errToken != nil {
		helper.Logger("error", "In Server: "+errToken.Error())
		return entities.LoginResponse{}, errToken
	}

	access := token["token"]

	ip := helper.GetClientIP(r)

	role, _ := helper.GetRoleByUID(dbDefault, users[0].Id)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf("%s [%s] - %s melakukan login pada waktu %s",
			ip,
			users[0].Email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return entities.LoginResponse{
		Id:      users[0].Id,
		Email:   users[0].Email,
		Enabled: users[0].Enabled,
		Verify:  users[0].Verify,
		Role:    roles[0].Name,
		Token:   access,
	}, nil
}

func Logout(r *http.Request, l *entities.Logout) (entities.LogoutResponse, error) {

	ip := helper.GetClientIP(r)

	role, _ := helper.GetRoleByEmail(dbDefault, l.Email)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf("%s [%s] - %s melakukan logout pada waktu %s",
			ip,
			l.Email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return entities.LogoutResponse{
		Email: l.Email,
	}, nil
}

func Register(rr *http.Request, r *entities.Register) (entities.RegisterResponse, error) {

	hashedPassword, errHasshed := helper.Hash(r.Password)
	if errHasshed != nil {
		helper.Logger("error", "In Server: "+errHasshed.Error())
		return entities.RegisterResponse{}, errors.New(errHasshed.Error())
	}

	users := []entities.LoginScan{}

	otp := helper.CodeOtpSecure()

	r.UserId = uuid.NewV4().String()

	queryUserExist := `
		SELECT 
			uid AS id, enabled, password, verify, email, phone
		FROM users
		WHERE email = ? OR phone = ?
	`

	errUserExist := dbDefault.
		Debug().
		Raw(queryUserExist, r.Email, r.Phone).
		Scan(&users).Error

	if errUserExist != nil {
		helper.Logger("error", "In Server: "+errUserExist.Error())
		return entities.RegisterResponse{}, errors.New(errUserExist.Error())
	}

	if len(users) > 0 {
		helper.Logger("error", "In Server: USER_ALREADY_EXIST")
		return entities.RegisterResponse{}, errors.New("USER_ALREADY_EXIST")
	}

	queryInsertUser := `INSERT INTO users (uid, email, phone, password, role, otp) VALUES (?, ?, ?, ?, ?, ?)`

	errInsertUser := dbDefault.Exec(queryInsertUser, r.UserId, r.Email, r.Phone, hashedPassword, 4, otp).Error

	if errInsertUser != nil {
		helper.Logger("error", "In Server: "+errInsertUser.Error())
		return entities.RegisterResponse{}, errors.New(errInsertUser.Error())
	}

	queryInsertProfile := `INSERT INTO profiles (user_id, fullname) VALUES (?, ?)`

	errInsertProfile := dbDefault.Exec(queryInsertProfile, r.UserId, r.Fullname).Error

	if errInsertProfile != nil {
		helper.Logger("error", "In Server: "+errInsertProfile.Error())
		return entities.RegisterResponse{}, errors.New(errInsertProfile.Error())
	}

	errEmail := helper.SendEmail(r.Email, "Fulusme", "Verification Account", otp, "-")
	if errEmail != nil {
		helper.Logger("error", "Failed to send email: "+errEmail.Error())
	}

	// if r.Role == "1" {

	// 	queryInsertProfile := `INSERT INTO profiles (user_id, fullname) VALUES (?, ?)`

	// 	errInsertProfile := dbDefault.Exec(queryInsertProfile, r.UserId, r.Fullname).Error

	// 	if errInsertProfile != nil {
	// 		helper.Logger("error", "In Server: "+errInsertProfile.Error())
	// 		return entities.RegisterResponse{}, errInsertProfile
	// 	}

	// 	queryInsertAccount := `INSERT INTO accounts (user_id, no, bank_name, bank_branch, bank_owner) VALUES (?, ?, ?, ?, ?)`

	// 	errInsertAccount := dbDefault.Exec(queryInsertAccount,
	// 		r.UserId, r.Investor.Bank.No, r.Investor.Bank.Name, r.Investor.Bank.Branch, r.Investor.Bank.Owner,
	// 	).Error

	// 	if errInsertAccount != nil {
	// 		helper.Logger("error", "In Server: "+errInsertAccount.Error())
	// 		return entities.RegisterResponse{}, errInsertAccount
	// 	}

	// 	queryInsertKtp := `INSERT INTO ktps (user_id, nik, place_and_datebirth) VALUES (?, ?, ?)`

	// 	errInsertKtp := dbDefault.Exec(queryInsertKtp, r.UserId, r.Investor.Ktp, r.Investor.AddressKtp).Error

	// 	if errInsertKtp != nil {
	// 		helper.Logger("error", "In Server: "+errInsertKtp.Error())
	// 		return entities.RegisterResponse{}, errInsertKtp
	// 	}

	// 	queryInsertJob := `INSERT INTO jobs (company_name, company_address, monthly_income, position, user_id)
	// 	VALUES (?, ?, ?, ?, ?)`

	// 	errInsertJob := dbDefault.Exec(queryInsertJob, r.Investor.Job.CompanyName, r.Investor.Job.CompanyAddress, r.Investor.Job.MonthlyIncome, r.Investor.Job.Position, r.UserId).Error

	// 	if errInsertJob != nil {
	// 		helper.Logger("error", "In Server: "+errInsertJob.Error())
	// 		return entities.RegisterResponse{}, errInsertJob
	// 	}
	// }

	// if r.Role == "2" {

	// 	queryInsertProfile := `INSERT INTO profiles (user_id, fullname) VALUES (?, ?)`

	// 	errInsertProfile := dbDefault.Exec(queryInsertProfile, r.UserId, r.Fullname).Error

	// 	if errInsertProfile != nil {
	// 		helper.Logger("error", "In Server: "+errInsertProfile.Error())
	// 		return entities.RegisterResponse{}, errInsertProfile
	// 	}

	// 	queryInsertCompany := `INSERT INTO companies (
	// 		company_data, company_name, company_nib,
	// 		deed_of_incorporation, latest_amendment_deed, sk_kumham, company_address, company_npwp,
	// 		total_employees, capital_structure, financial_statements, commissioner_name, commissioner_position,
	// 		commissioner_ktp, commissioner_npwp, director_name, director_position, director_ktp, director_npwp, user_id
	// 	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// 	errInsertCompany := dbDefault.Exec(queryInsertCompany,
	// 		r.Emiten.CompanyData, r.Emiten.CompanyName, r.Emiten.CompanyNib, r.Emiten.DeedOfIncorporation, r.Emiten.LatestAmendmentDeed, r.Emiten.SkKemenkumham, r.Emiten.CompanyAddress, r.Emiten.CompanyNpwp,
	// 		r.Emiten.TotalEmployees, r.Emiten.CapitalStructure, r.Emiten.FinancialStatements, r.Emiten.CommisionerName, r.Emiten.CommisionerPosition, r.Emiten.CommisionerKtp, r.Emiten.CommisionerNpwp,
	// 		r.Emiten.DirectorName, r.Emiten.DirectorPosition, r.Emiten.DirectorKtp, r.Emiten.DirectorNpwp, r.UserId,
	// 	).Error

	// 	if errInsertCompany != nil {
	// 		helper.Logger("error", "In Server: "+errInsertCompany.Error())
	// 		return entities.RegisterResponse{}, errInsertCompany
	// 	}

	// 	queryInsertBond := `INSERT INTO projects
	// 	(uid, user_id, title, type_of_bond, nominal_value, time_periode, interest_rate, interest_payment_schedule, principal_payment_schedule, use_of_funds, collateral_guarantee, desc_job, is_apbn)
	//  	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// 	isApbn := "0"
	// 	if r.Emiten.InfoBond.IsApbn {
	// 		isApbn = "1"
	// 	}

	// 	projectId := uuid.NewV4().String()

	// 	errInsertBond := dbDefault.Exec(queryInsertBond,
	// 		projectId,
	// 		r.UserId,
	// 		r.Emiten.InfoBond.Title,
	// 		r.Emiten.InfoBond.TypeOfBond,
	// 		r.Emiten.InfoBond.NominalValue,
	// 		r.Emiten.InfoBond.TimePeriode,
	// 		r.Emiten.InfoBond.InterestRate,
	// 		r.Emiten.InfoBond.InterestPaymentSchedule,
	// 		r.Emiten.InfoBond.PrincipalPaymentSchedule,
	// 		r.Emiten.InfoBond.UseOfFunds,
	// 		r.Emiten.InfoBond.CollateralGuarantee,
	// 		r.Emiten.InfoBond.DescJob,
	// 		isApbn,
	// 	).Error

	// 	if errInsertBond != nil {
	// 		helper.Logger("error", "In Server: "+errInsertBond.Error())
	// 		return entities.RegisterResponse{}, errInsertBond
	// 	}

	// 	queryInsertBondMedia := `INSERT INTO project_medias (project_id, path) VALUES (?, ?)`

	// 	errInsertBondMedia := dbDefault.Exec(queryInsertBondMedia, projectId, r.Emiten.InfoBond.Img).Error

	// 	if errInsertBondMedia != nil {
	// 		helper.Logger("error", "In Server: "+errInsertBondMedia.Error())
	// 		return entities.RegisterResponse{}, errInsertBondMedia
	// 	}

	// 	queryInsertProjectLoc := `INSERT INTO project_locations (project_id, name, url, lat, lng) VALUES (?, ?, ?, ?, ?)`

	// 	errInsertProjectLoc := dbDefault.Exec(queryInsertProjectLoc,
	// 		projectId, r.Emiten.InfoBond.Location.Name, r.Emiten.InfoBond.Location.Url,
	// 		r.Emiten.InfoBond.Location.Lat, r.Emiten.InfoBond.Location.Lng,
	// 	).Error

	// 	if errInsertProjectLoc != nil {
	// 		helper.Logger("error", "In Server: "+errInsertProjectLoc.Error())
	// 		return entities.RegisterResponse{}, errInsertProjectLoc
	// 	}

	// 	queryInsertProjectDoc := `INSERT INTO project_docs (project_id, path) VALUES (?, ?)`

	// 	errInsertProjectDoc := dbDefault.Exec(queryInsertProjectDoc,
	// 		projectId, r.Emiten.InfoBond.Doc,
	// 	).Error

	// 	if errInsertProjectDoc != nil {
	// 		helper.Logger("error", "In Server: "+errInsertProjectDoc.Error())
	// 		return entities.RegisterResponse{}, errInsertProjectDoc
	// 	}

	// }

	token, errToken := middleware.CreateToken(r.UserId)
	if errToken != nil {
		helper.Logger("error", "In Server: "+errToken.Error())
		return entities.RegisterResponse{}, errToken
	}

	access := token["token"]

	email, _ := helper.GetEmailByUID(dbDefault, r.UserId)
	role, _ := helper.GetRoleByUID(dbDefault, r.UserId)

	ip := helper.GetClientIP(rr)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s mendaftar pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return entities.RegisterResponse{
		Id:      r.UserId,
		Email:   r.Email,
		Enabled: false,
		Role:    "user",
		Verify:  false,
		Token:   access,
	}, nil
}

func RegisterAsEmiten(r *http.Request, ras *entities.RegisterAsEmiten) (map[string]any, error) {
	returnValue := map[string]any{}

	err := dbDefault.Transaction(func(tx *gorm.DB) error {
		queryUpdateProfile := `
			UPDATE profiles 
			SET fullname = ?, selfie = ?, position = ?, photo_ktp = ?, no_ktp = ?, no_npwp = ?
			WHERE user_id = ?`
		if err := tx.Exec(queryUpdateProfile, ras.Fullname, ras.PhotoSelfie, ras.Jabatan, ras.PhotoKtp, ras.NoKtp, ras.NoNpwp, ras.UserId).Error; err != nil {
			helper.Logger("error", "In Server (update profile): "+err.Error())
			return err
		}

		queryUpdateUser := `
			UPDATE users SET role = ? WHERE uid = ?`
		if err := tx.Exec(queryUpdateUser, ras.Role, ras.UserId).Error; err != nil {
			helper.Logger("error", "In Server (update user): "+err.Error())
			return err
		}

		queryInsertAdditionalDoc := `
			INSERT INTO additional_docs (path, type, user_id) 
			VALUES (?, ?, ?)`
		if err := tx.Exec(queryInsertAdditionalDoc, ras.SuratKuasa, "surat-kuasa", ras.UserId).Error; err != nil {
			helper.Logger("error", "In Server (insert doc): "+err.Error())
			return err
		}

		returnValue["message"] = "Register successful"
		return nil
	})

	if err != nil {
		return nil, err
	}

	email, _ := helper.GetEmailByUID(dbDefault, ras.UserId)
	role, _ := helper.GetRoleByUID(dbDefault, ras.UserId)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s mendaftar sebagai emiten pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format(time.Now().Format("2006-01-02 15:04:05")),
		),
		role,
	)

	return returnValue, nil
}

func ResendOtp(r *http.Request, rs *entities.ResendOtp) (map[string]any, error) {

	users := []entities.UserOtp{}
	query := `SELECT enabled, otp_date FROM users
	WHERE (email = ? OR phone = ?)`

	err := dbDefault.Raw(query, rs.Val, rs.Val).Scan(&users).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	isUserExist := len(users)

	if isUserExist == 0 {
		return nil, errors.New("USER_NOT_FOUND")
	}

	emailActive := users[0].Enabled
	createdAt := users[0].CreatedAt

	if emailActive == 1 {
		helper.Logger("error", "In Server: Account is already active")
		return nil, errors.New("ACCOUNT_IS_ALREADY_ACTIVE")
	}

	currentTime := time.Now()
	elapsed := currentTime.Sub(createdAt)

	otp := helper.CodeOtpSecure()

	if elapsed >= 1*time.Minute {

		queryUpdate := `UPDATE users SET otp = ?, created_at = NOW(), otp_date = NOW() WHERE email = ?`

		errUpdateResendOtp := dbDefault.Exec(queryUpdate, otp, rs.Val).Error

		if errUpdateResendOtp != nil {
			helper.Logger("error", "In Server: "+errUpdateResendOtp.Error())
			return nil, errors.New(errUpdateResendOtp.Error())
		}

		errEmail := helper.SendEmail(rs.Val, "Fulusme", "Verification Account", otp, "-")
		if errEmail != nil {
			helper.Logger("error", "Failed to send email: "+errEmail.Error())
		}
	}

	ip := helper.GetClientIP(r)

	role, _ := helper.GetRoleByEmail(dbDefault, rs.Val)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s mengirim ulang otp pada waktu %s",
			ip,
			rs.Val,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{
		"otp": otp,
	}, nil
}

func VerifyOtp(r *http.Request, u *entities.VerifyOtp) (entities.VerifyOtpResponse, error) {
	var user entities.UserOtp

	roles := []entities.CheckRole{}

	err := dbDefault.Raw(`
		SELECT uid as id, email, role, enabled, created_at 
		FROM users 
		WHERE (email = ? OR phone = ?) AND otp = ?`,
		u.Val, u.Val, u.Otp).First(&user).Error

	if err != nil {
		helper.Logger("error", "In Server: USER_OR_OTP_IS_INVALID")
		return entities.VerifyOtpResponse{}, errors.New("USER_OR_OTP_IS_INVALID")
	}

	if user.Enabled == 1 {
		helper.Logger("error", "In Server: Account is already active")
		return entities.VerifyOtpResponse{}, errors.New("ACCOUNT_IS_ALREADY_ACTIVE")
	}

	if time.Since(user.CreatedAt) > time.Minute {
		helper.Logger("error", "In Server: Otp is expired")
		return entities.VerifyOtpResponse{}, errors.New("OTP_IS_EXPIRED")
	}

	queryCheckRole := `SELECT id, name FROM roles WHERE id = ?`

	errCheckRole := dbDefault.Raw(queryCheckRole, user.Role).Scan(&roles).Error

	if errCheckRole != nil {
		helper.Logger("error", "In Server: "+errCheckRole.Error())
		return entities.VerifyOtpResponse{}, errors.New(errCheckRole.Error())
	}

	isCheckRoleExist := len(roles)

	if isCheckRoleExist == 0 {
		helper.Logger("error", "In Server: ROLE_NOT_FOUND")
		return entities.VerifyOtpResponse{}, errors.New("ROLE_NOT_FOUND")
	}

	errUpdate := dbDefault.Exec(`
		UPDATE users SET enabled = 1, email_active_date = NOW()
		WHERE uid = ?`, user.Id).Error
	if errUpdate != nil {
		helper.Logger("error", "In Server: "+errUpdate.Error())
		return entities.VerifyOtpResponse{}, errUpdate
	}

	token, errToken := middleware.CreateToken(user.Id)
	if errToken != nil {
		helper.Logger("error", "In Server: "+errToken.Error())
		return entities.VerifyOtpResponse{}, errToken
	}
	access := token["token"]

	email, _ := helper.GetEmailByUID(dbDefault, user.Id)
	role, _ := helper.GetRoleByUID(dbDefault, user.Id)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s melakukan aktivasi verifikasi otp pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return entities.VerifyOtpResponse{
		Id:      user.Id,
		Enabled: true,
		Role:    roles[0].Name,
		Email:   user.Email,
		Verify:  true,
		Token:   access,
	}, nil
}

func ForgotPassword(r *http.Request, fp *entities.ForgotPassword) (map[string]any, error) {
	users := []entities.ForgotPassword{}

	err := dbDefault.Raw(`SELECT email FROM users WHERE email = ?`, fp.Email).Scan(&users).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	isUserExist := len(users)

	if isUserExist == 0 {
		return nil, errors.New("USER_NOT_FOUND")
	}

	hashedPassword, errHashedPassword := helper.Hash(fp.NewPassword)

	if errHashedPassword != nil {
		helper.Logger("error", "In Server: "+errHashedPassword.Error())
		return nil, errHashedPassword
	}

	errUpdate := dbDefault.Exec(`UPDATE users SET password = ? WHERE email = ?`, hashedPassword, fp.Email).Error

	if errUpdate != nil {
		helper.Logger("error", "In Server: "+errUpdate.Error())
		return nil, errUpdate
	}

	email, _ := helper.GetEmailByEmail(dbDefault, fp.Email)
	role, _ := helper.GetRoleByEmail(dbDefault, fp.Email)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s melakukan aktivasi lupa kata sandi pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{}, nil
}

func AssignRole(r *http.Request, ar *entities.AssignRole) (map[string]any, error) {

	if ar.Role == "1" {

		queryUpdateRole := `UPDATE users SET role = ? WHERE uid = ?`

		errUpdateRole := dbDefault.Exec(queryUpdateRole, ar.Role, ar.UserId).Error

		if errUpdateRole != nil {
			helper.Logger("error", "In Server: "+errUpdateRole.Error())
			return nil, fmt.Errorf("failed to update role: %w", errUpdateRole)
		}

		// User Profile

		queryUpdateProfile := `UPDATE profiles SET gender = ?, avatar = ?, status_marital = ?, 
		last_education = ?, occupation = ?, province_name = ?, city_name = ?, district_name = ?, 
		subdistrict_name = ?, postal_code = ?, address_detail = ?, beneficiary_name = ?, beneficiary_phone = ?
		WHERE user_id = ?`

		errUpdateProfile := dbDefault.Exec(queryUpdateProfile, ar.Gender, ar.Avatar, ar.StatusMarital, ar.LastEducation,
			ar.Occupation, ar.ProvinceName, ar.CityName, ar.DistrictName, ar.SubdistrictName,
			ar.PostalCode, ar.AddressDetail, ar.NamaAhliWaris, ar.PhoneAhliWaris,
			ar.UserId,
		).Error

		if errUpdateProfile != nil {
			helper.Logger("error", "In Server: "+errUpdateProfile.Error())
			return nil, fmt.Errorf("failed to update profile: %w", errUpdateProfile)
		}

		// User Account

		queryInsertAccount := `INSERT INTO accounts (user_id, no, bank_name, bank_branch, bank_owner, rek_koran_path) VALUES (?, ?, ?, ?, ?, ?)`

		errInsertAccount := dbDefault.Exec(queryInsertAccount,
			ar.UserId, ar.Bank.No, ar.Bank.Name, ar.Bank.Branch, ar.Bank.Owner, ar.Bank.RekKoranPath,
		).Error

		if errInsertAccount != nil {
			helper.Logger("error", "In Server: "+errInsertAccount.Error())
			return nil, fmt.Errorf("failed to insert account: %w", errInsertAccount)
		}

		// User Security Account (optional)

		queryInsertSecurityAccount := `INSERT INTO security_accounts (account_name, account_no, account_sub_no, account_bank, user_id) 
		VALUES (?, ?, ?, ?, ?)`

		if ar.NamaRekeningEfek != "-" || ar.NomorRekeningEfek != "-" || ar.NomorSubRekeningEfek != "-" || ar.BankRekeningEfek != "-" {

			errInsertSecurityAccount := dbDefault.Exec(queryInsertSecurityAccount,
				ar.NamaRekeningEfek, ar.NomorRekeningEfek, ar.NomorSubRekeningEfek, ar.BankRekeningEfek, ar.UserId,
			).Error

			if errInsertSecurityAccount != nil {
				helper.Logger("error", "In Server: "+errInsertSecurityAccount.Error())
				return nil, fmt.Errorf("failed to insert security account: %w", errInsertSecurityAccount)
			}

		}

		// User Pay Slip

		queryInsertPaySlip := `INSERT INTO pay_slips (path, user_id) VALUES (?, ?)`

		errInsertPaySlip := dbDefault.Exec(queryInsertPaySlip, ar.SlipGaji, ar.UserId).Error

		if errInsertPaySlip != nil {
			helper.Logger("error", "In Server: "+errInsertPaySlip.Error())
			return nil, fmt.Errorf("failed to pay slip: %w", errInsertPaySlip)
		}

		// User KTP

		queryInsertKtp := `INSERT INTO ktps (user_id, name, nik, path, place_datebirth) VALUES (?, ?, ?, ?, ?)`

		errInsertKtp := dbDefault.Exec(queryInsertKtp, ar.UserId, ar.Ktp.Name, ar.Ktp.Nik, ar.Ktp.NikPath, ar.Ktp.PlaceDatebirth).Error

		if errInsertKtp != nil {
			helper.Logger("error", "In Server: "+errInsertKtp.Error())
			return nil, fmt.Errorf("failed to insert ktp: %w", errInsertKtp)
		}

		// User Job

		queryInsertJob := `INSERT INTO jobs (
			company_name, province_name, city_name, district_name, subdistrict_name,
			postal_code, company_address, monthly_income, annual_income, position,
			npwp, npwp_path, user_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		res := dbDefault.Exec(queryInsertJob,
			ar.Job.Company,
			ar.Job.ProvinceName,
			ar.Job.CityName,
			ar.Job.DistrictName,
			ar.Job.SubdistrictName,
			ar.Job.PostalCode,
			ar.Job.Address,
			ar.Job.MonthlyIncome,
			ar.Job.AnnualIncome,
			ar.Job.Position,
			ar.Job.Npwp,
			ar.Job.NpwpPath,
			ar.UserId,
		)
		if res.Error != nil {
			helper.Logger("error", "In Server: "+res.Error.Error())
			return nil, fmt.Errorf("failed to insert job: %w", res.Error)
		}

		// User Signature

		queryInsertUserSignature := `INSERT INTO user_signatures (user_id, path) VALUES (?, ?)`

		errInsertUserSignature := dbDefault.Exec(queryInsertUserSignature, ar.UserId, ar.SignaturePath).Error

		if errInsertUserSignature != nil {
			helper.Logger("error", "In Server: "+errInsertUserSignature.Error())
			return nil, fmt.Errorf("failed to insert user signature: %w", errInsertUserSignature)
		}

		// User Risk

		queryInsertUserRisk := `INSERT INTO user_risks (goal, tolerance, experience, capital_market_knowledge, user_id) 
		VALUES (?, ?, ?, ?, ?)`

		errInsertUserRisk := dbDefault.Exec(queryInsertUserRisk, ar.Risk.Goal, ar.Risk.Tolerance, ar.Risk.Experience, ar.Risk.PengetahuanPasarModal, ar.UserId).Error

		if errInsertUserRisk != nil {
			helper.Logger("error", "In Server: "+errInsertUserRisk.Error())
			return nil, fmt.Errorf("failed to insert user signature: %w", errInsertUserRisk)
		}

		return map[string]any{}, nil

	}

	if ar.Role == "2" || ar.Role == "9" {

		// Update Role

		queryUpdateRole := `UPDATE users SET role = ? WHERE uid = ?`

		errUpdateRole := dbDefault.Exec(queryUpdateRole, ar.Role, ar.UserId).Error

		if errUpdateRole != nil {
			return nil, fmt.Errorf("failed to update role: %w", errUpdateRole)
		}

		// Insert Company

		companyId := uuid.NewV4().String()

		queryInsertCompany := `INSERT INTO companies (
		uid, company_name, company_nib, company_nib_path, siup, tdp, est, site,
		deed_of_incorporation, latest_amendment_deed, latest_amendment_deed_path, sk_kumham, sk_kumham_last, sk_kumham_path,
		npwp_path, npwp, total_employees, financial_statement, bank_statement, email, phone,
		bank_name, bank_account, bank_owner_company, user_id, certificate_of_company_est, type_of_company, type_of_business, type_of_company_place
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		err := dbDefault.Exec(queryInsertCompany,
			companyId,
			ar.CompanyName, ar.CompanyNib, ar.CompanyNibPath, ar.Siup, ar.Tdp, ar.Didirikan, ar.Site,
			ar.AktaPendirian, ar.AktaPerubahanTerahkir, ar.AktaPerubahanTerahkirPath, ar.SkKumham, ar.SkKumhamTerahkir, ar.SkKumhamPath,
			ar.NpwpPath, ar.Npwp, ar.TotalEmployees, ar.LaporanKeuanganPath, ar.RekeningKoranPath,
			ar.Email, ar.Phone, ar.BankName, ar.BankAccount, ar.BankOwner,
			ar.UserId, ar.SkPendirianPerusahaan, ar.JenisPerusahaan, ar.JenisUsaha, ar.StatusKantor,
		).Error

		if err != nil {
			return nil, fmt.Errorf("failed to insert company: %w", err)
		}

		// Insert Security Account

		if ar.Role == "9" {

			querySecAccount := `INSERT INTO security_accounts (account_name, account_no, account_sub_no, account_bank, user_id) 
			VALUES (?, ?, ?, ?, ?)`

			errSecAccount := dbDefault.Exec(querySecAccount,
				ar.NamaRekeningEfek, ar.NomorRekeningEfek, "-",
				ar.BankRekeningEfek, ar.UserId,
			).Error

			if errSecAccount != nil {
				return nil, fmt.Errorf("failed to insert security account: %w", errSecAccount)
			}

		}

		// Insert Company Address

		queryInsertCompanyAddress := `INSERT INTO company_addresses (
			name, address, postal_code, company_id, province_name, city_name, district_name, subdistrict_name
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

		for _, a := range ar.Address {

			err = dbDefault.Exec(queryInsertCompanyAddress,
				a.Name, a.Detail, a.PostalCode, companyId,
				a.ProvinceName, a.CityName, a.DistrictName, a.SubdistrictName,
			).Error

			if err != nil {
				return nil, fmt.Errorf("failed to insert company address: %w", err)
			}

		}

		for _, p := range ar.Komisaris {
			queryInsertCompanyPosition := `INSERT INTO positions (title, name, position, ktp, ktp_path, npwp, npwp_path, company_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

			err = dbDefault.Exec(queryInsertCompanyPosition,
				p.Title, p.Name, p.Position,
				p.Ktp, p.KtpPath, p.Npwp, p.NpwpPath, companyId,
			).Error

			if err != nil {
				return nil, fmt.Errorf("failed to insert position: %w", err)
			}
		}

		for _, p := range ar.Directors {
			queryInsertCompanyPosition := `INSERT INTO positions (title, name, position, ktp, ktp_path, npwp, npwp_path, company_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

			err = dbDefault.Exec(queryInsertCompanyPosition,
				p.Title, p.Name, p.Position,
				p.Ktp, p.KtpPath, p.Npwp, p.NpwpPath, companyId,
			).Error

			if err != nil {
				return nil, fmt.Errorf("failed to insert position: %w", err)
			}
		}

		return map[string]any{
			"company_id": companyId,
		}, nil
	}

	email, _ := helper.GetEmailByUID(dbDefault, ar.UserId)
	role, _ := helper.GetRoleByUID(dbDefault, ar.UserId)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] melakukan pendaftaran role sebagai %s pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{"message": "no role matched"}, nil
}
