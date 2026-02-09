package entities

type OrderProjectCallback struct {
	ProjectId string `json:"project_id"`
}

type Order struct {
	Price         int    `json:"price"`
	PaymentMethod string `json:"payment_method"`
	ProjectId     string `json:"project_id"`
	UserId        string `json:"user_id"`
	ReceiverId    string `json:"receiver_id"`
}

type OrderResult struct {
	Price   int                `json:"price"`
	Invoice string             `json:"invoice"`
	Payment OrderPaymentResult `json:"payment"`
	Project OrderProjectResult `json:"project"`
	Inbox   OrderInbox         `json:"inbox"`
}

type ProjectOrder struct {
	Title  string `json:"id"`
	UserId string `json:"user_id"`
}

type OrderPaymentResult struct {
	Logo   string `json:"logo"`
	Name   string `json:"name"`
	Fee    int    `json:"fee"`
	Access string `json:"access"`
	Expire string `json:"expire"`
	Type   string `json:"type"`
}

type OrderProjectResult struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

type OrderScan struct {
	No int `json:"no"`
}

type OrderInbox struct {
	Id int `json:"id"`
}
