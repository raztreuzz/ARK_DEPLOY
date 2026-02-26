package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ark_deploy/internal/redis"
)

type Instance struct {
	ID          string            `json:"id"`
	ProductID   string            `json:"product_id"`
	DeviceID    string            `json:"device_id"`
	Environment string            `json:"environment"`
	Status      string            `json:"status"`
	URL         string            `json:"url"`
	Builds      map[string]string `json:"builds"`
	CreatedAt   time.Time         `json:"created_at"`
}

type InstanceStore struct{}

func NewInstanceStore() *InstanceStore {
	return &InstanceStore{}
}

func instanceKey(id string) string {
	return fmt.Sprintf("instance:%s", id)
}

func (s *InstanceStore) Create(i Instance) error {
	ctx := context.Background()
	key := instanceKey(i.ID)

	exists, err := redis.Client.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return errors.New("instance already exists")
	}

	if i.CreatedAt.IsZero() {
		i.CreatedAt = time.Now().UTC()
	}

	data, err := json.Marshal(i)
	if err != nil {
		return err
	}

	return redis.Client.Set(ctx, key, data, 0).Err()
}

func (s *InstanceStore) GetAll() []Instance {
	ctx := context.Background()

	var cursor uint64
	result := make([]Instance, 0)

	for {
		keys, next, err := redis.Client.Scan(ctx, cursor, "instance:*", 100).Result()
		if err != nil {
			return []Instance{}
		}

		for _, key := range keys {
			data, err := redis.Client.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			var i Instance
			if err := json.Unmarshal([]byte(data), &i); err != nil {
				continue
			}

			result = append(result, i)
		}

		cursor = next
		if cursor == 0 {
			break
		}
	}

	return result
}

func (s *InstanceStore) GetByID(id string) (Instance, error) {
	ctx := context.Background()
	key := instanceKey(id)

	data, err := redis.Client.Get(ctx, key).Result()
	if err != nil {
		return Instance{}, errors.New("instance not found")
	}

	var i Instance
	if err := json.Unmarshal([]byte(data), &i); err != nil {
		return Instance{}, err
	}

	return i, nil
}

func (s *InstanceStore) UpdateStatus(id string, status string) error {
	ctx := context.Background()
	key := instanceKey(id)

	data, err := redis.Client.Get(ctx, key).Result()
	if err != nil {
		return errors.New("instance not found")
	}

	var instance Instance
	if err := json.Unmarshal([]byte(data), &instance); err != nil {
		return err
	}

	instance.Status = status

	newData, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	return redis.Client.Set(ctx, key, newData, 0).Err()
}

func (s *InstanceStore) Delete(id string) error {
	ctx := context.Background()
	key := instanceKey(id)

	exists, err := redis.Client.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return errors.New("instance not found")
	}

	return redis.Client.Del(ctx, key).Err()
}