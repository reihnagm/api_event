package entities

type MidtransResponse struct {
	Data struct {
		Data struct {
			Actions []struct {
				Url string `json:"url"`
			} `json:"actions"`
			VANumber string `json:"vaNumber"`
		} `json:"data"`
		Expire string `json:"expire"`
	} `json:"data"`
}

type MidtransErrorResponse struct {
	Message string `json:"message"`
}
