package controllers

import (
	"net/http"
	helper "superapps/helpers"
	service "superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func ProjectEmitenList(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	tokenHeader := r.Header.Get("Authorization")

	token := helper.DecodeJwt(tokenHeader)

	claims, _ := token.Claims.(jwt.MapClaims)

	userId, _ := claims["id"].(string)

	result, err := service.ProjectEmitenList(userId, status)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Project Emiten List success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
