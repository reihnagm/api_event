package helper

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	entities "superapps/entities"
)

func MessageSuccess(status int, err bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "error": err, "message": message}
}

func MessageError(status int, err bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "error": err, "message": message}
}

func Response(w http.ResponseWriter, status int, err bool, message string, data any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	jsonEncoder := json.NewEncoder(w)
	errs := jsonEncoder.Encode(&entities.Response{Status: status, Error: err, Message: message, Data: data})
	if errs != nil {
		Logger("error", errs.Error())
	}
}

func ResponseWithPagination(w http.ResponseWriter, status int, err bool, message string, total any, perPage any, prevPage any, nextPage any, currentPage any, nextUrl any, prevUrl, data any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	jsonEncoder := json.NewEncoder(w)
	errs := jsonEncoder.Encode(&entities.ResponseWithPagination{Status: status, Error: err, Message: message, Total: total, PerPage: perPage, PrevPage: prevPage, NextPage: nextPage, CurrentPage: currentPage, NextUrl: nextUrl, PrevUrl: prevUrl, Data: data})
	if errs != nil {
		Logger("error", errs.Error())
	}
}

func FormatError(err string) error {

	if strings.Contains(err, "email") {
		Logger("error", "In Server: Email Already Taken")
		return errors.New("email already taken")
	}

	if strings.Contains(err, "phone") {
		Logger("error", "In Server: Phone Already Taken")
		return errors.New("phone already taken")
	}

	if strings.Contains(err, "hashedPassword") {
		Logger("error", "In Server: Incorrect Password")
		return errors.New("incorrect password")
	}

	return errors.New(err)
}
