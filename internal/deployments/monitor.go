package deployments

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ark_deploy/internal/jenkins"
)

func (h *Handler) BuildStatus(c *gin.Context) {
	job := c.Param("job")
	buildStr := c.Param("build")

	n, err := strconv.Atoi(buildStr)
	if err != nil || n <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid build number"})
		return
	}

	client := jenkins.NewClient(h.cfg.JenkinsBaseURL, h.cfg.JenkinsUser, h.cfg.JenkinsAPIToken)

	building, result, err := client.ReadBuildStatus(job, n)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	status := "finished"
	if building {
		status = "running"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   status,
		"building": building,
		"result":   result,
	})
}

func (h *Handler) BuildLogs(c *gin.Context) {
	job := c.Param("job")
	buildStr := c.Param("build")

	n, err := strconv.Atoi(buildStr)
	if err != nil || n <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid build number"})
		return
	}

	client := jenkins.NewClient(h.cfg.JenkinsBaseURL, h.cfg.JenkinsUser, h.cfg.JenkinsAPIToken)

	logs, err := client.ReadBuildLogs(job, n)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	c.String(http.StatusOK, logs)
}

func (h *Handler) PendingJobs(c *gin.Context) {
	client := jenkins.NewClient(h.cfg.JenkinsBaseURL, h.cfg.JenkinsUser, h.cfg.JenkinsAPIToken)

	queueItems, err := client.ReadQueueItems()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(queueItems),
		"items": queueItems,
	})
}