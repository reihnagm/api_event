package services

import (
	"errors"
	"superapps/entities"
	helper "superapps/helpers"
)

func UpdateAccount(up *entities.UpdateAccount) (map[string]any, error) {

	query := `UPDATE accounts SET no = ? WHERE user_id = ?`

	err := dbDefault.Exec(query, up.No).Error

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	return map[string]any{
		"data": up,
	}, nil
}
