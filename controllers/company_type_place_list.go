package controllers

import (
	"net/http"
	helper "superapps/helpers"
	service "superapps/services"
)

func CompanyTypePlaceList(w http.ResponseWriter, r *http.Request) {

	result, err := service.CompnayPlaceTypeList()

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Company Type Place List success")
	helper.Response(w, http.StatusOK, false, "Successfully", result["data"])
}
