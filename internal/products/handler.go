package products

import (
	"net/http"

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
		store = storage.NewProductStore()
	}
	return &Handler{store: store}
}

type CreateProductRequest struct {
	ID          string            `json:"id" binding:"required"`
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	DeployJobs  map[string]string `json:"deploy_jobs"`
	DeleteJob   string            `json:"delete_job"`
	Jobs        map[string]string `json:"jobs"`
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	product := storage.Product{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		DeployJobs:  req.DeployJobs,
		DeleteJob:   req.DeleteJob,
		Jobs:        req.Jobs,
	}

	if len(product.DeployJobs) == 0 && len(product.Jobs) > 0 {
		product.DeployJobs = product.Jobs
	}
	if len(product.DeployJobs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "deploy_jobs is required"})
		return
	}
	if product.DeleteJob == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "delete_job is required"})
		return
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
	id := c.Param("id")

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
	DeployJobs  map[string]string `json:"deploy_jobs"`
	DeleteJob   string            `json:"delete_job"`
	Jobs        map[string]string `json:"jobs"`
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	product := storage.Product{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		DeployJobs:  req.DeployJobs,
		DeleteJob:   req.DeleteJob,
		Jobs:        req.Jobs,
	}

	if len(product.DeployJobs) == 0 && len(product.Jobs) > 0 {
		product.DeployJobs = product.Jobs
	}
	if len(product.DeployJobs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "deploy_jobs is required"})
		return
	}
	if product.DeleteJob == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "delete_job is required"})
		return
	}

	if err := h.store.Update(id, product); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}
