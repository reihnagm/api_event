package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	service "superapps/services"

	"github.com/gorilla/mux"
)

func AdminUpdateProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	data := &entities.AdminUpdateProject{}

	data.ProjectId = id

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	result, err := service.AdminUpdateProject(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Update Admin Project success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
