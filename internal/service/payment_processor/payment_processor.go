package paymentprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

type paymentProcessorService struct {
	baseUrl     string
	processorId ProcessorType
}

type IPaymentProcessorService interface {
	GetProcessorId() ProcessorType
	GetHealth() (*ServiceHealthResponse, error)
	PostPayment(correlationId string, amount float32, requestedAt time.Time) error
}

func NewPaymentProcessorService(baseUrl string, processorId ProcessorType) IPaymentProcessorService {
	return &paymentProcessorService{baseUrl, processorId}
}

func (p *paymentProcessorService) GetProcessorId() ProcessorType {
	return p.processorId
}

func (p *paymentProcessorService) GetHealth() (*ServiceHealthResponse, error) {
	statusCode, body, err := fasthttp.Get(nil, fmt.Sprintf("%s/payments/service-health", p.baseUrl))
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("couldn't get processor status. Returned status code: %d", statusCode)
	}

	response := new(ServiceHealthResponse)
	err = json.Unmarshal(body, response)
	return response, err
}

func (p *paymentProcessorService) PostPayment(correlationId string, amount float32, requestedAt time.Time) error {
	reqBody := map[string]any{
		"correlationId": correlationId,
		"amount":        amount,
		"requestedAt":   requestedAt.UTC().Format(time.RFC3339Nano),
	}

	reqBodyStr, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(fmt.Sprintf("%s/payments", p.baseUrl))
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(reqBodyStr)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- fasthttp.Do(req, resp)
	}()

	select {
	case err := <-done:
		if err != nil {
			return err
		}
		if resp.StatusCode() != fasthttp.StatusOK {
			return fmt.Errorf("error posting payment. status code: %d", resp.StatusCode())
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout de %s atingido ao chamar o processador default", timeout)
	}
}
