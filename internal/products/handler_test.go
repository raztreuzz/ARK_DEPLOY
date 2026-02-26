package products

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"ark_deploy/internal/storage"
)

type mockProductStore struct {
	mu       sync.RWMutex
	products map[string]storage.Product
}

func newMockProductStore() *mockProductStore {
	return &mockProductStore{products: make(map[string]storage.Product)}
}

func (s *mockProductStore) Create(p storage.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[p.ID]; exists {
		return assert.AnError
	}
	s.products[p.ID] = p
	return nil
}

func (s *mockProductStore) GetAll() []storage.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]storage.Product, 0, len(s.products))
	for _, p := range s.products {
		result = append(result, p)
	}
	return result
}

func (s *mockProductStore) GetByID(id string) (storage.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, exists := s.products[id]
	if !exists {
		return storage.Product{}, assert.AnError
	}
	return p, nil
}

func (s *mockProductStore) Update(id string, p storage.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		return assert.AnError
	}
	p.ID = id
	s.products[id] = p
	return nil
}

func (s *mockProductStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		return assert.AnError
	}
	delete(s.products, id)
	return nil
}

func setupTest() (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	store := newMockProductStore()
	handler := NewHandler(store)
	router := gin.New()
	return router, handler
}

func TestCreateProduct(t *testing.T) {
	router, handler := setupTest()
	router.POST("/products", handler.Create)

	tests := []struct {
		name       string
		payload    CreateProductRequest
		wantStatus int
		wantError  bool
	}{
		{
			name: "crear producto válido",
			payload: CreateProductRequest{
				ID:          "task-manager",
				Name:        "Task Manager",
				Description: "Sistema de tareas",
				DeployJobs: map[string]string{
					"prod": "deploy-task-manager-prod",
					"dev":  "deploy-task-manager-dev",
					"test": "deploy-task-manager-test",
				},
				DeleteJob: "delete-task-manager",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name: "crear producto sin ID",
			payload: CreateProductRequest{
				Name:        "Task Manager",
				Description: "Sistema de tareas",
				DeployJobs: map[string]string{
					"prod": "deploy-task-manager-prod",
					"dev":  "deploy-task-manager-dev",
					"test": "deploy-task-manager-test",
				},
				DeleteJob: "delete-task-manager",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "crear producto sin jobs",
			payload: CreateProductRequest{
				ID:          "task-manager",
				Name:        "Task Manager",
				Description: "Sistema de tareas",
				DeleteJob:   "delete-task-manager",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "crear producto sin delete job",
			payload: CreateProductRequest{
				ID:          "task-manager",
				Name:        "Task Manager",
				Description: "Sistema de tareas",
				DeployJobs: map[string]string{
					"prod": "deploy-task-manager-prod",
					"dev":  "deploy-task-manager-dev",
					"test": "deploy-task-manager-test",
				},
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "crear producto sin envs requeridos",
			payload: CreateProductRequest{
				ID:          "task-manager-2",
				Name:        "Task Manager 2",
				Description: "Sistema de tareas",
				DeployJobs: map[string]string{
					"prod": "deploy-task-manager-prod",
				},
				DeleteJob: "delete-task-manager",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantError {
				var response storage.Product
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.payload.ID, response.ID)
				assert.Equal(t, tt.payload.Name, response.Name)
				assert.Equal(t, tt.payload.DeleteJob, response.DeleteJob)
				assert.Equal(t, "deploy-task-manager-prod", response.DeployJobs["prod"])
				assert.Equal(t, "deploy-task-manager-dev", response.DeployJobs["dev"])
				assert.Equal(t, "deploy-task-manager-test", response.DeployJobs["test"])
			}
		})
	}
}

func TestListProducts(t *testing.T) {
	router, handler := setupTest()
	router.GET("/products", handler.List)

	handler.store.Create(storage.Product{
		ID:   "prod-1",
		Name: "Product 1",
		DeployJobs: map[string]string{
			"prod": "job-1-prod",
			"dev":  "job-1-dev",
			"test": "job-1-test",
		},
		DeleteJob: "delete-job-1",
	})
	handler.store.Create(storage.Product{
		ID:   "prod-2",
		Name: "Product 2",
		DeployJobs: map[string]string{
			"prod": "job-2-prod",
			"dev":  "job-2-dev",
			"test": "job-2-test",
		},
		DeleteJob: "delete-job-2",
	})

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["total"])
}

func TestGetProduct(t *testing.T) {
	router, handler := setupTest()
	router.GET("/products/:id", handler.Get)

	handler.store.Create(storage.Product{
		ID:          "task-manager",
		Name:        "Task Manager",
		Description: "Sistema de tareas",
		DeployJobs: map[string]string{
			"prod": "deploy-task-manager-prod",
			"dev":  "deploy-task-manager-dev",
			"test": "deploy-task-manager-test",
		},
		DeleteJob: "delete-task-manager",
	})

	tests := []struct {
		name       string
		productID  string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "obtener producto existente",
			productID:  "task-manager",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "obtener producto inexistente",
			productID:  "no-existe",
			wantStatus: http.StatusNotFound,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/products/"+tt.productID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantError {
				var response storage.Product
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.productID, response.ID)
			}
		})
	}
}

func TestUpdateProduct(t *testing.T) {
	router, handler := setupTest()
	router.PUT("/products/:id", handler.Update)

	handler.store.Create(storage.Product{
		ID:   "task-manager",
		Name: "Task Manager Old",
		DeployJobs: map[string]string{
			"prod": "old-job-prod",
			"dev":  "old-job-dev",
			"test": "old-job-test",
		},
		DeleteJob: "delete-task-manager",
	})

	payload := UpdateProductRequest{
		Name:        "Task Manager Updated",
		Description: "Nueva descripción",
		DeployJobs: map[string]string{
			"prod": "new-job-prod",
			"dev":  "new-job-dev",
			"test": "new-job-test",
		},
		DeleteJob: "delete-task-manager",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/products/task-manager", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response storage.Product
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Task Manager Updated", response.Name)
	assert.Equal(t, "new-job-prod", response.DeployJobs["prod"])
}

func TestDeleteProduct(t *testing.T) {
	router, handler := setupTest()
	router.DELETE("/products/:id", handler.Delete)

	handler.store.Create(storage.Product{
		ID:   "task-manager",
		Name: "Task Manager",
		DeployJobs: map[string]string{
			"prod": "job-prod",
			"dev":  "job-dev",
			"test": "job-test",
		},
		DeleteJob: "delete-task-manager",
	})

	req := httptest.NewRequest(http.MethodDelete, "/products/task-manager", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	_, err := handler.store.GetByID("task-manager")
	assert.Error(t, err)
}