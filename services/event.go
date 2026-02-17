package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	helper "superapps/helpers"

	"github.com/google/uuid"
)

type eventRow struct {
	ID        int64      `json:"id"`
	UID       string     `json:"uid"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	UserID    string     `json:"user_id"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`

	UserEmail   *string `json:"-"`
	UserPhone   *string `json:"-"`
	ProfileName *string `json:"-"`
	ImageConcat *string `json:"-"`
}

type eventImageRow struct {
	Path string `json:"path"`
}

type createEventPayload struct {
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	StartDate *string   `json:"start_date"`
	EndDate   *string   `json:"end_date"`
	Images    *[]string `json:"images"`
}

type updateEventPayload struct {
	Title     *string   `json:"title"`
	Content   *string   `json:"content"`
	StartDate *string   `json:"start_date"`
	EndDate   *string   `json:"end_date"`
	Images    *[]string `json:"images"`
}

func getAuthUIDFromRequest(r *http.Request) string {
	// sesuaikan dengan sistem auth kamu
	if v := strings.TrimSpace(r.Header.Get("X-User-UID")); v != "" {
		return v
	}
	return ""
}

func getEventUIDParam(r *http.Request) string {
	q := r.URL.Query()
	if v := strings.TrimSpace(q.Get("uid")); v != "" {
		return v
	}
	if v := strings.TrimSpace(q.Get("event_id")); v != "" {
		return v
	}
	return ""
}

func parseDateFlexible(s string) (*time.Time, error) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return nil, nil
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", raw, time.Local); err == nil {
		return &t, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t, nil
	}
	return nil, errors.New("INVALID_DATE_FORMAT")
}

func buildAuthorMap(e eventRow) map[string]any {
	fullname := ""
	if e.ProfileName != nil {
		fullname = strings.TrimSpace(*e.ProfileName)
	}
	email := ""
	if e.UserEmail != nil {
		email = strings.TrimSpace(*e.UserEmail)
	}
	phone := ""
	if e.UserPhone != nil {
		phone = strings.TrimSpace(*e.UserPhone)
	}

	return map[string]any{
		"uid":      e.UserID,
		"email":    email,
		"phone":    phone,
		"fullname": fullname,
	}
}

func splitImages(concat *string) []string {
	images := make([]string, 0)
	if concat == nil {
		return images
	}
	raw := strings.TrimSpace(*concat)
	if raw == "" {
		return images
	}
	parts := strings.Split(raw, "||")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			images = append(images, p)
		}
	}
	return images
}

// =====================================================
// GET LIST EVENT
// Query params optional:
// - page (default 1)
// - limit (default 10, max 200)
// - q (search by title)
// - mine=1 (only my events)
// =====================================================
func GetListEvent(r *http.Request) (map[string]any, error) {
	qp := r.URL.Query()

	page, _ := strconv.Atoi(qp.Get("page"))
	limit, _ := strconv.Atoi(qp.Get("limit"))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 10
	}
	offset := (page - 1) * limit

	search := strings.TrimSpace(qp.Get("q"))
	mine := strings.TrimSpace(qp.Get("mine")) == "1"

	var conds []string
	var args []any
	conds = append(conds, "1=1")

	if search != "" {
		conds = append(conds, "LOWER(e.title) LIKE LOWER(?)")
		args = append(args, "%"+search+"%")
	}

	authUID := getAuthUIDFromRequest(r)
	if mine {
		if authUID == "" {
			return map[string]any{"message": "UNAUTHORIZED"}, errors.New("UNAUTHORIZED")
		}
		conds = append(conds, "e.user_id = ?")
		args = append(args, authUID)
	}

	whereSQL := "WHERE " + strings.Join(conds, " AND ")

	// COUNT
	var totalRows []struct {
		Total int64 `json:"total"`
	}
	countSQL := fmt.Sprintf(`SELECT COUNT(*) AS total FROM events e %s`, whereSQL)

	if err := dbDefault.Raw(countSQL, args...).Scan(&totalRows).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	total := int64(0)
	if len(totalRows) > 0 {
		total = totalRows[0].Total
	}

	// DATA + JOIN users + profiles + GROUP_CONCAT images
	dataSQL := fmt.Sprintf(`
		SELECT
			e.id, e.uid, e.title, e.content, e.user_id,
			e.start_date, e.end_date, e.created_at, e.updated_at,

			u.email AS user_email,
			u.phone AS user_phone,
			pr.fullname AS profile_name,

			GROUP_CONCAT(ei.path SEPARATOR '||') AS image_concat
		FROM events e
		INNER JOIN users u ON u.uid = e.user_id
		LEFT JOIN profiles pr ON pr.user_id = u.uid
		LEFT JOIN event_images ei ON ei.event_id = e.uid
		%s
		GROUP BY
			e.id, e.uid, e.title, e.content, e.user_id,
			e.start_date, e.end_date, e.created_at, e.updated_at,
			u.email, u.phone, pr.fullname
		ORDER BY e.start_date DESC, e.id DESC
		LIMIT ? OFFSET ?
	`, whereSQL)

	var rows []eventRow
	dataArgs := append(append([]any{}, args...), limit, offset)

	if err := dbDefault.Raw(dataSQL, dataArgs...).Scan(&rows).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	list := make([]map[string]any, 0, len(rows))
	for _, e := range rows {
		list = append(list, map[string]any{
			"id":         e.ID,
			"uid":        e.UID,
			"title":      e.Title,
			"content":    e.Content,
			"start_date": e.StartDate,
			"end_date":   e.EndDate,
			"created_at": e.CreatedAt,
			"updated_at": e.UpdatedAt,
			"author":     buildAuthorMap(e),
			"images":     splitImages(e.ImageConcat),
		})
	}

	return map[string]any{
		"page":  page,
		"limit": limit,
		"total": total,
		"data":  list,
	}, nil
}

// =====================================================
// GET DETAIL EVENT
// Query: ?uid=EVENT_UID
// =====================================================
func GetDetailEvent(r *http.Request) (map[string]any, error) {
	eventUID := getEventUIDParam(r)
	if eventUID == "" {
		return map[string]any{"message": "BAD_REQUEST"}, errors.New("BAD_REQUEST")
	}

	// event + author join
	var e eventRow
	queryEvent := `
		SELECT
			e.id, e.uid, e.title, e.content, e.user_id,
			e.start_date, e.end_date, e.created_at, e.updated_at,
			u.email AS user_email,
			u.phone AS user_phone,
			pr.fullname AS profile_name
		FROM events e
		INNER JOIN users u ON u.uid = e.user_id
		LEFT JOIN profiles pr ON pr.user_id = u.uid
		WHERE e.uid = ?
		LIMIT 1
	`
	if err := dbDefault.Raw(queryEvent, eventUID).Scan(&e).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}
	if strings.TrimSpace(e.UID) == "" {
		return map[string]any{"message": "EVENT_NOT_FOUND"}, errors.New("EVENT_NOT_FOUND")
	}

	// images
	var imgs []eventImageRow
	queryImgs := `SELECT path FROM event_images WHERE event_id = ? ORDER BY id ASC`
	if err := dbDefault.Raw(queryImgs, eventUID).Scan(&imgs).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	images := make([]string, 0, len(imgs))
	for _, it := range imgs {
		p := strings.TrimSpace(it.Path)
		if p != "" {
			images = append(images, p)
		}
	}

	return map[string]any{
		"id":         e.ID,
		"uid":        e.UID,
		"title":      e.Title,
		"content":    e.Content,
		"start_date": e.StartDate,
		"end_date":   e.EndDate,
		"created_at": e.CreatedAt,
		"updated_at": e.UpdatedAt,
		"author":     buildAuthorMap(e),
		"images":     images,
	}, nil
}

// =====================================================
// DELETE EVENT (ownership: events.user_id must == authUID)
// Query: ?uid=EVENT_UID
// =====================================================
func DeleteEvent(r *http.Request) (map[string]any, error) {
	eventUID := getEventUIDParam(r)
	if eventUID == "" {
		return map[string]any{"message": "BAD_REQUEST"}, errors.New("BAD_REQUEST")
	}

	authUID := getAuthUIDFromRequest(r)
	if authUID == "" {
		return map[string]any{"message": "UNAUTHORIZED"}, errors.New("UNAUTHORIZED")
	}

	// cek event + ownership
	var e eventRow
	qEvent := `SELECT uid, user_id, title FROM events WHERE uid = ? LIMIT 1`
	if err := dbDefault.Raw(qEvent, eventUID).Scan(&e).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}
	if strings.TrimSpace(e.UID) == "" {
		return map[string]any{"message": "EVENT_NOT_FOUND"}, errors.New("EVENT_NOT_FOUND")
	}
	if e.UserID != authUID {
		return map[string]any{"message": "FORBIDDEN"}, errors.New("FORBIDDEN")
	}

	tx := dbDefault.Begin()
	if tx.Error != nil {
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	// delete images
	if err := tx.Exec(`DELETE FROM event_images WHERE event_id = ?`, eventUID).Error; err != nil {
		tx.Rollback()
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	// delete event
	if err := tx.Exec(`DELETE FROM events WHERE uid = ?`, eventUID).Error; err != nil {
		tx.Rollback()
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	if err := tx.Commit().Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	return map[string]any{
		"message": "DELETE_EVENT_SUCCESS",
		"uid":     eventUID,
	}, nil
}

// =====================================================
// CREATE EVENT (auth required)
// Body: { title, content, start_date?, end_date?, images?[] }
// Ownership: events.user_id = authUID
// =====================================================
func CreateEvent(r *http.Request) (map[string]any, error) {
	authUID := getAuthUIDFromRequest(r)
	if authUID == "" {
		return map[string]any{"message": "UNAUTHORIZED"}, errors.New("UNAUTHORIZED")
	}

	var payload createEventPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return map[string]any{"message": "INVALID_JSON"}, errors.New("INVALID_JSON")
	}

	title := strings.TrimSpace(payload.Title)
	content := strings.TrimSpace(payload.Content)

	if title == "" {
		return map[string]any{"message": "BAD_REQUEST", "reason": "TITLE_REQUIRED"}, errors.New("BAD_REQUEST")
	}

	var startAt *time.Time
	var endAt *time.Time
	var err error

	if payload.StartDate != nil {
		startAt, err = parseDateFlexible(*payload.StartDate)
		if err != nil {
			return map[string]any{"message": err.Error()}, err
		}
	}
	if payload.EndDate != nil {
		endAt, err = parseDateFlexible(*payload.EndDate)
		if err != nil {
			return map[string]any{"message": err.Error()}, err
		}
	}

	// optional: validasi range
	if startAt != nil && endAt != nil && endAt.Before(*startAt) {
		return map[string]any{"message": "BAD_REQUEST", "reason": "END_BEFORE_START"}, errors.New("BAD_REQUEST")
	}

	eventUID := uuid.NewString()

	tx := dbDefault.Begin()
	if tx.Error != nil {
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	// insert event
	if err := tx.Exec(`
		INSERT INTO events (uid, title, content, user_id, start_date, end_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, eventUID, title, content, authUID, startAt, endAt).Error; err != nil {
		tx.Rollback()
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	// insert images (optional)
	images := []string{}
	if payload.Images != nil {
		images = *payload.Images
	}

	if len(images) > 0 {
		vals := make([]string, 0, len(images))
		iargs := make([]any, 0, len(images)*2)

		for _, p := range images {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			vals = append(vals, "(?, ?)")
			iargs = append(iargs, eventUID, p)
		}

		if len(vals) > 0 {
			insSQL := `INSERT INTO event_images (event_id, path) VALUES ` + strings.Join(vals, ", ")
			if err := tx.Exec(insSQL, iargs...).Error; err != nil {
				tx.Rollback()
				helper.Logger("error", "In Server: "+err.Error())
				return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	return map[string]any{
		"message": "CREATE_EVENT_SUCCESS",
		"uid":     eventUID,
	}, nil
}

// =====================================================
// UPDATE EVENT (ownership: events.user_id must == authUID)
// Query: ?uid=EVENT_UID
// Body: { title?, content?, start_date?, end_date?, images?[] }
// =====================================================
func UpdateEvent(r *http.Request) (map[string]any, error) {
	eventUID := getEventUIDParam(r)
	if eventUID == "" {
		return map[string]any{"message": "BAD_REQUEST"}, errors.New("BAD_REQUEST")
	}

	authUID := getAuthUIDFromRequest(r)
	if authUID == "" {
		return map[string]any{"message": "UNAUTHORIZED"}, errors.New("UNAUTHORIZED")
	}

	var payload updateEventPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return map[string]any{"message": "INVALID_JSON"}, errors.New("INVALID_JSON")
	}

	// cek event + ownership
	var e eventRow
	qEvent := `SELECT uid, user_id FROM events WHERE uid = ? LIMIT 1`
	if err := dbDefault.Raw(qEvent, eventUID).Scan(&e).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}
	if strings.TrimSpace(e.UID) == "" {
		return map[string]any{"message": "EVENT_NOT_FOUND"}, errors.New("EVENT_NOT_FOUND")
	}
	if e.UserID != authUID {
		return map[string]any{"message": "FORBIDDEN"}, errors.New("FORBIDDEN")
	}

	setCols := make([]string, 0, 8)
	args := make([]any, 0, 8)

	if payload.Title != nil {
		setCols = append(setCols, "title = ?")
		args = append(args, strings.TrimSpace(*payload.Title))
	}
	if payload.Content != nil {
		setCols = append(setCols, "content = ?")
		args = append(args, *payload.Content)
	}
	if payload.StartDate != nil {
		t, err := parseDateFlexible(*payload.StartDate)
		if err != nil {
			return map[string]any{"message": err.Error()}, err
		}
		setCols = append(setCols, "start_date = ?")
		args = append(args, t)
	}
	if payload.EndDate != nil {
		t, err := parseDateFlexible(*payload.EndDate)
		if err != nil {
			return map[string]any{"message": err.Error()}, err
		}
		setCols = append(setCols, "end_date = ?")
		args = append(args, t)
	}

	hasEventUpdate := len(setCols) > 0
	hasImagesUpdate := payload.Images != nil

	if !hasEventUpdate && !hasImagesUpdate {
		return map[string]any{"message": "NOTHING_TO_UPDATE"}, errors.New("NOTHING_TO_UPDATE")
	}

	tx := dbDefault.Begin()
	if tx.Error != nil {
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	// update events
	if hasEventUpdate {
		setCols = append(setCols, "updated_at = NOW()")
		updateSQL := fmt.Sprintf(`UPDATE events SET %s WHERE uid = ?`, strings.Join(setCols, ", "))
		args = append(args, eventUID)

		if err := tx.Exec(updateSQL, args...).Error; err != nil {
			tx.Rollback()
			helper.Logger("error", "In Server: "+err.Error())
			return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
		}
	}

	// replace images
	if hasImagesUpdate {
		if err := tx.Exec(`DELETE FROM event_images WHERE event_id = ?`, eventUID).Error; err != nil {
			tx.Rollback()
			helper.Logger("error", "In Server: "+err.Error())
			return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
		}

		paths := *payload.Images
		if len(paths) > 0 {
			vals := make([]string, 0, len(paths))
			iargs := make([]any, 0, len(paths)*2)

			for _, p := range paths {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				vals = append(vals, "(?, ?)")
				iargs = append(iargs, eventUID, p)
			}

			if len(vals) > 0 {
				insSQL := `INSERT INTO event_images (event_id, path) VALUES ` + strings.Join(vals, ", ")
				if err := tx.Exec(insSQL, iargs...).Error; err != nil {
					tx.Rollback()
					helper.Logger("error", "In Server: "+err.Error())
					return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	return map[string]any{
		"message": "UPDATE_EVENT_SUCCESS",
		"uid":     eventUID,
	}, nil
}
