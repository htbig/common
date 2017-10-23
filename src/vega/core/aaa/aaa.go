package aaa

import (
	"vega/core/aaa/localusers"
	"vega/core/aaa/radius"
)

type Config struct {
	RADIUS     radius.Config     `json:"radius"`
	LocalUsers localusers.Config `json:"localusers"`
}

func (config *Config) Legacy(legacyRoot string) {
	config.RADIUS.Legacy(legacyRoot)
	config.LocalUsers.Legacy(legacyRoot)
}

func (config *Config) CopyFrom(otherConfig Config) {
	config.RADIUS.CopyFrom(otherConfig.RADIUS)
	config.LocalUsers.CopyFrom(otherConfig.LocalUsers)
}

func (config *Config) CopyFromInterface(data interface{}) bool {
	otherConfig, ok := data.(*Config)
	if !ok {
		return false
	}

	config.CopyFrom(*otherConfig)
	return true
}

func (config *Config) CloneInterface() interface{} {
	return config.Clone()
}

func (config *Config) Clone() *Config {
	newConfig := new(Config)
	newConfig.CopyFrom(*config)

	return newConfig
}

func (config *Config) Factory() {
	// set defaults
	config.RADIUS.Factory()
	config.LocalUsers.Factory()
}

func (config *Config) SaveInterface(data interface{}) (bool, []error) {
	oldConfig, ok := data.(*Config)
	if !ok {
		return false, nil
	}

	return true, config.Save(*oldConfig)
}

func (config *Config) Save(oldConfig Config) []error {
	errs := []error{}

	errs = append(errs, config.RADIUS.Save(oldConfig.RADIUS)...)
	errs = append(errs, config.LocalUsers.Save(oldConfig.LocalUsers)...)

	return errs
}

func (config *Config) Tag() string {
	return `aaa`
}

func (config *Config) Verify() []error {
	errs := []error{}

	errs = append(errs, config.RADIUS.Verify()...)
	errs = append(errs, config.LocalUsers.Verify()...)

	return errs
}
