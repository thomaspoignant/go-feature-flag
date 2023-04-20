package config

import "fmt"

type NotifierConf struct {
	Kind            NotifierKind      `mapstructure:"notifier" koanf:"notifier"`
	SlackWebhookURL string            `mapstructure:"slackWebhookUrl" koanf:"slackWebhookUrl"`
	EndpointURL     string            `mapstructure:"endpointUrl" koanf:"endpointUrl"`
	Secret          string            `mapstructure:"secret" koanf:"secret"`
	Meta            map[string]string `mapstructure:"meta" koanf:"meta"`
}

func (c *NotifierConf) IsValid() error {
	if err := c.Kind.IsValid(); err != nil {
		return err
	}
	if c.Kind == SlackNotifier && c.SlackWebhookURL == "" {
		return fmt.Errorf("invalid notifier: no \"slackWebhookUrl\" property found for kind \"%s\"", c.Kind)
	}
	if c.Kind == WebhookNotifier && c.EndpointURL == "" {
		return fmt.Errorf("invalid notifier: no \"endpointUrl\" property found for kind \"%s\"", c.Kind)
	}
	return nil
}

type NotifierKind string

const (
	SlackNotifier   NotifierKind = "slack"
	WebhookNotifier NotifierKind = "webhook"
)

// IsValid is checking if the value is part of the enum
func (r NotifierKind) IsValid() error {
	switch r {
	case SlackNotifier, WebhookNotifier:
		return nil
	}
	return fmt.Errorf("invalid notifier: kind \"%s\" is not supported", r)
}
