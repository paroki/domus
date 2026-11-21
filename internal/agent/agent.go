package agent

import (
	"log/slog"

	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/platform/setting"
)

// Feature wires AI agent configuration, drivers, and settings registration.
type Feature struct {
	config  *ConfigStore
	drivers *driver.Registry
}

// New constructs the agent Feature with its ConfigStore and driver Registry.
func New(drivers *driver.Registry, logger *slog.Logger) *Feature {
	if drivers == nil {
		drivers = driver.NewRegistry(nil)
	}
	return &Feature{
		config:  NewConfigStore(drivers, logger),
		drivers: drivers,
	}
}

// Config returns the hot-reloaded ConfigStore.
func (f *Feature) Config() *ConfigStore {
	return f.config
}

// Drivers returns the multi-provider LLM and embedding Registry.
func (f *Feature) Drivers() *driver.Registry {
	return f.drivers
}

// RegisterSettings registers all AI agent settings descriptors and hot-reload listeners on s.
func (f *Feature) RegisterSettings(s *setting.Store) {
	f.config.RegisterSettings(s)
}
