package controllers

import (
	"net/http"
	helper "superapps/helpers"
	service "superapps/services"
)

func ProjectTypeList(w http.ResponseWriter, r *http.Request) {

	result, err := service.ProjectTypeList()

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Project Type List success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
