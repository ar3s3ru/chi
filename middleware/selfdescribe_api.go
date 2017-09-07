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

type route struct {
	Method      string `json:"method"`
	Path        string `json:"uri"`
	Description string `json:"description,omitempty"`
}

func SelfDescribe(options ...HijackOptions) func(http.Handler) http.Handler {
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
			var routes []route
			u := r.RequestURI
			err := chi.Walk(ctx.Routes,
				func(m string, r string, h http.Handler, mw ...func(http.Handler) http.Handler) error {
					r = strings.Replace(r, "/*/", "/", -1)
					lr, lu := len(r), len(u)
					if lr >= lu && r[:lu] == u {
						routes = append(routes, route{Method: m, Path: r})
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
