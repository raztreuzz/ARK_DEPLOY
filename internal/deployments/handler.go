package deployments

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ark_deploy/internal/config"
	"ark_deploy/internal/jenkins"
	"ark_deploy/internal/storage"
)

type Handler struct {
	cfg   config.Config
	store *storage.ProductStore
}

func NewHandler(cfg config.Config, store *storage.ProductStore) *Handler {
	return &Handler{
		cfg:   cfg,
		store: store,
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
		product, err := h.store.GetByID(req.ProductID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": "product not found: " + err.Error()})
			return
		}
		
		jobName = product.Jobs[req.Environment]
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

	if resolved {
		c.JSON(http.StatusAccepted, gin.H{
			"status":       "queued_resolved",
			"job_name":     jobName,
			"queue_url":    queueURL,
			"build_number": buildNumber,
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":    "queued",
		"job_name":  jobName,
		"queue_url": queueURL,
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
