package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"
)

func ContractLetterPaymentUpload(w http.ResponseWriter, r *http.Request) {

	data := &entities.ContractLetterPaymentUpload{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]any{})
		return
	}

	result, err := services.ContractLetterPaymentUpload(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Contract Letter Payment Upload")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
