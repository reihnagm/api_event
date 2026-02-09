package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	service "superapps/services"
)

func ProjectInquiry(w http.ResponseWriter, r *http.Request) {

	data := &entities.ProjectInquiry{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	result, err := service.ProjectInquiry(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Project Inquiry success")
	helper.Response(w, http.StatusOK, false, "Successfully", map[string]any{
		"total_amount": result["total_amount"],
		"info":         result["data"],
	})
}
