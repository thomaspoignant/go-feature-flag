package flagset

import "fmt"

type Manager struct {
	FlagSets map[string]*FlagSet
}

func NewManager() *Manager {
	return &Manager{
		FlagSets: make(map[string]*FlagSet),
	}
}

func (f *Manager) AddFlagSet(name string, flagSet *FlagSet) error {
	if _, ok := f.FlagSets[name]; ok {
		return fmt.Errorf("flagset %s already exists", name)
	}
	f.FlagSets[name] = flagSet
	return nil
}
