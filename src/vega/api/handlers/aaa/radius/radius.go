// Copyright (c) 2016, Virtual Gateway Labs. All rights reserved.

package radius

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"vega/api/auth"
	"vega/api/handlers"
	"vega/core/aaa/radius"
)

func Get(ctx handlers.Context) {
	ctx.Encode(ctx.Config.AAA.RADIUS)
}

func GetEnable(ctx handlers.Context) {
	ctx.Encode(ctx.Config.AAA.RADIUS.Enabled)
}

func GetFallback(ctx handlers.Context) {
	ctx.Encode(ctx.Config.AAA.RADIUS.Fallback)
}

func GetServers(ctx handlers.Context) {
	ctx.Encode(ctx.Config.AAA.RADIUS.Servers)
}

func Patch(ctx handlers.Context) {
	ctx.MapDecodeVerifySave(&ctx.Config.AAA.RADIUS)
}

func PutEnable(ctx handlers.Context) {
	cfg := ctx.Config.AAA.RADIUS.Clone()
	var en bool
	if ctx.Decode(&en) {
		cfg.Enabled = en

		ctx.VerifySave(cfg, &ctx.Config.AAA.RADIUS)
	}
}

func PutFallback(ctx handlers.Context) {
	cfg := ctx.Config.AAA.RADIUS.Clone()
	if ctx.Decode(&cfg.Fallback) {
		ctx.VerifySave(cfg, &ctx.Config.AAA.RADIUS)
	}
}

func PutServers(ctx handlers.Context) {
	cfg := ctx.Config.AAA.RADIUS.Clone()
	var serverList []radius.Server

	if ctx.Decode(&serverList) {
		cfg.Servers = serverList

		ctx.VerifySave(cfg, &ctx.Config.AAA.RADIUS)
	}
}

func PostServers(ctx handlers.Context) {
	cfg := ctx.Config.AAA.RADIUS.Clone()
	var serverList []radius.Server

	if ctx.Decode(&serverList) {
		cfg.Servers = append(cfg.Servers, serverList...)

		ctx.VerifySave(cfg, &ctx.Config.AAA.RADIUS)
	}
}

func DeleteServers(ctx handlers.Context) {
	serverList := ctx.Request.URL.Query().Get("servers")

	s := strings.FieldsFunc(serverList, func(c rune) bool {
		return c == ','
	})

	if len(s) > 0 {
		deleteServersByName(ctx, s)
	} else {
		//delete all servers
		cfg := ctx.Config.AAA.RADIUS.Clone()
		cfg.Servers = []radius.Server{}
		ctx.VerifySave(cfg, &ctx.Config.AAA.RADIUS)
	}
}

func deleteServersByName(ctx handlers.Context, toRemove []string) {
	cfg := ctx.Config.AAA.RADIUS.Clone()
	results, _, notRemoved := removeSelectServers(toRemove, cfg.Servers)
	if len(notRemoved) > 0 {
		errs := []error{}
		for _, server := range notRemoved {
			errs = append(errs, errors.New("["+server+"] is not configured"))
		}
		ctx.EncodeErrors(http.StatusNotFound, errs...)
	} else {
		cfg.Servers = results

		ctx.VerifySave(cfg, &ctx.Config.AAA.RADIUS)
	}
}

func removeSelectServers(toRemove []string, from []radius.Server) (results []radius.Server, removed, notRemoved []string) {
	results = []radius.Server{}
	removed = []string{}
	notRemoved = []string{}
	removedSet := make(map[string]bool)

outer:
	for _, ofFrom := range from {
		for _, ofToRemove := range toRemove {
			if host, port, err := getHostPort(ofToRemove); err == nil {
				if port == "" {
					if ofFrom.IPaddr == host {
						removedSet[ofToRemove] = true
						continue outer
					}
				} else {
					portAsString := strconv.Itoa(int(ofFrom.Port))
					if ofFrom.IPaddr == host && portAsString == port {
						removedSet[ofToRemove] = true
						continue outer
					}
				}
			}

		}
		results = append(results, ofFrom)
	}

	for _, remove := range toRemove {
		if removedSet[remove] {
			removed = append(removed, remove)
		} else {
			notRemoved = append(notRemoved, remove)
		}
	}

	return
}

func getHostPort(hostport string) (host string, port string, err error) {
	if (strings.Contains(hostport, ":") && strings.Contains(hostport, ".")) || strings.Contains(hostport, "]:") {
		return net.SplitHostPort(hostport)
	} else {
		return hostport, "", nil
	}
}

func LocalEnabled(ctx handlers.Context) {
	cfg := ctx.Config.AAA.RADIUS

	json.NewEncoder(ctx.Writer).Encode(cfg.Enabled)
}

func LocalAuthenticate(ctx handlers.Context) {

	username, password, _ := ctx.Request.BasicAuth()

	isRadius := true
	isFallback := ctx.Config.AAA.RADIUS.Fallback
	checkPrivilege := true
	authenticated, authorized, errs := auth.AuthenticateAPI(isRadius, isFallback, checkPrivilege, username, password)

	if len(errs) > 0 {
		if errs[0].Error() == radius.GatewayTimeoutError {
			ctx.EncodeErrors(http.StatusGatewayTimeout, errs...)
		} else if errs[0].Error() == radius.RadiusAuthError {
			ctx.EncodeErrors(http.StatusBadGateway, errs...)
		} else {
			ctx.EncodeInternalServerErrors(errs...)
		}
	} else {
		ctx.Encode(
			struct {
				Authenticated bool
				Privileged    bool
			}{
				authenticated,
				authorized,
			},
		)
	}

	//	username, password, _ := ctx.Request.BasicAuth()

	//	privilege, ok, errs := radius.RadiusAuthenticate(username, password)

	//	if len(errs) > 0 {
	//		if errs[0].Error() == radius.GatewayTimeoutError {
	//			ctx.EncodeErrors(http.StatusGatewayTimeout, errs...)
	//		} else if errs[0].Error() == radius.RadiusAuthError {
	//			ctx.EncodeErrors(http.StatusBadGateway, errs...)
	//		} else {
	//			syslogger.ErrErrors(errs...)

	//			ctx.EncodeInternalServerErrors(errs...)
	//		}

	//		return
	//	}

	//		json.NewEncoder(ctx.Writer).Encode(
	//			struct {
	//				Authenticated bool
	//				Privileged    bool
	//			}{
	//				ok,
	//				privilege,
	//			},
	//		)
}
