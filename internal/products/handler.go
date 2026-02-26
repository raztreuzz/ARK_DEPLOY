package products

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"ark_deploy/internal/storage"
)

type Store interface {
	Create(p storage.Product) error
	GetAll() []storage.Product
	GetByID(id string) (storage.Product, error)
	Update(id string, p storage.Product) error
	Delete(id string) error
}

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	if store == nil {
		panic("products store is required")
	}
	return &Handler{store: store}
}

type CreateProductRequest struct {
	ID          string            `json:"id" binding:"required"`
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	DeployJobs  map[string]string `json:"deploy_jobs" binding:"required"`
	DeleteJob   string            `json:"delete_job" binding:"required"`
	WebService  string            `json:"web_service"`
	WebPort     int               `json:"web_port"`
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	if err := validateProductFields(req.ID, req.Name, req.DeployJobs, req.DeleteJob, req.WebService, req.WebPort); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	product := storage.Product{
		ID:          strings.TrimSpace(req.ID),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		DeployJobs:  normalizeDeployJobs(req.DeployJobs),
		DeleteJob:   strings.TrimSpace(req.DeleteJob),
		WebService:  strings.TrimSpace(req.WebService),
		WebPort:     req.WebPort,
	}

	if err := h.store.Create(product); err != nil {
		c.JSON(http.StatusConflict, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func (h *Handler) List(c *gin.Context) {
	products := h.store.GetAll()
	c.JSON(http.StatusOK, gin.H{
		"total":    len(products),
		"products": products,
	})
}

func (h *Handler) Get(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	product, err := h.store.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

type UpdateProductRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	DeployJobs  map[string]string `json:"deploy_jobs" binding:"required"`
	DeleteJob   string            `json:"delete_job" binding:"required"`
	WebService  string            `json:"web_service"`
	WebPort     int               `json:"web_port"`
}

func (h *Handler) Update(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	if err := validateProductFields(id, req.Name, req.DeployJobs, req.DeleteJob, req.WebService, req.WebPort); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	product := storage.Product{
		ID:          id,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		DeployJobs:  normalizeDeployJobs(req.DeployJobs),
		DeleteJob:   strings.TrimSpace(req.DeleteJob),
		WebService:  strings.TrimSpace(req.WebService),
		WebPort:     req.WebPort,
	}

	if err := h.store.Update(id, product); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *Handler) Delete(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	if err := h.store.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}

func validateProductFields(id, name string, deployJobs map[string]string, deleteJob, webService string, webPort int) error {
	if strings.TrimSpace(id) == "" {
		return errString("id is required")
	}
	if strings.TrimSpace(name) == "" {
		return errString("name is required")
	}
	if len(deployJobs) == 0 {
		return errString("deploy_jobs is required")
	}
	if strings.TrimSpace(deleteJob) == "" {
		return errString("delete_job is required")
	}
	if !isSafeJenkinsJobName(deleteJob) {
		return errString("delete_job contains invalid characters")
	}

	n := normalizeDeployJobs(deployJobs)
	for _, k := range []string{"prod", "dev", "test"} {
		if strings.TrimSpace(n[k]) == "" {
			return errString("deploy_jobs must include keys: prod, dev, test")
		}
		if !isSafeJenkinsJobName(n[k]) {
			return errString("deploy_jobs contains invalid job name for env: " + k)
		}
	}

	ws := strings.TrimSpace(webService)
	if ws == "" {
		ws = "web"
	}
	if !isSafeServiceName(ws) {
		return errString("web_service contains invalid characters")
	}

	if webPort == 0 {
		webPort = 80
	}
	if webPort < 1 || webPort > 65535 {
		return errString("web_port must be between 1 and 65535")
	}

	return nil
}

func normalizeDeployJobs(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[strings.TrimSpace(strings.ToLower(k))] = strings.TrimSpace(v)
	}
	return out
}

func isSafeServiceName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
		if !ok {
			return false
		}
	}
	return true
}

func isSafeJenkinsJobName(s string) bool {
	if s == "" {
		return false
	}
	if strings.Contains(s, "..") {
		return false
	}
	if strings.ContainsAny(s, "\\\n\r\t") {
		return false
	}
	_, err := url.Parse(s)
	if err != nil {
		return false
	}
	if strings.Contains(s, " ") {
		return false
	}
	return true
}

type errString string

func (e errString) Error() string { return string(e) }