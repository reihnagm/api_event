package controllers

import (
	"net/http"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/gorilla/mux"
)

func ProjectDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	result, err := services.ProjectDetail(id)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	if result == nil || result["data"] == nil {
		helper.Response(w, http.StatusBadRequest, true, "project not found", map[string]any{})
		return
	}

	helper.Logger("info", "Project Detail success")
	helper.Response(w, http.StatusOK, false, "Successfully",
		result["data"],
	)
}
