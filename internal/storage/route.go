package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	arkredis "ark_deploy/internal/redis"
)

type Route struct {
	InstanceID string    `json:"instance_id"`
	TargetHost string    `json:"target_host"`
	TargetPort int       `json:"target_port"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type RouteStore struct{}

func NewRouteStore() *RouteStore {
	return &RouteStore{}
}

func routeKey(instanceID string) string {
	return fmt.Sprintf("route:%s", instanceID)
}

func (s *RouteStore) PutRoute(instanceID string, host string, port int) error {
	ctx := context.Background()

	record := Route{
		InstanceID: strings.TrimSpace(instanceID),
		TargetHost: strings.TrimSpace(host),
		TargetPort: port,
		UpdatedAt:  time.Now(),
	}

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return arkredis.Client.Set(ctx, routeKey(record.InstanceID), data, 0).Err()
}

func (s *RouteStore) GetRoute(instanceID string) (host string, port int, ok bool, err error) {
	ctx := context.Background()

	data, err := arkredis.Client.Get(ctx, routeKey(strings.TrimSpace(instanceID))).Result()
	if err != nil {
		if err == redis.Nil {
			return "", 0, false, nil
		}
		return "", 0, false, err
	}

	var record Route
	if err := json.Unmarshal([]byte(data), &record); err != nil {
		return "", 0, false, err
	}

	return record.TargetHost, record.TargetPort, true, nil
}

func (s *RouteStore) DeleteRoute(instanceID string) error {
	ctx := context.Background()
	return arkredis.Client.Del(ctx, routeKey(strings.TrimSpace(instanceID))).Err()
}
