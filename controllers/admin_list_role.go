package controllers

import (
	"net/http"
	helper "superapps/helpers"
	"superapps/services"
)

func AdminListRole(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")

	result, err := services.AdminListRole(page, limit)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Admin List Role success")
	helper.ResponseWithPagination(w, http.StatusOK, false, "Successfully",
		result["total"],
		result["per_page"],
		result["prev_page"],
		result["next_page"],
		result["current_page"],
		result["next_url"],
		result["prev_url"],
		result["data"],
	)
}
