package entities

type MinimalPortfolioItem struct {
	ProjectUID               string `json:"project_uid"`
	ProjectTitle             string `json:"project_title"`
	FundingStatus            string `json:"funding_status"`
	TargetAmountIDR          uint64 `json:"target_amount_idr"`
	UserPaidIDR              int64  `json:"user_paid_idr"`
	UserPendingIDR           int64  `json:"user_pending_idr"`
	ProjectPaidAmountIDR     uint64 `json:"project_paid_amount_idr"`
	ProjectReservedAmountIDR uint64 `json:"project_reserved_amount_idr"`
}

type PortfolioOnlyResponse struct {
	Portfolio []MinimalPortfolioItem `json:"portfolio"`
}
