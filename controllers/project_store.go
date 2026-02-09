package controllers

import (
	"encoding/json"
	"net/http"
	"strings"
	"superapps/entities"
	helper "superapps/helpers"
	service "superapps/services"
)

func ProjectStore(w http.ResponseWriter, r *http.Request) {

	data := &entities.ProjectStore{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	CompanyId := strings.TrimSpace(data.CompanyId)
	Title := strings.TrimSpace(data.Title)
	Deskripsi := strings.TrimSpace(data.Deskripsi)

	if CompanyId == "" {
		helper.Logger("error", "In Server: company_id is required")
		helper.Response(w, 400, true, "company_id is required", map[string]any{})
		return
	}

	if Title == "" {
		helper.Logger("error", "In Server: title is required")
		helper.Response(w, 400, true, "title is required", map[string]any{})
		return
	}

	if Deskripsi == "" {
		helper.Logger("error", "In Server: deskripsi is required")
		helper.Response(w, 400, true, "deskripsi is required", map[string]any{})
		return
	}

	result, err := service.ProjectStore(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Project Store success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
