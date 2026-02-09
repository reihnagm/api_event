package controllers

import (
	"net/http"

	helper "superapps/helpers"
	"superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func PortfolioInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tokenHeader := r.Header.Get("Authorization")

	token := helper.DecodeJwt(tokenHeader)

	claims, _ := token.Claims.(jwt.MapClaims)

	userId, _ := claims["id"].(string)

	if userId == "" {
		helper.Logger("warn", "PortfolioDetail: missing user_id")
		helper.Response(w, http.StatusBadRequest, true, "user_id is required", map[string]any{
			"portfolio": []any{},
		})
		return
	}

	result, err := services.PortfolioInfo(ctx, userId)
	if err != nil {
		helper.Logger("error", "PortfolioOnly error: "+err.Error())
		helper.Response(w, http.StatusBadRequest, true, "Internal server error ("+err.Error()+")", map[string]any{
			"portfolio": []any{},
		})
		return
	}

	helper.Logger("info", "Portfolio detail success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
