package deployments

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ark_deploy/internal/jenkins"
)

func (h *Handler) QueueToBuild(c *gin.Context) {
	queueURL := c.Query("url")
	if queueURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "missing query param: url"})
		return
	}

	client := jenkins.NewClient(h.cfg.JenkinsBaseURL, h.cfg.JenkinsUser, h.cfg.JenkinsAPIToken)

	buildNumber, cancelled, err := client.ReadQueueItem(queueURL)
	if err != nil {
		if strings.Contains(err.Error(), "status=404") {
			queueID, ok := extractQueueID(queueURL)
			if ok {
				n, e := client.ReadBuildNumberByQueueID(h.cfg.JenkinsJob, queueID)
				if e != nil {
					c.JSON(http.StatusBadGateway, gin.H{"detail": e.Error()})
					return
				}
				if n > 0 {
					c.JSON(http.StatusOK, gin.H{
						"status":       "queue_expired_but_resolved",
						"job_name":     h.cfg.JenkinsJob,
						"build_number": n,
					})
					return
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"status":   "queue_expired",
				"job_name": h.cfg.JenkinsJob,
			})
			return
		}

		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	if cancelled {
		c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
		return
	}

	if buildNumber == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "queued"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "started",
		"job_name":     h.cfg.JenkinsJob,
		"build_number": buildNumber,
	})
}

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

func (h *Handler) ListJobs(c *gin.Context) {
	client := jenkins.NewClient(h.cfg.JenkinsBaseURL, h.cfg.JenkinsUser, h.cfg.JenkinsAPIToken)

	jobs, err := client.ReadAllJobs()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(jobs),
		"jobs":  jobs,
	})
}
