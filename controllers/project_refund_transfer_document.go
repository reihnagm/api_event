package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"
)

func ProjectRefundTransferDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.Response(w, http.StatusMethodNotAllowed, true, "Method not allowed", map[string]any{})
		return
	}

	data := &entities.PaymentProjectReqRefundTransDoc{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]any{})
		return
	}

	result, err := services.ProjectRefundTransferDocument(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Payment Project success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
