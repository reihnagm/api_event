package controllers

import (
	"encoding/json"
	"net/http"

	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"
)

func ProjectPaymentCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.Response(w, http.StatusMethodNotAllowed, true, "Method not allowed", nil)
		return
	}

	var payload entities.Callback
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&payload); err != nil {
		helper.Logger("error", "callback decode: "+err.Error())
		helper.Response(w, 400, true, "Invalid JSON ("+err.Error()+")", nil)
		return
	}

	// TODO: (disarankan) verifikasi signature/HMAC dari Midtrans di sini.

	if err := services.HandleProjectPaymentCallback(&payload); err != nil {
		helper.Logger("error", "callback process: "+err.Error())
		helper.Response(w, 400, true, err.Error(), nil)
		return
	}

	helper.Response(w, 200, false, "OK", map[string]any{
		"order_id": payload.OrderId,
		"status":   payload.Status,
	})
}
