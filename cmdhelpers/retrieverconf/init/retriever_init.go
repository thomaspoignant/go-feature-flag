package init

import (
	"fmt"
	"time"

	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/retriever"
)

type DefaultRetrieverConfig struct {
	Timeout    time.Duration
	HTTPMethod string
	GitBranch  string
}

// InitRetriever initialize the retriever based on the configuration
func InitRetriever(
	c *retrieverconf.RetrieverConf, defaultRetrieverConfig DefaultRetrieverConfig) (retriever.Retriever, error) {
	if c.Timeout != 0 {
		defaultRetrieverConfig.Timeout = time.Duration(c.Timeout) * time.Millisecond
	}
	retrieverFactory, exists := retrieverFactories[c.Kind]
	if !exists {
		return nil, fmt.Errorf("invalid retriever: kind \"%s\" is not supported", c.Kind)
	}
	return retrieverFactory(c, &defaultRetrieverConfig)
}
