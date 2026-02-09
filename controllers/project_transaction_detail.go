package controllers

import (
	"net/http"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func ProjectTransactionDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.Response(w, http.StatusMethodNotAllowed, true, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	tokenHeader := r.Header.Get("Authorization")

	token := helper.DecodeJwt(tokenHeader)

	claims, _ := token.Claims.(jwt.MapClaims)

	userId, _ := claims["id"].(string)

	detail, err := services.ProjectTransactionDetail(r.Context(), userId, id)
	if err != nil {
		helper.Response(w, 404, true, err.Error(), nil)
		return
	}
	helper.Response(w, 200, false, "OK", detail)
}
