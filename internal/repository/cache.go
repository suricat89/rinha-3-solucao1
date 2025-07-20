package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type cacheRepository struct {
	key    string
	ctx    context.Context
	client *redis.Client
}

type ICacheRepository interface {
	AddPayment(processorId string, correlationId string, requestedAt time.Time, responseAt time.Time, amount float32) error
	GetPayments(fromTime time.Time, toTime time.Time) []*SummaryItem
	PurgePayments() error
}

func NewCacheRepository(ctx context.Context, host string, port int) ICacheRepository {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: "",
		DB:       0,
	})

	key := "payments"

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Couldn't connect to Redis: %v", err)
	}

	return &cacheRepository{key, ctx, client}
}

func (c *cacheRepository) AddPayment(
	processorId string,
	correlationId string,
	requestedAt time.Time,
	responseAt time.Time,
	amount float32,
) error {
	summaryItem, err := json.Marshal(&SummaryItem{
		ProcessorId:   processorId,
		CorrelationId: correlationId,
		RequestedAt:   requestedAt,
		ResponseAt:    responseAt,
		Amount:        float64(amount),
	})
	if err != nil {
		fmt.Printf("Error marshalling summary item: %v", err)
		return err
	}
	return c.client.HSet(c.ctx, c.key, correlationId, summaryItem).Err()
}

func (c *cacheRepository) GetPayments(fromTime time.Time, toTime time.Time) []*SummaryItem {
	items, err := c.client.HGetAll(c.ctx, c.key).Result()
	if err != nil {
		fmt.Printf("Error fetching transaction list from Redis: %v", err)
		return nil
	}

	summaryItems := make([]*SummaryItem, 0)
	for _, summaryItemStr := range items {
		var summaryItem SummaryItem
		err := json.Unmarshal([]byte(summaryItemStr), &summaryItem)
		if err != nil {
			continue
		}

		if summaryItem.RequestedAt.After(fromTime) && summaryItem.RequestedAt.Before(toTime) {
			summaryItems = append(summaryItems, &summaryItem)
		}
	}

	return summaryItems
}

func (c *cacheRepository) PurgePayments() error {
	_, err := c.client.Del(c.ctx, c.key).Result()
	return err
}
