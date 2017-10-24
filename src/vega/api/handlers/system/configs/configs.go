// Copyright (c) 2016, Virtual Gateway Labs. All rights reserved.

package configs

import (
	"net/http"

	"vega/api/handlers"
	"vega/core"
	"github.com/htbig/common/src/vega/syslogger"
)

func GetDefault(ctx handlers.Context) {
	cfg := core.NewConfig()

	if err := cfg.LoadDefault(); err != nil {
		ctx.EncodeInternalServerErrors(err)
	} else {
		ctx.Encode(cfg.Map())
	}
}

func PatchDefault(ctx handlers.Context) {
	cfg := core.NewConfig()

	if err := cfg.LoadDefault(); err != nil {
		ctx.EncodeInternalServerErrors(err)
	}

	if !ctx.MapDecode(&cfg) {
		return
	}

	if errorMap := cfg.Verify(); len(errorMap) > 0 {
		ctx.EncodeErrorMap(http.StatusBadRequest, errorMap)
		return
	}

	if err := cfg.SaveDefault(); err != nil {
		syslogger.Err("API Save Default Error:", err)
		ctx.EncodeInternalServerErrors(err)
		return
	}
}

func GetRunning(ctx handlers.Context) {
	ctx.Encode(ctx.Config.Map())
}

func PatchRunning(ctx handlers.Context) {
	configCopy, cloneError := ctx.Config.Clone()
	if cloneError != nil {
		ctx.EncodeInternalServerErrors(cloneError)
	}

	if !ctx.MapDecode(&configCopy) {
		return
	}

	if errorMap := configCopy.Verify(); len(errorMap) > 0 {
		ctx.EncodeErrorMap(http.StatusBadRequest, errorMap)
		return
	}

	configCopy.Save(*ctx.Config)

	// success
	if err := ctx.Config.CopyFrom(*configCopy); err != nil {
		ctx.EncodeInternalServerErrors(err)
		return
	}
}

func GetStartup(ctx handlers.Context) {
	cfg := core.NewConfig()

	if err := cfg.LoadStartup(); err != nil {
		ctx.EncodeInternalServerErrors(err)
	} else {
		ctx.Encode(cfg.Map())
	}
}

func PatchStartup(ctx handlers.Context) {
	cfg := core.NewConfig()

	if err := cfg.LoadStartup(); err != nil {
		ctx.EncodeInternalServerErrors(err)
	}

	if !ctx.MapDecode(&cfg) {
		return
	}

	if errorMap := cfg.Verify(); len(errorMap) > 0 {
		syslogger.Err("API Verify Startup Error:", errorMap)
		ctx.EncodeErrorMap(http.StatusBadRequest, errorMap)
		return
	}

	if err := cfg.SaveStartup(); err != nil {
		syslogger.Err("API Save Startup Error:", err)
		ctx.EncodeInternalServerErrors(err)
		return
	}
}

func SaveStartup(ctx handlers.Context) {
	if err := ctx.Config.SaveStartup(); err != nil {
		syslogger.Err("API Save Startup Error:", err)
		ctx.EncodeInternalServerErrors(err)
	}
}
