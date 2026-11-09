package metric

import (
	"strconv"

	prom "github.com/prometheus/client_golang/prometheus"
	ffclient "github.com/thomaspoignant/go-feature-flag"
)

// FlagsetSource gives access to the flagsets currently loaded by the relay proxy.
type FlagsetSource interface {
	AllFlagSets() (map[string]*ffclient.GoFeatureFlag, error)
}

type flagLoadedCollector struct {
	source FlagsetSource
	desc   *prom.Desc
}

func newFlagLoadedCollector(source FlagsetSource) *flagLoadedCollector {
	return &flagLoadedCollector{
		source: source,
		desc: prom.NewDesc(
			prom.BuildFQName("", GOFFSubSystem, "flag_loaded"),
			"Flags currently loaded in the relay proxy cache, one series per flag with value 1.",
			[]string{"flagset", flagNameLabel, "disabled"},
			nil,
		),
	}
}

func (c *flagLoadedCollector) Describe(ch chan<- *prom.Desc) {
	ch <- c.desc
}

func (c *flagLoadedCollector) Collect(ch chan<- prom.Metric) {
	flagsets, err := c.source.AllFlagSets()
	if err != nil {
		return
	}
	for flagsetName, flagset := range flagsets {
		if flagset == nil {
			continue
		}
		flags, err := flagset.GetFlagsFromCache()
		if err != nil {
			continue
		}
		for flagName, f := range flags {
			ch <- prom.MustNewConstMetric(
				c.desc, prom.GaugeValue, 1, flagsetName, flagName, strconv.FormatBool(f.IsDisable()),
			)
		}
	}
}

// RegisterFlagLoaded exposes gofeatureflag_flag_loaded, computed from the flag cache on every scrape.
func (m *Metrics) RegisterFlagLoaded(source FlagsetSource) error {
	if m.Registry == nil || source == nil {
		return nil
	}
	return m.Registry.Register(newFlagLoadedCollector(source))
}
