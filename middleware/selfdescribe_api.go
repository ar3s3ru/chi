package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

var defaultHijackOptions = HijackOptions{
	ContentType: "application/json",
	Render:      json.Marshal,
}

type RenderFn func(v interface{}) ([]byte, error)

type HijackOptions struct {
	ContentType string
	Render      RenderFn
}

func SelfDescribeAPI(options ...HijackOptions) func(http.Handler) http.Handler {
	opt := defaultHijackOptions
	if len(options) > 0 {
		opt = options[0]
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, log := chi.RouteContext(r.Context()), GetLogEntry(r)
			if ctx == nil || r.Method != "OPTIONS" {
				// Just proxy to the next handler
				h.ServeHTTP(w, r)
				return
			}
			// Hijack request
			var routes APIRoutes
			u := r.RequestURI
			err := chi.Walk(ctx.Routes,
				func(m string, r string, h http.Handler, mw ...func(http.Handler) http.Handler) error {
					r = strings.Replace(r, "/*/", "/", -1)
					lr, lu := len(r), len(u)
					if lr >= lu && r[:lu] == u {
						diff := r[lu:]
						if len(diff) <= 0 {
							diff = "/"
						}
						routes = append(routes, APIRouteInfo{Method: m, Path: diff})
					}
					return nil
				})
			raw, err := opt.Render(routes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Panic(fmt.Sprintf("rendering OPTIONS description failed: %s", err), nil)
				return
			}
			w.WriteHeader(200)
			w.Header().Add("Content-Type", opt.ContentType)
			w.Write(raw)
		})
	}
}

type APIRouteInfo struct {
	Method      string `json:"method"`
	Path        string `json:"uri"`
	Description string `json:"description,omitempty"`
}

type APIRoutes []APIRouteInfo

func (r APIRoutes) Len() int {
	return len(r)
}

func (r APIRoutes) Less(i int, j int) bool {
	return strings.Compare(r[i].Path, r[j].Path) == -1
}

func (r APIRoutes) Swap(i int, j int) {
	r[i], r[j] = r[j], r[i]
}
