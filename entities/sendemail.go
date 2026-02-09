package entities

type SendEmailRequest struct {
	To      string `json:"to"`
	App     string `json:"app"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Type    string `json:"type"`
}
