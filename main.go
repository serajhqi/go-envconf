package envconf

import (
	"fmt"
	"reflect"
)

var ErrNotLoaded = fmt.Errorf("envconf: config not loaded")

// Load initializes the singleton config for T.
//
// Supported options:
//   - WithDotEnvPath(path): read variables from a specific .env file path.
//   - WithoutDotEnv(): disable .env file loading entirely.
//
// If no option is provided, Load attempts to read "./.env" (missing file is
// ignored). Values are resolved in this order: OS env > .env > `default` tag.
func Load[T any](opts ...Option) (*T, error) {
	cfg := new(T)
	cfgType, err := validateTarget(cfg)
	if err != nil {
		return nil, err
	}

	resolved := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&resolved)
		}
	}
	fp := fingerprint(resolved)

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.initialized {
		if state.cfgType != cfgType || state.optionsFingerprint != fp {
			return nil, fmt.Errorf("envconf: singleton already initialized with type %s and options %q", state.cfgType, state.optionsFingerprint)
		}
		existing, ok := state.cfg.(*T)
		if !ok {
			return nil, fmt.Errorf("envconf: singleton type mismatch: stored %T", state.cfg)
		}
		return existing, nil
	}

	dotEnvValues, err := loadDotEnv(resolved)
	if err != nil {
		return nil, err
	}

	if err := populateStruct(cfg, dotEnvValues); err != nil {
		return nil, err
	}

	state.initialized = true
	state.cfg = cfg
	state.cfgType = cfgType
	state.optionsFingerprint = fp

	return cfg, nil
}

// MustLoad is like Load but panics on error.
func MustLoad[T any](opts ...Option) *T {
	cfg, err := Load[T](opts...)
	if err != nil {
		panic(err)
	}
	return cfg
}

// Get returns the already-loaded singleton config for T.
// It returns ErrNotLoaded when Load has not been called yet.
func Get[T any]() (*T, error) {
	state.mu.Lock()
	defer state.mu.Unlock()

	if !state.initialized {
		return nil, ErrNotLoaded
	}

	cfgType := reflect.TypeOf((*T)(nil))
	if state.cfgType != cfgType {
		return nil, fmt.Errorf("envconf: singleton already initialized with type %s", state.cfgType)
	}

	existing, ok := state.cfg.(*T)
	if !ok {
		return nil, fmt.Errorf("envconf: singleton type mismatch: stored %T", state.cfg)
	}
	return existing, nil
}

// MustGet is like Get but panics on error.
func MustGet[T any]() *T {
	cfg, err := Get[T]()
	if err != nil {
		panic(err)
	}
	return cfg
}
