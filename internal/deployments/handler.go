package deployments

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ark_deploy/internal/config"
	"ark_deploy/internal/jenkins"
	"ark_deploy/internal/storage"
)

type ProductStore interface {
	GetByID(id string) (storage.Product, error)
}

type InstanceStore interface {
	Create(i storage.Instance) error
	GetAll() []storage.Instance
	GetByID(id string) (storage.Instance, error)
	Delete(id string) error
}

type Handler struct {
	cfg           config.Config
	productStore  ProductStore
	instanceStore InstanceStore
}

func NewHandler(cfg config.Config, productStore ProductStore, instanceStore InstanceStore) *Handler {
	return &Handler{
		cfg:           cfg,
		productStore:  productStore,
		instanceStore: instanceStore,
	}
}

type CreateDeploymentRequest struct {
	ProductID    string `json:"product_id"`
	Environment  string `json:"environment"`
	JobName      string `json:"job_name"` // Opcional, para retrocompatibilidad
	AppName      string `json:"app_name" binding:"required"`
	Version      string `json:"version" binding:"required"`
	TargetHost   string `json:"target_host" binding:"required"`
	SimulateFail bool   `json:"simulate_fail"`
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	client := jenkins.NewClient(h.cfg.JenkinsBaseURL, h.cfg.JenkinsUser, h.cfg.JenkinsAPIToken)

	// Determinar el job_name
	var jobName string

	// Opción 1: Usar product_id + environment
	if req.ProductID != "" && req.Environment != "" {
		product, err := h.productStore.GetByID(req.ProductID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": "product not found: " + err.Error()})
			return
		}

		jobName = product.DeployJobs[req.Environment]
		if jobName == "" && len(product.DeployJobs) == 0 {
			jobName = product.Jobs[req.Environment]
		}
		if jobName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "no job configured for environment: " + req.Environment})
			return
		}
	} else if req.JobName != "" {
		// Opción 2: Usar job_name directamente (retrocompatibilidad)
		jobName = req.JobName
	} else {
		// Opción 3: Usar el job por defecto de la configuración
		jobName = h.cfg.JenkinsJob
	}

	queueURL, err := client.TriggerJobWithParams(jobName, map[string]string{
		"APP_NAME":      req.AppName,
		"VERSION":       req.Version,
		"TARGET_HOST":   req.TargetHost,
		"SIMULATE_FAIL": boolToString(req.SimulateFail),
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	buildNumber, resolved := h.tryResolveBuildNumber(client, jobName, queueURL)

	instanceID := uuid.New().String()
	instance := storage.Instance{
		ID:          instanceID,
		ProductID:   req.ProductID,
		DeviceID:    req.TargetHost,
		Environment: req.Environment,
		Status:      "provisioning",
		URL:         fmt.Sprintf("http://%s:3000", req.TargetHost),
		Builds:      map[string]string{jobName: strconv.Itoa(buildNumber)},
		CreatedAt:   time.Now(),
	}

	if err := h.instanceStore.Create(instance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to save instance: " + err.Error()})
		return
	}

	if resolved {
		c.JSON(http.StatusAccepted, gin.H{
			"instance_id":  instanceID,
			"url":          instance.URL,
			"status":       "queued_resolved",
			"job_name":     jobName,
			"queue_url":    queueURL,
			"build_number": buildNumber,
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"instance_id": instanceID,
		"url":         instance.URL,
		"status":      "queued",
		"job_name":    jobName,
		"queue_url":   queueURL,
	})
}

func (h *Handler) List(c *gin.Context) {
	instances := h.instanceStore.GetAll()

	// Ordenar por CreatedAt descendente (más nuevas primero)
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].CreatedAt.After(instances[j].CreatedAt)
	})

	c.JSON(http.StatusOK, gin.H{
		"total":     len(instances),
		"instances": instances,
	})
}

func (h *Handler) Delete(c *gin.Context) {
	instanceID := c.Param("id")

	instance, err := h.instanceStore.GetByID(instanceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "instance not found"})
		return
	}

	// TODO: Agregar limpieza de contenedor Docker en el nodo (deviceID)
	// Por ahora solo eliminamos del store

	if err := h.instanceStore.Delete(instanceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to delete instance: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "instance deleted",
		"instance_id": instanceID,
		"device_id":   instance.DeviceID,
	})
}

func (h *Handler) GetLogs(c *gin.Context) {
	instanceID := c.Param("id")

	instance, err := h.instanceStore.GetByID(instanceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "instance not found"})
		return
	}

	if len(instance.Builds) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"instance_id": instanceID,
			"logs":        map[string]string{},
		})
		return
	}

	client := jenkins.NewClient(h.cfg.JenkinsBaseURL, h.cfg.JenkinsUser, h.cfg.JenkinsAPIToken)
	logsMap := make(map[string]string)

	for jobName, buildNumber := range instance.Builds {
		log, err := client.GetBuildLog(jobName, buildNumber)
		if err != nil {
			logsMap[jobName] = fmt.Sprintf("Error fetching log: %v", err)
		} else {
			logsMap[jobName] = log
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"instance_id": instanceID,
		"device_id":   instance.DeviceID,
		"product_id":  instance.ProductID,
		"status":      instance.Status,
		"logs":        logsMap,
	})
}

func (h *Handler) tryResolveBuildNumber(client *jenkins.Client, jobName string, queueURL string) (int, bool) {
	queueID, ok := extractQueueID(queueURL)
	if !ok {
		return 0, false
	}

	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		buildNumber, cancelled, err := client.ReadQueueItem(queueURL)
		if err == nil {
			if cancelled {
				return 0, false
			}
			if buildNumber > 0 {
				return buildNumber, true
			}
		} else {
			if strings.Contains(err.Error(), "status=404") {
				n, e := client.ReadBuildNumberByQueueID(jobName, queueID)
				if e == nil && n > 0 {
					return n, true
				}
				return 0, false
			}
		}

		time.Sleep(350 * time.Millisecond)
	}

	return 0, false
}

func extractQueueID(queueURL string) (int, bool) {
	s := strings.TrimSpace(queueURL)
	s = strings.TrimSuffix(s, "/")
	parts := strings.Split(s, "/")
	if len(parts) == 0 {
		return 0, false
	}
	last := parts[len(parts)-1]
	n, err := strconv.Atoi(last)
	if err != nil {
		return 0, false
	}
	return n, true
}

func boolToString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
