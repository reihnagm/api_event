package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	service "superapps/services"
)

func ResendOtp(w http.ResponseWriter, r *http.Request) {

	data := &entities.ResendOtp{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]any{})
		return
	}

	val := data.Val

	if val == "" {
		helper.Logger("error", "In Server: email field is required")
		helper.Response(w, 400, true, "val field is required", map[string]any{})
		return
	}

	result, err := service.ResendOtp(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Resend otp success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
