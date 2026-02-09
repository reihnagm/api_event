package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func AdminRevokeRole(w http.ResponseWriter, r *http.Request) {

	data := &entities.AssignRole{}

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

	data.UserId = userId

	result, err := services.AssignRole(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Login success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
