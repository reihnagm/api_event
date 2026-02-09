package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	service "superapps/services"

	"github.com/gorilla/mux"
)

func ProjectCostOfFundTemplateUpdate(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		helper.Logger("error", "In Server: id is required")
		helper.Response(w, 400, true, "id is required", map[string]any{})
		return
	}

	data := &entities.ProjectCostOfFundTemplateUpdate{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]any{})
		return
	}

	data.Id = id

	result, err := service.ProjectCostOfFundTemplateUpdate(data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Project Cost Of Fund Template Update success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
