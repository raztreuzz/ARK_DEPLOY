package deployments

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"ark_deploy/internal/config"
	"ark_deploy/internal/storage"
)

func setupDeploymentTest() (*gin.Engine, *Handler, *MockProductStore) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		JenkinsBaseURL:  "http://jenkins-test.local",
		JenkinsUser:     "test-user",
		JenkinsAPIToken: "test-token",
		JenkinsJob:      "default-job",
	}

	productStore := NewMockProductStore()
	instanceStore := NewMockInstanceStore()
	handler := NewHandler(cfg, productStore, instanceStore)
	router := gin.New()

	return router, handler, productStore
}

func TestCreateDeployment_WithProduct(t *testing.T) {
	router, handler, store := setupDeploymentTest()
	router.POST("/deployments", handler.Create)

	// Crear un producto de prueba
	store.Create(storage.Product{
		ID:   "task-manager",
		Name: "Task Manager",
		DeployJobs: map[string]string{
			"prod": "deploy-task-manager-prod",
			"dev":  "deploy-task-manager-dev",
		},
		DeleteJob: "delete-task-manager",
	})

	// Nota: Este test fallará sin un Jenkins real o mock,
	// pero valida la estructura del request
	payload := CreateDeploymentRequest{
		ProductID:   "task-manager",
		Environment: "prod",
		AppName:     "task-manager",
		Version:     "v1.0.0",
		TargetHost:  "server-01",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/deployments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Como no hay Jenkins real, esperamos un error 502 (BadGateway)
	// pero esto valida que el producto se encontró correctamente
	assert.Contains(t, []int{http.StatusBadGateway, http.StatusAccepted}, w.Code)
}

func TestCreateDeployment_WithJobName(t *testing.T) {
	router, handler, _ := setupDeploymentTest()
	router.POST("/deployments", handler.Create)

	// Test usando job_name directamente (retrocompatibilidad)
	payload := CreateDeploymentRequest{
		JobName:    "deploy-custom-job",
		AppName:    "custom-app",
		Version:    "v2.0.0",
		TargetHost: "server-02",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/deployments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Sin Jenkins real, esperamos error de gateway
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestCreateDeployment_ProductNotFound(t *testing.T) {
	router, handler, _ := setupDeploymentTest()
	router.POST("/deployments", handler.Create)

	payload := CreateDeploymentRequest{
		ProductID:   "producto-inexistente",
		Environment: "prod",
		AppName:     "app",
		Version:     "v1.0.0",
		TargetHost:  "server-01",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/deployments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["detail"], "product not found")
}

func TestCreateDeployment_EnvironmentNotConfigured(t *testing.T) {
	router, handler, store := setupDeploymentTest()
	router.POST("/deployments", handler.Create)

	// Crear producto sin ambiente "staging"
	store.Create(storage.Product{
		ID:   "task-manager",
		Name: "Task Manager",
		DeployJobs: map[string]string{
			"prod": "deploy-task-manager-prod",
			"dev":  "deploy-task-manager-dev",
		},
		DeleteJob: "delete-task-manager",
	})

	payload := CreateDeploymentRequest{
		ProductID:   "task-manager",
		Environment: "staging", // No existe en el producto
		AppName:     "app",
		Version:     "v1.0.0",
		TargetHost:  "server-01",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/deployments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["detail"], "no job configured for environment")
}

func TestCreateDeployment_InvalidRequest(t *testing.T) {
	router, handler, _ := setupDeploymentTest()
	router.POST("/deployments", handler.Create)

	tests := []struct {
		name    string
		payload interface{}
	}{
		{
			name: "sin app_name",
			payload: map[string]interface{}{
				"version":     "v1.0.0",
				"target_host": "server-01",
			},
		},
		{
			name: "sin version",
			payload: map[string]interface{}{
				"app_name":    "app",
				"target_host": "server-01",
			},
		},
		{
			name: "sin target_host",
			payload: map[string]interface{}{
				"app_name": "app",
				"version":  "v1.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/deployments", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
