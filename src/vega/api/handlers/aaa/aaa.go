// Copyright (c) 2016, Virtual Gateway Labs. All rights reserved.

package aaa

import "vega/api/handlers"

func Get(ctx handlers.Context) {
	ctx.Encode(ctx.Config.AAA)
}

func Patch(ctx handlers.Context) {
	ctx.DecodeVerifySave(&ctx.Config.AAA)
}
