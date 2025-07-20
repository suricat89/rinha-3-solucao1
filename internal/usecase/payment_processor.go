package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/sony/gobreaker"
	"github.com/suricat89/rinha-3-solucao1/internal/repository"
	"github.com/suricat89/rinha-3-solucao1/internal/service/payment_processor"
)

type paymentProcessorUseCase struct {
	defaultService  paymentprocessor.IPaymentProcessorService
	fallbackService paymentprocessor.IPaymentProcessorService
	cacheRepository repository.ICacheRepository
	cb              *gobreaker.CircuitBreaker
}

type IPaymentProcessorUseCase interface {
	ProcessPayment(correlationId string, amount float32) error
	GetPayments(fromTime time.Time, toTime time.Time) map[string]*SummaryResult
	PurgePayments() error
}

func NewPaymentProcessorUseCase(
	defaultService, fallbackService paymentprocessor.IPaymentProcessorService,
	cacheRepository repository.ICacheRepository,
) IPaymentProcessorUseCase {
	cbSettings := gobreaker.Settings{
		Name:        "PaymentProcessorCB",
		MaxRequests: 1,               // just 1 success request to switch to closed state
		Timeout:     1 * time.Second, // switch to semi open state every second
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 1 // just 1 failed request to switch to open state
		},
	}

	cb := gobreaker.NewCircuitBreaker(cbSettings)

	return &paymentProcessorUseCase{
		defaultService,
		fallbackService,
		cacheRepository,
		cb,
	}
}

func (p *paymentProcessorUseCase) ProcessPayment(correlationId string, amount float32) error {
	requestedAt := time.Now()
	processorId := "default"
	_, err := p.cb.Execute(func() (any, error) {
		timeout := 10 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		done := make(chan error, 1)

		go func() {
			done <- p.defaultService.PostPayment(correlationId, amount, requestedAt)
		}()

		select {
		case err := <-done:
			return nil, err
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout de %s atingido ao chamar o processador default", timeout)
		}
	})

	if err != nil {
		requestedAt = time.Now()
		processorId = "fallback"
		err = p.fallbackService.PostPayment(correlationId, amount, requestedAt)
		if err != nil {
			return err
		}
	}

	p.cacheRepository.AddPayment(processorId, correlationId, requestedAt, amount)
	return nil
}

func (p *paymentProcessorUseCase) GetPayments(fromTime time.Time, toTime time.Time) map[string]*SummaryResult {
	summaryItems := p.cacheRepository.GetPayments(fromTime, toTime)
	summary := map[string]*SummaryResult{
		"default":  {TotalRequests: 0, TotalAmount: 0.0},
		"fallback": {TotalRequests: 0, TotalAmount: 0.0},
	}

	for _, summaryItem := range summaryItems {
		summary[summaryItem.ProcessorId].TotalRequests++
		summary[summaryItem.ProcessorId].TotalAmountCents += int64(math.Round(summaryItem.Amount * 100))
	}

	for _, summaryResult := range summary {
		summaryResult.TotalAmount = float64(summaryResult.TotalAmountCents) / 100
	}

	return summary
}

func (p *paymentProcessorUseCase) PurgePayments() error {
	return p.cacheRepository.PurgePayments()
}
