package controllers

import (
	"net/http"
	helper "superapps/helpers"
	service "superapps/services"

	"github.com/gorilla/mux"
)

func ProjectCostOfFundTemplateDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	result, err := service.ProjectCostOfFundTemplateDelete(id)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Project Cost Of Fund Template Delete success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
