package services

import (
	"superapps/entities"
	helper "superapps/helpers"
)

func CompnayTypeList() (map[string]any, error) {
	var companyTypeListScan []entities.CompanyTypeListScan

	query := `SELECT id, name FROM type_of_companies`

	err := dbDefault.Raw(query).Scan(&companyTypeListScan).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{
		"data": companyTypeListScan,
	}, nil
}

func CompnayPlaceTypeList() (map[string]any, error) {
	var companyTypeListScan []entities.CompanyTypeListScan

	query := `SELECT id, name FROM type_of_company_places`

	err := dbDefault.Raw(query).Scan(&companyTypeListScan).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{
		"data": companyTypeListScan,
	}, nil
}
