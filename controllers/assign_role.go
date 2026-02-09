package controllers

import (
	"encoding/json"
	"net/http"
	"superapps/entities"
	helper "superapps/helpers"
	"superapps/services"

	"github.com/dgrijalva/jwt-go"
)

func AssignRole(w http.ResponseWriter, r *http.Request) {

	data := &entities.AssignRole{}

	err := json.NewDecoder(r.Body).Decode(data)

	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		helper.Response(w, 400, true, "Internal server error ("+err.Error()+")", map[string]any{})
		return
	}

	tokenHeader := r.Header.Get("Authorization")

	token := helper.DecodeJwt(tokenHeader)

	claims, _ := token.Claims.(jwt.MapClaims)

	userId, _ := claims["id"].(string)

	data.UserId = userId

	if data.Role == "" {
		helper.Logger("error", "In Server: role is required")
		helper.Response(w, 400, true, "role is required", map[string]any{})
		return
	}

	if data.Role == "1" {

		if data.Ktp.Name == "" {
			helper.Logger("error", "In Server: ktp.name is required")
			helper.Response(w, 400, true, "ktp.name is required", map[string]any{})
			return
		}

		if data.Ktp.Nik == "" {
			helper.Logger("error", "In Server: ktp.nik is required")
			helper.Response(w, 400, true, "ktp.nik is required", map[string]any{})
			return
		}

	}

	if data.Role == "2" {

		if data.CompanyName == "" {
			helper.Logger("error", "In Server: company_name is required")
			helper.Response(w, 400, true, "company_name is required", map[string]any{})
			return
		}

		if data.CompanyNib == "" {
			helper.Logger("error", "In Server: company_nib is required")
			helper.Response(w, 400, true, "company_nib is required", map[string]any{})
			return
		}

		if data.CompanyNibPath == "" {
			helper.Logger("error", "In Server: company_nib_path is required")
			helper.Response(w, 400, true, "company_nib_path is required", map[string]any{})
			return
		}

		if data.AktaPendirian == "" {
			helper.Logger("error", "In Server: akta_pendirian is required")
			helper.Response(w, 400, true, "akta_pendirian is required", map[string]any{})
			return
		}

		if data.AktaPendirianPath == "" {
			helper.Logger("error", "In Server: akta_pendirian_path is required")
			helper.Response(w, 400, true, "akta_pendirian_path is required", map[string]any{})
			return
		}

		if data.AktaPerubahanTerahkir == "" {
			helper.Logger("error", "In Server: akta_perubahan_terahkir is required")
			helper.Response(w, 400, true, "akta_perubahan_terahkir is required", map[string]any{})
			return
		}

		if data.AktaPerubahanTerahkirPath == "" {
			helper.Logger("error", "In Server: akta_perubahan_terahkir_path is required")
			helper.Response(w, 400, true, "akta_perubahan_terahkir_path is required", map[string]any{})
			return
		}

		if data.SkKumham == "" {
			helper.Logger("error", "In Server: sk_kumham is required")
			helper.Response(w, 400, true, "sk_kumham is required", map[string]any{})
			return
		}

		if data.SkKumhamPath == "" {
			helper.Logger("error", "In Server: sk_kumham_path is required")
			helper.Response(w, 400, true, "sk_kumham_path is required", map[string]any{})
			return
		}

		if data.NpwpPath == "" {
			helper.Logger("error", "In Server: npwp_path is required")
			helper.Response(w, 400, true, "npwp_path is required", map[string]any{})
			return
		}

		if data.TotalEmployees == "" {
			helper.Logger("error", "In Server: total_employees is required")
			helper.Response(w, 400, true, "total_employees is required", map[string]any{})
			return
		}

		if data.LaporanKeuanganPath == "" {
			helper.Logger("error", "In Server: laporan_keuangan_path is required")
			helper.Response(w, 400, true, "laporan_keuangan_path is required", map[string]any{})
			return
		}

		if len(data.Address) == 0 {
			helper.Logger("error", "In Server: Address is required")
			helper.Response(w, 400, true, "Address is required", map[string]any{})
			return
		}

		if len(data.Komisaris) == 0 {
			helper.Logger("error", "In Server: Komisaris is required")
			helper.Response(w, 400, true, "Komisaris is required", map[string]any{})
			return
		}

		if len(data.Directors) == 0 {
			helper.Logger("error", "In Server: Directors is required")
			helper.Response(w, 400, true, "Directors is required", map[string]any{})
			return
		}

	}

	result, err := services.AssignRole(r, data)

	if err != nil {
		helper.Response(w, 400, true, err.Error(), map[string]any{})
		return
	}

	helper.Logger("info", "Login success")
	helper.Response(w, http.StatusOK, false, "Successfully", result)
}
