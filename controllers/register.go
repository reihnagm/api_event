package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"
)

func Register(w http.ResponseWriter, r *http.Request) {

	data := &entities.Register{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	Email := data.Email
	Fullname := data.Fullname
	Phone := data.Phone
	Password := data.Password

	if Email == "" {
		helper.Logger("error", "In Server: email is required")
		helper.Response(w, 400, true, "email is required", map[string]any{})
		return
	}

	if Fullname == "" {
		helper.Logger("error", "In Server: fullname is required")
		helper.Response(w, 400, true, "fullname is required", map[string]any{})
		return
	}

	if Phone == "" {
		helper.Logger("error", "In Server: phone is required")
		helper.Response(w, 400, true, "phone is required", map[string]any{})
		return
	}

	if Password == "" {
		helper.Logger("error", "In Server: password is required")
		helper.Response(w, 400, true, "password is required", map[string]any{})
		return
	}

	result, err := services.Register(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Register success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
