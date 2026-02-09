package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func UpdateValUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	Type := vars["type"]

	data := &entities.UpdateValUser{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]any{})
		return
	}

	tokenHeader := r.Header.Get("Authorization")

	token := helper.DecodeJwt(tokenHeader)

	claims, _ := token.Claims.(jwt.MapClaims)

	userId, _ := claims["id"].(string)

	data.Type = Type
	data.UserId = userId

	Val := data.Val

	if Val == "" {
		helper.Logger("error", "In Server: val is required")
		helper.Response(w, 400, true, "val is required", map[string]any{})
		return
	}

	result, err := services.UpdateValUser(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Document update")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
