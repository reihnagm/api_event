package controllers

import (
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	service "superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func GetProfile(w http.ResponseWriter, r *http.Request) {

	getProfile := &entities.GetProfile{}

	tokenHeader := r.Header.Get("Authorization")

	token := helper.DecodeJwt(tokenHeader)

	claims, _ := token.Claims.(jwt.MapClaims)

	userId, _ := claims["id"].(string)

	getProfile.UserId = userId

	result, err := service.GetProfile(getProfile)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Get Profile success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
