package config

import (
	"fmt"
	"strings"

	"go.uber.org/zap/zapcore"
)

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
	if err := c.validateRetrievers(); err != nil {
		return err
	}
	if err := c.validateExporters(); err != nil {
		return err
	}
	return c.validateNotifiers()
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
	// Validate retrievers
	if flagset.Retriever == nil && flagset.Retrievers == nil {
		return fmt.Errorf("no retriever available in the flagset configuration")
	}
	if flagset.Retriever != nil {
		if err := flagset.Retriever.IsValid(); err != nil {
			return err
		}
	}

	if flagset.Retrievers != nil {
		for i, retriever := range *flagset.Retrievers {
			if err := retriever.IsValid(); err != nil {
				return fmt.Errorf("retriever at index %d validation failed: %w", i, err)
			}
		}
	}

	// Validate exporters
	if flagset.Exporter != nil {
		if err := flagset.Exporter.IsValid(); err != nil {
			return err
		}
	}
	if flagset.Exporters != nil {
		for i, exporter := range *flagset.Exporters {
			if err := exporter.IsValid(); err != nil {
				return fmt.Errorf("exporter at index %d validation failed: %w", i, err)
			}
		}
	}

	// Validate notifiers
	if flagset.Notifiers != nil {
		for i, notifier := range flagset.Notifiers {
			if err := notifier.IsValid(); err != nil {
				return fmt.Errorf("notifier at index %d validation failed: %w", i, err)
			}
		}
	}

	return nil
}

// validateRetrievers validates the retrievers
func (c *Config) validateRetrievers() error {
	if c.Retriever == nil && c.Retrievers == nil {
		return fmt.Errorf("no retriever available in the configuration")
	}
	if c.Retriever != nil {
		if err := c.Retriever.IsValid(); err != nil {
			return err
		}
	}

	if c.Retrievers != nil {
		for _, retriever := range *c.Retrievers {
			if err := retriever.IsValid(); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateExporters validates the exporters
func (c *Config) validateExporters() error {
	if c.Exporter != nil {
		if err := c.Exporter.IsValid(); err != nil {
			return err
		}
	}
	if c.Exporters != nil {
		for _, exporter := range *c.Exporters {
			if err := exporter.IsValid(); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateNotifiers validates the notifiers
func (c *Config) validateNotifiers() error {
	if c.Notifiers != nil {
		for _, notif := range c.Notifiers {
			if err := notif.IsValid(); err != nil {
				return err
			}
		}
	}
	return nil
}
