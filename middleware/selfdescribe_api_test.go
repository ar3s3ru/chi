package middleware_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"testing"
	"time"

	"github.com/ar3s3ru/chi/middleware"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestSelfDescribeAPI(t *testing.T) {
	checkURLs := func(t *testing.T, s string) *url.URL {
		u, err := url.Parse(s)
		assert.NoError(t, err)
		return u
	}

	tests := []struct {
		i *http.Request
		o middleware.APIRoutes
	}{
		{
			i: &http.Request{Method: "OPTIONS", URL: checkURLs(t, "http://localhost:8080/route/test2/inner")},
			o: middleware.APIRoutes{middleware.APIRouteInfo{Method: "GET", Path: "/"}},
		},
		{
			i: &http.Request{Method: "OPTIONS", URL: checkURLs(t, "http://localhost:8080/route/test2")},
			o: middleware.APIRoutes{middleware.APIRouteInfo{Method: "GET", Path: "/inner"}},
		},
		{
			i: &http.Request{Method: "OPTIONS", URL: checkURLs(t, "http://localhost:8080/route/test")},
			o: middleware.APIRoutes{middleware.APIRouteInfo{Method: "GET", Path: "/hello/{id}"}},
		},
		{
			i: &http.Request{Method: "OPTIONS", URL: checkURLs(t, "http://localhost:8080/route/test/hello")},
			o: middleware.APIRoutes{middleware.APIRouteInfo{Method: "GET", Path: "/{id}"}},
		},
		{
			i: &http.Request{Method: "OPTIONS", URL: checkURLs(t, "http://localhost:8080/route")},
			o: middleware.APIRoutes{
				middleware.APIRouteInfo{Method: "GET", Path: "/test/hello/{id}"},
				middleware.APIRouteInfo{Method: "GET", Path: "/test2/inner"}, middleware.APIRouteInfo{Method: "GET", Path: "/get"},
				middleware.APIRouteInfo{Method: "POST", Path: "/post"}, middleware.APIRouteInfo{Method: "PUT", Path: "/put"},
				middleware.APIRouteInfo{Method: "PATCH", Path: "/patch"}, middleware.APIRouteInfo{Method: "DELETE", Path: "/delete"},
			},
		},
		{
			i: &http.Request{Method: "OPTIONS", URL: checkURLs(t, "http://localhost:8080")},
			o: middleware.APIRoutes{
				middleware.APIRouteInfo{Method: "GET", Path: "/route/test/hello/{id}"},
				middleware.APIRouteInfo{Method: "GET", Path: "/route/test2/inner"}, middleware.APIRouteInfo{Method: "GET", Path: "/get"},
				middleware.APIRouteInfo{Method: "POST", Path: "/route/post"}, middleware.APIRouteInfo{Method: "PUT", Path: "/route/put"},
				middleware.APIRouteInfo{Method: "PATCH", Path: "/route/patch"}, middleware.APIRouteInfo{Method: "DELETE", Path: "/route/delete"},
				middleware.APIRouteInfo{Method: "GET", Path: "/get"}, middleware.APIRouteInfo{Method: "POST", Path: "/post"},
				middleware.APIRouteInfo{Method: "PUT", Path: "/put"}, middleware.APIRouteInfo{Method: "PATCH", Path: "/patch"},
				middleware.APIRouteInfo{Method: "DELETE", Path: "/delete"},
			},
		},
	}

	go selfDescribeTestRunServer(8080)
	<-time.After(100 * time.Millisecond) // Just to be sure the server is running
	for i, test := range tests {
		v, err := http.DefaultClient.Do(test.i)
		assert.NoError(t, err, "test %d: error while making request", i)

		defer v.Body.Close()
		body, err := ioutil.ReadAll(v.Body)
		assert.NoError(t, err, "test %d: error while reading response body", i)

		var result middleware.APIRoutes
		assert.NoError(t, json.Unmarshal(body, &result), "test %d: error while unmarshaling results", i)
		// Sorting both cases to ensure coherence
		sort.Sort(test.o)
		sort.Sort(result)
		assert.Equal(t, test.o, result, "test %d: results mismatch", i)
	}
}

// HTTP server used in TestSelfDescribeAPI
func selfDescribeTestRunServer(port int) {
	helloWorld := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	}

	r := chi.NewRouter()
	r.Use(middleware.SelfDescribeAPI())

	r.Get("/get", helloWorld)
	r.Post("/post", helloWorld)
	r.Put("/put", helloWorld)
	r.Patch("/patch", helloWorld)
	r.Delete("/delete", helloWorld)

	r.Route("/route", func(r chi.Router) {
		r.Get("/get", helloWorld)
		r.Post("/post", helloWorld)
		r.Put("/put", helloWorld)
		r.Patch("/patch", helloWorld)
		r.Delete("/delete", helloWorld)
		r.Route("/test", func(r chi.Router) {
			r.Get("/hello/{id}", helloWorld)
		})
		r.Mount("/test2", chi.NewRouter().Group(func(r chi.Router) {
			r.Get("/inner", helloWorld)
		}))
	})

	log.Print(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
