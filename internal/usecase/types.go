package usecase

import (
	"time"

	paymentprocessor "github.com/suricat89/rinha-3-solucao1/internal/service/payment_processor"
)

type SummaryResult struct {
	TotalRequests     int       `json:"totalRequests"`
	TotalAmount       float64   `json:"totalAmount"`
	TotalAmountCents  int64     `json:"-"`
	LatestRequestedAt time.Time `json:"latestRequestedAt"`
	LatestResponseAt  time.Time `json:"latestResponseAt"`
}

type ServiceStatus map[paymentprocessor.ProcessorType]*paymentprocessor.ServiceHealthResponse
