package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func ProjectPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.Response(w, http.StatusMethodNotAllowed, true, "Method not allowed", map[string]any{})
		return
	}

	data := &entities.PaymentProjectReq{}

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

	result, err := services.ProjectPayment(r, data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Payment Project success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
