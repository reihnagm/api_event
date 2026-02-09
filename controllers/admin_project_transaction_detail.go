package controllers

import (
	"net/http"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/gorilla/mux"
)

func AdminProjectTransactionDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.Response(w, http.StatusMethodNotAllowed, true, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	detail, err := services.AdminProjectGetTransactionDetail(r.Context(), id)
	if err != nil {
		helper.Response(w, 404, true, err.Error(), nil)
		return
	}
	helper.Response(w, 200, false, "OK", detail)
}
