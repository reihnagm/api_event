package controllers

import (
	"net/http"
	helper "superapps/helpers"
	service "superapps/services"

	"github.com/gorilla/mux"
)

func ProjectCostOfFundTemplateWithoutDetail(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		helper.Logger("error", "In Server: id is required")
		helper.Response(w, 400, true, "id is required", map[string]any{})
		return
	}

	result, err := service.ProjectCostOfFundTemplateWithoutDetail(id)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Project Cost Of Fund Template Without Detail success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
