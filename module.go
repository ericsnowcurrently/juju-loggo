// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package loggo

import (
	"strings"
	"sync"
)

// Do not change rootModuleName: modules.resolve() will misbehave if it isn't "".
const (
	rootName       = "<root>"
	rootModuleName = ""
)

func newRootModule() *loggerState {
	// The root module can't be unspecified (see SubLogger.EffectiveLogLevel).
	// So we set a default level.
	return &loggerState{
		name:         rootName,
		level:        defaultRootLevel,
		defaultLevel: defaultRootLevel,
	}
}

// newSubmodule returns a new submodule for the given info.
//
// The name should not be the empty string.
// A parent must always be provided.
func newSubmodule(name string, parent *loggerState, level Level) *loggerState {
	name = strings.ToLower(name)
	return &loggerState{
		name:   name,
		level:  level,
		parent: parent,
	}
}

type modules struct {
	mu           sync.Mutex
	rootLevel    Level
	defaultLevel Level
	all          map[string]*loggerState
}

// Initially the modules map only contains the root module.
func newModules(rootLevel Level) *modules {
	m := &modules{
		rootLevel:    rootLevel,
		defaultLevel: defaultLevel,
	}
	m.initUnlocked()
	return m
}

func (m *modules) initUnlocked() {
	if m.rootLevel == UNSPECIFIED {
		// The root level cannot be UNSPECIFIED.
		m.rootLevel = defaultRootLevel
	}
	if m.defaultLevel == UNSPECIFIED {
		m.defaultLevel = defaultLevel
	}
	root := newRootModule()
	root.level = m.rootLevel
	m.all = map[string]*loggerState{
		rootModuleName: root,
	}
}

func (m *modules) maybeInitUnlocked() {
	if m.all == nil {
		m.initUnlocked()
	}
}

// get returns a Logger for the given module name,
// creating it and its parents if necessary.
func (m *modules) get(name string) *loggerState {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maybeInitUnlocked() // guarantee we have a root module

	// Lowercase the module name, and look for it in the modules map.
	name = strings.ToLower(name)
	return m.resolveUnlocked(name)
}

func (m *modules) resolveUnlocked(name string) *loggerState {
	// m must already be initialized (e.g. newModules()).
	if name == rootName {
		name = rootModuleName
	}
	if impl, found := m.all[name]; found {
		return impl
	}
	parentName := rootModuleName
	if i := strings.LastIndex(name, "."); i >= 0 {
		parentName = name[0:i]
	}
	// Since there is always a root module, we always get a parent here.
	parent := m.resolveUnlocked(parentName)
	impl := newSubmodule(name, parent, m.defaultLevel)
	m.all[name] = impl
	return impl
}

// config returns the current configuration of the modules. Modules
// with UNSPECIFIED level will not be included.
func (m *modules) config() LoggersConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maybeInitUnlocked() // guarantee we have a root module

	cfg := make(LoggersConfig)
	for _, module := range m.all {
		if module.MinLogLevel() == UNSPECIFIED {
			continue
		}
		cfg[module.name] = module.config()
	}
	return cfg
}

// resetLevels iterates through the known modules and sets the levels of all
// to UNSPECIFIED, except for <root> which is set to WARNING.
func (m *modules) resetLevels() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, module := range m.all {
		if name == rootModuleName {
			module.level.set(m.rootLevel)
		} else {
			module.level.set(m.defaultLevel)
		}
	}
}
