package controllers

import (
	"net/http"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/gorilla/mux"
)

func AdminDetailProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	result, err := services.AdminDetailProject(id)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Admin Detail Project success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
