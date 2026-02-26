package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"ark_deploy/internal/redis"
)

type Product struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	ReleaseTag  string            `json:"release_tag,omitempty"`
	DeployJobs  map[string]string `json:"deploy_jobs"`
	DeleteJob   string            `json:"delete_job"`
	WebService  string            `json:"web_service,omitempty"`
	WebPort     int               `json:"web_port,omitempty"`
	Jobs        map[string]string `json:"jobs,omitempty"`
}

type ProductStore struct{}

func NewProductStore() *ProductStore {
	return &ProductStore{}
}

func productKey(id string) string {
	return fmt.Sprintf("product:%s", id)
}

func (s *ProductStore) Create(p Product) error {
	ctx := context.Background()
	key := productKey(p.ID)

	exists, err := redis.Client.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return errors.New("product already exists")
	}

	p = normalizeProduct(p)
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return redis.Client.Set(ctx, key, data, 0).Err()
}

func (s *ProductStore) GetAll() []Product {
	ctx := context.Background()
	keys, err := redis.Client.Keys(ctx, "product:*").Result()
	if err != nil {
		return []Product{}
	}

	result := make([]Product, 0, len(keys))
	for _, key := range keys {
		data, err := redis.Client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var p Product
		if err := json.Unmarshal([]byte(data), &p); err != nil {
			continue
		}
		p = normalizeProduct(p)
		result = append(result, p)
	}

	return result
}

func (s *ProductStore) GetByID(id string) (Product, error) {
	ctx := context.Background()
	key := productKey(id)

	data, err := redis.Client.Get(ctx, key).Result()
	if err != nil {
		return Product{}, errors.New("product not found")
	}

	var p Product
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return Product{}, err
	}

	return normalizeProduct(p), nil
}

func (s *ProductStore) Update(id string, p Product) error {
	ctx := context.Background()
	key := productKey(id)

	exists, err := redis.Client.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return errors.New("product not found")
	}

	p.ID = id
	p = normalizeProduct(p)

	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return redis.Client.Set(ctx, key, data, 0).Err()
}

func (s *ProductStore) Delete(id string) error {
	ctx := context.Background()
	key := productKey(id)

	exists, err := redis.Client.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return errors.New("product not found")
	}

	return redis.Client.Del(ctx, key).Err()
}

func normalizeProduct(p Product) Product {
	if len(p.DeployJobs) == 0 && len(p.Jobs) > 0 {
		p.DeployJobs = p.Jobs
	}

	if p.DeployJobs != nil {
		n := make(map[string]string, len(p.DeployJobs))
		for k, v := range p.DeployJobs {
			n[strings.TrimSpace(strings.ToLower(k))] = strings.TrimSpace(v)
		}
		p.DeployJobs = n
	}

	p.ID = strings.TrimSpace(p.ID)
	p.Name = strings.TrimSpace(p.Name)
	p.Description = strings.TrimSpace(p.Description)
	p.ReleaseTag = strings.TrimSpace(p.ReleaseTag)
	p.DeleteJob = strings.TrimSpace(p.DeleteJob)
	p.WebService = strings.TrimSpace(p.WebService)

	if p.WebService == "" {
		p.WebService = "web"
	}
	if p.WebPort == 0 {
		p.WebPort = 80
	}

	return p
}
