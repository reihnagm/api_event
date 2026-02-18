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

	// join user + profile
	UserEmail   *string `json:"-"`
	UserPhone   *string `json:"-"`
	ProfileName *string `json:"-"`
	ImageConcat *string `json:"-"`
}

type eventImageRow struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
}

type updateEventPayload struct {
	Title     *string   `json:"title"`
	Content   *string   `json:"content"`
	StartDate *string   `json:"start_date"` // "2006-01-02 15:04:05" / RFC3339
	EndDate   *string   `json:"end_date"`
	Images    *[]string `json:"images"` // replace all images if provided
}

type createEventPayload struct {
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	StartDate *string   `json:"start_date"`
	EndDate   *string   `json:"end_date"`
	Images    *[]string `json:"images"`
}

type addImagePayload struct {
	Path string `json:"path"`
}

type replaceImagePayload struct {
	Path string `json:"path"`
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

func getImageIDParam(r *http.Request) (int64, bool) {
	q := r.URL.Query()
	raw := strings.TrimSpace(q.Get("image_id"))
	if raw == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
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

func ensureEventOwner(eventUID, authUID string) (eventRow, error) {
	var e eventRow
	if err := dbDefault.Raw(`SELECT uid, user_id FROM events WHERE uid = ? LIMIT 1`, eventUID).Scan(&e).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return eventRow{}, errors.New("INTERNAL_SERVER_ERROR")
	}
	if strings.TrimSpace(e.UID) == "" {
		return eventRow{}, errors.New("EVENT_NOT_FOUND")
	}
	if strings.TrimSpace(authUID) == "" {
		return eventRow{}, errors.New("UNAUTHORIZED")
	}
	if e.UserID != authUID {
		return eventRow{}, errors.New("FORBIDDEN")
	}
	return e, nil
}

// =====================================================
// CREATE EVENT (auth via JWT Bearer token)
// Body: { title, content?, start_date?, end_date?, images?[] }
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

	if startAt != nil && endAt != nil && endAt.Before(*startAt) {
		return map[string]any{"message": "BAD_REQUEST", "reason": "END_BEFORE_START"}, errors.New("BAD_REQUEST")
	}

	eventUID := uuid.NewString()

	tx := dbDefault.Begin()
	if tx.Error != nil {
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	if err := tx.Exec(`
		INSERT INTO events (uid, title, content, user_id, start_date, end_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, eventUID, title, content, authUID, startAt, endAt).Error; err != nil {
		tx.Rollback()
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

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
// GET LIST EVENT
// Query params:
// - page (default 1)
// - limit (default 10, max 200)
// - q (search by title)
// - mine=1 (only my events) -> auth via JWT
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

	var imgs []eventImageRow
	queryImgs := `SELECT id, path FROM event_images WHERE event_id = ? ORDER BY id ASC`
	if err := dbDefault.Raw(queryImgs, eventUID).Scan(&imgs).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	images := make([]map[string]any, 0, len(imgs))
	for _, it := range imgs {
		p := strings.TrimSpace(it.Path)
		if p == "" {
			continue
		}
		images = append(images, map[string]any{
			"id":   it.ID,
			"path": p,
		})
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
		"images":     images, // <-- sekarang detail balikin id + path biar FE bisa update 1 image by id
	}, nil
}

// =====================================================
// DELETE EVENT (ownership)
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

	if err := tx.Exec(`DELETE FROM event_images WHERE event_id = ?`, eventUID).Error; err != nil {
		tx.Rollback()
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

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
// UPDATE EVENT (ownership)
// - kalau images dikirim -> replace all (strategy lama tetap ada)
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

	// ownership check
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

	// replace all images (existing strategy)
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

// =====================================================
// IMAGE OPS (update 1 image / add 1 / delete 1)
// Auth: JWT
// Ownership: event.user_id must == authUID
// =====================================================

// POST /api/v1/event/image/add?uid=EVENT_UID
// Body: { "path": "/uploads/x.jpg" }
func AddEventImage(r *http.Request) (map[string]any, error) {
	eventUID := getEventUIDParam(r)
	if eventUID == "" {
		return map[string]any{"message": "BAD_REQUEST"}, errors.New("BAD_REQUEST")
	}

	authUID := getAuthUIDFromRequest(r)
	if authUID == "" {
		return map[string]any{"message": "UNAUTHORIZED"}, errors.New("UNAUTHORIZED")
	}

	_, err := ensureEventOwner(eventUID, authUID)
	if err != nil {
		return map[string]any{"message": err.Error()}, err
	}

	var payload addImagePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return map[string]any{"message": "INVALID_JSON"}, errors.New("INVALID_JSON")
	}

	path := strings.TrimSpace(payload.Path)
	if path == "" {
		return map[string]any{"message": "BAD_REQUEST", "reason": "PATH_REQUIRED"}, errors.New("BAD_REQUEST")
	}

	if err := dbDefault.Exec(`INSERT INTO event_images (event_id, path) VALUES (?, ?)`, eventUID, path).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	return map[string]any{
		"message": "ADD_EVENT_IMAGE_SUCCESS",
	}, nil
}

// PUT /api/v1/event/image/replace?uid=EVENT_UID&image_id=123
// Body: { "path": "/uploads/new.jpg" }
func ReplaceEventImage(r *http.Request) (map[string]any, error) {
	eventUID := getEventUIDParam(r)
	if eventUID == "" {
		return map[string]any{"message": "BAD_REQUEST"}, errors.New("BAD_REQUEST")
	}

	imageID, ok := getImageIDParam(r)
	if !ok {
		return map[string]any{"message": "BAD_REQUEST", "reason": "IMAGE_ID_REQUIRED"}, errors.New("BAD_REQUEST")
	}

	authUID := getAuthUIDFromRequest(r)
	if authUID == "" {
		return map[string]any{"message": "UNAUTHORIZED"}, errors.New("UNAUTHORIZED")
	}

	_, err := ensureEventOwner(eventUID, authUID)
	if err != nil {
		return map[string]any{"message": err.Error()}, err
	}

	// ensure image belongs to event
	var img struct {
		ID      int64  `json:"id"`
		EventID string `json:"event_id"`
		Path    string `json:"path"`
	}
	if err := dbDefault.Raw(`SELECT id, event_id, path FROM event_images WHERE id = ? LIMIT 1`, imageID).
		Scan(&img).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}
	if img.ID == 0 {
		return map[string]any{"message": "IMAGE_NOT_FOUND"}, errors.New("IMAGE_NOT_FOUND")
	}
	if img.EventID != eventUID {
		return map[string]any{"message": "FORBIDDEN"}, errors.New("FORBIDDEN")
	}

	var payload replaceImagePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return map[string]any{"message": "INVALID_JSON"}, errors.New("INVALID_JSON")
	}

	newPath := strings.TrimSpace(payload.Path)
	if newPath == "" {
		return map[string]any{"message": "BAD_REQUEST", "reason": "PATH_REQUIRED"}, errors.New("BAD_REQUEST")
	}

	if err := dbDefault.Exec(`UPDATE event_images SET path = ? WHERE id = ?`, newPath, imageID).Error; err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}

	return map[string]any{
		"message":   "REPLACE_EVENT_IMAGE_SUCCESS",
		"image_id":  imageID,
		"old_path":  img.Path,
		"new_path":  newPath,
		"event_uid": eventUID,
	}, nil
}

// DELETE /api/v1/event/image/delete?uid=EVENT_UID&image_id=123
func DeleteEventImage(r *http.Request) (map[string]any, error) {
	eventUID := getEventUIDParam(r)
	if eventUID == "" {
		return map[string]any{"message": "BAD_REQUEST"}, errors.New("BAD_REQUEST")
	}

	imageID, ok := getImageIDParam(r)
	if !ok {
		return map[string]any{"message": "BAD_REQUEST", "reason": "IMAGE_ID_REQUIRED"}, errors.New("BAD_REQUEST")
	}

	authUID := getAuthUIDFromRequest(r)
	if authUID == "" {
		return map[string]any{"message": "UNAUTHORIZED"}, errors.New("UNAUTHORIZED")
	}

	_, err := ensureEventOwner(eventUID, authUID)
	if err != nil {
		return map[string]any{"message": err.Error()}, err
	}

	res := dbDefault.Exec(`DELETE FROM event_images WHERE id = ? AND event_id = ?`, imageID, eventUID)
	if res.Error != nil {
		helper.Logger("error", "In Server: "+res.Error.Error())
		return map[string]any{}, errors.New("INTERNAL_SERVER_ERROR")
	}
	if res.RowsAffected == 0 {
		return map[string]any{"message": "IMAGE_NOT_FOUND"}, errors.New("IMAGE_NOT_FOUND")
	}

	return map[string]any{
		"message":   "DELETE_EVENT_IMAGE_SUCCESS",
		"image_id":  imageID,
		"event_uid": eventUID,
	}, nil
}
