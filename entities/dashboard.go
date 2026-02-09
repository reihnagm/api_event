package entities

type InvestorSummary struct {
	JobID             int64   `json:"job_id"`
	AnnualIncomeIDR   uint64  `json:"annual_income_idr"`
	AnnualQuotaIDR    *uint64 `json:"annual_quota_idr,omitempty"` // <- pointer + omitempty
	UsedThisYearIDR   uint64  `json:"used_this_year_idr"`
	PaidAllTimeIDR    uint64  `json:"paid_all_time_idr"`
	PaidThisYearIDR   uint64  `json:"paid_this_year_idr"`
	RemainingQuotaIDR *uint64 `json:"remaining_quota_idr,omitempty"` // <- pointer + omitempty
	ProjectsCount     int64   `json:"projects_count"`
	ActiveInvoices    int64   `json:"active_invoices"`
	QuotaEnforced     bool    `json:"quota_enforced"`
}

type InvestorDashboard struct {
	VerifiedInvestor   bool                          `json:"verified_investor"`
	RekEfek            bool                          `json:"rek_efek"`
	IsInstitusi        bool                          `json:"is_institusi"`
	Summary            InvestorSummary               `json:"summary"`
	ActiveInvoices     []InvestorInvoiceItem         `json:"active_invoices,omitempty"`
	Portfolio          []ProjectPortfolioItem        `json:"portfolio,omitempty"`
	MonthlyPaid        []MonthlyPoint                `json:"monthly_paid,omitempty"`
	RecentTransactions []InvestorTransactionListItem `json:"recent_transactions,omitempty"`
}

type InvestorTransactionListItem struct {
	PaymentID     uint64  `json:"payment_id"`
	ProjectUID    string  `json:"project_uid"`
	ProjectTitle  string  `json:"project_title"`
	Amount        int64   `json:"amount"`
	PaymentStatus string  `json:"payment_status"`
	CreatedAt     string  `json:"created_at"`
	PaidAt        *string `json:"paid_at,omitempty"`
	OrderID       *string `json:"order_id,omitempty"`
	Provider      *string `json:"provider,omitempty"`
	InvoiceStatus *string `json:"invoice_status,omitempty"`
	ChannelCode   *string `json:"channel_code,omitempty"`
	ChannelRef    *string `json:"channel_ref,omitempty"`
	ExpiresAt     *string `json:"expires_at,omitempty"`
	PaymentURL    *string `json:"payment_url,omitempty"`
}

type InvestorInvoiceItem struct {
	InvoiceID     uint64  `json:"invoice_id"`
	Provider      string  `json:"provider"`
	OrderID       string  `json:"order_id"`
	Amount        int64   `json:"amount"`
	InvoiceStatus string  `json:"invoice_status"`
	ChannelCode   *string `json:"channel_code,omitempty"`
	ChannelRef    *string `json:"channel_ref,omitempty"`
	ExpiresAt     *string `json:"expires_at,omitempty"`
	PaidAt        *string `json:"paid_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
	PaymentURL    *string `json:"payment_url,omitempty"`
}

type ProjectPortfolioItem struct {
	ProjectUID               string                        `json:"project_uid"`
	ProjectTitle             string                        `json:"project_title"`
	FundingStatus            string                        `json:"funding_status"`
	TargetAmount             uint64                        `json:"target_amount"`
	UserPaidIdr              int64                         `json:"user_paid_idr"`
	UserPending              int64                         `json:"user_pending_idr"`
	ProjectPaidAmountIdr     uint64                        `json:"project_paid_amount_idr"`
	ProjectReservedAmountIdr uint64                        `json:"project_reserved_amount_idr"`
	RecentTransactions       []InvestorTransactionListItem `json:"recent_transactions"`
}

type MonthlyPoint struct {
	Month  string `json:"month"`
	Amount int64  `json:"amount_idr"`
}
