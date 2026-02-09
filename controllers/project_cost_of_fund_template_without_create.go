package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	service "superapps/services"
)

func ProjectCostOfFundTemplateWithoutCreate(w http.ResponseWriter, r *http.Request) {

	data := &entities.ProjectCostOfFundTemplateWithoutStore{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]any{})
		return
	}

	result, err := service.ProjectCostOfFundTemplateWithoutStore(data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Project Cost Of Fund Template Without Store success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
