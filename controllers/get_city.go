package controllers

import (
	"net/http"
	helper "superapps/helpers"
	service "superapps/services"

	"github.com/gorilla/mux"
)

func GetCity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provinceId := vars["province_id"]

	result, err := service.GetCity(provinceId)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Get City success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
