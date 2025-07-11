package service

import (
	"context"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

type GoFeatureFlagService struct {
	goffClient map[string]*ffclient.GoFeatureFlag
}

// NewGoFeatureFlag creates a new GoFeatureFlag service.
func NewGoFeatureFlagService(
	ctx context.Context, proxyConf *config.Config, logger zap.Logger) (GoFeatureFlagService, error) {
	clients := make(map[string]*ffclient.GoFeatureFlag)
	for _, flagSet := range proxyConf.FlagSets {
		goff, err := NewGoFeatureFlagClient(&flagSet, &logger, []notifier.Notifier{})
		if err != nil {
			return GoFeatureFlagService{}, err
		}
		clients[flagSet.Name] = goff
	}

	return GoFeatureFlagService{
		goffClient: clients,
	}, nil
}
