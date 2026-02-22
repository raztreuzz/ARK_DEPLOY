package products

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"ark_deploy/internal/storage"
)

func setupTest() (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	store := storage.NewProductStore()
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
					"PROD": "deploy-task-manager-prod",
					"DEV":  "deploy-task-manager-dev",
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
					"PROD": "deploy-task-manager-prod",
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
					"PROD": "deploy-task-manager-prod",
				},
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
			}
		})
	}
}

func TestListProducts(t *testing.T) {
	router, handler := setupTest()
	router.GET("/products", handler.List)

	// Crear algunos productos primero
	handler.store.Create(storage.Product{
		ID:   "prod-1",
		Name: "Product 1",
		DeployJobs: map[string]string{"PROD": "job-1"},
		DeleteJob:  "delete-job-1",
	})
	handler.store.Create(storage.Product{
		ID:   "prod-2",
		Name: "Product 2",
		DeployJobs: map[string]string{"PROD": "job-2"},
		DeleteJob:  "delete-job-2",
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

	// Crear un producto
	handler.store.Create(storage.Product{
		ID:          "task-manager",
		Name:        "Task Manager",
		Description: "Sistema de tareas",
		DeployJobs: map[string]string{
			"PROD": "deploy-task-manager-prod",
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

	// Crear un producto
	handler.store.Create(storage.Product{
		ID:   "task-manager",
		Name: "Task Manager Old",
		DeployJobs: map[string]string{"PROD": "old-job"},
		DeleteJob:  "delete-task-manager",
	})

	payload := UpdateProductRequest{
		Name:        "Task Manager Updated",
		Description: "Nueva descripción",
		DeployJobs: map[string]string{
			"PROD": "new-job-prod",
			"DEV":  "new-job-dev",
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
	assert.Equal(t, "new-job-prod", response.DeployJobs["PROD"])
}

func TestDeleteProduct(t *testing.T) {
	router, handler := setupTest()
	router.DELETE("/products/:id", handler.Delete)

	// Crear un producto
	handler.store.Create(storage.Product{
		ID:   "task-manager",
		Name: "Task Manager",
		DeployJobs: map[string]string{"PROD": "job"},
		DeleteJob:  "delete-task-manager",
	})

	req := httptest.NewRequest(http.MethodDelete, "/products/task-manager", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verificar que fue eliminado
	_, err := handler.store.GetByID("task-manager")
	assert.Error(t, err)
}
