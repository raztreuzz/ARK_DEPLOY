package config

import (
	"errors"
	"os"
)

type Config struct {
	Port             string
	JenkinsBaseURL   string
	JenkinsUser      string
	JenkinsAPIToken  string
	JenkinsJob       string
	TailscaleAPIKey  string
	TailscaleTailnet string
	ARKPublicHost    string // Public host for reverse proxy URLs (e.g., "100.103.47.3:3000")
}

func Load() (Config, error) {
	cfg := Config{
		Port:             os.Getenv("ARK_PORT"),
		JenkinsBaseURL:   os.Getenv("JENKINS_BASE_URL"),
		JenkinsUser:      os.Getenv("JENKINS_USER"),
		JenkinsAPIToken:  os.Getenv("JENKINS_API_TOKEN"),
		JenkinsJob:       os.Getenv("JENKINS_JOB"),
		TailscaleAPIKey:  os.Getenv("TAILSCALE_API_KEY"),
		TailscaleTailnet: os.Getenv("TAILSCALE_TAILNET"),
		ARKPublicHost:    os.Getenv("ARK_PUBLIC_HOST"), // Optional: defaults to request host
	}

	if cfg.Port == "" {
		cfg.Port = "5050"
	}

	if cfg.JenkinsBaseURL == "" || cfg.JenkinsUser == "" || cfg.JenkinsAPIToken == "" || cfg.JenkinsJob == "" {
		return Config{}, errors.New("missing required env vars: JENKINS_BASE_URL, JENKINS_USER, JENKINS_API_TOKEN, JENKINS_JOB")
	}

	if cfg.TailscaleAPIKey == "" || cfg.TailscaleTailnet == "" {
		return Config{}, errors.New("missing required env vars: TAILSCALE_API_KEY, TAILSCALE_TAILNET")
	}

	return cfg, nil
}
