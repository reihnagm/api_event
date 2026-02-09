package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func RegisterAsInvestorInstitute(w http.ResponseWriter, r *http.Request) {

	tokenHeader := r.Header.Get("Authorization")

	token := helper.DecodeJwt(tokenHeader)

	claims, _ := token.Claims.(jwt.MapClaims)

	userId, _ := claims["id"].(string)

	data := &entities.RegisterAsEmiten{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]interface{}{})
		return
	}

	data.UserId = userId

	Fullname := data.Fullname
	Jabatan := data.Jabatan
	PhotoSelfie := data.PhotoSelfie
	NoKtp := data.NoKtp
	NoNpwp := data.NoNpwp
	PhotoKtp := data.PhotoKtp

	if Fullname == "" {
		helper.Logger("error", "In Server: fullname is required")
		helper.Response(w, 400, true, "fullname is required", map[string]any{})
		return
	}

	if Jabatan == "" {
		helper.Logger("error", "In Server: jabatan is required")
		helper.Response(w, 400, true, "jabatan is required", map[string]any{})
		return
	}

	if NoKtp == "" {
		helper.Logger("error", "In Server: no_ktp is required")
		helper.Response(w, 400, true, "no_ktp is required", map[string]any{})
		return
	}

	if NoNpwp == "" {
		helper.Logger("error", "In Server: no_npwp is required")
		helper.Response(w, 400, true, "no_npwp is required", map[string]any{})
		return
	}

	if PhotoSelfie == "" {
		helper.Logger("error", "In Server: photo_selfie is required")
		helper.Response(w, 400, true, "photo_selfie is required", map[string]any{})
		return
	}

	if PhotoKtp == "" {
		helper.Logger("error", "In Server: photo_ktp is required")
		helper.Response(w, 400, true, "photo_ktp is required", map[string]any{})
		return
	}

	result, err := services.RegisterAsEmiten(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Register as Emiten success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
