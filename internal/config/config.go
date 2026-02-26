package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Port             string
	JenkinsBaseURL   string
	JenkinsUser      string
	JenkinsAPIToken  string
	JenkinsJob       string
	TailscaleAPIKey  string
	TailscaleTailnet string
	ARKPublicHost    string
}

func Load() (Config, error) {
	cfg := Config{
		Port:             strings.TrimSpace(os.Getenv("ARK_PORT")),
		JenkinsBaseURL:   strings.TrimSpace(os.Getenv("JENKINS_BASE_URL")),
		JenkinsUser:      strings.TrimSpace(os.Getenv("JENKINS_USER")),
		JenkinsAPIToken:  strings.TrimSpace(os.Getenv("JENKINS_API_TOKEN")),
		JenkinsJob:       strings.TrimSpace(os.Getenv("JENKINS_JOB")),
		TailscaleAPIKey:  strings.TrimSpace(os.Getenv("TAILSCALE_API_KEY")),
		TailscaleTailnet: strings.TrimSpace(os.Getenv("TAILSCALE_TAILNET")),
		ARKPublicHost:    strings.TrimSpace(os.Getenv("ARK_PUBLIC_HOST")),
	}

	if cfg.Port == "" {
		cfg.Port = "5050"
	}

	var missing []string

	if cfg.JenkinsBaseURL == "" {
		missing = append(missing, "JENKINS_BASE_URL")
	}
	if cfg.JenkinsUser == "" {
		missing = append(missing, "JENKINS_USER")
	}
	if cfg.JenkinsAPIToken == "" {
		missing = append(missing, "JENKINS_API_TOKEN")
	}
	if cfg.JenkinsJob == "" {
		missing = append(missing, "JENKINS_JOB")
	}
	if cfg.TailscaleAPIKey == "" {
		missing = append(missing, "TAILSCALE_API_KEY")
	}
	if cfg.TailscaleTailnet == "" {
		missing = append(missing, "TAILSCALE_TAILNET")
	}
	if cfg.ARKPublicHost == "" {
		missing = append(missing, "ARK_PUBLIC_HOST")
	}

	if len(missing) > 0 {
		return Config{}, errors.New("missing required env vars: " + strings.Join(missing, ", "))
	}

	var err error

	cfg.ARKPublicHost, err = normalizeBaseURL(cfg.ARKPublicHost, "ARK_PUBLIC_HOST")
	if err != nil {
		return Config{}, err
	}

	cfg.JenkinsBaseURL, err = normalizeBaseURL(cfg.JenkinsBaseURL, "JENKINS_BASE_URL")
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func normalizeBaseURL(raw string, envName string) (string, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimRight(raw, "/")

	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("%s must be a valid URL with scheme (http:// or https://), got: %q", envName, raw)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("%s must use http or https, got: %q", envName, u.Scheme)
	}

	if u.User != nil {
		return "", fmt.Errorf("%s must not include userinfo, got: %q", envName, raw)
	}

	u.Fragment = ""
	u.RawQuery = ""
	u.Path = ""

	return u.String(), nil
}