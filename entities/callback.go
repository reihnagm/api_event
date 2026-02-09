package entities

type Callback struct {
	Status   string `json:"status"`
	Platform string `json:"platform"`
	OrderId  string `json:"order_id"`
}
