package middleware

import (
	"encoding/json"
	"log"
	"net/http"

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
			ctx := chi.RouteContext(r.Context())
			if ctx == nil || r.Method != "OPTIONS" {
				// Just proxy to the next handler
				log.Println("proxying", r)
				h.ServeHTTP(w, r)
				return
			}
			// Hijack request
			var routes []route
			for _, v := range ctx.Routes.Routes() {
				for k := range v.Handlers {
					if k == "OPTIONS" {
						continue
					}
					routes = append(routes, route{Method: k, Path: v.Pattern})
				}
			}
			raw, err := opt.Render(routes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err)
				return
			}
			w.WriteHeader(200)
			w.Header().Add("Content-Type", opt.ContentType)
			w.Write(raw)
		})
	}
}
