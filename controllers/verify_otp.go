package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	service "superapps/services"
)

func VerifyOtp(w http.ResponseWriter, r *http.Request) {

	data := &entities.VerifyOtp{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]any{})
		return
	}

	val := data.Val
	otp := data.Otp

	if val == "" {
		helper.Logger("error", "In Server: val field is required")
		helper.Response(w, 400, true, "val field is required", map[string]any{})
		return
	}

	if otp == "" {
		helper.Logger("error", "In Server: otp field is required")
		helper.Response(w, 400, true, "otp field is required", map[string]any{})
		return
	}

	result, err := service.VerifyOtp(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Verify account success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
