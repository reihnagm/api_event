package services

import (
	"errors"
	"superapps/entities"
	helper "superapps/helpers"
)

func GetProvince() (map[string]any, error) {
	var province []entities.GetProvince

	// ambil 1 id per province secara deterministik
	query := `
		SELECT
			MIN(id) AS id,
			province_name AS name
		FROM jne_destinations
		WHERE province_name IS NOT NULL AND province_name <> ''
		GROUP BY province_name
		ORDER BY province_name
	`

	err := dbDefault.Raw(query).Scan(&province).Error
	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	return map[string]any{"data": province}, nil
}

func GetCity(provinceName string) (map[string]any, error) {
	var city []entities.GetCity

	query := `
		SELECT
			MIN(id) AS id,
			city_name AS name
		FROM jne_destinations
		WHERE province_name = ?
		  AND city_name IS NOT NULL AND city_name <> ''
		GROUP BY city_name
		ORDER BY city_name
	`

	err := dbDefault.Raw(query, provinceName).Scan(&city).Error
	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	return map[string]any{"data": city}, nil
}

func GetDistrict(cityName string) (map[string]any, error) {
	var district []entities.GetDistrict

	query := `
		SELECT
			MIN(id) AS id,
			district_name AS name
		FROM jne_destinations
		WHERE city_name = ?
		  AND district_name IS NOT NULL AND district_name <> ''
		GROUP BY district_name
		ORDER BY district_name
	`

	err := dbDefault.Raw(query, cityName).Scan(&district).Error
	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	return map[string]any{"data": district}, nil
}

func GetSubdistrict(districtName string) (map[string]any, error) {
	var subdistrict []entities.GetSubdistrict

	// zip_code juga harus deterministik kalau ada grouping
	// pilih MAX(zip_code) (atau MIN) sesuai kebutuhan
	query := `
		SELECT
			MIN(id) AS id,
			subdistrict_name AS name,
			MAX(zip_code) AS zip_code
		FROM jne_destinations
		WHERE district_name = ?
		  AND subdistrict_name IS NOT NULL AND subdistrict_name <> ''
		GROUP BY subdistrict_name
		ORDER BY subdistrict_name
	`

	err := dbDefault.Raw(query, districtName).Scan(&subdistrict).Error
	if err != nil {
		helper.Logger("error", "In Server: "+err.Error())
		return nil, errors.New(err.Error())
	}

	return map[string]any{"data": subdistrict}, nil
}
