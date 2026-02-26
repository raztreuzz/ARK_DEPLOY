package storage

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type MockInstanceStore struct {
	mu        sync.RWMutex
	instances map[string]Instance
}

func NewMockInstanceStore() *MockInstanceStore {
	return &MockInstanceStore{
		instances: make(map[string]Instance),
	}
}

func (s *MockInstanceStore) Create(i Instance) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.instances[i.ID]; exists {
		return errors.New("instance already exists")
	}

	s.instances[i.ID] = i
	return nil
}

func (s *MockInstanceStore) GetByID(id string) (Instance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	i, exists := s.instances[id]
	if !exists {
		return Instance{}, errors.New("instance not found")
	}
	return i, nil
}

func (s *MockInstanceStore) GetAll() []Instance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Instance, 0, len(s.instances))
	for _, i := range s.instances {
		result = append(result, i)
	}
	return result
}

func (s *MockInstanceStore) UpdateStatus(id string, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	instance, exists := s.instances[id]
	if !exists {
		return errors.New("instance not found")
	}

	instance.Status = status
	s.instances[id] = instance
	return nil
}

func (s *MockInstanceStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.instances[id]; !exists {
		return errors.New("instance not found")
	}

	delete(s.instances, id)
	return nil
}

func TestInstanceStore_CRUD(t *testing.T) {
	store := NewMockInstanceStore()

	instance := Instance{
		ID:          "test-instance-1",
		ProductID:   "ecommerce",
		DeviceID:    "192.168.1.100",
		Environment: "prod",
		Status:      "running",
		URL:         "http://192.168.1.100:3000",
		Builds:      map[string]string{"backend": "123", "frontend": "124"},
		CreatedAt:   time.Now(),
	}

	if err := store.Create(instance); err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}

	retrieved, err := store.GetByID("test-instance-1")
	if err != nil {
		t.Fatalf("Failed to get instance: %v", err)
	}

	if retrieved.ID != instance.ID || retrieved.ProductID != instance.ProductID {
		t.Errorf("Retrieved instance data mismatch")
	}

	all := store.GetAll()
	if len(all) < 1 {
		t.Errorf("Expected at least 1 instance, got %d", len(all))
	}

	if err := store.UpdateStatus("test-instance-1", "stopped"); err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	updated, _ := store.GetByID("test-instance-1")
	if updated.Status != "stopped" {
		t.Errorf("Status not updated, got %s", updated.Status)
	}

	if err := store.Delete("test-instance-1"); err != nil {
		t.Fatalf("Failed to delete instance: %v", err)
	}

	_, err = store.GetByID("test-instance-1")
	if err == nil {
		t.Errorf("Instance should be deleted")
	}
}

func TestInstanceStore_DuplicateCreate(t *testing.T) {
	store := NewMockInstanceStore()

	instance := Instance{
		ID:        "dup-test",
		ProductID: "test",
		Status:    "running",
		CreatedAt: time.Now(),
	}

	store.Create(instance)

	if err := store.Create(instance); err == nil {
		t.Errorf("Expected error for duplicate instance, got nil")
	}
}
