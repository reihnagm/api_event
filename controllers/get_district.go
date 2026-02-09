package controllers

import (
	"net/http"
	helper "superapps/helpers"
	service "superapps/services"

	"github.com/gorilla/mux"
)

func GetDistrict(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cityId := vars["city_id"]

	result, err := service.GetDistrict(cityId)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Get District success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
