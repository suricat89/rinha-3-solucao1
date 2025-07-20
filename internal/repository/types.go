package repository

import "time"

type SummaryItem struct {
	ProcessorId   string    `json:"processorId"`
	CorrelationId string    `json:"correlationId"`
	RequestedAt   time.Time `json:"requestedAt"`
	ResponseAt    time.Time `json:"responseAt"`
	Amount        float64   `json:"amount"`
}
