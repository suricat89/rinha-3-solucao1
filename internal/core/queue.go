package core

type queue struct {
	items chan *PaymentRequest
}

type IQueue interface {
	Publish(p *PaymentRequest)
	Consume() *PaymentRequest
}

func NewQueue(bufferSize int) IQueue {
	return &queue{
		items: make(chan *PaymentRequest, bufferSize),
	}
}

func (q *queue) Publish(p *PaymentRequest) {
	q.items <- p
}

func (q *queue) Consume() *PaymentRequest {
	return <-q.items
}
