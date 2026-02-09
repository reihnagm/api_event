package controllers

import (
	"net/http"
	"strconv"

	helper "superapps/helpers"
	"superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func DashboardInvestor(w http.ResponseWriter, r *http.Request) {

	tokenHeader := r.Header.Get("Authorization")

	token := helper.DecodeJwt(tokenHeader)

	claims, _ := token.Claims.(jwt.MapClaims)

	userId, _ := claims["id"].(string)

	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	resp, err := services.DashboardInvestor(r.Context(), userId, limit)
	if err != nil {
		helper.Response(w, 400, true, err.Error(), nil)
		return
	}
	helper.Response(w, 200, false, "OK", resp)
}
