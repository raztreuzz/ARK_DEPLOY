package jenkins

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type queueItemResp struct {
	Executable *struct {
		Number int    `json:"number"`
		URL    string `json:"url"`
	} `json:"executable"`
	Cancelled bool `json:"cancelled"`
}

func (c *Client) ReadQueueItem(queueURL string) (int, bool, error) {
	u := strings.TrimRight(queueURL, "/") + "/api/json"

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return 0, false, err
	}
	req.SetBasicAuth(c.user, c.token)

	resp, err := c.httpc.Do(req)
	if err != nil {
		return 0, false, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return 0, false, fmt.Errorf("queue api failed: status=%d body=%s", resp.StatusCode, string(b))
	}

	var qi queueItemResp
	if err := json.NewDecoder(resp.Body).Decode(&qi); err != nil {
		return 0, false, err
	}

	if qi.Cancelled {
		return 0, true, nil
	}

	if qi.Executable == nil {
		return 0, false, nil
	}

	return qi.Executable.Number, false, nil
}

type buildStatusResp struct {
	Building bool   `json:"building"`
	Result   string `json:"result"`
	Number   int    `json:"number"`
}

func (c *Client) ReadBuildStatus(jobName string, buildNumber int) (bool, string, error) {
	u := fmt.Sprintf("%s/job/%s/%d/api/json", c.baseURL, jobName, buildNumber)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return false, "", err
	}
	req.SetBasicAuth(c.user, c.token)

	resp, err := c.httpc.Do(req)
	if err != nil {
		return false, "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return false, "", fmt.Errorf("build status failed: status=%d body=%s", resp.StatusCode, string(b))
	}

	var bs buildStatusResp
	if err := json.NewDecoder(resp.Body).Decode(&bs); err != nil {
		return false, "", err
	}

	return bs.Building, bs.Result, nil
}

func (c *Client) ReadBuildLogs(jobName string, buildNumber int) (string, error) {
	u := fmt.Sprintf("%s/job/%s/%d/consoleText", c.baseURL, jobName, buildNumber)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.user, c.token)

	resp, err := c.httpc.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("build logs failed: status=%d body=%s", resp.StatusCode, string(b))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (c *Client) ReadBuildNumberByQueueID(jobName string, queueID int) (int, error) {
	u := fmt.Sprintf("%s/job/%s/api/json?tree=builds[number,queueId]{0,20}", c.baseURL, jobName)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(c.user, c.token)

	resp, err := c.httpc.Do(req)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			
		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("job api failed: status=%d body=%s", resp.StatusCode, string(b))
	}

	var data struct {
		Builds []struct {
			Number  int `json:"number"`
			QueueID int `json:"queueId"`
		} `json:"builds"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	for _, b := range data.Builds {
		if b.QueueID == queueID {
			return b.Number, nil
		}
	}

	return 0, nil
}

type QueueItemInfo struct {
	ID      int    `json:"id"`
	Task    string `json:"task"`
	Why     string `json:"why"`
	Blocked bool   `json:"blocked"`
	Stuck   bool   `json:"stuck"`
}

func (c *Client) ReadQueueItems() ([]QueueItemInfo, error) {
	u := fmt.Sprintf("%s/queue/api/json?tree=items[id,task[name],why,blocked,stuck]", c.baseURL)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.user, c.token)

	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("queue api failed: status=%d body=%s", resp.StatusCode, string(b))
	}

	var data struct {
		Items []struct {
			ID      int  `json:"id"`
			Blocked bool `json:"blocked"`
			Stuck   bool `json:"stuck"`
			Why     string `json:"why"`
			Task    struct {
				Name string `json:"name"`
			} `json:"task"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	result := make([]QueueItemInfo, 0, len(data.Items))
	for _, item := range data.Items {
		result = append(result, QueueItemInfo{
			ID:      item.ID,
			Task:    item.Task.Name,
			Why:     item.Why,
			Blocked: item.Blocked,
			Stuck:   item.Stuck,
		})
	}

	return result, nil
}

type JobInfo struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}

func (c *Client) ReadAllJobs() ([]JobInfo, error) {
	u := fmt.Sprintf("%s/api/json?tree=jobs[name,url,color]", c.baseURL)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.user, c.token)

	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jobs api failed: status=%d body=%s", resp.StatusCode, string(b))
	}

	var data struct {
		Jobs []struct {
			Name  string `json:"name"`
			URL   string `json:"url"`
			Color string `json:"color"`
		} `json:"jobs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	result := make([]JobInfo, 0, len(data.Jobs))
	for _, job := range data.Jobs {
		result = append(result, JobInfo{
			Name:  job.Name,
			URL:   job.URL,
			Color: job.Color,
		})
	}

	return result, nil
}
