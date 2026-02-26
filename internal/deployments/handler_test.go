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
		ARKPublicHost:   "http://ark-test.local",
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

	store.Create(storage.Product{
		ID:   "task-manager",
		Name: "Task Manager",
		DeployJobs: map[string]string{
			"PROD": "deploy-task-manager-prod",
			"DEV":  "deploy-task-manager-dev",
		},
		DeleteJob: "delete-task-manager",
	})

	payload := CreateDeploymentRequest{
		ProductID:   "task-manager",
		Environment: "PROD",
		TargetHost:  "server-01",
		SSHUser:     "testuser",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/deployments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["detail"], "no deploy job configured")
}

func TestCreateDeployment_WithJobName(t *testing.T) {
	router, handler, _ := setupDeploymentTest()
	router.POST("/deployments", handler.Create)

	payload := CreateDeploymentRequest{
		JobName:    "deploy-custom-job",
		AppName:    "custom-app",
		Version:    "v2.0.0",
		TargetHost: "server-02",
		SSHUser:    "testuser",
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

func TestCreateDeployment_ProductNotFound(t *testing.T) {
	router, handler, _ := setupDeploymentTest()
	router.POST("/deployments", handler.Create)

	payload := CreateDeploymentRequest{
		ProductID:   "producto-inexistente",
		Environment: "prod",
		AppName:     "app",
		Version:     "v1.0.0",
		TargetHost:  "server-01",
		SSHUser:     "testuser",
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

	store.Create(storage.Product{
		ID:   "task-manager",
		Name: "Task Manager",
		DeployJobs: map[string]string{
			"PROD": "deploy-task-manager-prod",
			"DEV":  "deploy-task-manager-dev",
		},
		DeleteJob: "delete-task-manager",
	})

	payload := CreateDeploymentRequest{
		ProductID:   "task-manager",
		Environment: "STAGING",
		TargetHost:  "server-01",
		SSHUser:     "testuser",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/deployments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["detail"], "environment must be prod, dev, or test")
}

func TestCreateDeployment_InvalidRequest(t *testing.T) {
	router, handler, _ := setupDeploymentTest()
	router.POST("/deployments", handler.Create)

	tests := []struct {
		name    string
		payload interface{}
	}{
		{
			name: "sin product_id ni app_name",
			payload: map[string]interface{}{
				"environment": "PROD",
				"target_host": "server-01",
				"ssh_user":    "testuser",
			},
		},
		{
			name: "sin environment ni version",
			payload: map[string]interface{}{
				"product_id":  "test",
				"target_host": "server-01",
				"ssh_user":    "testuser",
			},
		},
		{
			name: "sin target_host",
			payload: map[string]interface{}{
				"product_id":  "test",
				"environment": "PROD",
				"ssh_user":    "testuser",
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

				if tt.name == "sin environment ni version" {
					assert.Equal(t, http.StatusNotFound, w.Code)
					return
				}
				assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestCreateDeployment_NewFormat(t *testing.T) {
	router, handler, store := setupDeploymentTest()
	router.POST("/deployments", handler.Create)

	store.Create(storage.Product{
		ID:   "ARKCHANNEL",
		Name: "ARK Channel",
		DeployJobs: map[string]string{
			"PROD": "deploy-arkchannel-prod",
			"DEV":  "deploy-arkchannel-dev",
		},
		DeleteJob: "delete-arkchannel",
	})

	payload := map[string]interface{}{
		"product_id":  "ARKCHANNEL",
		"environment": "PROD",
		"target_host": "100.103.96.26",
		"ssh_user":    "testuser",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/deployments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["detail"], "no deploy job configured")
}

func TestCreateDeployment_LegacyFormat(t *testing.T) {
	router, handler, store := setupDeploymentTest()
	router.POST("/deployments", handler.Create)

	store.Create(storage.Product{
		ID:   "vault",
		Name: "Vault",
		DeployJobs: map[string]string{
			"PROD": "deploy-vault-prod",
			"DEV":  "deploy-vault-dev",
		},
		DeleteJob: "delete-vault",
	})

	payload := map[string]interface{}{
		"app_name":    "vault",
		"version":     "prod",
		"target_host": "100.103.96.26",
		"ssh_user":    "testuser",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/deployments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["detail"], "no deploy job configured")
}
