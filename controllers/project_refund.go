package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"
)

func ProjectRefund(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.Response(w, http.StatusMethodNotAllowed, true, "Method not allowed", map[string]interface{}{})
		return
	}

	data := &entities.PaymentProjectReqRefund{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	result, err := services.ProjectRefund(r, data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Refund Payment Project success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
