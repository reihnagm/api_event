package services

import (
	"errors"
	"fmt"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"time"
)

func Callback(r *http.Request, c *entities.Callback) (map[string]any, error) {
	var callback entities.OrderProjectCallback

	if c.Status == "PAID" {

		// Update Order
		queryUpdateOrder := `UPDATE orders SET status = 4 WHERE invoice = ?`
		errUpdateOrder := dbDefault.Exec(queryUpdateOrder, c.OrderId).Error

		if errUpdateOrder != nil {
			helper.Logger("error", "In Server: "+errUpdateOrder.Error())
			return nil, errors.New(errUpdateOrder.Error())
		}

		// Get Project Id By Order
		queryOrder := `SELECT project_id FROM orders WHERE invoice = ?`

		errOrder := dbDefault.Raw(queryOrder, c.OrderId).Scan(&callback).Error
		if errOrder != nil {
			helper.Logger("error", "In Server: "+errOrder.Error())
			return nil, errors.New(errOrder.Error())
		}

		// Update Project
		queryUpdateProject := `UPDATE projects SET status = 4 WHERE uid = ?`
		errUpdateProject := dbDefault.Exec(queryUpdateProject, callback.ProjectId).Error

		if errUpdateProject != nil {
			helper.Logger("error", "In Server: "+errUpdateProject.Error())
			return nil, errors.New(errUpdateProject.Error())
		}

		// Update Inbox
		queryUpdateInbox := `UPDATE inboxes SET status = 4 WHERE field_8 = ?`
		errUpdateInbox := dbDefault.Exec(queryUpdateInbox, callback.ProjectId).Error

		if errUpdateInbox != nil {
			helper.Logger("error", "In Server: "+errUpdateInbox.Error())
			return nil, errors.New(errUpdateInbox.Error())
		}

		title, _ := helper.GetTitleByProjectId(dbDefault, callback.ProjectId)
		email, _ := helper.GetEmailByProjectId(dbDefault, callback.ProjectId)
		price, _ := helper.GetPriceByProjectId(dbDefault, callback.ProjectId)
		role, _ := helper.GetRoleByProjectId(dbDefault, callback.ProjectId)

		ip := helper.GetClientIP(r)

		helper.SendLogs(
			dbDefault,
			fmt.Sprintf(
				"%s [%s] - %s melakukan pembayaran project [%s] sebesar %s dengan status %s pada waktu %s",
				ip,
				email,
				role,
				title,
				helper.FormatIDR(price),
				c.Status,
				time.Now().Format("2006-01-02 15:04:05"),
			),
			role,
		)

	}

	return map[string]any{}, nil
}
