package services

import (
	"superapps/entities"
)

func InitializeFcm(data *entities.Fcm) (map[string]any, error) {
	query := `
		INSERT INTO fcms (user_id, token)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE
			token = VALUES(token),
			updated_at = NOW()
	`

	tx := dbDefault.Exec(query, data.UserId, data.Token)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return map[string]any{
		"data":          data,
		"rows_affected": tx.RowsAffected,
	}, nil
}
