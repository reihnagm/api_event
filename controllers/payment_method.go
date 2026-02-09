package controllers

import (
	"net/http"
	helper "superapps/helpers"
	"superapps/services"
)

func PaymentMethod(w http.ResponseWriter, r *http.Request) {

	result, err := services.PaymentMethod()

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	helper.Logger("info", "Payment Method success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
