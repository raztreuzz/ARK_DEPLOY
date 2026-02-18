package products

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ark_deploy/internal/storage"
)

type Handler struct {
	store *storage.ProductStore
}

func NewHandler(store *storage.ProductStore) *Handler {
	return &Handler{store: store}
}

type CreateProductRequest struct {
	ID          string            `json:"id" binding:"required"`
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Jobs        map[string]string `json:"jobs" binding:"required"`
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
		Jobs:        req.Jobs,
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
	Jobs        map[string]string `json:"jobs" binding:"required"`
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
		Jobs:        req.Jobs,
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
