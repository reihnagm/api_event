package entities

type GetProvince struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type GetCity struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type GetDistrict struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type GetSubdistrict struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	ZipCode string `json:"zip_code"`
}
