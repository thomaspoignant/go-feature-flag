package config

import (
	"fmt"
	"strings"

	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"go.uber.org/zap/zapcore"
)

// IsValid contains all the validation of the configuration.
func (c *Config) IsValid() error {
	if c == nil {
		return fmt.Errorf("empty config")
	}
	if c.GetServerPort(nil) == 0 {
		return fmt.Errorf("invalid port %d", c.GetServerPort(nil))
	}
	if err := validateLogLevel(c.LogLevel); err != nil {
		return err
	}
	if err := validateLogFormat(c.LogFormat); err != nil {
		return err
	}
	if err := c.Server.Validate(); err != nil {
		return err
	}
	if len(c.FlagSets) > 0 {
		return c.validateFlagSets()
	}
	return c.validateDefaultMode()
}

// validateLogFormat validates the log format
func validateLogFormat(logFormat string) error {
	switch strings.ToLower(logFormat) {
	case "json", "logfmt", "":
		return nil
	default:
		return fmt.Errorf("invalid log format %s", logFormat)
	}
}

// validateLogLevel validates the log level
func validateLogLevel(logLevel string) error {
	if logLevel == "" {
		return nil
	}
	if _, err := zapcore.ParseLevel(logLevel); err != nil {
		return err
	}
	return nil
}

// validateDefaultMode validates the default mode
func (c *Config) validateDefaultMode() error {
	if err := validateRetrievers(c.Retriever, c.Retrievers); err != nil {
		return err
	}
	if err := validateExporters(c.Exporter, c.Exporters); err != nil {
		return err
	}
	return validateNotifiers(c.Notifiers)
}

// validateFlagSets validates all configured flagsets
func (c *Config) validateFlagSets() error {
	if len(c.FlagSets) == 0 {
		return fmt.Errorf("no flagsets configured")
	}

	// Track API keys to ensure no duplicates across flagsets
	apiKeySet := make(map[string]string) // apiKey -> flagsetName

	for _, flagset := range c.FlagSets {
		// Validate API keys
		if len(flagset.APIKeys) == 0 {
			return fmt.Errorf("flagset %s has no API keys", flagset.Name)
		}

		// Check for duplicate API keys across flagsets
		for _, apiKey := range flagset.APIKeys {
			if existingFlagset, exists := apiKeySet[apiKey]; exists {
				return fmt.Errorf("API key %s is used by multiple flagsets: %s and %s", apiKey, existingFlagset, flagset.Name)
			}
			apiKeySet[apiKey] = flagset.Name
		}

		// Validate the CommonFlagSet embedded in the flagset
		if err := c.validateFlagSetCommonConfig(&flagset); err != nil {
			return fmt.Errorf("flagset %s: %w", flagset.Name, err)
		}
	}

	return nil
}

// validateFlagSetCommonConfig validates the CommonFlagSet configuration for a single flagset
func (c *Config) validateFlagSetCommonConfig(flagset *FlagSet) error {
	if err := validateRetrievers(flagset.Retriever, flagset.Retrievers); err != nil {
		return err
	}

	if err := validateExporters(flagset.Exporter, flagset.Exporters); err != nil {
		return err
	}

	// Validate notifiers
	if err := validateNotifiers(flagset.Notifiers); err != nil {
		return err
	}
	return nil
}

// validateRetrievers validates the retrievers
func validateRetrievers(retriever *retrieverconf.RetrieverConf, retrievers *[]retrieverconf.RetrieverConf) error {
	if retriever == nil && retrievers == nil {
		return fmt.Errorf("no retriever available in the configuration")
	}
	if retriever != nil {
		if err := retriever.IsValid(); err != nil {
			return err
		}
	}

	if retrievers != nil {
		for i, retriever := range *retrievers {
			if err := retriever.IsValid(); err != nil {
				return fmt.Errorf("retriever at index %d validation failed: %w", i, err)
			}
		}
	}
	return nil
}

// validateExporters validates the exporters
func validateExporters(exporter *ExporterConf, exporters *[]ExporterConf) error {
	if exporter != nil {
		if err := exporter.IsValid(); err != nil {
			return err
		}
	}
	if exporters != nil {
		for i, exporter := range *exporters {
			if err := exporter.IsValid(); err != nil {
				return fmt.Errorf("exporter at index %d validation failed: %w", i, err)
			}
		}
	}
	return nil
}

// validateNotifiers validates the notifiers
func validateNotifiers(notifiers []NotifierConf) error {
	if len(notifiers) > 0 {
		for i, notif := range notifiers {
			if err := notif.IsValid(); err != nil {
				return fmt.Errorf("notifier at index %d validation failed: %w", i, err)
			}
		}
	}
	return nil
}
