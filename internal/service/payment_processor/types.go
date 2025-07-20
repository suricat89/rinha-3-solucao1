package paymentprocessor

type ServiceHealthResponse struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

type ProcessorType string

const (
	DefaultProcessor  ProcessorType = "default"
	FallbackProcessor ProcessorType = "fallback"
)
