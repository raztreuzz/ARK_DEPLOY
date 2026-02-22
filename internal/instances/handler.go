package instances

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type RouteStore interface {
	GetRoute(instanceID string) (host string, port int, ok bool, err error)
	PutRoute(instanceID string, host string, port int) error
	DeleteRoute(instanceID string) error
}

type RegisterReq struct {
	InstanceID string `json:"instance_id" binding:"required"`
	TargetHost string `json:"target_host" binding:"required"`
	TargetPort int    `json:"target_port" binding:"required"`
}

func RegisterRoutes(r *gin.Engine, store RouteStore) {
	r.POST("/instances/register", func(c *gin.Context) {
		var req RegisterReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		req.InstanceID = strings.TrimSpace(req.InstanceID)
		req.TargetHost = strings.TrimSpace(req.TargetHost)

		if req.TargetPort <= 0 || req.TargetPort > 65535 {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid target_port"})
			return
		}

		if req.InstanceID == "" || req.TargetHost == "" {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "instance_id and target_host are required"})
			return
		}

		if err := store.PutRoute(req.InstanceID, req.TargetHost, req.TargetPort); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.DELETE("/instances/:id", func(c *gin.Context) {
		id := strings.TrimSpace(c.Param("id"))
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "instance id is required"})
			return
		}

		if err := store.DeleteRoute(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	proxy := func(c *gin.Context) {
		id := strings.TrimSpace(c.Param("id"))
		host, port, ok, err := store.GetRoute(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"detail": "instance not found"})
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

	r.Any("/instances/:id/*path", proxy)
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
