package controllers

import (
	"net/http"

	helper "superapps/helpers"
	"superapps/services"
)

func httpStatusFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch err.Error() {
	case "BAD_REQUEST", "INVALID_JSON", "INVALID_DATE_FORMAT", "NOTHING_TO_UPDATE", "IMAGE_NOT_FOUND":
		return http.StatusBadRequest
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "FORBIDDEN":
		return http.StatusForbidden
	case "EVENT_NOT_FOUND":
		return http.StatusNotFound
	case "INTERNAL_SERVER_ERROR":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func msgFromError(err error) string {
	if err == nil {
		return "Successfully"
	}

	switch err.Error() {
	case "BAD_REQUEST":
		return "Bad request"
	case "INVALID_JSON":
		return "Invalid JSON"
	case "INVALID_DATE_FORMAT":
		return "Invalid date format"
	case "NOTHING_TO_UPDATE":
		return "Nothing to update"
	case "IMAGE_NOT_FOUND":
		return "Image not found"
	case "UNAUTHORIZED":
		return "Unauthorized"
	case "FORBIDDEN":
		return "Forbidden"
	case "EVENT_NOT_FOUND":
		return "Event not found"
	case "INTERNAL_SERVER_ERROR":
		return "Internal server error"
	default:
		return "Internal server error"
	}
}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	data, err := services.CreateEvent(r)
	if err != nil {
		helper.Logger("error", "CreateEvent: "+err.Error())
		helper.Response(w, httpStatusFromError(err), true, msgFromError(err), data)
		return
	}
	helper.Response(w, http.StatusOK, false, "Successfully", data)
}

func GetListEvent(w http.ResponseWriter, r *http.Request) {
	data, err := services.GetListEvent(r)
	if err != nil {
		helper.Logger("error", "GetListEvent: "+err.Error())
		helper.Response(w, httpStatusFromError(err), true, msgFromError(err), data)
		return
	}
	helper.Response(w, http.StatusOK, false, "Successfully", data)
}

func GetDetailEvent(w http.ResponseWriter, r *http.Request) {
	data, err := services.GetDetailEvent(r)
	if err != nil {
		helper.Logger("error", "GetDetailEvent: "+err.Error())
		helper.Response(w, httpStatusFromError(err), true, msgFromError(err), data)
		return
	}
	helper.Response(w, http.StatusOK, false, "Successfully", data)
}

func DeleteEvent(w http.ResponseWriter, r *http.Request) {
	data, err := services.DeleteEvent(r)
	if err != nil {
		helper.Logger("error", "DeleteEvent: "+err.Error())
		helper.Response(w, httpStatusFromError(err), true, msgFromError(err), data)
		return
	}
	helper.Response(w, http.StatusOK, false, "Successfully", data)
}

func UpdateEvent(w http.ResponseWriter, r *http.Request) {
	data, err := services.UpdateEvent(r)
	if err != nil {
		helper.Logger("error", "UpdateEvent: "+err.Error())
		helper.Response(w, httpStatusFromError(err), true, msgFromError(err), data)
		return
	}
	helper.Response(w, http.StatusOK, false, "Successfully", data)
}
