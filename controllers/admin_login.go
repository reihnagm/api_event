package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"
)

func AdminLogin(w http.ResponseWriter, r *http.Request) {

	data := &entities.AdminLogin{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	Val := data.Val
	Password := data.Password

	if Val == "" {
		helper.Logger("error", "In Server: val is required")
		helper.Response(w, 400, true, "val is required", map[string]any{})
		return
	}

	if Password == "" {
		helper.Logger("error", "In Server: password is required")
		helper.Response(w, 400, true, "password is required", map[string]any{})
		return
	}

	result, err := services.AdminLogin(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Admin Login success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
