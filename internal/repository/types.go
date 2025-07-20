package repository

import "time"

type SummaryItem struct {
	ProcessorId   string
	CorrelationId string
	RequestedAt   time.Time
	ResponseAt    time.Time
	Amount        float64
}
