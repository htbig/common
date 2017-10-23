// Copyright (c) 2016, Virtual Gateway Labs. All rights reserved.

package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"runtime/debug"
	"sort"
	"strings"

	"vega/api/auth"
	"vega/api/handlers"
	"vega/api/locker"
	"vega/api/tasks"
	"vega/core"
	"vega/core/aaa/radius"
	"vega/syslogger"

	"github.com/julienschmidt/httprouter"
)

const ContentTypeJSON = "application/json"

type handlerWrapper func(handlers.Handler) handlers.Handler

type handler struct {
	ctx     handlers.Context
	handler handlers.Handler
}

type Recorder struct {
	length int
	status int
	rw     http.ResponseWriter
}

func (r *Recorder) Header() http.Header {
	return r.rw.Header()
}

func (r *Recorder) WriteHeader(status int) {
	r.status = status
	r.rw.WriteHeader(status)
}

func (r *Recorder) Write(data []byte) (int, error) {
	r.length += len(data)
	return r.rw.Write(data)
}

func (h *handler) handle(w http.ResponseWriter, r *http.Request, p handlers.Params) {
	recorder := Recorder{status: http.StatusOK, rw: w}
	var ctx handlers.Context
	ctx.Params = p
	ctx.Writer = &recorder
	ctx.Request = r
	ctx.Lock = h.ctx.Lock
	ctx.Tasks = h.ctx.Tasks
	ctx.Config = h.ctx.Config
	ctx.BasePath = h.ctx.BasePath
	ctx.Writer.Header().Set("Content-Type", ContentTypeJSON)
	f := func() {
		defer func() {
			if r := recover(); r != nil {
				syslogger.Err(string(debug.Stack()))
				ctx.Writer.WriteHeader(http.StatusInternalServerError)
			}
		}()

		h.handler(ctx)
	}

	f()

	if recorder.length == 0 && recorder.status == http.StatusOK {
		w.WriteHeader(http.StatusNoContent)
	}

}

type chain struct {
	ctx      handlers.Context
	wrappers []handlerWrapper
	tasks    *tasks.Manager
}

func (c chain) wrap(h handlers.Handler) handler {
	for i := len(c.wrappers) - 1; i >= 0; i-- {
		h = c.wrappers[i](h)
	}
	c.ctx.Tasks = c.tasks
	return handler{c.ctx, h}
}

func (c chain) wrap_wrappers(h handlers.Handler, wrappers ...handlerWrapper) handler {
	for i := len(wrappers) - 1; i >= 0; i-- {
		h = wrappers[i](h)
	}
	c.ctx.Tasks = c.tasks
	return handler{c.ctx, h}
}

func (c *chain) add(wrappers ...handlerWrapper) {
	for _, w := range wrappers {
		c.wrappers = append(c.wrappers, w)
	}
}

func wrapLocker(handler handlers.Handler) handlers.Handler {
	return func(ctx handlers.Context) {
		syslogger.Info(ctx.Request.URL)
		syslogger.Info(ctx.Request.RequestURI)
		syslogger.Info(ctx.Request)

		if ctx.TryLock() {
			defer ctx.Lock.Unlock()
			handler(ctx)
		}
	}
}

func wrapValidJSON(handler handlers.Handler) handlers.Handler {
	return func(ctx handlers.Context) {

		if ctx.Request.Header.Get("Content-Type") != ContentTypeJSON {
			ctx.Writer.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		buf, _ := ioutil.ReadAll(ctx.Request.Body)
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
		var jsonTest interface{}
		err := json.NewDecoder(bytes.NewBuffer(buf)).Decode(&jsonTest)
		if err != nil {
			ctx.Writer.WriteHeader(http.StatusUnsupportedMediaType)
		} else {
			handler(ctx)
		}
	}
}

func wrapLocal(handler handlers.Handler) handlers.Handler {
	return func(ctx handlers.Context) {
		host, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
		switch {
		case err != nil:
			panic(err)
		case host == "127.0.0.1":
			fallthrough
		case host == "::1":
			handler(ctx)
		default:
			ctx.Writer.WriteHeader(http.StatusNotFound)
		}
	}
}

func wrapAuth(checkPrivilege bool) func(handlers.Handler) handlers.Handler {
	return func(handler handlers.Handler) handlers.Handler {
		return func(ctx handlers.Context) {
			var authenticated, authorized bool
			var errs []error

			request := ctx.Request
			if host, _, err := net.SplitHostPort(request.RemoteAddr); err == nil {
				if host == "127.0.0.1" || host == "::1" {
					// skip authentication for local requests
					authorized, authenticated = true, true
				} else if !checkPrivilege && strings.HasPrefix(host, "172.17.0.") {
					// skip authentication for unprivileged container requests
					authorized, authenticated = true, true
				} else {
					username, password, ok := request.BasicAuth()
					if ok {
						radius := ctx.Config.AAA.RADIUS.Enabled
						fallback := ctx.Config.AAA.RADIUS.Fallback

						authenticated, authorized, errs = auth.AuthenticateAPI(radius, fallback, checkPrivilege, username, password)
					}
				}
			} else {
				ctx.EncodeInternalServerErrors(err)
			}

			if authenticated {
				if authorized {
					handler(ctx)
				} else {
					ctx.Writer.WriteHeader(http.StatusForbidden)
				}
			} else {
				if len(errs) > 0 {
					if errs[0].Error() == radius.GatewayTimeoutError {
						ctx.EncodeErrors(http.StatusGatewayTimeout, errs...)
					} else if errs[0].Error() == radius.RadiusAuthError {
						ctx.EncodeErrors(http.StatusBadGateway, errs...)
					} else {
						ctx.EncodeInternalServerErrors(errs...)
					}
				} else {
					ctx.Writer.WriteHeader(http.StatusUnauthorized)
				}
			}
		}
	}
}

//func wrapAuthAdmin(handler handlers.Handler) handlers.Handler {
//	return func(ctx handlers.Context) {
//		var authorized, authenticated bool
//		var err error
//		var errs []error

//		request := ctx.Request
//		host, _, err := net.SplitHostPort(request.RemoteAddr)
//		if err == nil {
//			if host == "127.0.0.1" || host == "::1" {
//				// skip authentication for local requests
//				authenticated = true
//				authorized = true
//			} else {
//				username, password, ok := request.BasicAuth()
//				if ok {
//					radius := ctx.Config.AAA.RADIUS.Enabled
//					authenticated, authorized, errs = authenticateAdmin(radius, username, password)
//				}
//			}
//		}

//		if authenticated {
//			if authorized {
//				handler(ctx)
//			} else {
//				ctx.Writer.WriteHeader(http.StatusForbidden)
//			}
//		} else {
//			if err != nil {
//				log.Println("Internal Authenticate Error: ", err)
//				ctx.Writer.WriteHeader(http.StatusInternalServerError)
//				json.NewEncoder(ctx.Writer).Encode(err)
//			} else if len(errs) > 0 {
//				if errs[0].Error() == radius.GatewayTimeoutError {
//					ctx.EncodeErrors(http.StatusGatewayTimeout, errs...)
//				} else if errs[0].Error() == radius.RadiusAuthError {
//					ctx.EncodeErrors(http.StatusBadGateway, errs...)
//				} else {
//					ctx.EncodeInternalServerErrors(errs...)
//				}
//			} else {
//				ctx.Writer.WriteHeader(http.StatusUnauthorized)
//			}
//		}
//	}
//}

//func wrapAuthUser(handler handlers.Handler) handlers.Handler {
//	return func(ctx handlers.Context) {
//		var authenticated bool
//		var err error
//		var errs []error

//		request := ctx.Request
//		host, _, err := net.SplitHostPort(request.RemoteAddr)
//		if err == nil {
//			// docker or localhost bypass
//			if strings.HasPrefix(host, "172.17.0.") || host == "127.0.0.1" || host == "::1" {
//				// skip authentication for local requests
//				authenticated = true
//			} else {
//				username, password, ok := request.BasicAuth()
//				if ok {
//					radius := ctx.Config.AAA.RADIUS.Enabled
//					authenticated, errs = authenticateUser(radius, username, password)
//				}
//			}
//		}

//		if authenticated {
//			handler(ctx)
//		} else {
//			if err != nil {
//				log.Println("Internal Authenticate Error: ", err)
//				ctx.Writer.WriteHeader(http.StatusInternalServerError)
//				json.NewEncoder(ctx.Writer).Encode(err)
//			} else if len(errs) > 0 {
//				if errs[0].Error() == radius.GatewayTimeoutError {
//					ctx.EncodeErrors(http.StatusGatewayTimeout, errs...)
//				} else if errs[0].Error() == radius.RadiusAuthError {
//					ctx.EncodeErrors(http.StatusBadGateway, errs...)
//				} else {
//					ctx.EncodeInternalServerErrors(errs...)
//				}
//			} else {
//				ctx.Writer.WriteHeader(http.StatusUnauthorized)
//			}
//		}
//	}
//}

func wrapRouter(h handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		h.handle(w, r, p)
	}
}

func handlePanic(w http.ResponseWriter, r *http.Request, err interface{}) {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(http.StatusInternalServerError)

	errorMap := make(map[string]interface{})
	errorMap["errors"] = err

	json.NewEncoder(w).Encode(errorMap)
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(http.StatusNotFound)
}

func newChain(ctx handlers.Context) chain {
	return chain{ctx: ctx}
}

func router() (http.Handler, handlers.Context) {
	ctx := handlers.Context{
		Lock:     locker.New(),
		BasePath: "",
		Config:   core.NewConfig(),
	}

	r := httprouter.New()
	ctx.Config.LoadStartup() // ignore error here
	cfg_factory := core.NewConfig()
	ctx.Config.Save(*cfg_factory)

	// local routes
	localRouting := localRoutes(ctx)
	for method, paths := range localRouting {
		for path, handle := range paths {
			r.Handle(method, ctx.BasePath+path, wrapRouter(handle))
		}
	}

	// public routes
	endpoints := make(map[string][]string)
	publicRouting := publicRoutes(ctx)
	for method, paths := range publicRouting {
		for path, handle := range paths {
			endpoints[method] = append(endpoints[method], path)
			r.Handle(method, ctx.BasePath+path, wrapRouter(handle))
		}
		// sort paths
		sort.Strings(endpoints[method])
	}

	r.PanicHandler = handlePanic
	r.NotFound = http.HandlerFunc(handleNotFound)

	// endpoint that shows all endpoints

	root := func(ctx handlers.Context) {
		ctx.Writer.Header().Set("Content-Type", ContentTypeJSON)
		json.NewEncoder(ctx.Writer).Encode(endpoints)
	}
	r.GET(ctx.BasePath+"/endpoints", wrapRouter(handler{ctx, wrapAuth(true)(root)}))

	// endpoint for ping
	ping := func(ctx handlers.Context) {
		ctx.Writer.WriteHeader(http.StatusNoContent)
	}
	//	r.GET(ctx.BasePath, wrapRouter(handler{ctx, ping}))
	r.GET("/", wrapRouter(handler{ctx, ping}))

	return r, ctx
}
