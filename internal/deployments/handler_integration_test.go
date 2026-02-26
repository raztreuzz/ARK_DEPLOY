package deployments

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"ark_deploy/internal/config"
	"ark_deploy/internal/storage"
)

type MockInstanceStore struct {
	mu        sync.RWMutex
	instances map[string]storage.Instance
}

func NewMockInstanceStore() *MockInstanceStore {
	return &MockInstanceStore{
		instances: make(map[string]storage.Instance),
	}
}

func (s *MockInstanceStore) Create(i storage.Instance) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.instances[i.ID]; exists {
		return errors.New("instance already exists")
	}

	s.instances[i.ID] = i
	return nil
}

func (s *MockInstanceStore) GetByID(id string) (storage.Instance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	i, exists := s.instances[id]
	if !exists {
		return storage.Instance{}, errors.New("instance not found")
	}
	return i, nil
}

func (s *MockInstanceStore) GetAll() []storage.Instance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]storage.Instance, 0, len(s.instances))
	for _, i := range s.instances {
		result = append(result, i)
	}
	return result
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

type MockProductStore struct {
	mu       sync.RWMutex
	products map[string]storage.Product
}

func NewMockProductStore() *MockProductStore {
	return &MockProductStore{
		products: make(map[string]storage.Product),
	}
}

func (s *MockProductStore) Create(p storage.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[p.ID]; exists {
		return errors.New("product already exists")
	}

	s.products[p.ID] = p
	return nil
}

func (s *MockProductStore) GetByID(id string) (storage.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, exists := s.products[id]
	if !exists {
		return storage.Product{}, errors.New("product not found")
	}
	return p, nil
}

func setupTestRouter(productStore ProductStore, instanceStore InstanceStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := config.Config{
		JenkinsBaseURL:  "http://jenkins-test.local",
		JenkinsUser:     "test-user",
		JenkinsAPIToken: "test-token",
		ARKPublicHost:   "http://ark-test.local",
	}

	h := NewHandler(cfg, productStore, instanceStore)

	r.GET("/deployments", h.List)
	r.DELETE("/deployments/:id", h.Delete)

	return r
}

func TestDeploymentsList_Empty(t *testing.T) {
	productStore := NewMockProductStore()
	instanceStore := NewMockInstanceStore()
	router := setupTestRouter(productStore, instanceStore)

	req, _ := http.NewRequest("GET", "/deployments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["total"].(float64) != 0 {
		t.Errorf("Expected total 0, got %v", response["total"])
	}
}

func TestDeploymentsList_WithInstances(t *testing.T) {
	productStore := NewMockProductStore()
	instanceStore := NewMockInstanceStore()

	instance := storage.Instance{
		ID:          "test-instance",
		ProductID:   "test-product",
		DeviceID:    "192.168.1.100",
		Environment: "prod",
		Status:      "running",
		URL:         "http://192.168.1.100:3000",
		Builds:      map[string]string{"backend": "123"},
		CreatedAt:   time.Now(),
	}
	instanceStore.Create(instance)

	router := setupTestRouter(productStore, instanceStore)

	req, _ := http.NewRequest("GET", "/deployments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["total"].(float64) != 1 {
		t.Errorf("Expected total 1, got %v", response["total"])
	}
}

func TestDeploymentsDelete_Success(t *testing.T) {
	productStore := NewMockProductStore()
	instanceStore := NewMockInstanceStore()

	instance := storage.Instance{
		ID:        "delete-test",
		ProductID: "test-product",
		DeviceID:  "192.168.1.100",
		Status:    "running",
		URL:       "http://192.168.1.100:3000",
		CreatedAt: time.Now(),
	}
	instanceStore.Create(instance)

	router := setupTestRouter(productStore, instanceStore)

	req, _ := http.NewRequest("DELETE", "/deployments/delete-test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "instance deleted" {
		t.Errorf("Expected 'instance deleted', got %v", response["message"])
	}

	_, err := instanceStore.GetByID("delete-test")
	if err == nil {
		t.Errorf("Instance should be deleted")
	}
}

func TestDeploymentsDelete_NotFound(t *testing.T) {
	productStore := NewMockProductStore()
	instanceStore := NewMockInstanceStore()
	router := setupTestRouter(productStore, instanceStore)

	req, _ := http.NewRequest("DELETE", "/deployments/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}