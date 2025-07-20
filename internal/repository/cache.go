package repository

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
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
	score := float64(requestedAt.UnixNano())
	member := fmt.Sprintf(
		"%s|%s|%s|%s|%.2f",
		processorId,
		correlationId,
		requestedAt.Format(time.RFC3339Nano),
		responseAt.Format(time.RFC3339Nano),
		amount,
	)

	return c.client.ZAdd(c.ctx, c.key, &redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

func (c *cacheRepository) GetPayments(fromTime time.Time, toTime time.Time) []*SummaryItem {
	fromMillis := strconv.FormatInt(fromTime.UnixNano(), 10)
	toMillis := strconv.FormatInt(toTime.UnixNano(), 10)

	members, err := c.client.ZRangeByScore(c.ctx, c.key, &redis.ZRangeBy{
		Min: fromMillis,
		Max: toMillis,
	}).Result()

	if err != nil {
		log.Printf("Error fetching payment summary: %v", err)
		return nil
	}

	summaryItems := make([]*SummaryItem, 0)
	for _, member := range members {
		parts := strings.Split(member, "|")
		if len(parts) != 5 {
			continue
		}
		processorId := parts[0]
		correlationId := parts[1]
		requestedAt, err := time.Parse(time.RFC3339Nano, parts[2])
		if err != nil {
			continue
		}
		responseAt, err := time.Parse(time.RFC3339Nano, parts[3])
		if err != nil {
			continue
		}
		amount, err := strconv.ParseFloat(parts[4], 64)
		if err != nil {
			continue
		}

		summaryItems = append(summaryItems, &SummaryItem{
			ProcessorId:   processorId,
			CorrelationId: correlationId,
			RequestedAt:   requestedAt,
			ResponseAt:    responseAt,
			Amount:        amount,
		})
	}

	return summaryItems

}

func (c *cacheRepository) PurgePayments() error {
	_, err := c.client.Del(c.ctx, c.key).Result()
	return err
}
