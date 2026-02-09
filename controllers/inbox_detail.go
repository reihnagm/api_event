package controllers

import (
	"net/http"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/gorilla/mux"
)

func InboxDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	result, err := services.InboxDetail(id)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Inbox Detail success")
	helper.Response(w, http.StatusOK, false, "Successfully",
		result["data"],
	)
}
