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
	AppName      string `json:"app_name"`
	Version      string `json:"version"`
	TargetHost   string `json:"target_host" binding:"required"`
	SimulateFail bool   `json:"simulate_fail"`
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	// Normalize and validate request
	req.ProductID = strings.TrimSpace(req.ProductID)
	req.Environment = strings.TrimSpace(req.Environment)
	req.AppName = strings.TrimSpace(req.AppName)
	req.Version = strings.TrimSpace(req.Version)
	req.TargetHost = strings.TrimSpace(req.TargetHost)

	// Determine productID (new format or legacy fallback)
	productID := req.ProductID
	if productID == "" {
		productID = req.AppName
	}
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "product_id or app_name is required"})
		return
	}

	// Determine environment (new format or legacy mapping)
	env := strings.ToUpper(req.Environment)
	if env == "" {
		// Map version to environment
		version := strings.ToLower(req.Version)
		if version == "" {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "environment or version is required"})
			return
		}
		switch version {
		case "prod", "production":
			env = "PROD"
		case "dev", "development":
			env = "DEV"
		default:
			env = "PROD" // default fallback
		}
	}

	client := jenkins.NewClient(h.cfg.JenkinsBaseURL, h.cfg.JenkinsUser, h.cfg.JenkinsAPIToken)

	// Resolve Jenkins job from product catalog
	var jobName string
	if req.JobName != "" {
		// Direct job_name override (for advanced use)
		jobName = req.JobName
	} else {
		// Fetch product and resolve job
		product, err := h.productStore.GetByID(productID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": "product not found"})
			return
		}

		// Try deploy_jobs first, then fallback to legacy jobs field
		jobName = product.DeployJobs[env]
		if jobName == "" && len(product.DeployJobs) == 0 {
			jobName = product.Jobs[env]
		}
		if jobName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"detail": fmt.Sprintf("no deploy job configured for product %s in environment %s", productID, env)})
			return
		}
	}

	// Use normalized values for Jenkins parameters
	appNameParam := req.AppName
	if appNameParam == "" {
		appNameParam = productID
	}
	versionParam := req.Version
	if versionParam == "" {
		versionParam = strings.ToLower(env)
	}

	instanceID := uuid.New().String()

	// Client jobs must publish an ephemeral host port and register:
	// instance_id -> target_host (tailscale) + target_port.
	// ARK routes traffic by path /instances/<instance_id>/ to that target.
	queueURL, err := client.TriggerJobWithParams(jobName, map[string]string{
		"INSTANCE_ID":   instanceID,
		"APP_NAME":      appNameParam,
		"VERSION":       versionParam,
		"TARGET_HOST":   req.TargetHost,
		"ARK_CALLBACK_URL": "http://100.103.47.3:5050/instances/register",
		"SIMULATE_FAIL": boolToString(req.SimulateFail),
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	buildNumber, resolved := h.tryResolveBuildNumber(client, jobName, queueURL)

	// Determine public host for reverse proxy URL
	publicHost := h.cfg.ARKPublicHost
	if publicHost == "" {
		publicHost = c.Request.Host
	}
	if publicHost == "" {
		publicHost = "localhost:3000"
	}

	// URL exposed via Nginx reverse proxy (path-based), not direct node access
	instanceURL := fmt.Sprintf("http://%s/instances/%s/", publicHost, instanceID)

	instance := storage.Instance{
		ID:          instanceID,
		ProductID:   productID,
		DeviceID:    req.TargetHost,
		Environment: env,
		Status:      "provisioning",
		URL:         instanceURL,
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
