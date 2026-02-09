package controllers

import (
	"net/http"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func AdminGetProfile(w http.ResponseWriter, r *http.Request) {

	tokenHeader := r.Header.Get("Authorization")

	token := helper.DecodeJwt(tokenHeader)

	claims, _ := token.Claims.(jwt.MapClaims)

	userId, _ := claims["id"].(string)

	result, err := services.AdminGetProfile(userId)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Admin List Role success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
