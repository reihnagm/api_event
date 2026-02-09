package controllers

import (
	"net/http"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func InboxList(w http.ResponseWriter, r *http.Request) {

	isAdmin := r.URL.Query().Get("is_admin")

	tokenHeader := r.Header.Get("Authorization")

	token := helper.DecodeJwt(tokenHeader)

	claims, _ := token.Claims.(jwt.MapClaims)

	userId, _ := claims["id"].(string)

	result, err := services.InboxList(userId, isAdmin)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Inbox List success")
	helper.Response(w, http.StatusOK, false, "Successfully",
		result["data"],
	)
}
