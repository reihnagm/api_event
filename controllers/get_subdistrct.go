package controllers

import (
	"net/http"
	helper "superapps/helpers"
	service "superapps/services"

	"github.com/gorilla/mux"
)

func GetSubdistrict(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	districtId := vars["district_id"]

	result, err := service.GetSubdistrict(districtId)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Get Subdistrict success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
