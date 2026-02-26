package storage

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"

	arkredis "ark_deploy/internal/redis"
)

type SSHUserStore struct{}

func NewSSHUserStore() *SSHUserStore {
	return &SSHUserStore{}
}

func sshUserKey(host string) string {
	return "sshuser:" + strings.TrimSpace(host)
}

func (s *SSHUserStore) Set(host, user string) error {
	ctx := context.Background()
	h := strings.TrimSpace(host)
	u := strings.TrimSpace(user)
	if h == "" {
		return nil
	}
	return arkredis.Client.Set(ctx, sshUserKey(h), u, 0).Err()
}

func (s *SSHUserStore) Get(host string) (string, bool, error) {
	ctx := context.Background()
	h := strings.TrimSpace(host)
	if h == "" {
		return "", false, nil
	}
	v, err := arkredis.Client.Get(ctx, sshUserKey(h)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", false, nil
		}
		return "", false, err
	}
	return strings.TrimSpace(v), true, nil
}

func (s *SSHUserStore) Delete(host string) error {
	ctx := context.Background()
	h := strings.TrimSpace(host)
	if h == "" {
		return nil
	}
	return arkredis.Client.Del(ctx, sshUserKey(h)).Err()
}

func (s *SSHUserStore) List() (map[string]string, error) {
	ctx := context.Background()
	out := make(map[string]string)
	var cursor uint64

	for {
		keys, next, err := arkredis.Client.Scan(ctx, cursor, "sshuser:*", 100).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			host := strings.TrimPrefix(key, "sshuser:")
			v, err := arkredis.Client.Get(ctx, key).Result()
			if err == nil {
				out[host] = strings.TrimSpace(v)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}

	return out, nil
}

