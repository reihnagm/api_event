package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func InitializeFcm(w http.ResponseWriter, r *http.Request) {

	tokenHeader := r.Header.Get("Authorization")

	token := helper.DecodeJwt(tokenHeader)

	claims, _ := token.Claims.(jwt.MapClaims)

	userId, _ := claims["id"].(string)

	data := &entities.Fcm{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]any{})
		return
	}

	UserId := userId
	Token := data.Token

	if UserId == "" {
		helper.Logger("error", "In Server: user_id is required")
		helper.Response(w, 400, true, "user_id is required", map[string]any{})
		return
	}

	if Token == "" {
		helper.Logger("error", "In Server: new_password is required")
		helper.Response(w, 400, true, "new_password is required", map[string]any{})
		return
	}

	data.UserId = userId

	result, err := services.InitializeFcm(data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Initialize FCM success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
