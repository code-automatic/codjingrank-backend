package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// JobPayload is the JSON payload stored in the Redis queue.
type JobPayload struct {
	SubmissionID string `json:"submissionId"`
}

// Queue wraps a Redis client for job queue operations.
type Queue struct {
	client *redis.Client
}

// New creates a new Queue with a connection to Redis at the given address.
func New(redisAddr string) (*Queue, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("queue: failed to connect to redis: %w", err)
	}

	return &Queue{client: client}, nil
}

// Enqueue pushes a submission job onto the Redis queue.
func (q *Queue) Enqueue(ctx context.Context, submissionID string) error {
	payload := JobPayload{SubmissionID: submissionID}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("queue: failed to marshal payload: %w", err)
	}

	if err := q.client.LPush(ctx, "submission_jobs", data).Err(); err != nil {
		return fmt.Errorf("queue: failed to enqueue: %w", err)
	}
	return nil
}

// Dequeue blocks until a submission job is available or the context is cancelled.
// Returns the submission ID, or empty string if no job was available.
func (q *Queue) Dequeue(ctx context.Context) (string, error) {
	result, err := q.client.BRPop(ctx, 1*time.Second, "submission_jobs").Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("queue: failed to dequeue: %w", err)
	}

	var payload JobPayload
	if err := json.Unmarshal([]byte(result[1]), &payload); err != nil {
		return "", fmt.Errorf("queue: failed to unmarshal payload: %w", err)
	}

	return payload.SubmissionID, nil
}

// Close closes the Redis connection.
func (q *Queue) Close() error {
	return q.client.Close()
}
