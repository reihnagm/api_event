package services

import (
	"errors"
	"fmt"
	"strings"

	"superapps/entities"
	helper "superapps/helpers"
)

func BroadcastList(userId string) ([]entities.BroadcastList, error) {
	// If no userId, return general/public broadcasts.
	if strings.TrimSpace(userId) == "" {
		const queryGeneral = `
			SELECT id, title, content, path, date, created_at
			FROM broadcasts
			WHERE cc IS NULL OR cc = 0
			ORDER BY date DESC
		`

		var list []entities.BroadcastList
		if err := dbDefault.Raw(queryGeneral).Scan(&list).Error; err != nil {
			helper.Logger("error", "In Server (list general broadcasts): "+err.Error())
			return nil, err
		}
		if len(list) == 0 {
			list = []entities.BroadcastList{}
		}
		return list, nil
	}

	// Otherwise, do role-based lookup.
	const queryUser = `SELECT role FROM users WHERE uid = ?`
	type roleRw struct {
		Role uint64 `gorm:"column:role"`
	}
	var rr roleRw
	if err := dbDefault.Raw(queryUser, userId).Scan(&rr).Error; err != nil {
		helper.Logger("error", "In Server (select user role): "+err.Error())
		return nil, err
	}
	if rr.Role == 0 {
		return nil, errors.New("user role not found")
	}

	const queryBroadcasts = `
		SELECT id, title, content, path, date, created_at
		FROM broadcasts
		WHERE cc = ?
		ORDER BY date DESC
	`
	var list []entities.BroadcastList
	if err := dbDefault.Raw(queryBroadcasts, rr.Role).Scan(&list).Error; err != nil {
		helper.Logger("error", "In Server (list broadcasts by role): "+err.Error())
		return nil, err
	}

	if len(list) == 0 {
		list = []entities.BroadcastList{}
	}
	return list, nil
}

func BroadcastDetail(Id string) (entities.BroadcastList, error) {
	if strings.TrimSpace(Id) == "" {
		return entities.BroadcastList{}, errors.New("empty id")
	}

	const queryBroadcasts = `
		SELECT id, title, content, path, date, created_at
		FROM broadcasts
		WHERE id = ?
		ORDER BY date DESC
	`
	var detail entities.BroadcastList
	if err := dbDefault.Raw(queryBroadcasts, Id).Scan(&detail).Error; err != nil {
		helper.Logger("error", "In Server (detail broadcast by role): "+err.Error())
		return entities.BroadcastList{}, err
	}

	return detail, nil
}
func BroadcastSend(bs *entities.BroadcastSend) (entities.BroadcastSend, error) {
	if bs == nil {
		return entities.BroadcastSend{}, errors.New("nil payload")
	}

	title := strings.TrimSpace(bs.Title)
	content := strings.TrimSpace(bs.Content)
	path := strings.TrimSpace(bs.Path)

	if title == "" || content == "" {
		return entities.BroadcastSend{}, errors.New("title/content cannot be empty")
	}

	seen := make(map[string]struct{}, len(bs.Cc))
	roleIDs := make([]uint64, 0, len(bs.Cc))
	ccNormalized := make([]string, 0, len(bs.Cc))

	// tokens that mean "general/public"
	generalTokens := map[string]struct{}{
		"general": {},
		"*":       {},
		"all":     {},
		"public":  {},
	}
	isGeneral := false

	const querySelRole = `SELECT id FROM roles WHERE LOWER(name) = ? LIMIT 1`

	for _, v := range bs.Cc {
		roleName := strings.ToLower(strings.TrimSpace(v))
		if roleName == "" {
			continue
		}
		if _, ok := seen[roleName]; ok {
			continue
		}

		// handle "general" targets
		if _, ok := generalTokens[roleName]; ok {
			isGeneral = true
			seen[roleName] = struct{}{}
			continue
		}

		// lookup role id
		var ids []uint64
		if err := dbDefault.Raw(querySelRole, roleName).Scan(&ids).Error; err != nil {
			helper.Logger("error", "In Server (select role): "+err.Error())
			return entities.BroadcastSend{}, err
		}
		if len(ids) == 0 {
			err := fmt.Errorf("role not found: %s", roleName)
			helper.Logger("warn", "In Server: "+err.Error())
			return entities.BroadcastSend{}, err
		}

		roleIDs = append(roleIDs, ids[0])
		ccNormalized = append(ccNormalized, roleName)
		seen[roleName] = struct{}{}
	}

	// If user didn't provide any CCs at all, treat as general/public broadcast.
	if len(bs.Cc) == 0 {
		isGeneral = true
	}

	// If after normalization we still have nothing to send to, error out.
	if len(roleIDs) == 0 && !isGeneral {
		return entities.BroadcastSend{}, errors.New("no valid roles or general audience provided")
	}

	tx := dbDefault.Begin()
	if err := tx.Error; err != nil {
		return entities.BroadcastSend{}, err
	}

	// Separate SQL so general writes a real NULL.
	const insertBroadcastGeneral = `
		INSERT INTO broadcasts (title, content, path, date, cc)
		VALUES (?, ?, ?, ?, NULL)
	`
	const insertBroadcastRole = `
		INSERT INTO broadcasts (title, content, path, date, cc)
		VALUES (?, ?, ?, ?, ?)
	`

	// Insert a "general" broadcast row (cc = NULL) if requested.
	if isGeneral {
		if err := tx.Exec(
			insertBroadcastGeneral,
			title,
			content,
			path,
			bs.Date,
		).Error; err != nil {
			tx.Rollback()
			helper.Logger("error", "In Server (insert general broadcast): "+err.Error())
			return entities.BroadcastSend{}, err
		}
		// reflect back that "general" was targeted
		ccNormalized = append(ccNormalized, "general")
	}

	// Insert a row per role id (same as before).
	for i := range roleIDs {
		if err := tx.Exec(
			insertBroadcastRole,
			title,
			content,
			path,
			bs.Date,
			roleIDs[i],
		).Error; err != nil {
			tx.Rollback()
			helper.Logger("error", "In Server (insert broadcast): "+err.Error())
			return entities.BroadcastSend{}, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		helper.Logger("error", "In Server (commit): "+err.Error())
		return entities.BroadcastSend{}, err
	}

	return entities.BroadcastSend{
		Title:   title,
		Content: content,
		Date:    bs.Date,
		Cc:      ccNormalized, // includes role names and/or "general"
	}, nil
}
