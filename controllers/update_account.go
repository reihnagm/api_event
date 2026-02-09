package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	service "superapps/services"
)

func UpdateAccount(w http.ResponseWriter, r *http.Request) {

	data := &entities.UpdateAccount{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	No := data.No

	if No == "" {
		helper.Logger("error", "In Server: no is required")
		helper.Response(w, 400, true, "no is required", map[string]any{})
		return
	}

	result, err := service.UpdateAccount(data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Update Profile success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
