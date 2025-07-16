package repository

type SummaryResult struct {
	TotalRequests    int     `json:"totalRequests"`
	TotalAmount      float64 `json:"totalAmount"`
	TotalAmountCents int64   `json:"-"`
}
