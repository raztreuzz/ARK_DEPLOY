package instances

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type RouteStore interface {
	GetRoute(instanceID string) (host string, port int, ok bool, err error)
	PutRoute(instanceID string, host string, port int) error
	DeleteRoute(instanceID string) error
}

type Handler struct {
	store RouteStore
}

func NewHandler(store RouteStore) *Handler {
	return &Handler{store: store}
}

type RegisterReq struct {
	InstanceID    string `json:"instance_id" binding:"required"`
	TargetHost    string `json:"target_host" binding:"required"`
	TargetPort    int    `json:"target_port" binding:"required"`
	ContainerName string `json:"container_name"`
	WebPort       string `json:"web_port"`
}

func (h *Handler) RegisterRoutes(r gin.IRoutes) {
	r.POST("/instances/register", h.register)
	r.DELETE("/instances/:id", h.delete)
	r.Any("/instances/:id/*path", h.proxy)
}

func (h *Handler) register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	req.InstanceID = strings.TrimSpace(req.InstanceID)
	req.TargetHost = strings.TrimSpace(req.TargetHost)

	if req.InstanceID == "" || req.TargetHost == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "instance_id and target_host are required"})
		return
	}

	if req.TargetPort <= 0 || req.TargetPort > 65535 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid target_port"})
		return
	}

	if err := h.store.PutRoute(req.InstanceID, req.TargetHost, req.TargetPort); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	upstreamURL := fmt.Sprintf("http://%s:%d/", req.TargetHost, req.TargetPort)
	reachable := checkUpstreamReachable(upstreamURL, 2*time.Second)

	c.JSON(http.StatusOK, gin.H{
		"status":             "ok",
		"upstream_reachable": reachable,
		"upstream_url":       upstreamURL,
	})
}

func (h *Handler) delete(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "instance id is required"})
		return
	}

	if err := h.store.DeleteRoute(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) proxy(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "instance id is required"})
		return
	}

	host, port, ok, err := h.store.GetRoute(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"detail": "instance not found"})
		return
	}
	if port <= 0 || port > 65535 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid target_port"})
		return
	}

	target, err := url.Parse("http://" + host + ":" + strconv.Itoa(port))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	rp := httputil.NewSingleHostReverseProxy(target)
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"detail":"upstream unreachable"}`))
	}

	origPath := c.Param("path")
	if origPath == "" {
		origPath = "/"
	}

	c.Request.URL.Path = singleJoiningSlash(target.Path, origPath)
	c.Request.Host = target.Host
	rp.ServeHTTP(c.Writer, c.Request)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")

	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	default:
		return a + b
	}
}

func checkUpstreamReachable(rawURL string, timeout time.Duration) bool {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 500
}
