package http2

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/mux"

	"github.com/moov-io/base/log"
)

const (
	PATH_VAR_REGEX = "[a-zA-Z0-9-_]{1,36}"
)

type (
	HandlerFunc = func(Request) Response
)

func NewRouter(logger log.Logger) *Router {
	return &Router{
		muxRouter: mux.NewRouter(),
		logger:    logger,
	}
}

func newHandler(h func(req Request) Response) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := Request{
			Request: r,
		}
		res := h(req)
		res.close(w)
	}
}

func (r *Router) Build() {
	logger := r.logger.Info().Set("ts", time.Now().String())
	logger.Log("Building routes for service")

	for _, route := range r.routes {
		for method, handler := range route.handlers {
			re := regexp.MustCompile(`{\w+`)

			// Add regexp validation to the path variable
			route.path = re.ReplaceAllString(route.path, fmt.Sprintf("$0:%s", PATH_VAR_REGEX))
			muxRoute := r.muxRouter.Path(route.path).Methods(method).Handler(newHandler(handler.handler))

			if handler.name != "" {
				muxRoute.Name(handler.name)
			}

			logger.Log(route.path)
		}
	}
}

type Router struct {
	muxRouter *mux.Router
	routes    []Route
	logger    log.Logger
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		var resp Response
		if recovered := recover(); r != nil {
			err, ok := recovered.(error)
			if !ok {
				resp = InternalServerError()
				return
			}

			switch err.(type) {
			case errPathVarNotFound,
				errHeaderNotFound:
				r.logger.LogError(err)
				resp = Error(err)
			default:
				switch err {
				case errInvalidJSON:
					r.logger.LogError(err)
					resp = Error(err)
				default:
					resp = InternalServerError()
				}
			}
		}

		resp.close(w)
	}()

	r.muxRouter.ServeHTTP(w, req)
}

func (r *Router) SetRoute(path string) *Route {
	route := &Route{
		path:     path,
		handlers: make(map[string]*handler),
	}
	r.routes = append(r.routes, *route)
	return route
}

type Route struct {
	path        string
	handlers    map[string]*handler // method
	lastHandler *handler
}

type handler struct {
	name    string
	handler HandlerFunc
}

func (r *Route) Get(h HandlerFunc) *Route {
	r.set(http.MethodGet, h)
	return r
}

func (r *Route) Post(h HandlerFunc) *Route {
	r.set(http.MethodPost, h)
	return r
}

func (r *Route) Put(h HandlerFunc) *Route {
	r.set(http.MethodPut, h)
	return r
}

func (r *Route) Delete(h HandlerFunc) *Route {
	r.set(http.MethodDelete, h)
	return r
}

func (r *Route) Name(name string) *Route {
	r.lastHandler.name = name
	return r
}

func (r *Route) set(method string, h HandlerFunc) {
	r.lastHandler = &handler{handler: h}
	r.handlers[method] = r.lastHandler
}

type CustomerRepo interface {
	Create() error
}

type customerRepo struct{}
