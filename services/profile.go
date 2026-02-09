package services

import (
	"errors"
	"fmt"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"time"
)

func GetProfile(gp *entities.GetProfile) (map[string]any, error) {
	var profileAddionalDoc entities.ProfileAdditionalDocScan
	var profileUserCompany entities.ProfileUserCompanyScan
	var profileInvestorBank entities.ProfileUserInvestorBankScan
	var profileInvestorJob entities.ProfileUserInvestorJobScan
	var profileInvestorRisk entities.ProfileuserInvestorRiskScan
	var profileInvestorKtp entities.ProfileUserInvestorKtpScan
	var profileSlipPay entities.ProfileSlipPay
	var profileSecurityAccount entities.ProfileSecurityAccount

	var user entities.ProfileScan

	// Get Profile
	queryUserExist := `
		SELECT p.user_id as id, u.phone, u.email, p.fullname, p.avatar, p.gender, p.last_education, 
		u.verify_emiten,
		u.verify_investor,
		p.beneficiary_name, 
		p.beneficiary_phone,
		r.name AS role,
		p.selfie, p.position, p.photo_ktp, p.no_ktp, p.no_npwp,
		p.province_name, p.city_name, p.district_name, p.subdistrict_name, 
		p.address_detail, p.postal_code, p.occupation, p.status_marital,
		p.created_at, p.updated_at
		FROM profiles p 
		INNER JOIN users u ON u.uid = p.user_id
		INNER JOIN roles r ON u.role = r.id
		WHERE p.user_id = ?
	`
	err := dbDefault.Raw(queryUserExist, gp.UserId).Scan(&user).Error
	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	if user.Id == "" {
		return nil, errors.New("USER_NOT_FOUND")
	}

	// Get Power Of Attorney
	queryAdditionalDoc := `SELECT path, type FROM additional_docs WHERE user_id = ?`

	errAdditionalDoc := dbDefault.Raw(queryAdditionalDoc, gp.UserId).Scan(&profileAddionalDoc).Error

	if errAdditionalDoc != nil {
		helper.Logger("error", "In Server: "+errAdditionalDoc.Error())
		return nil, errAdditionalDoc
	}

	// Get Slip Pay
	querySlipPay := `SELECT path FROM pay_slips 
		WHERE user_id = ?`

	errSlipPay := dbDefault.Raw(querySlipPay, gp.UserId).Scan(&profileSlipPay).Error

	if errSlipPay != nil {
		helper.Logger("error", "In Server: "+errSlipPay.Error())
		return nil, errors.New(errSlipPay.Error())
	}

	// Get Security Account
	querySecurityAccount := `SELECT account_name, account_no, account_sub_no, account_bank 
	FROM security_accounts 
	WHERE user_id = ?`

	errSecurityAccount := dbDefault.Raw(querySecurityAccount, gp.UserId).Scan(&profileSecurityAccount).Error

	if errSecurityAccount != nil {
		helper.Logger("error", "In Server: "+errSecurityAccount.Error())
		return nil, errors.New(errSecurityAccount.Error())
	}

	// Get Bank
	queryUserBank := `SELECT no, bank_name, bank_owner, bank_branch, rek_koran_path FROM accounts 
	WHERE user_id = ?`

	errUserBank := dbDefault.Raw(queryUserBank, gp.UserId).Scan(&profileInvestorBank).Error

	if errUserBank != nil {
		helper.Logger("error", "In Server (Accounts): "+errUserBank.Error())
		return nil, errUserBank
	}

	// Get KTP
	queryUserKtp := `SELECT name, nik, place_datebirth, path FROM ktps WHERE user_id = ?`

	errUserKtp := dbDefault.Raw(queryUserKtp, gp.UserId).Scan(&profileInvestorKtp).Error

	if errUserKtp != nil {
		helper.Logger("error", "In Server (KTP): "+errUserKtp.Error())
		return nil, errUserKtp
	}

	// Get Job
	queryUserJob := `SELECT province_name, city_name, district_name, subdistrict_name, 
	postal_code, company_name, company_address, monthly_income, annual_income, npwp, npwp_path, position 
	FROM jobs WHERE user_id = ?`

	errUserJob := dbDefault.Raw(queryUserJob, gp.UserId).Scan(&profileInvestorJob).Error

	if errUserJob != nil {
		helper.Logger("error", "In Server (KTP): "+errUserJob.Error())
		return nil, errUserJob
	}

	// Get Risk
	queryUserRisk := `SELECT goal, tolerance, experience, capital_market_knowledge 
	FROM user_risks WHERE user_id = ?`

	errUserRisk := dbDefault.Raw(queryUserRisk, gp.UserId).Scan(&profileInvestorRisk).Error

	if errUserRisk != nil {
		helper.Logger("error", "In Server (KTP): "+errUserRisk.Error())
		return nil, errUserRisk
	}

	// Get Company
	queryCompany := `
		SELECT c.uid AS id, c.company_name, c.company_nib, c.company_nib_path, c.deed_of_incorporation, c.latest_amendment_deed, c.latest_amendment_deed_path,
			c.sk_kumham, c.sk_kumham_last, c.sk_kumham_path, c.bank_name, c.bank_account, c.bank_owner_company, c.siup, c.site, c.email, 
			c.npwp, c.phone, c.tdp, c.npwp_path, c.total_employees, c.financial_statement, c.bank_statement, toc.name AS jenis_perusahaan,
			c.certificate_of_company_est AS sk_pendirian_perusahaan, c.latest_amendment_deed AS akta_perubahan_terahkir
		FROM companies c
		INNER JOIN type_of_companies toc ON toc.id = c.type_of_company
		WHERE c.user_id = ?
	`

	errCompany := dbDefault.Raw(queryCompany, user.Id).Scan(&profileUserCompany).Error
	if errCompany != nil {
		helper.Logger("error", "In Server (Company): "+errCompany.Error())
		return nil, errCompany
	}

	// Get Address for Company
	var companyAddresses []entities.ProfileUserAddressCompany
	queryAddress := `
	SELECT name, address AS detail, postal_code, company_id, province_name, city_name, district_name, subdistrict_name, created_at 
	FROM company_addresses 
	WHERE company_id = ?
`

	errAddress := dbDefault.Raw(queryAddress, profileUserCompany.Id).Scan(&companyAddresses).Error
	if errAddress != nil {
		helper.Logger("error", "In Server (Address): "+errAddress.Error())
		return nil, errAddress
	}

	var activeProj []entities.ActiveProj
	var canCreateProject bool

	terminal := []string{
		"APPROVED",
		"PUBLISH",
	}

	qActive := `
		SELECT p.uid AS id, p.title, ps.name AS status_name
		FROM projects p
		INNER JOIN project_statuses ps ON ps.id = p.status
		WHERE p.company_id = ?
		  AND ps.name IN (?)
		ORDER BY p.created_at DESC, p.uid DESC
		LIMIT 1`
	errActive := dbDefault.
		Raw(qActive, profileUserCompany.Id, terminal).
		Scan(&activeProj).Error
	if errActive != nil {
		helper.Logger("error", "In Server (Project Gate): "+errActive.Error())
		return nil, errActive
	}

	if len(activeProj) > 0 {
		canCreateProject = false
	} else {
		canCreateProject = true
	}

	// Get Position Director for Company
	var companyPositionDir []entities.ProfileUserDirector

	queryPositionDir := `SELECT id, title, name, position, ktp, ktp_path, npwp, npwp_path 
	FROM positions 
	WHERE company_id = ? AND title LIKE ?`

	errCompanyPositionDir := dbDefault.
		Raw(queryPositionDir, profileUserCompany.Id, "%Dire%").
		Scan(&companyPositionDir).Error

	if errCompanyPositionDir != nil {
		helper.Logger("error", "In Server (Director Position): "+errCompanyPositionDir.Error())
		return nil, errCompanyPositionDir
	}

	if companyPositionDir == nil {
		companyPositionDir = []entities.ProfileUserDirector{}
	}

	// Get Position Komisaris for Company
	var companyPositionKom []entities.ProfileUserKomisaris

	queryPositionKom := `
		SELECT id, title, name, position, ktp, ktp_path, npwp, npwp_path 
		FROM positions 
		WHERE company_id = ? AND title LIKE ?
	`

	errCompanyPositionKom := dbDefault.
		Raw(queryPositionKom, profileUserCompany.Id, "%Komi%").
		Scan(&companyPositionKom).Error

	if errCompanyPositionKom != nil {
		helper.Logger("error", "In Server (Director Position): "+errCompanyPositionKom.Error())
		return nil, errCompanyPositionKom
	}

	if companyPositionKom == nil {
		companyPositionKom = []entities.ProfileUserKomisaris{}
	}

	// Check Sec Account
	var hasSecAccount bool

	errHasSecAccount := dbDefault.Raw(`
		SELECT EXISTS(
			SELECT 1
			FROM security_accounts
			WHERE user_id = ?
		) AS has
	`, gp.UserId).Scan(&hasSecAccount).Error

	if errHasSecAccount != nil {
		helper.Logger("error", "In Server (Security Account): "+errHasSecAccount.Error())
		return nil, errHasSecAccount
	}

	// Get Project
	var companyProject []entities.ProfileUserProjectScan

	queryCompanyProject := `
		SELECT 
			p.uid AS id, p.title, top.name AS type_of_project, pos.name AS status,
			p.nominal_value, p.time_periode, p.interest_rate,
			p.interest_payment_schedule, p.principal_payment_schedule,
			p.desc_job, p.company_profile, p.is_apbn, p.is_approved,
			p.loa, p.spk,
			p.start_project, p.end_project,
			p.created_at, p.updated_at
			FROM projects p
			JOIN type_of_projects top ON p.type_of_project = top.id
			JOIN project_statuses pos ON pos.id = p.status
			WHERE p.uid = (
			SELECT uid
			FROM projects
			WHERE company_id = ?
			ORDER BY created_at DESC, uid DESC
			LIMIT 1
		)
	`

	errCompanyProject := dbDefault.
		Raw(queryCompanyProject, profileUserCompany.Id).
		Scan(&companyProject).Error

	if errCompanyProject != nil {
		helper.Logger("error", "In Server (Project): "+errCompanyProject.Error())
		return nil, errCompanyProject
	}

	var finalProjects []entities.ProfileUserProject

	for _, p := range companyProject {

		// Get Project Contract
		var companyProjectContract entities.ProfileUserProjectContractScan

		queryCompanyProjectContract := `
		SELECT value, path
		FROM project_contracts
		WHERE project_id = ?`

		errCompanyProjectContract := dbDefault.Raw(queryCompanyProjectContract, p.Id).Scan(&companyProjectContract).Error

		if errCompanyProjectContract != nil {
			helper.Logger("error", "In Server (Project): "+errCompanyProjectContract.Error())
			return nil, errCompanyProjectContract
		}

		// Get Project Penggunaan Data
		var companyProjectPenggunaanData []entities.ProfileUserProjectPenggunaanData

		queryCompanyProjectPenggunaanData := `
		SELECT id, name
		FROM use_of_funds
		WHERE project_id = ?`

		errCompanyProjectPenggunaanDana := dbDefault.Raw(queryCompanyProjectPenggunaanData, p.Id).Scan(&companyProjectPenggunaanData).Error

		if errCompanyProjectPenggunaanDana != nil {
			helper.Logger("error", "In Server (Project): "+errCompanyProjectPenggunaanDana.Error())
			return nil, errCompanyProjectPenggunaanDana
		}

		if companyProjectPenggunaanData == nil {
			companyProjectPenggunaanData = []entities.ProfileUserProjectPenggunaanData{}
		}

		// Get Project Jaminan Kolateral
		var companyProjectJaminanKolateral []entities.ProfileUserProjectJaminanKolateral

		queryCompanyProjectJaminanKolateral := `
		SELECT id, name
		FROM collateral_guarantees
		WHERE project_id = ?`

		errCompanyProjectJaminanKolateral := dbDefault.Raw(queryCompanyProjectJaminanKolateral, p.Id).Scan(&companyProjectJaminanKolateral).Error

		if errCompanyProjectJaminanKolateral != nil {
			helper.Logger("error", "In Server (Project): "+errCompanyProjectJaminanKolateral.Error())
			return nil, errCompanyProjectJaminanKolateral
		}

		if companyProjectJaminanKolateral == nil {
			companyProjectJaminanKolateral = []entities.ProfileUserProjectJaminanKolateral{}
		}

		// Get Project Media
		var companyProjectMedia = make([]entities.ProfileUserProjectMedia, 0)

		queryCompanyProjectMedia := `
		SELECT id, path
		FROM project_medias
		WHERE project_id = ?`

		errCompanyProjectMedia := dbDefault.Raw(queryCompanyProjectMedia, p.Id).Scan(&companyProjectMedia).Error

		if errCompanyProjectMedia != nil {
			helper.Logger("error", "In Server (Project): "+errCompanyProjectMedia.Error())
			return nil, errCompanyProjectMedia
		}

		project := entities.ProfileUserProject{
			Id:                    p.Id,
			Title:                 p.Title,
			Deskripsi:             p.DescJob,
			JenisProjek:           p.TypeOfProject,
			JumlahMinimal:         p.NominalValue,
			JangkaWaktu:           p.TimePeriode,
			Spk:                   p.Spk,
			Loa:                   p.Loa,
			MulaiProject:          p.StartProject,
			SelesaiProject:        p.EndProject,
			TingkatBunga:          p.InterestRate,
			JadwalPembayaranBunga: p.InterestPaymentSchedule,
			JadwalPembayaranPokok: p.PrincipalPaymentSchedule,
			CompanyProfile:        p.CompanyProfile,
			PenggunaanDana:        companyProjectPenggunaanData,
			Media:                 companyProjectMedia,
			JaminanKolateral:      companyProjectJaminanKolateral,
			NilaiKontrakPath:      companyProjectContract.Path,
			NilaiKontrak:          companyProjectContract.Value,
			IsApbn:                p.IsApbn,
			Status:                p.Status,
		}
		finalProjects = append(finalProjects, project)
	}

	response := entities.ProfileResponse{
		Profile: entities.Profile{
			Id:                     helper.DefaultIfEmpty(user.Id, "-"),
			Role:                   helper.DefaultIfEmpty(user.Role, "-"),
			Fullname:               helper.DefaultIfEmpty(user.Fullname, "-"),
			Avatar:                 helper.DefaultIfEmpty(user.Avatar, "-"),
			Selfie:                 helper.DefaultIfEmpty(user.Selfie, "-"),
			PhotoKtp:               helper.DefaultIfEmpty(user.PhotoKtp, "-"),
			Phone:                  helper.DefaultIfEmpty(user.Phone, "-"),
			Email:                  helper.DefaultIfEmpty(user.Email, "-"),
			NoKtp:                  helper.DefaultIfEmpty(user.NoKtp, "-"),
			Npwp:                   helper.DefaultIfEmpty(user.NoNpwp, "-"),
			CanCreateProject:       canCreateProject,
			Gender:                 helper.DefaultIfEmpty(user.Gender, "-"),
			LastEducation:          helper.DefaultIfEmpty(user.LastEducation, "-"),
			StatusMarital:          helper.DefaultIfEmpty(user.StatusMarital, "-"),
			AddressDetail:          helper.DefaultIfEmpty(user.AddressDetail, "-"),
			Occupation:             helper.DefaultIfEmpty(user.Occupation, "-"),
			Position:               helper.DefaultIfEmpty(user.Position, "-"),
			ProvinceName:           helper.DefaultIfEmpty(user.ProvinceName, "-"),
			CityName:               helper.DefaultIfEmpty(user.CityName, "-"),
			DistrictName:           helper.DefaultIfEmpty(user.DistrictName, "-"),
			SubdistrictName:        helper.DefaultIfEmpty(user.SubdistrictName, "-"),
			PostalCode:             helper.DefaultIfEmpty(user.PostalCode, "-"),
			NamaAhliWaris:          helper.DefaultIfEmpty(user.BeneficiaryName, "-"),
			PhoneAhliWaris:         helper.DefaultIfEmpty(user.BeneficiaryPhone, "-"),
			SlipGaji:               helper.DefaultIfEmpty(profileSlipPay.Path, "-"),
			ProfileSecurityAccount: profileSecurityAccount,
			Doc: entities.ProfileDoc{
				Path: profileAddionalDoc.Path,
				Type: profileAddionalDoc.Type,
			},
			Investor: entities.ProfileUserInvestor{
				Bank: entities.ProfileUserInvestorBankScan{
					No:           profileInvestorBank.No,
					BankName:     profileInvestorBank.BankName,
					BankOwner:    profileInvestorBank.BankOwner,
					BankBranch:   profileInvestorBank.BankBranch,
					RekKoranPath: profileInvestorBank.RekKoranPath,
					CreatedAt:    profileInvestorBank.CreatedAt,
				},
				Ktp: entities.ProfileUserInvestorKtpScan{
					Name:           profileInvestorKtp.Name,
					Nik:            profileInvestorKtp.Nik,
					PlaceDatebirth: profileInvestorKtp.PlaceDatebirth,
					Path:           profileInvestorKtp.Path,
					CreatedAt:      profileInvestorKtp.CreatedAt,
				},
				Job: entities.ProfileUserInvestorJobScan{
					ProvinceName:    profileInvestorJob.ProvinceName,
					CityName:        profileInvestorJob.CityName,
					DistrictName:    profileInvestorJob.DistrictName,
					SubdistrictName: profileInvestorJob.SubdistrictName,
					PostalCode:      profileInvestorJob.PostalCode,
					CompanyName:     profileInvestorJob.CompanyName,
					CompanyAddress:  profileInvestorJob.CompanyAddress,
					MonthlyIncome:   profileInvestorJob.MonthlyIncome,
					AnnualIncome:    profileInvestorJob.AnnualIncome,
					NpwpPath:        profileInvestorJob.NpwpPath,
					Npwp:            profileInvestorJob.Npwp,
					Position:        profileInvestorJob.Position,
				},
				Risk: entities.ProfileuserInvestorRiskScan{
					Goal:                   profileInvestorRisk.Goal,
					Tolerance:              profileInvestorRisk.Tolerance,
					Experience:             profileInvestorRisk.Experience,
					CapitalMarketKnowledge: profileInvestorRisk.CapitalMarketKnowledge,
				},
			},
			Company: entities.ProfileUserCompany{
				Id:                        profileUserCompany.Id,
				Name:                      profileUserCompany.CompanyName,
				Nib:                       profileUserCompany.CompanyNib,
				NibPath:                   profileUserCompany.CompanyNibPath,
				AktaPendirian:             profileUserCompany.DeedOfIncorporation,
				AktaPerubahanTerahkir:     profileUserCompany.LatestAmendmentDeed,
				AktaPerubahanTerahkirPath: profileUserCompany.LatestAmendmentDeedPath,
				SkKumham:                  profileUserCompany.SkKumham,
				SkKumhamTerahkir:          profileUserCompany.SkKumhamLast,
				SkKumhamPath:              profileUserCompany.SkKumhamPath,
				NpwpPath:                  profileUserCompany.NpwpPath,
				TotalEmployees:            profileUserCompany.TotalEmployees,
				LaporanKeuanganPath:       profileUserCompany.FinancialStatement,
				RekeningKoran:             profileUserCompany.BankStatement,
				SkPendirianPerusahaan:     profileUserCompany.SkPendirianPerusahaan,
				Siup:                      profileUserCompany.Siup,
				Tdp:                       profileUserCompany.Tdp,
				Site:                      profileUserCompany.Site,
				Email:                     profileUserCompany.Email,
				Npwp:                      profileUserCompany.Npwp,
				Phone:                     profileUserCompany.Phone,
				Bank: entities.ProfileBankCompany{
					Name:  profileUserCompany.BankName,
					No:    profileUserCompany.BankAccount,
					Owner: profileUserCompany.BankOwnerCompany,
				},
				JenisPerusahaan: profileUserCompany.JenisPerusahaan,
				Address:         companyAddresses,
				Directors:       companyPositionDir,
				Komisaris:       companyPositionKom,
				Projects:        finalProjects,
			},
			RekEfek:        hasSecAccount,
			VerifyEmiten:   user.VerifyEmiten,
			VerifyInvestor: user.VerifyInvestor,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
		},
	}

	return map[string]any{
		"data": response,
	}, nil
}

func UpdateProfile(r *http.Request, up *entities.UpdateProfile) (map[string]any, error) {
	updates := make(map[string]interface{})

	if up.Fullname != "" {
		updates["fullname"] = up.Fullname
	}
	if up.Avatar != "" {
		updates["avatar"] = up.Avatar
	}
	if up.Selfie != "" {
		updates["selfie"] = up.Selfie
	}
	if up.Gender != "" {
		updates["gender"] = up.Gender
	}
	if up.StatusMarital != "" {
		updates["status_marital"] = up.StatusMarital
	}
	if up.LastEducation != "" {
		updates["last_education"] = up.LastEducation
	}
	if up.ProvinceName != "" {
		updates["province_name"] = up.ProvinceName
	}
	if up.CityName != "" {
		updates["city_name"] = up.CityName
	}
	if up.DistrictName != "" {
		updates["district_name"] = up.DistrictName
	}
	if up.SubdistrictName != "" {
		updates["subdistrict_name"] = up.SubdistrictName
	}
	if up.PostalCode != "" {
		updates["postal_code"] = up.PostalCode
	}
	if up.AddressDetail != "" {
		updates["address_detail"] = up.AddressDetail
	}
	if up.PhotoKtp != "" {
		updates["photo_ktp"] = up.PhotoKtp
	}
	if up.NoKtp != "" {
		updates["no_ktp"] = up.NoKtp
	}
	if up.Position != "" {
		updates["position"] = up.Position
	}
	if up.Occupation != "" {
		updates["occupation"] = up.Occupation
	}
	if up.BeneficiaryName != "" {
		updates["beneficiary_name"] = up.BeneficiaryName
	}
	if up.BeneficiaryPhone != "" {
		updates["beneficiary_phone"] = up.BeneficiaryPhone
	}

	updates["updated_at"] = time.Now()

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	err := dbDefault.
		Model(&entities.Profile{}).
		Where("user_id = ?", up.UserId).
		Updates(updates).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	email, _ := helper.GetEmailByUID(dbDefault, up.UserId)
	role, _ := helper.GetRoleByUID(dbDefault, up.UserId)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s memperbarui profil akun pada waktu %s",
			ip,
			email,
			role,
			time.Now().Format(time.Now().Format("2006-01-02 15:04:05")),
		),
		role,
	)

	return map[string]any{
		"updated_fields": updates,
	}, nil
}
