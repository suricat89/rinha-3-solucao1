package main

import (
	"context"

	"github.com/suricat89/rinha-3-solucao1/internal/conf"
	"github.com/suricat89/rinha-3-solucao1/internal/core"
	"github.com/suricat89/rinha-3-solucao1/internal/repository"
	"github.com/suricat89/rinha-3-solucao1/internal/service/payment_processor"
	"github.com/suricat89/rinha-3-solucao1/internal/usecase"
)

func main() {
	ctx := context.Background()

	processorDefault := paymentprocessor.NewPaymentProcessorService(
		conf.Env.ProcessorDefaultBaseUrl,
	)
	processorFallback := paymentprocessor.NewPaymentProcessorService(
		conf.Env.ProcessorFallbackBaseUrl,
	)

	cacheRepository := repository.NewCacheRepository(
		ctx,
		conf.Env.RedisHost,
		conf.Env.RedisPort,
	)

	paymentProcessorUseCase := usecase.NewPaymentProcessorUseCase(
		processorDefault,
		processorFallback,
		cacheRepository,
	)

	queue := core.NewQueue(conf.Env.QueueBufferSize)
	producer := core.NewProducer(queue, paymentProcessorUseCase)
	consumer := core.NewConsumer(queue, paymentProcessorUseCase)

	consumer.StartConsumer(conf.Env.ConsumerGoroutines)
	producer.StartApi(conf.Env.Port)
}
