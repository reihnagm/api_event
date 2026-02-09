package controllers

import (
	"net/http"
	"strconv"
	helper "superapps/helpers"
	"superapps/services"
)

func AdminProjectTransactionList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.Response(w, http.StatusMethodNotAllowed, true, "Method not allowed", nil)
		return
	}

	q := r.URL.Query()
	statusCSV := q.Get("status")

	page, _ := strconv.Atoi(q.Get("page"))
	per, _ := strconv.Atoi(q.Get("limit"))
	if page <= 0 {
		page = 1
	}
	if per <= 0 {
		per = 10
	}

	resp, err := services.AdminProjectListTransactions(r.Context(), statusCSV, page, per)
	if err != nil {
		helper.Logger("error", "ListTransactions: "+err.Error())
		helper.Response(w, 400, true, err.Error(), nil)
		return
	}
	helper.Response(w, 200, false, "OK", resp)
}
