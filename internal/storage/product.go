package storage

import (
	"errors"
	"sync"
)

type Product struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Jobs        map[string]string `json:"jobs"` // environment -> job_name
}

type ProductStore struct {
	mu       sync.RWMutex
	products map[string]Product
}

func NewProductStore() *ProductStore {
	return &ProductStore{
		products: make(map[string]Product),
	}
}

func (s *ProductStore) Create(p Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[p.ID]; exists {
		return errors.New("product already exists")
	}

	s.products[p.ID] = p
	return nil
}

func (s *ProductStore) GetAll() []Product {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Product, 0, len(s.products))
	for _, p := range s.products {
		result = append(result, p)
	}
	return result
}

func (s *ProductStore) GetByID(id string) (Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, exists := s.products[id]
	if !exists {
		return Product{}, errors.New("product not found")
	}
	return p, nil
}

func (s *ProductStore) Update(id string, p Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		return errors.New("product not found")
	}

	p.ID = id // Asegurar que el ID no cambie
	s.products[id] = p
	return nil
}

func (s *ProductStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		return errors.New("product not found")
	}

	delete(s.products, id)
	return nil
}
