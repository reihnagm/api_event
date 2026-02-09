package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	service "superapps/services"
)

func UpdateAdminVerifyUser(w http.ResponseWriter, r *http.Request) {

	data := &entities.AdminVerifyUser{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	UserId := data.UserId

	if UserId == "" {
		helper.Logger("error", "In Server: user_id is required")
		helper.Response(w, 400, true, "user_id is required", map[string]any{})
		return
	}

	result, err := service.VerifyUser(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Update Admin Verify User success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}

func UpdateAdminVerifyUserEmiten(w http.ResponseWriter, r *http.Request) {

	data := &entities.AdminVerifyUser{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	UserId := data.UserId

	if UserId == "" {
		helper.Logger("error", "In Server: user_id is required")
		helper.Response(w, 400, true, "user_id is required", map[string]any{})
		return
	}

	result, err := service.VerifyUserEmiten(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Update Admin Verify User Emiten success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}

func UpdateAdminVerifyUserInvestor(w http.ResponseWriter, r *http.Request) {

	data := &entities.AdminVerifyUser{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	UserId := data.UserId

	if UserId == "" {
		helper.Logger("error", "In Server: user_id is required")
		helper.Response(w, 400, true, "user_id is required", map[string]any{})
		return
	}

	result, err := service.VerifyUserInvestor(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Update Admin Verify User Investor success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}

func UpdateAdminVerifyProject(w http.ResponseWriter, r *http.Request) {

	data := &entities.AdminVerifyProject{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	Id := data.Id

	if Id == "" {
		helper.Logger("error", "In Server: Id is required")
		helper.Response(w, 400, true, "Id is required", map[string]any{})
		return
	}

	result, err := service.VerifyProject(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Update Admin Verify Project success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
