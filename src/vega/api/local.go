// Copyright (c) 2016, Virtual Gateway Labs. All rights reserved.

package main

import (
	"vega/api/handlers"
	"vega/api/handlers/aaa/radius"
)

func localRoutes(ctx handlers.Context) map[string]map[string]handler {
	local := newChain(ctx)
	local.add(wrapLocal)

	r := map[string]map[string]handler{
		"GET": {
			"/local/radius/authenticate": local.wrap(radius.LocalAuthenticate),
			"/local/radius/enabled":      local.wrap(radius.LocalEnabled),
		},
	}

	return r
}
