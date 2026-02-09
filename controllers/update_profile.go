package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	service "superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func UpdateProfile(w http.ResponseWriter, r *http.Request) {

	data := &entities.UpdateProfile{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	tokenHeader := r.Header.Get("Authorization")

	token := helper.DecodeJwt(tokenHeader)

	claims, _ := token.Claims.(jwt.MapClaims)

	userId, _ := claims["id"].(string)

	data.UserId = userId

	Fullname := data.Fullname

	if Fullname == "" {
		helper.Logger("error", "In Server: fullname is required")
		helper.Response(w, 400, true, "fullname is required", map[string]any{})
		return
	}

	result, err := service.UpdateProfile(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Update Profile success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
