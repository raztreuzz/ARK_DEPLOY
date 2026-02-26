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
	JobName      string `json:"job_name"`
	AppName      string `json:"app_name"`
	Version      string `json:"version"`
	TargetHost   string `json:"target_host" binding:"required"`
	SSHUser      string `json:"ssh_user" binding:"required"`
	SimulateFail bool   `json:"simulate_fail"`
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	req.ProductID = strings.TrimSpace(req.ProductID)
	req.Environment = strings.TrimSpace(req.Environment)
	req.AppName = strings.TrimSpace(req.AppName)
	req.Version = strings.TrimSpace(req.Version)
	req.TargetHost = strings.TrimSpace(req.TargetHost)
	req.SSHUser = strings.TrimSpace(req.SSHUser)

	productID := req.ProductID
	if productID == "" {
		productID = req.AppName
	}
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "product_id or app_name is required"})
		return
	}

	env := strings.ToLower(strings.TrimSpace(req.Environment))
	if env == "" {
		version := strings.ToLower(strings.TrimSpace(req.Version))
		if version == "" {
			env = "prod"
		} else {
			switch version {
			case "prod", "production":
				env = "prod"
			case "dev", "development":
				env = "dev"
			case "test", "testing":
				env = "test"
			default:
				env = "prod"
			}
		}
	}

	if env != "prod" && env != "dev" && env != "test" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "environment must be prod, dev, or test"})
		return
	}

	client := jenkins.NewClient(h.cfg.JenkinsBaseURL, h.cfg.JenkinsUser, h.cfg.JenkinsAPIToken)

	var jobName string
	var product storage.Product

	if strings.TrimSpace(req.JobName) != "" {
		jobName = strings.TrimSpace(req.JobName)
	} else {
		p, err := h.productStore.GetByID(productID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": "product not found"})
			return
		}
		product = p

		jobName = strings.TrimSpace(product.DeployJobs[env])
		if jobName == "" {
			jobName = strings.TrimSpace(product.Jobs[env])
		}
		if jobName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"detail": fmt.Sprintf("no deploy job configured for product %s in environment %s", productID, env)})
			return
		}
	}

	if product.ID == "" {
		p, err := h.productStore.GetByID(productID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": "product not found"})
			return
		}
		product = p
	}

	webService := strings.TrimSpace(product.WebService)
	if webService == "" {
		webService = "web"
	}
	webPort := product.WebPort
	if webPort == 0 {
		webPort = 80
	}

	instanceID := uuid.New().String()

	publicBase := strings.TrimRight(h.cfg.ARKPublicHost, "/")
	callbackURL := publicBase + "/api/instances/register"

	queueURL, err := client.TriggerJobWithParams(jobName, map[string]string{
		"INSTANCE_ID":      instanceID,
		"PRODUCT_ID":       productID,
		"ENV":              env,
		"TARGET_HOST":      req.TargetHost,
		"SSH_USER":         req.SSHUser,
		"ARK_CALLBACK_URL": callbackURL,
		"WEB_SERVICE":      webService,
		"WEB_PORT":         strconv.Itoa(webPort),
		"SIMULATE_FAIL":    boolToString(req.SimulateFail),
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	buildNumber, resolved := h.tryResolveBuildNumber(client, jobName, queueURL)

	instanceURL := publicBase + "/instances/" + instanceID + "/"

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
			"target_host":  req.TargetHost,
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"instance_id": instanceID,
		"url":         instance.URL,
		"status":      "queued",
		"job_name":    jobName,
		"queue_url":   queueURL,
		"target_host": req.TargetHost,
	})
}

func (h *Handler) List(c *gin.Context) {
	instances := h.instanceStore.GetAll()

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
