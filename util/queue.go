package util

import (
	"context"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/go-redis/redis/v8"
	"time"
)

type KVRedisQueue struct {
	Key string
	Val string
}

type RedisQueue struct {
	redisClient *redis.Client
	queue       *goconcurrentqueue.FIFO
}

func NewRedisQueue(redisClient *redis.Client) *RedisQueue {
	queue := goconcurrentqueue.NewFIFO()

	q := &RedisQueue{
		redisClient: redisClient,
		queue:       queue,
	}

	// start dequeue in background
	go q.dequeue()

	return q
}

func (q *RedisQueue) Enqueue(kvQueue ...KVRedisQueue) {
	for _, kv := range kvQueue {
		q.queue.Enqueue(kv)
	}
}

func (q *RedisQueue) dequeue() {
	for {
		data, err := q.queue.DequeueOrWaitForNextElement()
		if err != nil {
			return
		}
		q.redisClient.Set(context.Background(), data.(KVRedisQueue).Key, data.(KVRedisQueue).Val, time.Hour*24*7)
	}
}
