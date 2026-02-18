package controllers

import (
	"net/http"

	helper "superapps/helpers"
	"superapps/services"
)

func authHTTPStatusFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch err.Error() {
	case "BAD_REQUEST", "INVALID_JSON":
		return http.StatusBadRequest
	case "EMAIL_EXISTS":
		return http.StatusBadRequest
	case "USER_NOT_FOUND":
		return http.StatusNotFound
	case "INVALID_CREDENTIALS":
		return http.StatusBadRequest
	case "INTERNAL_SERVER_ERROR":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func authMsgFromError(err error) string {
	if err == nil {
		return "Successfully"
	}
	switch err.Error() {
	case "BAD_REQUEST":
		return "Bad request"
	case "INVALID_JSON":
		return "Invalid JSON"
	case "EMAIL_EXISTS":
		return "Email already registered"
	case "USER_NOT_FOUND":
		return "User not found"
	case "INVALID_CREDENTIALS":
		return "Invalid credentials"
	case "INTERNAL_SERVER_ERROR":
		return "Internal server error"
	default:
		return "Internal server error"
	}
}

// POST /api/v1/auth/register
func Register(w http.ResponseWriter, r *http.Request) {
	data, err := services.Register(r)
	if err != nil {
		helper.Logger("error", "Register: "+err.Error())
		helper.Response(w, authHTTPStatusFromError(err), true, authMsgFromError(err), data)
		return
	}
	helper.Response(w, http.StatusOK, false, "Successfully", data)
}

// POST /api/v1/auth/login
func Login(w http.ResponseWriter, r *http.Request) {
	data, err := services.Login(r)
	if err != nil {
		helper.Logger("error", "Login: "+err.Error())
		helper.Response(w, authHTTPStatusFromError(err), true, authMsgFromError(err), data)
		return
	}
	helper.Response(w, http.StatusOK, false, "Successfully", data)
}

// POST /api/v1/auth/logout
func Logout(w http.ResponseWriter, r *http.Request) {
	data, err := services.Logout(r)
	if err != nil {
		helper.Logger("error", "Logout: "+err.Error())
		helper.Response(w, authHTTPStatusFromError(err), true, authMsgFromError(err), data)
		return
	}
	helper.Response(w, http.StatusOK, false, "Successfully", data)
}
