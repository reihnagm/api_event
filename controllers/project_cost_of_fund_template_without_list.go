package controllers

import (
	"net/http"
	helper "superapps/helpers"
	service "superapps/services"
)

func ProjectCostOfFundTemplateWithoutList(w http.ResponseWriter, r *http.Request) {

	result, err := service.ProjectCostOfFundTemplateWithoutList()

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Project Cost Of Fund Template Without List success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
