package controllers

import (
	"net/http"
	helper "superapps/helpers"
	service "superapps/services"

	"github.com/gorilla/mux"
)

func ProjectCostOfFundTemplateWithoutDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	result, err := service.ProjectCostOfFundTemplateWithoutDelete(id)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Project Cost Of Fund Template Without Delete success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
