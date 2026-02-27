package jenkins

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	baseURL string
	user    string
	token   string
	httpc   *http.Client
}

// NewClient crea una nueva instancia del cliente Jenkins con la URL base, usuario y token proporcionados.
// Retorna un puntero a Client.
func NewClient(baseURL, user, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		user:    user,
		token:   token,
		httpc:   &http.Client{},
	}
}
//Creamos el espacio para obtener el crumb de jenkins que es una especie de token de seguridad 
type crumbResp struct {
	CrumbRequestField string `json:"crumbRequestField"`
	Crumb             string `json:"crumb"`
}

// getCrumb obtiene el crumb de Jenkins necesario para realizar peticiones autenticadas que requieren protección CSRF.
// Retorna el nombre del campo crumb, el valor del crumb y un error si ocurre alguno.
func (c *Client) getCrumb() (string, string, error) {
	endpoint := c.baseURL + "/crumbIssuer/api/json"

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", "", err
	}
	req.SetBasicAuth(c.user, c.token)

	resp, err := c.httpc.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("crumb issuer failed: status=%d body=%s", resp.StatusCode, string(b))
	}

	var cr crumbResp
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", "", err
	}
	if cr.CrumbRequestField == "" || cr.Crumb == "" {
		return "", "", errors.New("crumb response missing fields")
	}

	return cr.CrumbRequestField, cr.Crumb, nil
}

// Mapeamos el endpoint con parametros para el trigger del job y obtenemos el crumb para la autenticacion  
func (c *Client) TriggerJobWithParams(jobName string, params map[string]string) (string, error) {
	endpoint := fmt.Sprintf("%s/job/%s/buildWithParameters", c.baseURL, url.PathEscape(jobName))

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	crumbField, crumb, err := c.getCrumb()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(c.user, c.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(crumbField, crumb)

	resp, err := c.httpc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("jenkins trigger failed: status=%d body=%s", resp.StatusCode, string(b))
	}

	queueURL := resp.Header.Get("Location")
	if queueURL == "" {
		return "", errors.New("jenkins did not return queue Location header")
	}

	return queueURL, nil
}

//Obtenemos logs  del build a traves del endpoint 

func (c *Client) GetBuildLog(jobName string, buildNumber string) (string, error) {
	endpoint := fmt.Sprintf("%s/job/%s/%s/consoleText", c.baseURL, url.PathEscape(jobName), buildNumber)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.user, c.token)

	resp, err := c.httpc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get build log: status=%d body=%s", resp.StatusCode, string(b))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
