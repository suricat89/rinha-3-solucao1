package core

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/suricat89/rinha-3-solucao1/internal/usecase"
	"github.com/valyala/fasthttp"
)

type producer struct {
	queue                   IQueue
	paymentProcessorUseCase usecase.IPaymentProcessorUseCase
}

type IProducer interface {
	StartApi(port int)
}

func NewProducer(queue IQueue, paymentProcessorUseCase usecase.IPaymentProcessorUseCase) IProducer {
	return &producer{queue, paymentProcessorUseCase}
}

func (p *producer) StartApi(port int) {
	addr := fmt.Sprintf(":%d", port)

	log.Printf("Producer listening on port %s...", addr)
	if err := fasthttp.ListenAndServe(addr, p.requestHandler); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func (p *producer) requestHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/payments":
		p.handlePostPayment(ctx)
	case "/payments-summary":
		p.handleGetPaymentsSummary(ctx)
	case "/purge-payments":
		p.handlePostPurgePayment(ctx)
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

func (p *producer) handlePostPayment(ctx *fasthttp.RequestCtx) {
	if !ctx.IsPost() {
		ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		fmt.Fprintf(ctx, "Method not allowed")
		return
	}

	req := new(PaymentRequest)
	err := json.Unmarshal(ctx.PostBody(), req)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		fmt.Fprintf(ctx, "Invalid request payload: %v", err)
		return
	}

	p.queue.Publish(req)
	ctx.SetStatusCode(fasthttp.StatusCreated)
}

func (p *producer) handleGetPaymentsSummary(ctx *fasthttp.RequestCtx) {
	if !ctx.IsGet() {
		ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		fmt.Fprintf(ctx, "Method not allowed")
		return
	}

	fromStr := string(ctx.QueryArgs().Peek("from"))
	toStr := string(ctx.QueryArgs().Peek("to"))

	if fromStr == "" {
		fromStr = "2000-01-01T00:00:00Z"
	}
	if toStr == "" {
		toStr = "2900-01-01T00:00:00Z"
	}

	fromTime, err := time.Parse(time.RFC3339Nano, fromStr)
	if err != nil {
		ctx.Error("Invalid date format in 'from' query param", fasthttp.StatusBadRequest)
		return
	}
	toTime, err := time.Parse(time.RFC3339Nano, toStr)
	if err != nil {
		ctx.Error("Invalid date format in 'to' query param", fasthttp.StatusBadRequest)
		return
	}

	summary := p.paymentProcessorUseCase.GetPayments(fromTime, toTime)

	responseBody, _ := json.Marshal(summary)
	log.Printf("[%s]: %s", ctx.URI().String(), string(responseBody))

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	json.NewEncoder(ctx).Encode(summary)
}

func (p *producer) handlePostPurgePayment(ctx *fasthttp.RequestCtx) {
	if !ctx.IsPost() {
		ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		fmt.Fprintf(ctx, "Method not allowed")
		return
	}

	err := p.paymentProcessorUseCase.PurgePayments()
	if err != nil {
		fmt.Fprintf(ctx, "Error purging payments: %v", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}
