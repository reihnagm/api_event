package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func Order(w http.ResponseWriter, r *http.Request) {

	data := &entities.Order{}

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

	ProjectId := data.ProjectId
	PaymentMethod := data.PaymentMethod
	Price := data.Price

	data.UserId = userId

	if ProjectId == "" {
		helper.Logger("error", "In Server: project_id is required")
		helper.Response(w, 400, true, "project_id is required", map[string]any{})
		return
	}

	if PaymentMethod == "" {
		helper.Logger("error", "In Server: payment_method is required")
		helper.Response(w, 400, true, "payment_method is required", map[string]any{})
		return
	}

	if Price == 0 {
		helper.Logger("error", "In Server: price is required")
		helper.Response(w, 400, true, "price is required", map[string]any{})
		return
	}

	result, err := services.Order(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Order success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
