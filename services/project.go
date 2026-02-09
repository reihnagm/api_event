package services

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"superapps/entities"
	helper "superapps/helpers"
	"time"

	uuid "github.com/satori/go.uuid"
)

func ProjectCostOfFundTemplateList() (map[string]any, error) {

	var ProjectCostOfFundTemplateList []entities.ProjectCostOfFundTemplateListScan

	query := `SELECT id, name, percentage, description, fixed_amount, payment_split, created_at, updated_at FROM cost_of_fund_templates`

	err := dbDefault.Raw(query).Scan(&ProjectCostOfFundTemplateList).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{
		"data": ProjectCostOfFundTemplateList,
	}, nil
}

func ProjectCostOfFundTemplateWithoutList() (map[string]any, error) {

	var ProjectCostOfFundTemplateList []entities.ProjectCostOfFundTemplateWithoutListScan

	query := `SELECT id, name, percentage, description, fixed_amount, created_at, updated_at 
	FROM cost_of_fund_template_without_divided
	WHERE enabled = '1'`

	err := dbDefault.Raw(query).Scan(&ProjectCostOfFundTemplateList).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{
		"data": ProjectCostOfFundTemplateList,
	}, nil
}

func ProjectCostOfFundTemplateDetail(id string) (map[string]any, error) {

	var ProjectCostOfFundTemplateDetail entities.ProjectCostOfFundTemplateDetailScan

	query := `SELECT id, name, percentage, description, fixed_amount, payment_split, created_at, updated_at 
	FROM cost_of_fund_templates 
	WHERE id = ?`

	err := dbDefault.Raw(query, id).Scan(&ProjectCostOfFundTemplateDetail).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{
		"data": ProjectCostOfFundTemplateDetail,
	}, nil
}

func ProjectCostOfFundTemplateWithoutDetail(id string) (map[string]any, error) {

	var ProjectCostOfFundTemplateDetail entities.ProjectCostOfFundTemplateWithoutDetailScan

	query := `SELECT id, name, percentage, description, fixed_amount, created_at, updated_at 
	FROM cost_of_fund_template_without_divided
	WHERE id = ?`

	err := dbDefault.Raw(query, id).Scan(&ProjectCostOfFundTemplateDetail).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{
		"data": ProjectCostOfFundTemplateDetail,
	}, nil
}

func ProjectCostOfFundTemplateStore(pcofts *entities.ProjectCostOfFundTemplateStore) (map[string]any, error) {

	query := `INSERT INTO cost_of_fund_templates (name, percentage, fixed_amount, payment_split) VALUES (?, ?, ?, ?)`

	err := dbDefault.Raw(query, pcofts.Name, pcofts.Percentage, pcofts.FixedAmount, pcofts.PaymentSplit).Exec(query).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{}, nil
}

func ProjectCostOfFundTemplateDelete(id string) (map[string]any, error) {

	query := `DELETE FROM cost_of_fund_templates WHERE id = ?`

	err := dbDefault.Raw(query, id).Exec(query).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{}, nil
}

func ProjectCostOfFundTemplateUpdate(pcofts *entities.ProjectCostOfFundTemplateUpdate) (map[string]any, error) {

	query := `UPDATE cost_of_fund_templates SET name = ?, percentage = ?, fixed_amount = ?, payment_split = ? WHERE id = ?`

	err := dbDefault.Raw(query, pcofts.Name, pcofts.Percentage, pcofts.FixedAmount, pcofts.PaymentSplit, pcofts.Id).Exec(query).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{}, nil
}

func ProjectCostOfFundTemplateWithoutStore(pcofts *entities.ProjectCostOfFundTemplateWithoutStore) (map[string]any, error) {

	query := `INSERT INTO cost_of_fund_template_without_divided (name, percentage, fixed_amount) VALUES (?, ?, ?)`

	err := dbDefault.Raw(query, pcofts.Name, pcofts.Percentage, pcofts.FixedAmount).Exec(query).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{}, nil
}

func ProjectCostOfFundTemplateWithoutDelete(id string) (map[string]any, error) {

	query := `DELETE FROM cost_of_fund_template_without_divided WHERE id = ?`

	err := dbDefault.Raw(query, id).Exec(query).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{}, nil
}

func ProjectCostOfFundTemplateWithoutUpdate(pcofts *entities.ProjectCostOfFundTemplateWithoutUpdate) (map[string]any, error) {

	query := `UPDATE cost_of_fund_template_without_divided 
	SET name = ?, description = ?, percentage = ?, fixed_amount = ? 
	WHERE id = ?`

	err := dbDefault.Raw(query, pcofts.Name, pcofts.Description, pcofts.Percentage, pcofts.FixedAmount, pcofts.Id).Exec(query).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{}, nil
}

func ProjectAuthorityTypeList() (map[string]any, error) {
	var ProjectAuthorityTypeList []entities.ProjectAuthorityTypeListScan

	query := `SELECT id, name FROM type_of_contracting_authorities`

	err := dbDefault.Raw(query).Scan(&ProjectAuthorityTypeList).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{
		"data": ProjectAuthorityTypeList,
	}, nil
}

func ProjectTypeList() (map[string]any, error) {
	var ProjectTypeList []entities.ProjectTypeListScan

	query := `SELECT id, name FROM type_of_projects`

	err := dbDefault.Raw(query).Scan(&ProjectTypeList).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{
		"data": ProjectTypeList,
	}, nil
}

func ProjectEmitenList(userId, status string) (map[string]any, error) {

	var project entities.ProjectListScan
	var dataProject = make([]entities.ProjectListResponse, 0)

	query := `SELECT p.uid AS id, p.title, p.goal, p.capital, p.code_effect, p.min_invest, p.unit_price, p.unit_total,
	p.number_of_unit, p.periode, ps.name AS status,  top.name AS type_of_project, p.nominal_value, p.time_periode, p.interest_rate, p.interest_payment_schedule,
	p.principal_payment_schedule, p.use_of_funds, p.collateral_guarantee, p.desc_job, p.is_apbn, p.is_approved
	FROM projects p
	INNER JOIN companies c ON c.uid = p.company_id 
	INNER JOIN project_statuses ps ON ps.id = p.status
	INNER JOIN type_of_projects top ON top.id = p.type_of_project
	WHERE c.user_id = ? AND ps.name = ?`

	var rows *sql.Rows
	var err error

	rows, err = dbDefault.Raw(query, userId, status).Rows()

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		errProjectRows := dbDefault.ScanRows(rows, &project)
		if errProjectRows != nil {
			helper.Logger("error", "In Server: "+errProjectRows.Error())
			return nil, errProjectRows
		}

		dataProjectMedia := make([]entities.ProjectMedia, 0)

		queryProjectMedia := `SELECT id, path FROM project_medias WHERE project_id = ?`

		errProjectMedia := dbDefault.
			Raw(queryProjectMedia, project.Id).
			Scan(&dataProjectMedia).Error

		if errProjectMedia != nil {
			helper.Logger("error", "In Server: "+errProjectMedia.Error())
			return nil, errProjectMedia
		}

		var dataProjectLoc entities.ProjectLocation

		queryProjectLoc := `SELECT id, url, name, lat, lng FROM project_locations WHERE project_id = ?`

		errProjectLoc := dbDefault.
			Raw(queryProjectLoc, project.Id).
			Scan(&dataProjectLoc).Error

		if errProjectLoc != nil {
			helper.Logger("error", "In Server: "+errProjectLoc.Error())
			return nil, errProjectLoc
		}

		var dataProjectDoc entities.ProjectDoc

		queryProjectDoc := `SELECT id, path FROM project_docs WHERE project_id = ?`

		errProjectDoc := dbDefault.
			Raw(queryProjectDoc, project.Id).
			Scan(&dataProjectDoc).Error

		if errProjectDoc != nil {
			helper.Logger("error", "In Server: "+errProjectDoc.Error())
			return nil, errProjectDoc
		}

		var dataProjectCollateralGuarantee []entities.ProjectCollateralGuarantee

		queryProjectCollateralGurantee := `SELECT id, name FROM collateral_guarantees WHERE project_id = ?`

		errProjectCollateralGurantee := dbDefault.
			Raw(queryProjectCollateralGurantee, project.Id).
			Scan(&dataProjectCollateralGuarantee).Error

		if errProjectCollateralGurantee != nil {
			helper.Logger("error", "In Server: "+errProjectCollateralGurantee.Error())
			return nil, errProjectCollateralGurantee
		}

		if dataProjectCollateralGuarantee == nil {
			dataProjectCollateralGuarantee = []entities.ProjectCollateralGuarantee{}
		}

		var dataProjectUseOfFunds []entities.ProjectUseOfFunds

		queryProjectUseOfFunds := `SELECT id, name FROM use_of_funds WHERE project_id = ?`

		errProjectUseOfFunds := dbDefault.
			Raw(queryProjectUseOfFunds, project.Id).
			Scan(&dataProjectUseOfFunds).Error

		if errProjectUseOfFunds != nil {
			helper.Logger("error", "In Server: "+errProjectUseOfFunds.Error())
			return nil, errProjectUseOfFunds
		}

		if dataProjectUseOfFunds == nil {
			dataProjectUseOfFunds = []entities.ProjectUseOfFunds{}
		}

		dataProject = append(dataProject, entities.ProjectListResponse{
			Id:      project.Id,
			Title:   project.Title,
			Goal:    project.Goal,
			Capital: project.Capital,
			Medias:  dataProjectMedia,
			Location: entities.ProjectLocation{
				Id:   dataProjectLoc.Id,
				Url:  helper.DefaultIfEmpty(dataProjectLoc.Url, "-"),
				Name: helper.DefaultIfEmpty(dataProjectLoc.Name, "-"),
				Lat:  helper.DefaultIfEmpty(dataProjectLoc.Lat, "-"),
				Lng:  helper.DefaultIfEmpty(dataProjectLoc.Lng, "-"),
			},
			Doc: entities.ProjectDoc{
				Id:   dataProjectDoc.Id,
				Path: helper.DefaultIfEmpty(dataProjectDoc.Path, "-"),
			},
			MinInvest:                project.MinInvest,
			HargaUnit:                project.UnitPrice,
			UnitTotal:                project.UnitTotal,
			JumlahUnit:               project.NumberOfUnit,
			KodeEfek:                 project.CodeEffect,
			Periode:                  project.Periode,
			TypeOfProject:            project.TypeOfProject,
			NominalValue:             project.NominalValue,
			TimePeriode:              project.TimePeriode,
			InterestRate:             project.InterestRate,
			InterestPaymentSchedule:  project.InterestPaymentSchedule,
			PrincipalPaymentSchedule: project.PrincipalPaymentSchedule,
			UseOfFunds:               dataProjectUseOfFunds,
			CollateralGuarantee:      dataProjectCollateralGuarantee,
			DescJob:                  project.DescJob,
			IsApbn:                   project.IsApbn,
			Status:                   project.Status,
			IsApproved:               project.IsApproved,
			CreatedAt:                project.CreatedAt,
			UpdatedAt:                project.UpdatedAt,
		})
	}

	return map[string]any{
		"data": dataProject,
	}, nil
}

func ProjectList() (map[string]any, error) {
	var project entities.ProjectListScan
	var dataProject = make([]entities.ProjectListResponse, 0)

	query := `SELECT
		p.uid AS id,
		p.title,
		p.goal,
		p.capital,
		p.min_invest,
		p.unit_price,
		p.unit_total,
		p.number_of_unit,
		p.periode,
		p.amount_shares_per_lot,
		top.name AS type_of_project,
		p.nominal_value,
		p.time_periode,
		p.interest_rate,
		p.interest_payment_schedule,
		p.principal_payment_schedule,
		p.use_of_funds,
		p.collateral_guarantee,
		p.desc_job,
		p.is_apbn,
		p.target_amount_idr AS target_amount,

		COALESCE(pyagg.user_paid, 0) AS user_paid,
		COALESCE(pyagg.investors_paid, 0) AS investors_paid,
		COALESCE(pyagg.can_refund, 0) AS can_refund,

		ps.name AS status,
		p.is_approved,
		p.profit_percentage,
		p.loan_term,
		p.code_effect,
		p.doc_prospect,
		p.start_project,
		p.end_project,
		c.user_id,

		p.provider_address,
		p.provider_province_name,
		p.provider_city_name,
		p.provider_district_name,
		p.provider_subdistrict_name,
		p.provider_postal_code,

		p.created_at,
		p.updated_at
		FROM projects p
		JOIN companies        c   ON c.uid = p.company_id
		JOIN type_of_projects top ON top.id = p.type_of_project
		JOIN project_statuses ps  ON ps.id  = p.status
		LEFT JOIN (
		SELECT
			py.project_uid,
			SUM(CASE WHEN py.payment_status='PAID' THEN py.amount_idr ELSE 0 END) AS user_paid,
			COUNT(DISTINCT CASE WHEN py.payment_status='PAID' THEN j.user_id END) AS investors_paid,
			MAX(
			CASE
				WHEN py.payment_status = 'PAID'
				AND py.paid_at IS NOT NULL
				AND py.paid_at >= (UTC_TIMESTAMP() - INTERVAL 2 HOUR)
				THEN 1 ELSE 0
			END
			) AS can_refund
		FROM payments py
		LEFT JOIN jobs j ON j.id = py.investor_job_id
		GROUP BY py.project_uid
		) pyagg ON pyagg.project_uid = p.uid
		WHERE p.status = 5
		ORDER BY p.created_at DESC;
	`

	var rows *sql.Rows
	var err error

	rows, err = dbDefault.Raw(query).Rows()
	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		errProjectRows := dbDefault.ScanRows(rows, &project)
		if errProjectRows != nil {
			helper.Logger("error", "In Server: "+errProjectRows.Error())
			return nil, errProjectRows
		}

		// --- Project Media ---
		var dataProjectMedia = make([]entities.ProjectMedia, 0)
		queryProjectMedia := `SELECT id, path FROM project_medias WHERE project_id = ?`
		if err := dbDefault.
			Raw(queryProjectMedia, project.Id).
			Scan(&dataProjectMedia).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return nil, err
		}

		// --- Project Location ---
		var dataProjectLoc entities.ProjectLocation
		queryProjectLoc := `SELECT id, url, name, lat, lng FROM project_locations WHERE project_id = ?`
		if err := dbDefault.
			Raw(queryProjectLoc, project.Id).
			Scan(&dataProjectLoc).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return nil, err
		}

		// --- Project Doc ---
		var dataProjectDoc entities.ProjectDoc
		queryProjectDoc := `SELECT id, path FROM project_docs WHERE project_id = ?`
		if err := dbDefault.
			Raw(queryProjectDoc, project.Id).
			Scan(&dataProjectDoc).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return nil, err
		}

		// --- Collateral Guarantee ---
		var dataProjectCollateralGuarantee []entities.ProjectCollateralGuarantee
		queryProjectCollateralGurantee := `SELECT id, name FROM collateral_guarantees WHERE project_id = ?`
		if err := dbDefault.
			Raw(queryProjectCollateralGurantee, project.Id).
			Scan(&dataProjectCollateralGuarantee).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return nil, err
		}
		if dataProjectCollateralGuarantee == nil {
			dataProjectCollateralGuarantee = []entities.ProjectCollateralGuarantee{}
		}

		// --- Use Of Funds ---
		var dataProjectUseOfFunds []entities.ProjectUseOfFunds
		queryProjectUseOfFunds := `SELECT id, name FROM use_of_funds WHERE project_id = ?`
		if err := dbDefault.
			Raw(queryProjectUseOfFunds, project.Id).
			Scan(&dataProjectUseOfFunds).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return nil, err
		}
		if dataProjectUseOfFunds == nil {
			dataProjectUseOfFunds = []entities.ProjectUseOfFunds{}
		}

		// --- Company ---
		var dataProjectCompany entities.ProjectCompany
		queryProjectCompany := `SELECT c.company_name, tob.name AS jenis_usaha
        FROM companies c
        INNER JOIN type_of_businesses tob ON tob.id = c.type_of_business  
        WHERE c.user_id = ?`
		if err := dbDefault.
			Raw(queryProjectCompany, project.UserId).
			Scan(&dataProjectCompany).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return nil, err
		}

		// --- expiry logic (45-day window based on updated_at) ---
		const expiryWindowDays = 45
		var remainingDays int
		var projectIsExpire bool

		if !project.UpdatedAt.IsZero() {
			now := time.Now().UTC()
			updated := project.UpdatedAt.UTC()
			elapsedDays := int(math.Floor(now.Sub(updated).Hours() / 24.0))

			// expired if strictly more than 45 days since updated_at
			projectIsExpire = elapsedDays > expiryWindowDays

			if elapsedDays >= expiryWindowDays {
				remainingDays = 0
			} else if elapsedDays >= 0 {
				remainingDays = expiryWindowDays - elapsedDays
			}
		} else {
			// policy: treat missing updated_at as expired
			projectIsExpire = true
			remainingDays = 0
		}
		// --- end expiry logic ---

		// --- Amount Shares Per Lot ---
		var TotalPaidAmountSharesPerLot int64
		queryAmountSharesPerLot := `SELECT
			COALESCE(SUM(paid_amount_shares_per_lot), 0) AS total_paid_amount_shares_per_lot
			FROM invoices
			WHERE project_uid = ?
			AND invoice_status = 'PAID';
		`
		if err := dbDefault.
			Raw(queryAmountSharesPerLot, project.Id).
			Scan(&TotalPaidAmountSharesPerLot).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return nil, err
		}

		var stockLot int64

		if project.AmountSharesPerLot == 0 {
			stockLot = 0 // or return error
		} else {
			stockLot = (project.NumberOfUnit / project.AmountSharesPerLot) - TotalPaidAmountSharesPerLot
		}

		dataProject = append(dataProject, entities.ProjectListResponse{
			Id:             project.Id,
			Title:          project.Title,
			UserPaidAmount: project.UserPaid,
			TargetAmount:   project.TargetAmount,
			InvestorPaid:   project.InvestorsPaid,
			Goal:           project.Goal,
			Capital:        project.Capital,
			Medias:         dataProjectMedia,
			Location: entities.ProjectLocation{
				Id:   dataProjectLoc.Id,
				Url:  helper.DefaultIfEmpty(dataProjectLoc.Url, "-"),
				Name: helper.DefaultIfEmpty(dataProjectLoc.Name, "-"),
				Lat:  helper.DefaultIfEmpty(dataProjectLoc.Lat, "-"),
				Lng:  helper.DefaultIfEmpty(dataProjectLoc.Lng, "-"),
			},
			Doc: entities.ProjectDoc{
				Id:   dataProjectDoc.Id,
				Path: helper.DefaultIfEmpty(dataProjectDoc.Path, "-"),
			},
			CanRefund:                project.CanRefund,
			MinInvest:                project.MinInvest,
			HargaUnit:                project.UnitPrice,
			UnitTotal:                project.UnitTotal,
			JumlahLot:                project.AmountSharesPerLot,
			JumlahUnit:               project.NumberOfUnit,
			StokLot:                  stockLot,
			KodeEfek:                 project.CodeEffect,
			Periode:                  helper.DefaultIfEmpty(project.Periode, "-"),
			TypeOfProject:            project.TypeOfProject,
			NominalValue:             project.NominalValue,
			TimePeriode:              project.TimePeriode,
			InterestRate:             project.InterestRate,
			InterestPaymentSchedule:  project.InterestPaymentSchedule,
			PrincipalPaymentSchedule: project.PrincipalPaymentSchedule,
			UseOfFunds:               dataProjectUseOfFunds,
			CollateralGuarantee:      dataProjectCollateralGuarantee,
			DescJob:                  project.DescJob,
			IsApbn:                   project.IsApbn,
			IsApproved:               project.IsApproved,
			MulaiProject:             project.StartProject,
			SelesaiProject:           project.EndProject,
			AlamatPenyediaProject:    project.ProviderAddress,
			AlamatPenyediaProvinsi:   project.ProviderProvinceName,
			AlamatPenyediaKota:       project.ProviderCityName,
			AlamatPenyediaDaerah:     project.ProviderDistrictName,
			AlamatPenyediaWilayah:    project.ProviderSubdistrictName,
			AlamatPenyediaKodePos:    project.ProviderPostalCode,
			RemainingDays:            remainingDays,
			ProjectIsExpire:          projectIsExpire,
			Status:                   project.Status,
			LoanTerm:                 project.LoanTerm,
			Roi:                      project.ProfitPercentage,
			DocProspect:              project.DocProspect,
			Company: entities.Company{
				Name:       helper.DefaultIfEmpty(dataProjectCompany.CompanyName, "-"),
				JenisUsaha: helper.DefaultIfEmpty(dataProjectCompany.JenisUsaha, "-"),
			},
			CreatedAt: project.CreatedAt,
			UpdatedAt: project.UpdatedAt,
		})
	}

	return map[string]any{
		"data": dataProject,
	}, nil
}

func ProjectDetail(id string) (map[string]any, error) {
	var project entities.ProjectListScan

	query := `SELECT
		p.uid AS id,
		p.title,
		p.goal,
		p.capital,
		p.min_invest,
		p.unit_price,
		p.unit_total,
		p.number_of_unit,
		p.periode,
		top.name AS type_of_project,
		p.nominal_value,
		p.time_periode,
		p.interest_rate,
		p.interest_payment_schedule,
		p.principal_payment_schedule,
		p.use_of_funds,
		p.amount_shares_per_lot,
		p.collateral_guarantee,
		p.desc_job,
		p.is_apbn,
		p.target_amount_idr AS target_amount,

		/* agregat hanya pembayaran PAID */
		COALESCE(SUM(CASE WHEN py.payment_status='PAID' THEN py.amount_idr END), 0)     AS user_paid,
		COALESCE(COUNT(DISTINCT CASE WHEN py.payment_status='PAID' THEN j.user_id END), 0) AS investors_paid,
		COALESCE(MAX(
		CASE
			WHEN py.payment_status = 'PAID'
			AND py.paid_at IS NOT NULL
			AND py.paid_at >= (UTC_TIMESTAMP() - INTERVAL 2 HOUR) -- < 2 jam
			THEN 1 ELSE 0
		END
		), 0) AS can_refund,

		ps.name AS status,
		p.is_approved,
		p.profit_percentage,
		p.loan_term,
		p.code_effect,
		p.doc_prospect,
		p.start_project,
		p.end_project,
		c.user_id,                              -- pemilik/company

		p.provider_address,
		p.provider_province_name,
		p.provider_city_name,
		p.provider_district_name,
		p.provider_subdistrict_name,
		p.provider_postal_code,

		p.created_at,
		p.updated_at
		FROM projects p
		JOIN companies        c   ON c.uid = p.company_id
		JOIN type_of_projects top ON top.id = p.type_of_project
		JOIN project_statuses ps  ON ps.id  = p.status
		LEFT JOIN payments    py  ON py.project_uid = p.uid        -- <— LEFT JOIN
		LEFT JOIN jobs        j   ON j.id = py.investor_job_id     -- untuk hitung investor (user_id)

		WHERE p.uid = ?

		GROUP BY p.uid
		ORDER BY p.created_at DESC;
`

	if err := dbDefault.Raw(query, id).Scan(&project).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	if strings.TrimSpace(project.Id) == "" {
		return map[string]any{
			"message": "project not found",
		}, nil
	}

	// Fetch project medias
	dataProjectMedia := make([]entities.ProjectMedia, 0)
	queryProjectMedia := `SELECT id, path FROM project_medias WHERE project_id = ?`
	if err := dbDefault.Raw(queryProjectMedia, project.Id).Scan(&dataProjectMedia).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	// Fetch project location
	var dataProjectLoc entities.ProjectLocation
	queryProjectLoc := `SELECT id, url, name, lat, lng FROM project_locations WHERE project_id = ?`
	if err := dbDefault.Raw(queryProjectLoc, project.Id).Scan(&dataProjectLoc).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	// Fetch project docs
	var dataProjectDoc entities.ProjectDoc
	queryProjectDoc := `SELECT id, path FROM project_docs WHERE project_id = ?`
	if err := dbDefault.Raw(queryProjectDoc, project.Id).Scan(&dataProjectDoc).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	// Fetch project company
	var dataProjectCompany entities.ProjectCompany
	queryProjectCompany := `SELECT c.company_name, tob.name AS jenis_usaha
	FROM companies c
	INNER JOIN type_of_businesses tob ON tob.id = c.type_of_business  
	WHERE c.user_id = ?`
	if err := dbDefault.Raw(queryProjectCompany, project.UserId).Scan(&dataProjectCompany).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	// --- expiry logic (45-day window based on updated_at) ---
	const expiryWindowDays = 45
	var remainingDays int
	var projectIsExpire bool

	if !project.UpdatedAt.IsZero() {
		now := time.Now().UTC()
		updated := project.UpdatedAt.UTC()
		elapsedDays := int(math.Floor(now.Sub(updated).Hours() / 24.0))

		projectIsExpire = elapsedDays > expiryWindowDays
		if elapsedDays >= expiryWindowDays {
			remainingDays = 0
		} else if elapsedDays >= 0 {
			remainingDays = expiryWindowDays - elapsedDays
		}
	} else {
		projectIsExpire = true
		remainingDays = 0
	}
	// --- end expiry logic ---

	// --- Amount Shares Per Lot ---
	var TotalPaidAmountSharesPerLot int64
	queryAmountSharesPerLot := `SELECT
			COALESCE(SUM(paid_amount_shares_per_lot), 0) AS total_paid_amount_shares_per_lot
			FROM invoices
			WHERE project_uid = ?
			AND invoice_status = 'PAID';
		`
	if err := dbDefault.
		Raw(queryAmountSharesPerLot, project.Id).
		Scan(&TotalPaidAmountSharesPerLot).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	var stockLot int64

	if project.AmountSharesPerLot == 0 {
		stockLot = 0 // or return error
	} else {
		stockLot = (project.NumberOfUnit / project.AmountSharesPerLot) - TotalPaidAmountSharesPerLot
	}

	projectDetail := entities.ProjectListResponse{
		Id:             project.Id,
		Title:          project.Title,
		TargetAmount:   project.TargetAmount,
		UserPaidAmount: project.UserPaid,
		InvestorPaid:   project.InvestorsPaid,
		Goal:           project.Goal,
		Capital:        project.Capital,
		Medias:         dataProjectMedia,
		Location: entities.ProjectLocation{
			Id:   dataProjectLoc.Id,
			Url:  helper.DefaultIfEmpty(dataProjectLoc.Url, "-"),
			Name: helper.DefaultIfEmpty(dataProjectLoc.Name, "-"),
			Lat:  helper.DefaultIfEmpty(dataProjectLoc.Lat, "-"),
			Lng:  helper.DefaultIfEmpty(dataProjectLoc.Lng, "-"),
		},
		Doc: entities.ProjectDoc{
			Id:   dataProjectDoc.Id,
			Path: helper.DefaultIfEmpty(dataProjectDoc.Path, "-"),
		},
		MinInvest:                project.MinInvest,
		HargaUnit:                project.UnitPrice,
		UnitTotal:                project.UnitTotal,
		JumlahUnit:               project.NumberOfUnit,
		JumlahLot:                project.AmountSharesPerLot,
		StokLot:                  stockLot,
		Periode:                  helper.DefaultIfEmpty(project.Periode, "-"),
		KodeEfek:                 project.CodeEffect,
		CanRefund:                project.CanRefund,
		TypeOfProject:            project.TypeOfProject,
		NominalValue:             project.NominalValue,
		TimePeriode:              project.TimePeriode,
		InterestRate:             project.InterestRate,
		InterestPaymentSchedule:  project.InterestPaymentSchedule,
		PrincipalPaymentSchedule: project.PrincipalPaymentSchedule,
		UseOfFunds:               []entities.ProjectUseOfFunds{},
		CollateralGuarantee:      []entities.ProjectCollateralGuarantee{},
		DescJob:                  project.DescJob,
		IsApbn:                   project.IsApbn,
		IsApproved:               project.IsApproved,
		MulaiProject:             project.StartProject,
		SelesaiProject:           project.EndProject,
		AlamatPenyediaProject:    project.ProviderAddress,
		AlamatPenyediaProvinsi:   project.ProviderProvinceName,
		AlamatPenyediaKota:       project.ProviderCityName,
		AlamatPenyediaDaerah:     project.ProviderDistrictName,
		AlamatPenyediaWilayah:    project.ProviderSubdistrictName,
		AlamatPenyediaKodePos:    project.ProviderPostalCode,
		RemainingDays:            remainingDays,
		ProjectIsExpire:          projectIsExpire,
		Status:                   project.Status,
		LoanTerm:                 project.LoanTerm,
		Roi:                      project.ProfitPercentage,
		DocProspect:              project.DocProspect,
		Company: entities.Company{
			Name:       helper.DefaultIfEmpty(dataProjectCompany.CompanyName, "-"),
			JenisUsaha: helper.DefaultIfEmpty(dataProjectCompany.JenisUsaha, "-"),
		},
		CreatedAt: project.CreatedAt,
		UpdatedAt: project.UpdatedAt,
	}

	return map[string]any{
		"data": projectDetail,
	}, nil
}

func ProjectStore(r *http.Request, ps *entities.ProjectStore) (map[string]any, error) {
	ps.Id = uuid.NewV4().String()

	queryInsertProject := `
	INSERT INTO projects (
		uid, company_id, title, desc_job, interest_rate, type_of_project, profit_percentage,
		time_periode, interest_payment_schedule, principal_payment_schedule,
		loa, spk, company_profile, expire_date, site, loan_term, is_apbn, capital,
		doc_bank_statement, doc_financial_statement, doc_contract, doc_prospect,
		contracting_authority, type_of_contracting_authority,
		start_project, end_project,
		provider_address, 
		provider_province_name, provider_city_name, 
		provider_district_name, provider_subdistrict_name, provider_postal_code, required_fund
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

	errInsertProject := dbDefault.Exec(
		queryInsertProject,
		ps.Id,
		ps.CompanyId,
		ps.Title,
		ps.Deskripsi,
		ps.TingkatBunga,
		ps.JenisProject,
		ps.PersentaseKeuntungan,
		ps.JangkaWaktu,
		ps.JadwalPembayaranBunga,
		ps.JadwalPembayaranPokok,
		ps.Loa,
		ps.Spk,
		ps.CompanyProfile,
		ps.BatasAkhirPengerjaan,
		ps.Website,
		ps.TenorPinjaman,
		ps.IsApbn,
		ps.Modal,
		ps.DocRekeningKoran,
		ps.DocLaporanKeuangan,
		ps.DocContract,
		ps.DocProspect,
		ps.InstansiPemberiProject,
		ps.JenisInstansiPemberiProject,
		ps.MulaiProject,
		ps.SelesaiProject,
		ps.AlamatPenyediaProject,
		ps.AlamatPenyediaProvinsi,
		ps.AlamatPenyediaKota,
		ps.AlamatPenyediaDaerah,
		ps.AlamatPenyediaWilayah,
		ps.AlamatPenyediaKodePos,
		ps.DanaYangDibutuhkan,
	).Error

	if errInsertProject != nil {
		helper.Logger("error", "In Server: "+errInsertProject.Error())
		return nil, errors.New(errInsertProject.Error())
	}

	// Store Project Jaminan Kolateral
	for _, p := range ps.JaminanKolateral {
		queryInsertCollateralGuarantees := `INSERT INTO collateral_guarantees (name, project_id) VALUES (?, ?)`

		if err := dbDefault.Exec(queryInsertCollateralGuarantees, p.Name, ps.Id).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return nil, fmt.Errorf("failed to project collateral: %w", err)
		}
	}

	// Store Project Penggunaan Dana
	for _, p := range ps.PenggunaanDana {
		queryInsertUseOfFunds := `INSERT INTO use_of_funds (name, project_id) VALUES (?, ?)`

		if err := dbDefault.Exec(queryInsertUseOfFunds, p.Name, ps.Id).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return nil, fmt.Errorf("failed to project use of funds: %w", err)
		}
	}

	// Store Project Media
	queryInsertProjectMedia := `INSERT INTO project_medias (project_id, path) VALUES (?, ?)`
	for _, v := range ps.Media {
		if err := dbDefault.Exec(queryInsertProjectMedia, ps.Id, v.Path).Error; err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return nil, errors.New(err.Error())
		}
	}

	// Store Project Location
	queryInsertProjectLocation := `INSERT INTO project_locations (project_id, name, url, lat, lng) VALUES (?, ?, ?, ?, ?)`
	if err := dbDefault.Exec(
		queryInsertProjectLocation,
		ps.Id, ps.Location.Name, ps.Location.Url, ps.Location.Lat, ps.Location.Lng,
	).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	// Store Project Contract
	queryInsertProjectContract := `INSERT INTO project_contracts (project_id, value, path) VALUES (?, ?, ?)`
	if err := dbDefault.Exec(queryInsertProjectContract, ps.Id, ps.NoContractValue, ps.NoContractPath).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	// === Broadcast: Pengingat untuk Project Analyst ===
	emails, _ := helper.GetEmailsByRoleProjectAnalyst(dbDefault)

	for _, email := range emails {
		subject := fmt.Sprintf(`Pengingat Project Analyst: Proyek "%s" membutuhkan peninjauan`, safe(ps.Title, "-"))

		body := fmt.Sprintf(`<!doctype html>
			<html lang="id">
			<head>
				<meta charset="utf-8">
				<title>%[1]s</title>
				<meta name="viewport" content="width=device-width, initial-scale=1">
				<meta name="x-apple-disable-message-reformatting">
			</head>
			<body style="margin:0;padding:0;background:#f6f7f9;font-family:Arial,Helvetica,sans-serif;">
				<!-- Preheader (disembunyikan) -->
				<div style="display:none;visibility:hidden;opacity:0;height:0;width:0;overflow:hidden;color:transparent;">
					Pengingat: Proyek membutuhkan peninjauan oleh Project Analyst.
				</div>

				<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background:#f6f7f9;padding:24px 0;">
				<tr>
					<td align="center">
					<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="600" style="background:#ffffff;border-radius:12px;overflow:hidden;border:1px solid #e5e7eb;">
						<!-- Header -->
						<tr>
							<td style="padding:20px 24px;background:#111827;color:#ffffff;">
								<h1 style="margin:0;font-size:18px;line-height:1.4;">Pengingat Project Analyst</h1>
								<p style="margin:6px 0 0 0;font-size:12px;opacity:.85;">Proyek membutuhkan peninjauan kelayakan</p>
							</td>
						</tr>

						<!-- Body -->
						<tr>
							<td style="padding:24px;">
								<p style="margin:0 0 14px 0;color:#374151;font-size:14px;line-height:1.7;">
									Halo Tim Project Analyst,
								</p>
								<p style="margin:0 0 18px 0;color:#374151;font-size:14px;line-height:1.7;">
									Ini adalah pengingat bahwa terdapat proyek baru yang telah disimpan dan membutuhkan peninjauan Anda.
								</p>

								<!-- Detail Proyek -->
								<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="border-collapse:separate;border-spacing:0 8px;">
									<tr>
										<td style="width:40%%;padding:10px 12px;background:#f9fafb;color:#6b7280;font-size:13px;border:1px solid #e5e7eb;border-right:0;border-radius:8px 0 0 8px;">Judul</td>
										<td style="padding:10px 12px;background:#ffffff;color:#111827;font-size:13px;border:1px solid #e5e7eb;border-radius:0 8px 8px 0;">%[3]s</td>
									</tr>
									<tr>
										<td style="width:40%%;padding:10px 12px;background:#f9fafb;color:#6b7280;font-size:13px;border:1px solid #e5e7eb;border-right:0;border-radius:8px 0 0 8px;">Dana Dibutuhkan</td>
										<td style="padding:10px 12px;background:#ffffff;color:#111827;font-size:13px;border:1px solid #e5e7eb;border-radius:0 8px 8px 0;">%[5]s</td>
									</tr>
									<tr>
										<td style="width:40%%;padding:10px 12px;background:#f9fafb;color:#6b7280;font-size:13px;border:1px solid #e5e7eb;border-right:0;border-radius:8px 0 0 8px;">Dibuat pada</td>
										<td style="padding:10px 12px;background:#ffffff;color:#111827;font-size:13px;border:1px solid #e5e7eb;border-radius:0 8px 8px 0;">%[6]s</td>
									</tr>
								</table>
							</td>
						</tr>

						<!-- Footer -->
						<tr>
							<td style="background:#f3f4f6;color:#6b7280;padding:14px 24px;text-align:center;font-size:12px;">
								&copy; %[7]d Fulusme. All rights reserved.
							</td>
						</tr>
					</table>
					</td>
				</tr>
				</table>
				</body>
			</html>`,
			subject,             // %[1]s
			safe(ps.Title, "-"), // %[2]s
			safe(helper.FormatIDRInt(ps.DanaYangDibutuhkan), "-"), // %[3]s
			safe(helper.FormatDate(time.Now()), "-"),              // %[4]s
			time.Now().Year(),                                     // %[5]d
		)

		if err := helper.SendEmail(email, "Fulusme", subject, body, "another-otp"); err != nil {
			helper.Logger("error", fmt.Sprintf("gagal kirim email ke %s: %v", email, err))
			continue
		}
		helper.Logger("info", fmt.Sprintf("pengingat Project Analyst terkirim ke %s", email))
	}

	email, _ := helper.GetEmailByProjectId(dbDefault, ps.Id)
	role, _ := helper.GetRoleByProjectId(dbDefault, ps.Id)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s melakukan aktivitas membuat project [%s] pada waktu %s",
			ip,
			email,
			role,
			ps.Title,
			time.Now().Format(time.Now().Format("2006-01-02 15:04:05")),
		),
		role,
	)

	return map[string]any{}, nil
}

func ProjectStoreMedia(pm *entities.ProjectStoreMedia) (map[string]any, error) {

	queryInsertProjectMedia := `INSERT INTO project_medias (id, project_id, path) VALUES (?, ?, ?)`

	errInsertProjectMedia := dbDefault.Exec(queryInsertProjectMedia, pm.Id, pm.ProjectId, pm.Path).Error

	if errInsertProjectMedia != nil {
		helper.Logger("error", "In Server: "+errInsertProjectMedia.Error())
		return nil, errors.New(errInsertProjectMedia.Error())
	}

	return map[string]any{
		"data": pm,
	}, nil
}

func ProjectStoreLocation(pl *entities.ProjectStoreLocation) (map[string]any, error) {

	queryInsertProjectLocation := `INSERT INTO project_locations (project_id, name, url, lat, lng) VALUES (?, ?, ?, ?, ?)`

	errInsertProjectLocation := dbDefault.Exec(queryInsertProjectLocation, pl.ProjectId, pl.Name, pl.Url, pl.Lat, pl.Lng).Error

	if errInsertProjectLocation != nil {
		helper.Logger("error", "In Server: "+errInsertProjectLocation.Error())
		return nil, errors.New(errInsertProjectLocation.Error())
	}

	return map[string]any{
		"data": pl,
	}, nil
}

func ProjectInquiry(r *http.Request, project *entities.ProjectInquiry) (map[string]any, error) {
	returnData := []struct {
		ID                  uint    `json:"id"`
		ProjectID           string  `json:"project_id"`
		TemplateID          uint    `json:"template_id"`
		TemplateName        string  `json:"template_name"`
		TemplateDescription string  `json:"template_description"`
		Percentage          float64 `json:"percentage"`
		FixedAmount         float64 `json:"fixed_amount"`
		CalculatedAmount    float64 `json:"calculated_amount"`
	}{}

	tx := dbDefault.Begin()

	var count int64
	if err := tx.Model(&entities.ProjectCostOfFund{}).
		Where("project_id = ?", project.ProjectId).
		Count(&count).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if count == 0 {
		// for i := 1; i <= 7; i++ {
		// 	if err := tx.Exec(
		// 		`INSERT INTO project_cost_of_funds (project_id, template_id) VALUES (?, ?)`,
		// 		project.ProjectId, i,
		// 	).Error; err != nil {
		// 		tx.Rollback()
		// 		return nil, errors.New(err.Error())
		// 	}
		// }
		for _, id := range []int{8, 9, 10} {
			if err := tx.Exec(
				`INSERT INTO project_cost_of_funds (project_id, template_id) VALUES (?, ?)`,
				project.ProjectId, id,
			).Error; err != nil {
				tx.Rollback()
				return nil, errors.New(err.Error())
			}
		}
	}

	err := tx.Raw(`
		SELECT 
			p.id,
			p.project_id,
			p.template_id,
			t.name AS template_name,
			t.description AS template_description,
			t.percentage,
			t.fixed_amount,
			COALESCE(p.calculated_amount, 0) AS calculated_amount
		FROM project_cost_of_funds p
		JOIN cost_of_fund_template_without_divided t ON p.template_id = t.id
		WHERE p.project_id = ?
		ORDER BY p.template_id
	`, project.ProjectId).Scan(&returnData).Error

	if err != nil {
		tx.Rollback()
		return nil, errors.New(err.Error())
	}

	tx.Commit()

	var totalAmount float64
	for _, item := range returnData {
		totalAmount += item.CalculatedAmount
	}

	email, _ := helper.GetEmailByProjectId(dbDefault, project.ProjectId)
	title, _ := helper.GetTitleByProjectId(dbDefault, project.ProjectId)
	role, _ := helper.GetRoleByProjectId(dbDefault, project.ProjectId)

	ip := helper.GetClientIP(r)

	helper.SendLogs(
		dbDefault,
		fmt.Sprintf(
			"%s [%s] - %s mengirim inquiry pada project [%s] pada waktu %s",
			ip,
			email,
			role,
			title,
			time.Now().Format("2006-01-02 15:04:05"),
		),
		role,
	)

	return map[string]any{
		"status":       "success",
		"total_amount": totalAmount,
		"data":         returnData,
	}, nil
}
