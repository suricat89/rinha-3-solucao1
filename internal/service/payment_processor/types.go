package paymentprocessor

type ServiceHealthResponse struct {
	Failing         bool  `json:"failing"`
	MinResponseTime int32 `json:"minResponseTime"`
}
