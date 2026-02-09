package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func InboxStore(w http.ResponseWriter, r *http.Request) {

	data := &entities.InboxStore{}

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

	Title := data.Title
	Content := data.Content
	ReceiverId := data.ReceiverId
	data.UserId = userId

	if Title == "" {
		helper.Logger("error", "In Server: title is required")
		helper.Response(w, 400, true, "title is required", map[string]any{})
		return
	}

	if Content == "" {
		helper.Logger("error", "In Server: content is required")
		helper.Response(w, 400, true, "content is required", map[string]any{})
		return
	}

	if ReceiverId == "" {
		helper.Logger("error", "In Server: receiver_id is required")
		helper.Response(w, 400, true, "receiver_id is required", map[string]any{})
		return
	}

	result, err := services.InboxStore(data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Inbox Detail success")
	helper.Response(w, http.StatusOK, false, "Successfully",
		result["data"],
	)
}
