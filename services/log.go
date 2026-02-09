package services

import (
	"superapps/entities"
	helper "superapps/helpers"
)

func LogList() ([]entities.LogList, error) {

	const queryLogList = `
		SELECT * FROM logs
	`
	var list []entities.LogList
	if err := dbDefault.Raw(queryLogList).Scan(&list).Error; err != nil {
		helper.Logger("error", "In Server (list log by role): "+err.Error())
		return nil, err
	}

	if len(list) == 0 {
		list = []entities.LogList{}
	}
	return list, nil
}
