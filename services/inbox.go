package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"superapps/entities"
	helper "superapps/helpers"
)

func InboxList(userId string, isAdmin string) (map[string]any, error) {
	var inbox []entities.InboxList

	// Base query
	query := `
		SELECT i.id, i.title, i.content, it.name AS type, ps.name AS status,
		i.field_1 as field1, i.field_2 as field2, i.field_3 as field3, i.field_4 as field4,
		i.field_5 as field5, i.field_6 as field6, i.field_7 as field7, i.field_8 as field8,
		i.is_read, i.data, i.created_at
		FROM inboxes i 
		INNER JOIN inbox_types it ON it.id = i.type
		INNER JOIN project_statuses ps ON ps.id = i.status
	`

	var err error
	if isAdmin == "true" {
		// Admin can see all inboxes
		err = dbDefault.Raw(query).Scan(&inbox).Error
	} else {
		// Regular user: filter by receiver_id
		query += " WHERE receiver_id = ?"
		err = dbDefault.Raw(query, userId).Scan(&inbox).Error
	}

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	return map[string]any{
		"data": inbox,
	}, nil
}

func InboxDetail(id string) (map[string]any, error) {
	var inbox entities.InboxDetail

	queryGetInbox := `SELECT i.id, i.title, i.content, it.name AS type, i.data, ps.name AS status, i.field_1 as field1, i.field_2 as field2, i.field_3 as field3, i.field_4 as field4, i.field_5 as field5, i.field_6 as field6,  i.field_7 as field7, i.field_8 as field8, i.is_read, i.created_at 
	FROM inboxes i 
	INNER JOIN inbox_types it ON it.id = i.type 
	INNER JOIN project_statuses ps ON ps.id = i.status
	WHERE i.id = ?`

	errGetInbox := dbDefault.Raw(queryGetInbox, id).Scan(&inbox).Error
	if errGetInbox != nil {
		helper.Logger("error", "In Server: "+errGetInbox.Error())
		return nil, errors.New(errGetInbox.Error())
	}

	queryUpdateInbox := `UPDATE inboxes SET is_read = 1 WHERE id = ?`
	errUpdateInbox := dbDefault.Exec(queryUpdateInbox, id).Error
	if errUpdateInbox != nil {
		helper.Logger("error", "In Server: "+errUpdateInbox.Error())
		return nil, errors.New(errUpdateInbox.Error())
	}

	return map[string]any{
		"data": inbox,
	}, nil
}

func InboxStore(i *entities.InboxStore) (map[string]any, error) {
	var cutPrice any

	if i.Field1 != "" {
		num, err := strconv.Atoi(i.Field1)
		if err != nil {
			fmt.Println("Error converting string to int:", err)
			return nil, errors.New("field_1 harus berupa angka")
		}
		cutPrice = num / 2
	} else {
		cutPrice = nil
	}

	var dataJSON any = nil
	if i.Data != nil {
		b, err := json.Marshal(i.Data)
		if err != nil {
			helper.Logger("error", "Failed to marshal Data: "+err.Error())
			return nil, errors.New("invalid data")
		}
		dataJSON = string(b)
	}

	queryInsertInbox := `INSERT INTO inboxes 
    (title, content, field_1, field_2, field_3, field_4, field_5, data, user_id, receiver_id) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	switch i.Field4 {
	case "ktp":
		err := dbDefault.
			Exec(`UPDATE ktps SET nik = NULL WHERE user_id = ?`, i.ReceiverId).Error
		if err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return nil, err
		}

	case "ktp-pic":
		err := dbDefault.
			Exec(`UPDATE ktps SET path = NULL WHERE user_id = ?`, i.ReceiverId).Error
		if err != nil {
			helper.Logger("error", "In Server: "+err.Error())
			return nil, err
		}
	}

	errInsertInbox := dbDefault.Exec(queryInsertInbox,
		i.Title,
		i.Content,
		cutPrice,
		i.Field2,
		i.Field3,
		i.Field4,
		i.Field5,
		dataJSON,
		i.UserId,
		i.ReceiverId,
	).Error

	if errInsertInbox != nil {
		helper.Logger("error", "In Server: "+errInsertInbox.Error())
		return nil, errors.New(errInsertInbox.Error())
	}

	return map[string]any{
		"data": i,
	}, nil
}
