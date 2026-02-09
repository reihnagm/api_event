// entities/transaction.go
package entities

import "encoding/json"

type TransactionListItem struct {
	PaymentId      uint64                       `json:"payment_id"`
	ProjectId      string                       `json:"project_id"`
	ProjectTitle   string                       `json:"project_title"`
	ProjectSku     string                       `json:"project_sku"`
	Amount         int64                        `json:"amount"`
	PaymentStatus  string                       `json:"payment_status"`
	CreatedAt      string                       `json:"created_at"`
	PaidAt         *string                      `json:"paid_at,omitempty"`
	ContractLetter PaymentContractLetter        `json:"contract_letter"`
	Investor       PaymentInvestor              `json:"investor"`
	Company        PaymentCompany               `json:"company"`
	Refund         PaymentProjectRefundTransDoc `json:"refund"`
	IsRefund       bool                         `json:"is_refund"`

	// invoice (latest)
	OrderId       *string `json:"order_id,omitempty"`
	Provider      *string `json:"provider,omitempty"`
	InvoiceStatus *string `json:"invoice_status,omitempty"`
	ChannelCode   *string `json:"channel_code,omitempty"`
	ChannelRef    *string `json:"channel_ref,omitempty"`
	ExpiresAt     *string `json:"expires_at,omitempty"`
}

type TransactionListResp struct {
	Items      []TransactionListItem `json:"items"`
	Page       int                   `json:"page"`
	PerPage    int                   `json:"per_page"`
	TotalItems int64                 `json:"total_items"`
}

type TransactionDetail struct {
	PaymentId      uint64                       `json:"payment_id"`
	ProjectId      string                       `json:"project_id"`
	ProjectTitle   string                       `json:"project_title"`
	ProjectSku     string                       `json:"project_sku"`
	Amount         int64                        `json:"amount"`
	PaymentStatus  string                       `json:"payment_status"`
	CreatedAt      string                       `json:"created_at"`
	PaidAt         *string                      `json:"paid_at,omitempty"`
	ContractLetter PaymentContractLetter        `json:"contract_letter"`
	Investor       PaymentInvestor              `json:"investor"`
	Company        PaymentCompany               `json:"company"`
	Refund         PaymentProjectRefundTransDoc `json:"refund"`
	IsRefund       bool                         `json:"is_refund"`

	// Ringkasan project
	ProjectFundingStatus string `json:"project_funding_status"`
	ProjectTarget        uint64 `json:"project_target_amount_idr"`
	ProjectPaid          uint64 `json:"project_paid_amount_idr"`
	ProjectReserved      uint64 `json:"project_reserved_amount_idr"`

	// Semua invoice terkait (terbaru dulu)
	Invoices []InvoiceItem `json:"invoices"`
}

type InvoiceItem struct {
	InvoiceId     uint64           `json:"invoice_id"`
	Provider      string           `json:"provider"`
	OrderId       string           `json:"order_id"`
	Amount        int64            `json:"amount"`
	RawResponse   *json.RawMessage `json:"raw_response,omitempty"`
	InvoiceStatus string           `json:"invoice_status"`
	PaymentMethod PaymentMethod    `json:"payment_method"`
	ChannelCode   *string          `json:"channel_code,omitempty"`
	ChannelRef    *string          `json:"channel_ref,omitempty"`
	ExpiresAt     *string          `json:"expires_at,omitempty"`
	PaidAt        *string          `json:"paid_at,omitempty"`
	CreatedAt     string           `json:"created_at"`
}
