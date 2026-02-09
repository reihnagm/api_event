package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"
)

func ForgotPassword(w http.ResponseWriter, r *http.Request) {

	data := &entities.ForgotPassword{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]any{})
		return
	}

	Email := data.Email
	NewPassword := data.NewPassword

	if Email == "" {
		helper.Logger("error", "In Server: email is required")
		helper.Response(w, 400, true, "email is required", map[string]any{})
		return
	}

	if NewPassword == "" {
		helper.Logger("error", "In Server: new_password is required")
		helper.Response(w, 400, true, "new_password is required", map[string]any{})
		return
	}

	result, err := services.ForgotPassword(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Forgot Password success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
