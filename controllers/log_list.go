package controllers

import (
	"net/http"
	helper "superapps/helpers"
	"superapps/services"
)

func LogList(w http.ResponseWriter, r *http.Request) {
	result, err := services.LogList()

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Log List success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
