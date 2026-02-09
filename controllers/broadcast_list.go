package controllers

import (
	"net/http"
	helper "superapps/helpers"
	"superapps/services"
)

func BroadcastList(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")

	result, err := services.BroadcastList(userId)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Broadcast List success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
