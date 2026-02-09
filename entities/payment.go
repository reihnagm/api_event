package entities

import "time"

type PaymentMethod struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	NameCode string `json:"name_code"`
	Logo     string `json:"logo"`
	Platform string `json:"platform"`
	Fee      int    `json:"fee"`
}

type PaymentContractLetter struct {
	Path string `json:"path"`
}

type PaymentInvestor struct {
	Id       string `json:"id"`
	Fullname string `json:"fullname"`
	Selfie   string `json:"selfie"`
	Sku      string `json:"sku"`
	Email    string `json:"email"`
	BankName string `json:"bank_name"`
	BankNo   string `json:"bank_no"`
}

type PaymentCompany struct {
	Id                        string `json:"id"`
	Name                      string `json:"name"`
	AktaPerubahanTerahkir     string `json:"akta_perubahan_terahkir"`
	AktaPerubahanTerahkirPath string `json:"akta_perubahan_terahkir_path"`
	BankName                  string `json:"bank_name"`
	BankNo                    string `json:"bank_no"`
}

type PaymentProjectReq struct {
	Id            uint64 `json:"id"`
	ProjectId     string `json:"project_id"`
	PaymentMethod string `json:"payment_method"`
	Amount        int    `json:"amount"`
	Lot           int    `json:"lot"`
	OrderId       string `json:"order_id"`
	Provider      string `json:"provider"`
	VANumber      string `json:"va_number"`
	ExpireAt      string `json:"expire_at"`
	PaymentURL    string `json:"payment_url"`
	UserId        string `json:"user_id"`
}

type PaymentProjectRefundTransDoc struct {
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

type PaymentProjectReqRefundTransDoc struct {
	Path          string `json:"path"`
	TransactionId string `json:"transaction_id"`
	InvestorId    string `json:"investor_id"`
	Amount        string `json:"amount"`
}

type PaymentProjectReqRefund struct {
	PaymentId string `json:"payment_id"`
}

type PaymentProjectResRefund struct {
	PaymentId string `json:"payment_id"`
}

type PaymentProjectJobScan struct {
	Id int64 `json:"id"`
}
