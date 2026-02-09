package services

import (
	"superapps/entities"
	helper "superapps/helpers"
)

func BusinessTypeList() (map[string]any, error) {
	var bussinesListScan []entities.BusinessTypeListScan

	query := `SELECT id, name FROM type_of_businesses`

	err := dbDefault.Raw(query).Scan(&bussinesListScan).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, err
	}

	return map[string]any{
		"data": bussinesListScan,
	}, nil
}
