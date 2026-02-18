package controllers

import (
	"net/http"

	helper "superapps/helpers"
	"superapps/services"
)

func AddEventImage(w http.ResponseWriter, r *http.Request) {
	data, err := services.AddEventImage(r)
	if err != nil {
		helper.Logger("error", "AddEventImage: "+err.Error())
		helper.Response(w, httpStatusFromError(err), true, msgFromError(err), data)
		return
	}
	helper.Response(w, http.StatusOK, false, "Successfully", data)
}

func ReplaceEventImage(w http.ResponseWriter, r *http.Request) {
	data, err := services.ReplaceEventImage(r)
	if err != nil {
		helper.Logger("error", "ReplaceEventImage: "+err.Error())
		helper.Response(w, httpStatusFromError(err), true, msgFromError(err), data)
		return
	}
	helper.Response(w, http.StatusOK, false, "Successfully", data)
}

func DeleteEventImage(w http.ResponseWriter, r *http.Request) {
	data, err := services.DeleteEventImage(r)
	if err != nil {
		helper.Logger("error", "DeleteEventImage: "+err.Error())
		helper.Response(w, httpStatusFromError(err), true, msgFromError(err), data)
		return
	}
	helper.Response(w, http.StatusOK, false, "Successfully", data)
}
