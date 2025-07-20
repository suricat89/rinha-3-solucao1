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
	AddPayment(processorId string, correlationId string, requestedAt time.Time, amount float32) error
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

func (c *cacheRepository) AddPayment(processorId string, correlationId string, requestedAt time.Time, amount float32) error {
	score := float64(requestedAt.UnixMilli())
	member := fmt.Sprintf("%s:%s:%.2f", processorId, correlationId, amount)

	return c.client.ZAdd(c.ctx, c.key, &redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

func (c *cacheRepository) GetPayments(fromTime time.Time, toTime time.Time) []*SummaryItem {
	fromMillis := strconv.FormatInt(fromTime.UnixMilli(), 10)
	toMillis := strconv.FormatInt(toTime.UnixMilli(), 10)

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
		parts := strings.Split(member, ":")
		if len(parts) != 3 {
			continue
		}
		processorId := parts[0]
		correlationId := parts[1]
		amount, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			continue
		}

		summaryItems = append(summaryItems, &SummaryItem{
			ProcessorId:   processorId,
			CorrelationId: correlationId,
			Amount:        amount,
		})
	}

	return summaryItems

}

func (c *cacheRepository) PurgePayments() error {
	_, err := c.client.Del(c.ctx, c.key).Result()
	return err
}
