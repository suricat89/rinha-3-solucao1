package core

import (
	"log"

	"github.com/suricat89/rinha-3-solucao1/internal/usecase"
)

type consumer struct {
	queue                   IQueue
	paymentProcessorUseCase usecase.IPaymentProcessorUseCase
}

type IConsumer interface {
	StartConsumer(amountGoroutines int)
}

func NewConsumer(
	queue IQueue,
	paymentProcessorUseCase usecase.IPaymentProcessorUseCase,
) IConsumer {
	return &consumer{
		queue,
		paymentProcessorUseCase,
	}
}

func (c *consumer) StartConsumer(amountGoroutines int) {
	go c.paymentProcessorUseCase.MonitorServiceHealth()
	for range amountGoroutines {
		go c.consume()
	}
	log.Printf("Consumer started with %d goroutines", amountGoroutines)
}

func (c *consumer) consume() {
	for {
		paymentRequest := c.queue.Consume()
		err := c.paymentProcessorUseCase.ProcessPayment(paymentRequest.CorrelationId, paymentRequest.Amount)
		if err != nil {
			c.queue.Publish(paymentRequest)
		}
	}
}

